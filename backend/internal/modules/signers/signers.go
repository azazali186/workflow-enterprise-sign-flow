// Package signers manages signers on a contract: add, patch, remove and list.
package signers

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Service is the signer use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*contracts.Signer, error)
	Patch(ctx context.Context, req PatchRequest) (*contracts.Signer, error)
	Delete(ctx context.Context, req ByIDRequest) error
	List(ctx context.Context, q pagination.Query) (*repo.Page[contracts.Signer], error)
}

type service struct {
	db    *gorm.DB
	repo  *repo.Repo[contracts.Signer]
	audit *audit.Service
}

// NewService wires the signer repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[contracts.Signer]{
			Slug:       "signer",
			Searchable: []string{"name", "email"},
			Filterable: map[string]string{"contract_id": "contract_id", "status": "status", "role": "role"},
			Sortable:   map[string]string{"name": "name", "order": "order"},
			DateFields: []string{"created_at", "signed_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&contracts.Signer{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a,
	}
}

// CreateRequest adds a signer to a contract.
type CreateRequest struct {
	ContractID string `json:"contract_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	Order      int    `json:"order"`
}

// PatchRequest updates a signer.
type PatchRequest struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
	Role  *string `json:"role"`
}

// ByIDRequest targets a signer via the body.
type ByIDRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*contracts.Signer, error) {
	if req.ContractID == "" || req.Name == "" {
		return nil, errs.ErrValidation
	}
	var contract contracts.Contract
	if err := s.db.WithContext(ctx).First(&contract, "id = ?", req.ContractID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	signer := contracts.Signer{
		ContractID: req.ContractID, Name: req.Name, Email: req.Email, Phone: req.Phone,
		Role: orDefault(req.Role, "signer"), Status: contracts.SignerPending, Order: req.Order,
	}
	if err := s.repo.Create(ctx, &signer); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "signer.created", "signer", signer.ID, nil, map[string]any{
		"contract_id": req.ContractID, "name": signer.Name,
	})
	return &signer, nil
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*contracts.Signer, error) {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	signer, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "signer.patched", "signer", req.ID, nil, map[string]any{"name": signer.Name})
	return signer, nil
}

func (s *service) Delete(ctx context.Context, req ByIDRequest) error {
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, "signer.deleted", "signer", req.ID, nil, map[string]any{"deleted": true})
	return nil
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[contracts.Signer], error) {
	return s.repo.List(ctx, q)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the signers handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/signers", "Create Signer", "API")
	reg.Register("PATCH", "/api/v1/signers", "Update Signer", "API")
	reg.Register("DELETE", "/api/v1/signers", "Delete Signer", "API")
	reg.Register("POST", "/api/v1/signers/list", "List Signers", "API")
	g.POST("/api/v1/signers", h.Create)
	g.PATCH("/api/v1/signers", h.Patch)
	g.DELETE("/api/v1/signers", h.Delete)
	g.POST("/api/v1/signers/list", h.List)
}

// Create godoc
// @Summary Add a signer to a contract
// @Description Adds a signer with a role and signing order.
// @Tags signers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Signer payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/signers [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	signer, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, signer)
}

// Patch godoc
// @Summary Update a signer
// @Description Partially updates a signer by id in the body.
// @Tags signers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/signers [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	signer, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, signer)
}

// Delete godoc
// @Summary Remove a signer
// @Description Removes a signer by id in the body.
// @Tags signers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body ByIDRequest true "Signer id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/signers [delete]
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
// @Summary List signers
// @Description Cursor-paginated list with filters, search, sorting, date-range and a status summary.
// @Tags signers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/signers/list [post]
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
