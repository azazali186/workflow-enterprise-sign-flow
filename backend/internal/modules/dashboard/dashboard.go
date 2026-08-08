// Package dashboard exposes cross-entity DB summary data for the dashboard.
package dashboard

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signatures"
	"github.com/aeroxe/sign-flow/backend/internal/modules/storages"
	"github.com/aeroxe/sign-flow/backend/internal/modules/templates"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/apikit"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/response"
	"github.com/aeroxe/sign-flow/backend/internal/registry"
)

// Service is the dashboard use-case boundary.
type Service interface {
	Summary(ctx context.Context, req SummaryRequest) (*Summary, error)
}

type service struct{ db *gorm.DB }

// NewService wires the dashboard service.
func NewService(db *gorm.DB) Service { return &service{db: db} }

// SummaryRequest optionally narrows the window.
type SummaryRequest struct {
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
}

// Summary is the dashboard payload.
type Summary struct {
	Contracts struct {
		Total    int64 `json:"total"`
		Draft    int64 `json:"draft"`
		Sent     int64 `json:"sent"`
		Signed   int64 `json:"signed"`
		Executed int64 `json:"executed"`
		Today    int64 `json:"created_today"`
	} `json:"contracts"`
	Signatures struct {
		Total    int64 `json:"total"`
		Captured int64 `json:"captured"`
		Verified int64 `json:"verified"`
		Today    int64 `json:"captured_today"`
	} `json:"signatures"`
	Signers struct {
		Pending int64 `json:"pending"`
		Signed  int64 `json:"signed"`
	} `json:"signers"`
	Templates    int64 `json:"templates"`
	ActiveUsers  int64 `json:"active_users"`
	StorageBytes int64 `json:"storage_bytes"`
}

func (s *service) Summary(ctx context.Context, req SummaryRequest) (*Summary, error) {
	out := &Summary{}
	db := s.db.WithContext(ctx)

	windowed := db.Model(&contracts.Contract{})
	if req.DateFrom != nil {
		windowed = windowed.Where("created_at >= ?", req.DateFrom)
	}
	if req.DateTo != nil {
		windowed = windowed.Where("created_at <= ?", req.DateTo)
	}
	windowed.Count(&out.Contracts.Total)
	db.Model(&contracts.Contract{}).Where("status = ?", contracts.StatusDraft).Count(&out.Contracts.Draft)
	db.Model(&contracts.Contract{}).Where("status IN ?", []string{contracts.StatusAwaitingSignature, contracts.StatusPartiallySigned}).Count(&out.Contracts.Sent)
	db.Model(&contracts.Contract{}).Where("status = ?", contracts.StatusSigned).Count(&out.Contracts.Signed)
	db.Model(&contracts.Contract{}).Where("status = ?", contracts.StatusExecuted).Count(&out.Contracts.Executed)
	db.Model(&contracts.Contract{}).Where("created_at >= ?", startOfDay()).Count(&out.Contracts.Today)

	db.Model(&signatures.Signature{}).Count(&out.Signatures.Total)
	db.Model(&signatures.Signature{}).Where("status = ?", signatures.StatusCaptured).Count(&out.Signatures.Captured)
	db.Model(&signatures.Signature{}).Where("status = ?", signatures.StatusVerified).Count(&out.Signatures.Verified)
	db.Model(&signatures.Signature{}).Where("signed_at >= ?", startOfDay()).Count(&out.Signatures.Today)

	db.Model(&contracts.Signer{}).Where("status = ?", contracts.SignerPending).Count(&out.Signers.Pending)
	db.Model(&contracts.Signer{}).Where("status = ?", contracts.SignerSigned).Count(&out.Signers.Signed)

	db.Model(&templates.Template{}).Count(&out.Templates)
	db.Model(&models.User{}).Where("status = ?", models.UserActive).Count(&out.ActiveUsers)
	db.Model(&storages.Storage{}).Select("COALESCE(SUM(size_bytes),0)").Scan(&out.StorageBytes)
	return out, nil
}

func startOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the dashboard handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/dashboard/summary", "Dashboard Summary", "API")
	g.POST("/api/v1/dashboard/summary", h.Summary)
}

// Summary godoc
// @Summary Get the dashboard summary
// @Description Returns cross-entity counts: contracts, signatures, signers, templates, active users and storage bytes, optionally windowed.
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body SummaryRequest true "Optional date window"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/dashboard/summary [post]
func (h *Handler) Summary(ctx context.Context, c *app.RequestContext) {
	var req SummaryRequest
	if !apikit.Bind(c, &req) {
		return
	}
	out, err := h.svc.Summary(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, out)
}
