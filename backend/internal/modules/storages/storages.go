// Package storages tracks document/object storage metadata. Object keys are
// encrypted at rest; only non-sensitive metadata is exposed via the API.
package storages

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

// Storage statuses.
const (
	StatusPending  = "pending"
	StatusUploaded = "uploaded"
	StatusFailed   = "failed"
)

// Storage is one stored object reference.
type Storage struct {
	model.Base
	EntityType  string     `gorm:"size:40;index" json:"entity_type"`
	EntityID    string     `gorm:"size:60;index" json:"entity_id"`
	Bucket      string     `gorm:"size:80" json:"bucket"`
	ObjectKey   string     `gorm:"size:255;serializer:enc" json:"-"`
	ContentType string     `gorm:"size:80" json:"content_type"`
	SizeBytes   int64      `json:"size_bytes"`
	Checksum    string     `gorm:"size:64" json:"checksum"`
	Status      string     `gorm:"size:20;default:pending;index" json:"status"`
	UploadedAt  *time.Time `json:"uploaded_at"`
}

// Service is the storage use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Storage, error)
	Patch(ctx context.Context, req PatchRequest) (*Storage, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Storage], error)
	Detail(ctx context.Context, req DetailRequest) (*Storage, error)
}

type service struct {
	repo  *repo.Repo[Storage]
	audit *audit.Service
}

// NewService wires the storage repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		repo: repo.New(db, c, repo.Options[Storage]{
			Slug:       "storage",
			Searchable: []string{"bucket"},
			Filterable: map[string]string{"entity_type": "entity_type", "entity_id": "entity_id", "status": "status"},
			Sortable:   map[string]string{"size_bytes": "size_bytes"},
			DateFields: []string{"created_at", "uploaded_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var out struct {
					Total     int64 `json:"total_objects"`
					TotalSize int64 `json:"total_size_bytes"`
					Uploaded  int64 `json:"uploaded_objects"`
				}
				tx.Model(&Storage{}).Count(&out.Total)
				tx.Model(&Storage{}).Select("COALESCE(SUM(size_bytes),0)").Scan(&out.TotalSize)
				tx.Model(&Storage{}).Where("status = ?", StatusUploaded).Count(&out.Uploaded)
				return out, nil
			},
		}),
		audit: a,
	}
}

// CreateRequest registers an object.
type CreateRequest struct {
	EntityType  string `json:"entity_type"`
	EntityID    string `json:"entity_id"`
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	Checksum    string `json:"checksum"`
}

// PatchRequest updates storage state (e.g. upload completion).
type PatchRequest struct {
	ID        string  `json:"id"`
	Status    *string `json:"status"`
	SizeBytes *int64  `json:"size_bytes"`
	Checksum  *string `json:"checksum"`
}

// DetailRequest targets a storage record via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Storage, error) {
	if req.EntityType == "" || req.EntityID == "" || req.ObjectKey == "" {
		return nil, errs.ErrValidation
	}
	st := Storage{
		EntityType: req.EntityType, EntityID: req.EntityID, Bucket: req.Bucket,
		ObjectKey: req.ObjectKey, ContentType: req.ContentType,
		SizeBytes: req.SizeBytes, Checksum: req.Checksum, Status: StatusPending,
	}
	if err := s.repo.Create(ctx, &st); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "storage.created", "storage", st.ID, nil, map[string]any{
		"entity_type": st.EntityType, "entity_id": st.EntityID,
	})
	return &st, nil
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*Storage, error) {
	updates := map[string]any{}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == StatusUploaded {
			updates["uploaded_at"] = time.Now()
		}
	}
	if req.SizeBytes != nil {
		updates["size_bytes"] = *req.SizeBytes
	}
	if req.Checksum != nil {
		updates["checksum"] = *req.Checksum
	}
	st, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "storage.patched", "storage", req.ID, nil, map[string]any{"status": st.Status})
	return st, nil
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Storage], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*Storage, error) {
	return s.repo.Get(ctx, req.ID)
}

// Handler exposes HTTP routes.
type Handler struct{ svc Service }

// NewHandler builds the storages handler.
func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds routes.
func (h *Handler) Register(reg *registry.Registry, g *route.RouterGroup) {
	reg.Register("POST", "/api/v1/storages", "Create Storage", "API")
	reg.Register("PATCH", "/api/v1/storages", "Update Storage", "API")
	reg.Register("POST", "/api/v1/storages/list", "List Storages", "API")
	reg.Register("POST", "/api/v1/storages/detail", "Storage Detail", "API")
	g.POST("/api/v1/storages", h.Create)
	g.PATCH("/api/v1/storages", h.Patch)
	g.POST("/api/v1/storages/list", h.List)
	g.POST("/api/v1/storages/detail", h.Detail)
}

// Create godoc
// @Summary Register a stored object
// @Description Registers object storage metadata. Object keys are encrypted at rest.
// @Tags storages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateRequest true "Storage metadata"
// @Success 201 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/storages [post]
func (h *Handler) Create(ctx context.Context, c *app.RequestContext) {
	var req CreateRequest
	if !apikit.Bind(c, &req) {
		return
	}
	st, err := h.svc.Create(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, st)
}

// Patch godoc
// @Summary Update storage state
// @Description Updates upload status and size/checksum by id in the body.
// @Tags storages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body PatchRequest true "Updates"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/storages [patch]
func (h *Handler) Patch(ctx context.Context, c *app.RequestContext) {
	var req PatchRequest
	if !apikit.Bind(c, &req) {
		return
	}
	st, err := h.svc.Patch(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, st)
}

// List godoc
// @Summary List stored objects
// @Description Cursor-paginated list with filters, search, sorting, date-range and a size summary.
// @Tags storages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body pagination.Query true "Pagination query"
// @Success 200 {object} response.Body
// @Failure 400 {object} response.Body
// @Router /api/v1/storages/list [post]
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
// @Summary Get a storage record
// @Description Returns one storage record by id in the body.
// @Tags storages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body DetailRequest true "Storage id"
// @Success 200 {object} response.Body
// @Failure 404 {object} response.Body
// @Router /api/v1/storages/detail [post]
func (h *Handler) Detail(ctx context.Context, c *app.RequestContext) {
	var req DetailRequest
	if !apikit.Bind(c, &req) {
		return
	}
	st, err := h.svc.Detail(ctx, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, st)
}
