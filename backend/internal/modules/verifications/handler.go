package verifications

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

// NewHandler builds the verifications handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/verifications", "Create Verification", "API")
	reg.Register("POST", "/api/v1/verifications/verify", "Verify Code", "API")
	reg.Register("POST", "/api/v1/verifications/list", "List Verifications", "API")
	reg.Register("POST", "/api/v1/verifications/detail", "Verification Detail", "API")
	g.POST("/api/v1/verifications", h.Create)
	g.POST("/api/v1/verifications/verify", h.Verify)
	g.POST("/api/v1/verifications/list", h.List)
	g.POST("/api/v1/verifications/detail", h.Detail)
}

// Create godoc
// @Summary Start a verification
// @Description Starts an OTP verification for a signature. The code is never stored or returned in plaintext.
// @Tags verifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Verification request"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/verifications [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	v, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, v)
}

// Verify godoc
// @Summary Verify an OTP code
// @Description Submits the OTP code for a verification record.
// @Tags verifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body VerifyRequest true "OTP submission"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/verifications/verify [post]
func (h *Handler) Verify(ctx context.Context, c *app.RequestContext) {
	var req VerifyRequest
	if !apikit.Bind(c, &req) {
		return
	}
	v, err := h.svc.Verify(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, v)
}

// List godoc
// @Summary List verifications
// @Description Cursor-paginated list with filters, sorting, date-range and a status summary.
// @Tags verifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/verifications/list [post]
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
// @Summary Get a verification
// @Description Returns one verification by id in the body.
// @Tags verifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Verification id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/verifications/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	v, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, v)
}
