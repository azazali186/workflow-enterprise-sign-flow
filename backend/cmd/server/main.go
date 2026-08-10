// Command server runs the AeroXe SignFlow API.
//
// @title AeroXe SignFlow API
// @version 1.0
// @description Digital signature and contract management platform. All data is
// exchanged in snake_case JSON bodies; only POST, PATCH and DELETE are used.
// @host localhost:8080
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/aeroxe/sign-flow/backend/docs"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/config"
	"github.com/aeroxe/sign-flow/backend/internal/database"
	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/middleware"
	"github.com/aeroxe/sign-flow/backend/internal/modules/auditlogs"
	"github.com/aeroxe/sign-flow/backend/internal/modules/auth"
	"github.com/aeroxe/sign-flow/backend/internal/modules/certificates"
	"github.com/aeroxe/sign-flow/backend/internal/modules/compliances"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/modules/dashboard"
	"github.com/aeroxe/sign-flow/backend/internal/modules/loginlogs"
	"github.com/aeroxe/sign-flow/backend/internal/modules/permissions"
	"github.com/aeroxe/sign-flow/backend/internal/modules/reports"
	"github.com/aeroxe/sign-flow/backend/internal/modules/roles"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signatures"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signers"
	"github.com/aeroxe/sign-flow/backend/internal/modules/storages"
	"github.com/aeroxe/sign-flow/backend/internal/modules/templates"
	"github.com/aeroxe/sign-flow/backend/internal/modules/users"
	"github.com/aeroxe/sign-flow/backend/internal/modules/verifications"
	"github.com/aeroxe/sign-flow/backend/internal/natsx"
	"github.com/aeroxe/sign-flow/backend/internal/outbox"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/crypto"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/httpx"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/jwt"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/retryutil"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
	"github.com/aeroxe/sign-flow/backend/internal/seed"
	"github.com/aeroxe/sign-flow/backend/internal/ws"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env, cfg.LogLevel)
	if cfg.Env == "production" {
		validateProductionSecrets(cfg)
	}
	if err := crypto.Init(cfg.EncryptionKey); err != nil {
		logger.L().Fatal("crypto init failed", zap.Error(err))
	}

	// Database may be briefly unreachable during boot (failover, container
	// ordering); retry with backoff before giving up.
	cfgDB := retryutil.Config{InitialInterval: 500 * time.Millisecond, MaxInterval: 5 * time.Second, MaxElapsedTime: 60 * time.Second}
	var db *gorm.DB
	err := retryutil.Do(context.Background(), cfgDB, func() error {
		d, derr := database.Init(cfg.DatabaseURL)
		if derr != nil {
			return derr
		}
		sqlDB, serr := d.DB()
		if serr != nil {
			return serr
		}
		return sqlDB.Ping()
	})
	if err != nil {
		logger.L().Fatal("database connect failed after retries", zap.Error(err))
	}
	db = database.DB
	if err := database.Migrate(db, database.AllModels()...); err != nil {
		logger.L().Fatal("migration failed", zap.Error(err))
	}
	logger.L().Info("migrations applied")

	var cacheClient cache.Cache
	cacheClient, err = cache.NewRedis(cfg.RedisURL)
	if err != nil {
		if cfg.Env == "production" {
			logger.L().Fatal("redis connect failed", zap.Error(err))
		}
		logger.L().Warn("redis unavailable, using in-memory cache", zap.Error(err))
		cacheClient = cache.NewMemory()
	}

	// NATS: connect immediately when possible, otherwise keep retrying in the
	// background. While disconnected, outbox events stay pending (never lost).
	natsClient := natsx.NewClient()
	natsCtx, natsCancel := context.WithCancel(context.Background())
	defer natsCancel()
	if err := natsClient.Connect(cfg.NATSURL); err != nil {
		logger.L().Warn("nats unavailable at boot, reconnecting in background", zap.Error(err))
		natsClient.StartReconnectLoop(natsCtx, cfg.NATSURL, 5*time.Second)
	}

	if err := seed.Run(db, seed.Options{AdminEmail: cfg.AdminEmail, AdminPassword: cfg.AdminPassword}); err != nil {
		logger.L().Fatal("seed failed", zap.Error(err))
	}

	met := metrics.New()
	auditSvc := audit.New(db)
	bus := events.NewBus()
	relay := outbox.NewRelay(db, events.NewNATSPublisher(natsClient), subjectFor, met, cfg.OutboxPollInterval)
	relay.Start()
	defer relay.Stop()

	jwtMgr := jwt.New(cfg.JWTSecret, cfg.JWTExpiry)

	h := server.New(
		server.WithHostPorts(":"+cfg.Port),
		server.WithMaxRequestBodySize(cfg.MaxBodySize),
		server.WithReadTimeout(cfg.ReadTimeout),
		server.WithWriteTimeout(cfg.WriteTimeout),
		server.WithIdleTimeout(cfg.IdleTimeout),
		server.WithExitWaitTime(cfg.ShutdownTimeout),
	)

	// Guards map is filled in by route registration below; middleware is
	// registered BEFORE routes because Hertz snapshots handler chains at
	// route-registration time. Order: request id -> recovery -> security
	// headers/CORS -> request log -> rate limit -> auth -> RBAC.
	guards := map[string]string{}
	h.Use(middleware.RequestID(), middleware.Recovery())
	h.Use(middleware.Security(cfg.CORSAllowedOrigins, cfg.Env == "production"))
	h.Use(middleware.RequestLog(met), middleware.RateLimit(cacheClient, cfg.RateLimitPerMin))
	h.Use(middleware.Auth(jwtMgr, cacheClient, guards), middleware.NewRBAC(db, cacheClient, guards, met).Middleware())

	reg := registry.New()
	// Handlers register full "/api/v1/..." paths, so mount them on the root.
	r := h.Group("/")
	wsHub := ws.NewHub(bus, cfg.WSMaxConnections, ws.OriginPolicy{Allowed: cfg.CORSAllowedOrigins, Enforce: cfg.Env == "production"})
	ws.Register(reg, h.Group("/"), wsHub)

	// Build handlers.
	authH := auth.NewHandler(auth.NewService(db, cacheClient, jwtMgr, auditSvc, cfg.JWTExpiry))
	userSvc := users.NewService(db, cacheClient, auditSvc)
	usersH := users.NewHandler(userSvc)
	roleSvc := roles.NewService(db, cacheClient, auditSvc)
	rolesH := roles.NewHandler(roleSvc)
	permsH := permissions.NewHandler(permissions.NewService(db, cacheClient))
	contractSvc := contracts.NewService(db, cacheClient, auditSvc, bus)
	contractsH := contracts.NewHandler(contractSvc)
	templatesH := templates.NewHandler(templates.NewService(db, cacheClient, auditSvc))
	signerSvc := signers.NewService(db, cacheClient, auditSvc)
	signersH := signers.NewHandler(signerSvc)
	sigSvc := signatures.NewService(db, cacheClient, auditSvc, bus)
	signaturesH := signatures.NewHandler(sigSvc)
	verifsH := verifications.NewHandler(verifications.NewService(db, cacheClient, auditSvc))
	storagesH := storages.NewHandler(storages.NewService(db, cacheClient, auditSvc))
	compliancesH := compliances.NewHandler(compliances.NewService(db, cacheClient, auditSvc))
	certsH := certificates.NewHandler(certificates.NewService(db, cacheClient, auditSvc))
	auditLogsH := auditlogs.NewHandler(auditlogs.NewService(db, cacheClient))
	loginLogsH := loginlogs.NewHandler(loginlogs.NewService(db, cacheClient))
	dashH := dashboard.NewHandler(dashboard.NewService(db))
	reportsH := reports.NewHandler(contractSvc, sigSvc, auditlogs.NewService(db, cacheClient), signerSvc)

	// Register routes (order matters: after all registrations we seed permissions).
	authH.Register(reg, r)
	usersH.Register(reg, r)
	rolesH.Register(reg, r)
	permsH.Register(reg, r)
	contractsH.Register(reg, r)
	templatesH.Register(reg, r)
	signersH.Register(reg, r)
	signaturesH.Register(reg, r)
	verifsH.Register(reg, r)
	storagesH.Register(reg, r)
	compliancesH.Register(reg, r)
	certsH.Register(reg, r)
	auditLogsH.Register(reg, r)
	loginLogsH.Register(reg, r)
	dashH.Register(reg, r)
	reportsH.Register(reg, r)

	// Public infra endpoints (also registered as PUBLIC so auth/RBAC skip them).
	h.GET("/swagger/*any", httpx.Adapt(httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json"))))
	h.GET("/metrics", httpx.Adapt(promhttp.Handler()))
	h.POST("/api/v1/health", healthHandler(db, cacheClient))
	reg.Register("GET", "/swagger/*any", "Swagger Documentation", "PUBLIC")
	reg.Register("GET", "/metrics", "Prometheus Metrics", "PUBLIC")
	reg.Register("POST", "/api/v1/health", "Health Check", "PUBLIC")

	// Populate the guard table now that every route is registered, then seed
	// permissions from the registered API routes (RBAC).
	for _, g := range reg.Guards() {
		guards[g.Key] = g.Guard
	}
	if err := reg.SeedPermissions(db, cacheClient); err != nil {
		logger.L().Fatal("permission seeding failed", zap.Error(err))
	}
	// Seeding may have changed the permission catalog, so stale per-user
	// grants cached by a previous process must be purged.
	purgeCtx, purgeCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := middleware.InvalidateAllRBAC(purgeCtx, cacheClient); err != nil {
		logger.L().Warn("rbac cache purge failed (non-fatal)", zap.Error(err))
	}
	purgeCancel()

	go h.Spin()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.L().Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	_ = h.Shutdown(shutdownCtx)
	logger.Sync()
}

// validateProductionSecrets refuses to boot with dev-only defaults in prod.
func validateProductionSecrets(cfg *config.Config) {
	weak := func(v string) bool {
		return v == "" || strings.Contains(v, "change-me") || strings.Contains(v, "dev-only")
	}
	if weak(cfg.JWTSecret) {
		logger.L().Fatal("JWT_SECRET must be set to a strong random value in production")
	}
	if weak(cfg.EncryptionKey) {
		logger.L().Fatal("ENCRYPTION_KEY must be set to a strong random value in production")
	}
	if strings.EqualFold(cfg.AdminPassword, "ChangeMe!123") || strings.Contains(strings.ToLower(cfg.AdminPassword), "changeme") {
		logger.L().Fatal("ADMIN_PASSWORD must be replaced with a strong value in production")
	}
	if slices.Contains(cfg.CORSAllowedOrigins, "*") {
		logger.L().Fatal("CORS_ALLOWED_ORIGINS must not contain '*' in production")
	}
}

// healthHandler reports liveness and readiness (public).
func healthHandler(db *gorm.DB, cacheClient cache.Cache) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		deps := map[string]string{"database": "ok", "cache": "ok"}
		status := http.StatusOK
		if sqlDB, err := db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
			deps["database"] = "down"
			status = http.StatusServiceUnavailable
		}
		if err := cacheClient.Ping(ctx); err != nil {
			deps["cache"] = "down"
			status = http.StatusServiceUnavailable
		}
		msg := "healthy"
		code := 0
		if status != http.StatusOK {
			msg = "degraded"
			code = 50300
		}
		c.JSON(status, map[string]any{"code": code, "message": msg, "data": map[string]any{
			"status":  msg,
			"deps":    deps,
			"time":    time.Now().Format(time.RFC3339),
			"version": "1.0.0",
		}})
	}
}

func subjectFor(ev *outbox.Event) string {
	return natsx.SubjectPrefix + "." + ev.EventType
}
