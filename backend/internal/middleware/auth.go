// Package middleware provides the HTTP middleware stack: auth, RBAC,
// rate limiting, request logging and panic recovery.
package middleware

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/jwt"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

const (
	tokenKeyPrefix = "auth:token:"
	userNamePrefix = "auth:user:"
	rbacPrefix     = "rbac:user:"
)

// Auth validates the bearer token against the distributed session store
// (single-login: a newer login invalidates older tokens). Public routes are
// skipped.
func Auth(mgr *jwt.Manager, cache cache.Cache, guards map[string]string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if guards[string(c.Request.Method())+" "+c.FullPath()] == GuardPublic {
			c.Next(ctx)
			return
		}
		authorization := strings.TrimSpace(c.Request.Header.Get("Authorization"))
		if !strings.HasPrefix(authorization, "Bearer ") || len(authorization) <= 7 {
			c.JSON(401, body(40100, "invalid token"))
			c.Abort()
			return
		}
		tokenStr := authorization[7:]
		claims, err := mgr.Parse(tokenStr)
		if err != nil {
			logger.L().Warn("token parse failed", zap.Error(err))
			c.JSON(401, body(40100, "invalid token"))
			c.Abort()
			return
		}
		userID := claims.UserID
		sessionKey := tokenKeyPrefix + userID
		stored, ttl, err := cache.GetWithTTL(ctx, sessionKey)
		if err != nil || stored != md5Hex(tokenStr+userID) {
			c.JSON(401, body(40101, "token expired"))
			c.Abort()
			return
		}
		// Renew the session when less than half the TTL remains.
		if ttl < mgr.TTL()/2 {
			_ = cache.Set(ctx, sessionKey, md5Hex(tokenStr+userID), mgr.TTL())
		}
		if name, err := cache.Get(ctx, userNamePrefix+userID); err == nil {
			ctx = ctxval.SetUserName(ctx, name)
		}
		ctx = ctxval.SetUserID(ctx, userID)
		ctx = ctxval.SetIP(ctx, clientIP(c))
		ctx = ctxval.SetUserAgent(ctx, string(c.Request.Header.UserAgent()))
		ctx = ctxval.SetRequestID(ctx, requestID(c))
		c.Next(ctx)
	}
}

func body(code int, message string) map[string]any {
	return map[string]any{"code": code, "message": message}
}

func md5Hex(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func clientIP(c *app.RequestContext) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	return string(c.Request.Header.Peek("X-Forwarded-For"))
}

func requestID(c *app.RequestContext) string {
	if id := string(c.Request.Header.Peek("X-Request-ID")); id != "" {
		return id
	}
	id, err := ctxval.NewRequestID()
	if err != nil {
		return ""
	}
	return id
}
