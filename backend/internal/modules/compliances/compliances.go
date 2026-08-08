// Package compliances implements compliance checks (gdpr, esign, retention)
// per contract with a review workflow.
package compliances

import (
	"context"
	"encoding/json"
	"time"

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

// Compliance statuses.
const (
	StatusPending      = "pending"
	StatusCompliant    = "compliant"
	StatusNonCompliant = "non_compliant"
	StatusUnderReview  = "under_review"
)

// Compliance is one compliance check.
type Compliance struct {
	model.Base
	ContractID string          `gorm:"size:60;index" json:"contract_id"`
	Type       string          `gorm:"size:40;index" json:"type"`
	Status     string          `gorm:"size:30;index;default:pending" json:"status"`
	Evidence   json.RawMessage `gorm:"type:jsonb" json:"evidence,omitempty"`
	ReviewedBy string          `gorm:"size:60" json:"reviewed_by"`
	ReviewedAt *time.Time      `json:"reviewed_at"`
}

// Service is the compliance use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Compliance, error)
	Patch(ctx context.Context, req PatchRequest) (*Compliance, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Compliance], error)
	Detail(ctx context.Context, req DetailRequest) (*Compliance, error)
}

type service struct {
	repo  *repo.Repo[Compliance]
	audit *audit.Service
}

// NewService wires the compliance repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[Compliance]{
			Slug:       "compliance",
			Searchable: []string{},
			Filterable: map[string]string{"contract_id": "contract_id", "type": "type", "status": "status"},
			Sortable:   map[string]string{},
			DateFields: []string{"created_at", "reviewed_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&Compliance{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a,
	}
}

// CreateRequest opens a compliance check.
type CreateRequest struct {
	ContractID string          `json:"contract_id"`
	Type       string          `json:"type"`
	Evidence   json.RawMessage `json:"evidence"`
}

// PatchRequest reviews a compliance check.
type PatchRequest struct {
	ID       string          `json:"id"`
	Status   *string         `json:"status"`
	Evidence json.RawMessage `json:"evidence"`
}

// DetailRequest targets a compliance record via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Compliance, error) {
	if req.ContractID == "" || req.Type == "" {
		return nil, errs.ErrValidation
	}
	c := Compliance{ContractID: req.ContractID, Type: req.Type, Status: StatusPending, Evidence: req.Evidence}
	if err := s.repo.Create(ctx, &c); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "compliance.created", "compliance", c.ID, nil, map[string]any{
		"contract_id": c.ContractID, "type": c.Type,
	})
	return &c, nil
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*Compliance, error) {
	updates := map[string]any{}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status != StatusPending {
			updates["reviewed_by"] = ctxval.UserID(ctx)
			updates["reviewed_at"] = time.Now()
		}
	}
	if req.Evidence != nil {
		updates["evidence"] = req.Evidence
	}
	c, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "compliance.patched", "compliance", req.ID, nil, map[string]any{"status": c.Status})
	return c, nil
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Compliance], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*Compliance, error) {
	return s.repo.Get(ctx, req.ID)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the compliances handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/compliances", "Create Compliance", "API")
	reg.Register("PATCH", "/api/v1/compliances", "Review Compliance", "API")
	reg.Register("POST", "/api/v1/compliances/list", "List Compliances", "API")
	reg.Register("POST", "/api/v1/compliances/detail", "Compliance Detail", "API")
	g.POST("/api/v1/compliances", h.Create)
	g.PATCH("/api/v1/compliances", h.Patch)
	g.POST("/api/v1/compliances/list", h.List)
	g.POST("/api/v1/compliances/detail", h.Detail)
}

// Create godoc
// @Summary Open a compliance check
// @Description Creates a compliance check (gdpr, esign, retention) for a contract.
// @Tags compliances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Compliance check"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/compliances [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	cc, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, cc)
}

// Patch godoc
// @Summary Review a compliance check
// @Description Reviews a check by id; sets status, reviewer and reviewed_at.
// @Tags compliances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Review"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/compliances [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	cc, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cc)
}

// List godoc
// @Summary List compliance checks
// @Description Cursor-paginated list with filters, sorting, date-range and a status summary.
// @Tags compliances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/compliances/list [post]
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
// @Summary Get a compliance check
// @Description Returns one compliance check by id in the body.
// @Tags compliances
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Compliance id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/compliances/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	cc, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, cc)
}
