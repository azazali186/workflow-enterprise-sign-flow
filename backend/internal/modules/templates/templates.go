// Package templates implements reusable contract templates.
package templates

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Template is a reusable document template.
type Template struct {
	model.Base
	Name        string `gorm:"size:160" json:"name"`
	Slug        string `gorm:"size:80;uniqueIndex" json:"slug"`
	Description string `gorm:"size:500" json:"description"`
	Content     string `gorm:"type:text" json:"content"`
	Version     int    `gorm:"default:1" json:"version"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	CreatedBy   string `gorm:"size:60;index" json:"created_by"`
}

// Service is the template use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Template, error)
	Patch(ctx context.Context, req PatchRequest) (*Template, error)
	Delete(ctx context.Context, req ByIDRequest) error
	Detail(ctx context.Context, req ByIDRequest) (*Template, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Template], error)
}

type service struct {
	repo  *repo.Repo[Template]
	audit *audit.Service
}

// NewService wires the template repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[Template]{
			Slug:       "template",
			Searchable: []string{"name", "slug", "description"},
			Filterable: map[string]string{"is_active": "is_active"},
			Sortable:   map[string]string{"name": "name", "version": "version"},
			Summary: func(tx *gorm.DB) (any, error) {
				var out struct {
					Total  int64 `json:"total_templates"`
					Active int64 `json:"active_templates"`
				}
				tx.Model(&Template{}).Count(&out.Total)
				tx.Model(&Template{}).Where("is_active = ?", true).Count(&out.Active)
				return out, nil
			},
		}),
		audit: a,
	}
}

// CreateRequest creates a template.
type CreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// PatchRequest updates a template.
type PatchRequest struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Content     *string `json:"content"`
	IsActive    *bool   `json:"is_active"`
}

// ByIDRequest targets a record via the body.
type ByIDRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Template, error) {
	if req.Name == "" || req.Slug == "" {
		return nil, errs.ErrValidation
	}
	t := Template{
		Name: req.Name, Slug: req.Slug, Description: req.Description,
		Content: req.Content, Version: 1, IsActive: true, CreatedBy: ctxval.UserID(ctx),
	}
	if err := s.repo.Create(ctx, &t); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "template.created", "template", t.ID, nil, map[string]any{"name": t.Name, "slug": t.Slug})
	return s.Detail(ctx, ByIDRequest{ID: t.ID})
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*Template, error) {
	before, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Content != nil {
		updates["content"] = *req.Content
		updates["version"] = before.Version + 1
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	after, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "template.patched", "template", req.ID,
		map[string]any{"name": before.Name, "version": before.Version},
		map[string]any{"name": after.Name, "version": after.Version})
	return after, nil
}

func (s *service) Delete(ctx context.Context, req ByIDRequest) error {
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, "template.deleted", "template", req.ID, nil, map[string]any{"deleted": true})
	return nil
}

func (s *service) Detail(ctx context.Context, req ByIDRequest) (*Template, error) {
	return s.repo.Get(ctx, req.ID)
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Template], error) {
	return s.repo.List(ctx, q)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the templates handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/templates", "Create Template", "API")
	reg.Register("PATCH", "/api/v1/templates", "Update Template", "API")
	reg.Register("DELETE", "/api/v1/templates", "Delete Template", "API")
	reg.Register("POST", "/api/v1/templates/list", "List Templates", "API")
	reg.Register("POST", "/api/v1/templates/detail", "Template Detail", "API")
	g.POST("/api/v1/templates", h.Create)
	g.PATCH("/api/v1/templates", h.Patch)
	g.DELETE("/api/v1/templates", h.Delete)
	g.POST("/api/v1/templates/list", h.List)
	g.POST("/api/v1/templates/detail", h.Detail)
}

// Create godoc
// @Summary Create a template
// @Description Creates a reusable contract template.
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Template payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/templates [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	t, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, t)
}

// Patch godoc
// @Summary Update a template
// @Description Partially updates a template by id; content changes bump the version.
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/templates [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	t, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

// Delete godoc
// @Summary Delete a template
// @Description Soft-deletes a template by id in the body.
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Template id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/templates [delete]
func (h *Handler) Delete(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	if err := h.svc.Delete(ctx, req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// List godoc
// @Summary List templates
// @Description Cursor-paginated list with filters, search, sorting, date-range and a summary.
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/templates/list [post]
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
// @Summary Get a template
// @Description Returns one template by id in the body.
// @Tags templates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Template id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/templates/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	t, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}
