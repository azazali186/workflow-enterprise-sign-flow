// Package loginlogs exposes authentication attempt records.
package loginlogs

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Service is the login log use-case boundary.
type Service interface {
	List(ctx context.Context, q pagination.Query) (*repo.Page[audit.LoginLog], error)
}

type service struct {
	repo *repo.Repo[audit.LoginLog]
}

// NewService wires the login log repository.
func NewService(db *gorm.DB, c cache.Cache) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[audit.LoginLog]{
			Slug:       "login_log",
			Searchable: []string{"username"},
			Filterable: map[string]string{"success": "success", "username": "username"},
			Sortable:   map[string]string{"username": "username"},
			DateFields: []string{"created_at", "login_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var out struct {
					Total   int64 `json:"total_attempts"`
					Success int64 `json:"successful"`
					Failed  int64 `json:"failed"`
				}
				tx.Model(&audit.LoginLog{}).Count(&out.Total)
				tx.Model(&audit.LoginLog{}).Where("success = ?", true).Count(&out.Success)
				tx.Model(&audit.LoginLog{}).Where("success = ?", false).Count(&out.Failed)
				return out, nil
			},
		}),
	}
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[audit.LoginLog], error) {
	return s.repo.List(ctx, q)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the login logs handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/login_logs/list", "List Login Logs", "API")
	g.POST("/api/v1/login_logs/list", h.List)
}

// List godoc
// @Summary List login logs
// @Description Cursor-paginated authentication attempt records with filters, search, date-range and a success summary.
// @Tags login_logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/login_logs/list [post]
func (h *Handler) List(ctx context.Context, c *app.RequestContext) {
	var q pagination.Query
	if !apikit.Bind(c, &q) {
		return
	}
	page, err := h.svc.List(ctx, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, page)
}
