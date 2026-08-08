package signatures

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Handler exposes the signature HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the signatures handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds signature routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/signatures/capture", "Capture Signature", "API")
	reg.Register("POST", "/api/v1/signatures/list", "List Signatures", "API")
	reg.Register("POST", "/api/v1/signatures/detail", "Signature Detail", "API")
	g.POST("/api/v1/signatures/capture", h.Capture)
	g.POST("/api/v1/signatures/list", h.List)
	g.POST("/api/v1/signatures/detail", h.Detail)
}

// Capture godoc
// @Summary Capture a signature
// @Description Records a captured signature for a signer and publishes an event.
// @Tags signatures
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CaptureRequest true "Signature capture"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/signatures/capture [post]
func (h *Handler) Capture(ctx context.Context, c *app.RequestContext) {
	var req CaptureRequest
	if !apikit.Bind(c, &req) {
		return
	}
	sig, err := h.svc.Capture(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, sig)
}

// List godoc
// @Summary List signatures
// @Description Cursor-paginated list with filters, search, sorting, date-range and a summary.
// @Tags signatures
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/signatures/list [post]
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
// @Summary Get a signature
// @Description Returns one signature by id in the body.
// @Tags signatures
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Signature id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/signatures/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	sig, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sig)
}
