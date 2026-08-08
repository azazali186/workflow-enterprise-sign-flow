package users

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

// NewHandler builds the users handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/users", "Create User", "API")
	reg.Register("PATCH", "/api/v1/users", "Update User", "API")
	reg.Register("DELETE", "/api/v1/users", "Delete User", "API")
	reg.Register("POST", "/api/v1/users/list", "List Users", "API")
	reg.Register("POST", "/api/v1/users/detail", "User Detail", "API")
	reg.Register("PATCH", "/api/v1/users/assign_roles", "Assign User Roles", "API")
	g.POST("/api/v1/users", h.Create)
	g.PATCH("/api/v1/users", h.Patch)
	g.DELETE("/api/v1/users", h.Delete)
	g.POST("/api/v1/users/list", h.List)
	g.POST("/api/v1/users/detail", h.Detail)
	g.PATCH("/api/v1/users/assign_roles", h.AssignRoles)
}

// Create godoc
// @Summary Create a user
// @Description Creates a new user account with optional roles.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "User payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 409 {object} response.Body
// @Router /api/v1/users [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	user, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, user)
}

// Patch godoc
// @Summary Update a user
// @Description Partially updates a user by id in the body.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/users [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	user, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// Delete godoc
// @Summary Delete a user
// @Description Soft-deletes a user by id in the body.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "User id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/users [delete]
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
// @Summary List users
// @Description Cursor-paginated list with filters, search, sorting, date-range and a status summary.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/users/list [post]
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
// @Summary Get a user
// @Description Returns one user by id in the body.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "User id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/users/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	user, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}

// AssignRoles godoc
// @Summary Assign roles to a user
// @Description Replaces the user's role set.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body AssignRolesRequest true "Role assignment"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/users/assign_roles [patch]
func (h *Handler) AssignRoles(ctx context.Context, c *app.RequestContext) {
	var req AssignRolesRequest
	if !apikit.Bind(c, &req) {
		return
	}
	user, err := h.svc.AssignRoles(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, user)
}
