// Package auth implements login/logout/me with Redis-backed JWT sessions and
// login audit logging. No passwords are ever logged or returned.
package auth

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/jwt"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
	"go.uber.org/zap"
)

// Service is the auth use-case boundary.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResult, error)
	Logout(ctx context.Context) error
	Me(ctx context.Context) (*models.User, error)
}

type service struct {
	db    *gorm.DB
	cache cache.Cache
	jwt   *jwt.Manager
	audit *audit.Service
	ttl   time.Duration
}

// NewService builds the auth service.
func NewService(db *gorm.DB, c cache.Cache, mgr *jwt.Manager, a *audit.Service, ttl time.Duration) Service {
	return &service{db: db, cache: c, jwt: mgr, audit: a, ttl: ttl}
}

// LoginRequest carries credentials in the request body (no query params).
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResult is the token payload returned on success.
type LoginResult struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        *models.User `json:"user"`
}

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResult, error) {
	if req.Email == "" || req.Password == "" {
		s.audit.RecordLogin(ctx, req.Email, false, "missing credentials")
		return nil, errs.ErrValidation
	}
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Roles").First(&user, "email = ?", req.Email).Error; err != nil {
		s.audit.RecordLogin(ctx, req.Email, false, "unknown user")
		if err == gorm.ErrRecordNotFound {
			return nil, errs.New(401, 40110, "invalid credentials")
		}
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		s.audit.RecordLogin(ctx, req.Email, false, "invalid password")
		return nil, errs.New(401, 40110, "invalid credentials")
	}
	if user.Status != models.UserActive {
		s.audit.RecordLogin(ctx, req.Email, false, "account suspended")
		return nil, errs.New(403, 40310, "account is suspended")
	}
	token, ttl, err := s.jwt.Issue(user.ID)
	if err != nil {
		return nil, err
	}
	sessionKey := "auth:token:" + user.ID
	if err := s.cache.Set(ctx, sessionKey, md5Hex(token+user.ID), ttl); err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, "auth:user:"+user.ID, user.Name, ttl)
	now := time.Now()
	_ = s.db.WithContext(ctx).Model(&user).Update("last_login_at", now)
	s.audit.RecordLogin(ctx, req.Email, true, "login success")
	logger.L().Info("user logged in", zap.String("user_id", user.ID))
	return &LoginResult{
		AccessToken: token, TokenType: "bearer", ExpiresIn: int64(ttl.Seconds()), User: &user,
	}, nil
}

func (s *service) Logout(ctx context.Context) error {
	userID := ctxval.UserID(ctx)
	if userID == "" {
		return errs.ErrUnauthorized
	}
	_ = s.cache.Del(ctx, "auth:token:"+userID, "auth:user:"+userID, "rbac:user:"+userID)
	s.audit.Record(ctx, "auth.logout", "user", userID, nil, map[string]any{"action": "logout"})
	return nil
}

func (s *service) Me(ctx context.Context) (*models.User, error) {
	userID := ctxval.UserID(ctx)
	if userID == "" {
		return nil, errs.ErrUnauthorized
	}
	var user models.User
	if err := s.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", userID).Error; err != nil {
		return nil, errs.ErrNotFound
	}
	return &user, nil
}

func md5Hex(s string) string {
	// local helper to avoid importing crypto/md5 twice
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// Handler wires the HTTP routes.
type Handler struct {
	svc Service
}

// NewHandler builds the auth handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes to the registry and engine.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/auth/login", "Login", "PUBLIC")
	reg.Register("POST", "/api/v1/auth/logout", "Logout", "API")
	reg.Register("POST", "/api/v1/auth/me", "Current User", "API")
	g.POST("/api/v1/auth/login", h.Login)
	g.POST("/api/v1/auth/logout", h.Logout)
	g.POST("/api/v1/auth/me", h.Me)
}

// Login godoc
// @Summary Login with credentials
// @Description Authenticates a user and returns a bearer token (JWT). Sessions are stored server-side for single sign-out.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Credentials"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 401 {object} response.Body
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(ctx context.Context, c *app.RequestContext) {
	var req LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Error(c, errs.Wrap(400, 40000, "validation failed", err))
		return
	}
	res, err := h.svc.Login(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, res)
}

// Logout godoc
// @Summary Logout and invalidate the session
// @Description Revokes the current token server-side.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(ctx context.Context, c *app.RequestContext) {
	if err := h.svc.Logout(ctx); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// Me godoc
// @Summary Get the current user
// @Description Returns the authenticated user with roles.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Body
// @Failure 401 {object} response.Body
// @Router /api/v1/auth/me [post]
func (h *Handler) Me(ctx context.Context, c *app.RequestContext) {
	user, err := h.svc.Me(ctx)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}
