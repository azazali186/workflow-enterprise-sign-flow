package contracts

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Handler exposes the contract HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the contracts handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds contract routes to the registry and engine.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/contracts", "Create Contract", "API")
	reg.Register("PATCH", "/api/v1/contracts", "Update Contract", "API")
	reg.Register("DELETE", "/api/v1/contracts", "Delete Contract", "API")
	reg.Register("POST", "/api/v1/contracts/list", "List Contracts", "API")
	reg.Register("POST", "/api/v1/contracts/detail", "Contract Detail", "API")
	reg.Register("POST", "/api/v1/contracts/send_signature_request", "Send Signature Request", "API")
	reg.Register("POST", "/api/v1/contracts/execute", "Execute Contract", "API")
	reg.Register("POST", "/api/v1/contracts/cancel", "Cancel Contract", "API")
	g.POST("/api/v1/contracts", h.Create)
	g.PATCH("/api/v1/contracts", h.Patch)
	g.DELETE("/api/v1/contracts", h.Delete)
	g.POST("/api/v1/contracts/list", h.List)
	g.POST("/api/v1/contracts/detail", h.Detail)
	g.POST("/api/v1/contracts/send_signature_request", h.SendSignatureRequest)
	g.POST("/api/v1/contracts/execute", h.Execute)
	g.POST("/api/v1/contracts/cancel", h.Cancel)
}

// Create godoc
// @Summary Create a contract
// @Description Creates a contract from a template or free-form payload.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Contract payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/contracts [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, contract)
}

// Patch godoc
// @Summary Update a contract
// @Description Partially updates a contract by id in the body.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, contract)
}

// Delete godoc
// @Summary Delete a contract
// @Description Soft-deletes a contract by id in the body.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Contract id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts [delete]
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
// @Summary List contracts
// @Description Cursor-paginated list with filters, search, sorting, date-range and a status summary.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/contracts/list [post]
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
// @Summary Get a contract
// @Description Returns one contract by id in the body.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Contract id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, contract)
}

// SendSignatureRequest godoc
// @Summary Send a signature request
// @Description Sends signature requests to all pending signers of a contract and publishes an event.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Contract id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts/send_signature_request [post]
func (h *Handler) SendSignatureRequest(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.SendSignatureRequest(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, contract)
}

// Execute godoc
// @Summary Execute a contract
// @Description Marks a fully signed contract as executed.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Contract id"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts/execute [post]
func (h *Handler) Execute(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.Execute(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, contract)
}

// Cancel godoc
// @Summary Cancel a contract
// @Description Cancels a contract by id in the body.
// @Tags contracts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Contract id"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/contracts/cancel [post]
func (h *Handler) Cancel(ctx context.Context, c *app.RequestContext) {
	var req ByIDRequest
	if !apikit.Bind(c, &req) {
		return
	}
	contract, err := h.svc.Cancel(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, contract)
}
