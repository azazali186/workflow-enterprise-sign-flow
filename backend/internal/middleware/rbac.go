package middleware

import (
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/metrics"
)

// Guard values for the route registry.
const (
	GuardPublic = "PUBLIC"
	GuardAPI    = "API"
)

// GuardedRoute describes a registered route for RBAC enforcement.
type GuardedRoute struct {
	Key   string // "METHOD /path"
	Guard string
}

// RBAC enforces permissions per route using a cached user-role-permission set.
type RBAC struct {
	db      *gorm.DB
	cache   cache.Cache
	guards  map[string]string // populated after route registration; maps are refs
	metrics *metrics.Collectors
	ttl     time.Duration
}

// NewRBAC builds the enforcer. The guards map is shared with the caller and
// populated after route registration (it is only read per request).
func NewRBAC(db *gorm.DB, c cache.Cache, guards map[string]string, met *metrics.Collectors) *RBAC {
	return &RBAC{db: db, cache: c, guards: guards, metrics: met, ttl: 15 * time.Minute}
}

// Middleware enforces the permission keyed by "METHOD path" of every request.
func (r *RBAC) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		key := string(c.Request.Method()) + " " + c.FullPath()
		guard, known := r.guards[key]
		if !known {
			r.metrics.RBACChecks.WithLabelValues("unknown_route").Inc()
			c.JSON(403, body(40300, "forbidden: route not registered"))
			c.Abort()
			return
		}
		if guard == GuardPublic {
			r.metrics.RBACChecks.WithLabelValues("public").Inc()
			c.Next(ctx)
			return
		}
		userID := ctxval.UserID(ctx)
		allowed, err := r.check(ctx, userID, key)
		if err != nil {
			c.JSON(500, body(50000, "internal server error"))
			c.Abort()
			return
		}
		if allowed {
			r.metrics.RBACChecks.WithLabelValues("allow").Inc()
			c.Next(ctx)
			return
		}
		r.metrics.RBACChecks.WithLabelValues("deny").Inc()
		c.JSON(403, body(40301, "forbidden: missing permission "+key))
		c.Abort()
	}
}

// check resolves the user's permission keys with caching.
func (r *RBAC) check(ctx context.Context, userID, key string) (bool, error) {
	cacheKey := rbacPrefix + userID
	if raw, err := r.cache.Get(ctx, cacheKey); err == nil && raw != "" {
		return userAllows(raw, key), nil
	}
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Roles.Permissions").First(&user, "id = ?", userID).Error; err != nil {
		return false, err
	}
	if user.IsSuperAdmin() {
		_ = r.cache.Set(ctx, cacheKey, `["*"]`, r.ttl)
		return true, nil
	}
	keys := user.PermissionKeys()
	raw, _ := json.Marshal(keys)
	_ = r.cache.Set(ctx, cacheKey, raw, r.ttl)
	return userAllows(string(raw), key), nil
}

func userAllows(raw, key string) bool {
	if raw == `["*"]` {
		return true
	}
	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return false
	}
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
