// Package registry collects every registered route at startup and seeds it as
// an RBAC permission — mirroring the reference "store new permissions" flow.
package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/middleware"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
	"go.uber.org/zap"
)

const redisRouteKey = "api-gateway-permissions"

// Route describes one registered API route.
type Route struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Path   string `json:"path"`
	Guard  string `json:"guard"`
	Method string `json:"method"`
}

// Registry accumulates routes during registration.
type Registry struct {
	routes []Route
}

// New builds an empty registry.
func New() *Registry { return &Registry{} }

// Register records a route. Public routes are excluded from permission seeding.
func (r *Registry) Register(method, path, name, guard string) {
	r.routes = append(r.routes, Route{
		Name: name, URL: path, Path: path, Guard: guard, Method: method,
	})
}

// All returns a copy of the registered routes.
func (r *Registry) All() []Route {
	out := make([]Route, len(r.routes))
	copy(out, r.routes)
	return out
}

// Guards returns the route table for the RBAC middleware.
func (r *Registry) Guards() []middleware.GuardedRoute {
	var out []middleware.GuardedRoute
	for _, rt := range r.routes {
		out = append(out, middleware.GuardedRoute{Key: rt.Method + " " + rt.Path, Guard: rt.Guard})
	}
	return out
}

// SeedPermissions upserts permissions for all API-guarded routes and persists
// the route table to routes.json + Redis (reference behaviour).
func (r *Registry) SeedPermissions(db *gorm.DB, c cache.Cache) error {
	var routesToStore []Route
	for _, rt := range r.routes {
		if rt.Guard == middleware.GuardPublic {
			logger.L().Info("route excluded from permissions", zap.String("key", rt.Method+" "+rt.Path))
			continue
		}
		routesToStore = append(routesToStore, rt)
	}
	jsonData, err := json.MarshalIndent(routesToStore, "", "    ")
	if err != nil {
		return err
	}
	// routes.json is reference/debug output only: failure must not prevent
	// the server from booting (e.g. read-only containers).
	if err := os.WriteFile("routes.json", jsonData, 0o644); err != nil {
		logger.L().Warn("failed to write routes.json (non-fatal)", zap.Error(err))
	}
	if err := c.Set(context.Background(), redisRouteKey, jsonData, 0); err != nil {
		logger.L().Warn("failed to store routes in redis", zap.Error(err))
	}
	for _, rt := range routesToStore {
		if err := upsertPermission(db, rt); err != nil {
			return err
		}
	}
	logger.L().Info("permissions seeded", zap.Int("count", len(routesToStore)))
	return nil
}

func upsertPermission(db *gorm.DB, rt Route) error {
	uniqueRoute := fmt.Sprintf("%s %s", rt.Method, rt.Path)
	var existing models.Permission
	err := db.Where("route = ?", uniqueRoute).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		perm := models.Permission{
			Name: rt.Name, Route: uniqueRoute, Path: rt.Path,
			Method: rt.Method, Service: "api-gateway",
		}
		if err := db.Create(&perm).Error; err != nil {
			return err
		}
		logger.L().Info("new permission inserted", zap.String("route", uniqueRoute))
		return nil
	}
	needUpdate := false
	if existing.Name != rt.Name {
		existing.Name, needUpdate = rt.Name, true
	}
	if existing.Path != rt.Path {
		existing.Path, needUpdate = rt.Path, true
	}
	if existing.Method != rt.Method {
		existing.Method, needUpdate = rt.Method, true
	}
	if existing.Service != "api-gateway" {
		existing.Service, needUpdate = "api-gateway", true
	}
	if needUpdate {
		return db.Save(&existing).Error
	}
	return nil
}

// FormatRouteName title-cases a path for the permission display name.
func FormatRouteName(path string) string {
	cleaned := strings.Replace(path, "/api/v1", "", 1)
	cleaned = strings.ReplaceAll(cleaned, "/", " ")
	return cases.Title(language.English).String(strings.TrimSpace(cleaned))
}
