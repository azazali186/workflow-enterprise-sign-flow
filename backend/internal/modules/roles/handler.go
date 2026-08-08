package roles

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the roles handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/roles", "Create Role", "API")
	reg.Register("PATCH", "/api/v1/roles", "Update Role", "API")
	reg.Register("DELETE", "/api/v1/roles", "Delete Role", "API")
	reg.Register("POST", "/api/v1/roles/list", "List Roles", "API")
	reg.Register("POST", "/api/v1/roles/detail", "Role Detail", "API")
	reg.Register("PATCH", "/api/v1/roles/assign_permissions", "Assign Role Permissions", "API")
	g.POST("/api/v1/roles", h.Create)
	g.PATCH("/api/v1/roles", h.Patch)
	g.DELETE("/api/v1/roles", h.Delete)
	g.POST("/api/v1/roles/list", h.List)
	g.POST("/api/v1/roles/detail", h.Detail)
	g.PATCH("/api/v1/roles/assign_permissions", h.AssignPermissions)
}

// Create godoc
// @Summary Create a role
// @Description Creates a new RBAC role.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Role payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/roles [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	role, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, role)
}

// Patch godoc
// @Summary Update a role
// @Description Partially updates a role by id in the body. System roles are protected.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/roles [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	role, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}

// Delete godoc
// @Summary Delete a role
// @Description Deletes a role by id in the body. System roles are protected.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Role id"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/roles [delete]
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
// @Summary List roles
// @Description Cursor-paginated list with filters, search, sorting, date-range and a summary.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/roles/list [post]
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
// @Summary Get a role
// @Description Returns one role by id in the body, including its permissions.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Role id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/roles/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	role, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}

// AssignPermissions godoc
// @Summary Assign permissions to a role
// @Description Replaces the role's permission set. System roles are protected.
// @Tags roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body AssignPermissionsRequest true "Permission assignment"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/roles/assign_permissions [patch]
func (h *Handler) AssignPermissions(ctx context.Context, c *app.RequestContext) {
	var req AssignPermissionsRequest
	if !apikit.Bind(c, &req) {
		return
	}
	role, err := h.svc.AssignPermissions(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, role)
}
