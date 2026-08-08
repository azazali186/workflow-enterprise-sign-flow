// Package auditlogs exposes the audit trail for compliance and forensics.
// Records are written by the audit service and are read-only here.
package auditlogs

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

// Service is the audit log use-case boundary.
type Service interface {
	List(ctx context.Context, q pagination.Query) (*repo.Page[audit.Log], error)
	Detail(ctx context.Context, req DetailRequest) (*audit.Log, error)
}

type service struct {
	repo *repo.Repo[audit.Log]
}

// NewService wires the audit log repository.
func NewService(db *gorm.DB, c cache.Cache) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[audit.Log]{
			Slug:       "audit_log",
			Searchable: []string{"action", "entity_type", "actor_name"},
			Filterable: map[string]string{
				"action": "action", "entity_type": "entity_type", "entity_id": "entity_id",
				"actor_user_id": "actor_user_id", "request_id": "request_id",
			},
			Sortable:   map[string]string{"action": "action"},
			DateFields: []string{"created_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Action string `gorm:"column:action" json:"action"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&audit.Log{}).Select("action, count(*) as count").Group("action").Scan(&rows).Error
				return rows, err
			},
		}),
	}
}

// DetailRequest targets an audit log via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[audit.Log], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*audit.Log, error) {
	return s.repo.Get(ctx, req.ID)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the audit logs handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/audit_logs/list", "List Audit Logs", "API")
	reg.Register("POST", "/api/v1/audit_logs/detail", "Audit Log Detail", "API")
	g.POST("/api/v1/audit_logs/list", h.List)
	g.POST("/api/v1/audit_logs/detail", h.Detail)
}

// List godoc
// @Summary List audit logs
// @Description Cursor-paginated audit trail with filters, search, sorting, date-range and an action summary.
// @Tags audit_logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/audit_logs/list [post]
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

// Detail godoc
// @Summary Get an audit log entry
// @Description Returns one audit log entry by id in the body.
// @Tags audit_logs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Audit log id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/audit_logs/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	entry, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, entry)
}
