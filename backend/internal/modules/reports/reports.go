// Package reports exposes cursor-paginated report endpoints with the same
// filters, sorting, date-range and summary capabilities as entity lists.
package reports

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/aeroxe/sign-flow/backend/internal/modules/auditlogs"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signatures"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signers"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Handler exposes the report HTTP routes.
type Handler struct {
	contracts  contracts.Service
	signatures signatures.Service
	auditlogs  auditlogs.Service
	signers    signers.Service
}

// NewHandler wires the report handler with entity services.
func NewHandler(contracts contracts.Service, signatures signatures.Service, auditlogs auditlogs.Service, signers signers.Service) *Handler {
	return &Handler{contracts: contracts, signatures: signatures, auditlogs: auditlogs, signers: signers}
}

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/reports/contracts", "Contract Report", "API")
	reg.Register("POST", "/api/v1/reports/signatures", "Signature Report", "API")
	reg.Register("POST", "/api/v1/reports/signers", "Signer Report", "API")
	reg.Register("POST", "/api/v1/reports/audit_logs", "Audit Log Report", "API")
	g.POST("/api/v1/reports/contracts", h.Contracts)
	g.POST("/api/v1/reports/signatures", h.Signatures)
	g.POST("/api/v1/reports/signers", h.Signers)
	g.POST("/api/v1/reports/audit_logs", h.AuditLogs)
}

// Contracts godoc
// @Summary Contract report
// @Description Cursor-paginated contract report with filters, search, sorting, date-range and a summary.
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/reports/contracts [post]
func (h *Handler) Contracts(ctx context.Context, c *app.RequestContext) {
	var q pagination.Query
	if !apikit.Bind(c, &q) {
		return
	}
	page, err := h.contracts.List(ctx, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, page)
}

// Signatures godoc
// @Summary Signature report
// @Description Cursor-paginated signature report with filters, search, sorting, date-range and a summary.
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/reports/signatures [post]
func (h *Handler) Signatures(ctx context.Context, c *app.RequestContext) {
	var q pagination.Query
	if !apikit.Bind(c, &q) {
		return
	}
	page, err := h.signatures.List(ctx, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, page)
}

// Signers godoc
// @Summary Signer report
// @Description Cursor-paginated signer report with filters, search, sorting, date-range and a summary.
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/reports/signers [post]
func (h *Handler) Signers(ctx context.Context, c *app.RequestContext) {
	var q pagination.Query
	if !apikit.Bind(c, &q) {
		return
	}
	page, err := h.signers.List(ctx, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, page)
}

// AuditLogs godoc
// @Summary Audit log report
// @Description Cursor-paginated audit log report with filters, search, sorting, date-range and a summary.
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/reports/audit_logs [post]
func (h *Handler) AuditLogs(ctx context.Context, c *app.RequestContext) {
	var q pagination.Query
	if !apikit.Bind(c, &q) {
		return
	}
	page, err := h.auditlogs.List(ctx, q)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, page)
}
