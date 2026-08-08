// Package certificates issues completion certificates for executed contracts.
// Certificate payloads (PEM etc.) are encrypted at rest.
package certificates

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Certificate statuses.
const (
	StatusValid   = "valid"
	StatusRevoked = "revoked"
	StatusExpired = "expired"
)

// Certificate is a completion certificate.
type Certificate struct {
	model.Base
	ContractID   string     `gorm:"size:60;index" json:"contract_id"`
	Subject      string     `gorm:"size:200" json:"subject"`
	Issuer       string     `gorm:"size:200" json:"issuer"`
	SerialNumber string     `gorm:"size:80;uniqueIndex" json:"serial_number"`
	NotBefore    *time.Time `json:"not_before"`
	NotAfter     *time.Time `json:"not_after"`
	Data         string     `gorm:"type:text;serializer:enc" json:"-"`
	Status       string     `gorm:"size:20;index;default:valid" json:"status"`
}

// Service is the certificate use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Certificate, error)
	Patch(ctx context.Context, req PatchRequest) (*Certificate, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Certificate], error)
	Detail(ctx context.Context, req DetailRequest) (*Certificate, error)
}

type service struct {
	repo  *repo.Repo[Certificate]
	audit *audit.Service
}

// NewService wires the certificate repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[Certificate]{
			Slug:       "certificate",
			Searchable: []string{"subject", "issuer", "serial_number"},
			Filterable: map[string]string{"contract_id": "contract_id", "status": "status"},
			Sortable:   map[string]string{"subject": "subject"},
			DateFields: []string{"created_at", "not_before", "not_after"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&Certificate{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a,
	}
}

// CreateRequest issues a certificate.
type CreateRequest struct {
	ContractID   string     `json:"contract_id"`
	Subject      string     `json:"subject"`
	Issuer       string     `json:"issuer"`
	SerialNumber string     `json:"serial_number"`
	NotBefore    *time.Time `json:"not_before"`
	NotAfter     *time.Time `json:"not_after"`
	Data         string     `json:"data"`
}

// PatchRequest updates certificate state.
type PatchRequest struct {
	ID     string  `json:"id"`
	Status *string `json:"status"`
}

// DetailRequest targets a certificate via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Certificate, error) {
	if req.ContractID == "" || req.SerialNumber == "" {
		return nil, errs.ErrValidation
	}
	c := Certificate{
		ContractID: req.ContractID, Subject: req.Subject, Issuer: req.Issuer,
		SerialNumber: req.SerialNumber, NotBefore: req.NotBefore, NotAfter: req.NotAfter,
		Data: req.Data, Status: StatusValid,
	}
	if err := s.repo.Create(ctx, &c); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "certificate.created", "certificate", c.ID, nil, map[string]any{
		"contract_id": c.ContractID, "serial_number": c.SerialNumber,
	})
	return &c, nil
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*Certificate, error) {
	updates := map[string]any{}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	c, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "certificate.patched", "certificate", req.ID, nil, map[string]any{"status": c.Status})
	return c, nil
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Certificate], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*Certificate, error) {
	return s.repo.Get(ctx, req.ID)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the certificates handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/certificates", "Issue Certificate", "API")
	reg.Register("PATCH", "/api/v1/certificates", "Update Certificate", "API")
	reg.Register("POST", "/api/v1/certificates/list", "List Certificates", "API")
	reg.Register("POST", "/api/v1/certificates/detail", "Certificate Detail", "API")
	g.POST("/api/v1/certificates", h.Create)
	g.PATCH("/api/v1/certificates", h.Patch)
	g.POST("/api/v1/certificates/list", h.List)
	g.POST("/api/v1/certificates/detail", h.Detail)
}

// Create godoc
// @Summary Issue a certificate
// @Description Issues a completion certificate for an executed contract. Payload is encrypted at rest.
// @Tags certificates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Certificate payload"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/certificates [post]
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
// @Summary Update a certificate
// @Description Updates certificate status (e.g. revoke) by id in the body.
// @Tags certificates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/certificates [patch]
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
// @Summary List certificates
// @Description Cursor-paginated list with filters, search, sorting, date-range and a status summary.
// @Tags certificates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/certificates/list [post]
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
// @Summary Get a certificate
// @Description Returns one certificate by id in the body.
// @Tags certificates
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Certificate id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/certificates/detail [post]
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
