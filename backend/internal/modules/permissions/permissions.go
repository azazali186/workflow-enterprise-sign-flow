// Package permissions exposes the seeded permission catalog. Permissions are
// created automatically from registered routes at server start.
package permissions

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Service is the permission use-case boundary.
type Service interface {
	List(ctx context.Context, q pagination.Query) (*repo.Page[models.Permission], error)
	Detail(ctx context.Context, req DetailRequest) (*models.Permission, error)
}

type service struct {
	repo *repo.Repo[models.Permission]
}

// NewService wires the permission repository.
func NewService(db *gorm.DB, c cache.Cache) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[models.Permission]{
			Slug:       "permission",
			Searchable: []string{"name", "route"},
			Filterable: map[string]string{"method": "method", "service": "service"},
			Sortable:   map[string]string{"name": "name", "route": "route"},
			Preloads:   []string{"Roles"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Method string `gorm:"column:method" json:"method"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&models.Permission{}).Select("method, count(*) as count").Group("method").Scan(&rows).Error
				return rows, err
			},
		}),
	}
}

// DetailRequest targets a permission via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[models.Permission], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*models.Permission, error) {
	return s.repo.Get(ctx, req.ID)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the permissions handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/permissions/list", "List Permissions", "API")
	reg.Register("POST", "/api/v1/permissions/detail", "Permission Detail", "API")
	g.POST("/api/v1/permissions/list", h.List)
	g.POST("/api/v1/permissions/detail", h.Detail)
}

// List godoc
// @Summary List permissions
// @Description Cursor-paginated catalog of permissions seeded from registered routes.
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/permissions/list [post]
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
// @Summary Get a permission
// @Description Returns one permission by id in the body.
// @Tags permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Permission id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/permissions/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	perm, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, perm)
}
