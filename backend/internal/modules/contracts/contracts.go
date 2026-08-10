// Package contracts implements the core contract lifecycle: creation,
// signature request, execution and cancellation — all with transactional
// outbox events, distributed locking and audit logging.
package contracts

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/lock"
	"github.com/aeroxe/sign-flow/backend/internal/outbox"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
)

// Service is the contract use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Contract, error)
	Patch(ctx context.Context, req PatchRequest) (*Contract, error)
	Delete(ctx context.Context, req ByIDRequest) error
	Detail(ctx context.Context, req ByIDRequest) (*Contract, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Contract], error)
	SendSignatureRequest(ctx context.Context, req ByIDRequest) (*Contract, error)
	Execute(ctx context.Context, req ByIDRequest) (*Contract, error)
	Cancel(ctx context.Context, req ByIDRequest) (*Contract, error)
}

type service struct {
	db    *gorm.DB
	repo  *repo.Repo[Contract]
	audit *audit.Service
	bus   *events.Bus
	cache cache.Cache
}

// NewService wires the contract repository, outbox and event bus.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service, bus *events.Bus) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[Contract]{
			Slug:       "contract",
			Searchable: []string{"title", "reference_no", "description"},
			Filterable: map[string]string{"status": "status", "template_id": "template_id", "created_by": "created_by"},
			Sortable:   map[string]string{"title": "title", "reference_no": "reference_no"},
			DateFields: []string{"created_at", "updated_at", "sent_at", "executed_at", "expires_at"},
			Preloads:   []string{"Signers"},
			Summary: func(tx *gorm.DB) (any, error) {
				var byStatus []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				if err := tx.Model(&Contract{}).Select("status, count(*) as count").Group("status").Scan(&byStatus).Error; err != nil {
					return nil, err
				}
				var out struct {
					TotalContracts  int64 `json:"total_contracts"`
					TotalSigners    int64 `json:"total_signers"`
					SignedContracts int64 `json:"signed_contracts"`
					ByStatus        any   `json:"by_status"`
				}
				tx.Model(&Contract{}).Count(&out.TotalContracts)
				tx.Model(&Signer{}).Count(&out.TotalSigners)
				tx.Model(&Contract{}).Where("status = ?", StatusSigned).Count(&out.SignedContracts)
				out.ByStatus = byStatus
				return out, nil
			},
		}),
		audit: a, bus: bus, cache: c,
	}
}

// CreateRequest, PatchRequest and ByIDRequest live in requests.go.

func (s *service) Create(ctx context.Context, req CreateRequest) (*Contract, error) {
	if req.Title == "" || len(req.Signers) == 0 {
		return nil, errs.ErrValidation
	}
	if req.ReferenceNo == "" {
		id, _ := model.NewID()
		req.ReferenceNo = "CT-" + id[:8]
	}
	contract := Contract{
		Title: req.Title, ReferenceNo: req.ReferenceNo, Description: req.Description,
		TemplateID: req.TemplateID, CreatedBy: ctxval.UserID(ctx),
		ExpiresAt: req.ExpiresAt, Metadata: req.Metadata, Status: StatusDraft,
	}
	for i, in := range req.Signers {
		contract.Signers = append(contract.Signers, Signer{
			Name: in.Name, Email: in.Email, Phone: in.Phone, Role: orDefault(in.Role, "signer"),
			Status: SignerPending, Order: orDefaultI(in.Order, i+1),
		})
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateTx(ctx, tx, &contract); err != nil {
			return err
		}
		return outbox.Enqueue(ctx, tx, "contract", contract.ID, "contract_created", map[string]any{
			"contract_id": contract.ID, "reference_no": contract.ReferenceNo, "signers": len(contract.Signers),
		})
	})
	if err != nil {
		return nil, err
	}
	s.bus.Publish(events.Event{EventType: "contract_created", Data: json.RawMessage(fmt.Sprintf(`{"contract_id":%q}`, contract.ID))})
	s.audit.Record(ctx, "contract.created", "contract", contract.ID, nil, map[string]any{
		"title": contract.Title, "reference_no": contract.ReferenceNo, "signers": len(contract.Signers),
	})
	return s.Detail(ctx, ByIDRequest{ID: contract.ID})
}

func (s *service) Patch(ctx context.Context, req PatchRequest) (*Contract, error) {
	before, err := s.Detail(ctx, ByIDRequest{ID: req.ID})
	if err != nil {
		return nil, err
	}
	if before.Status != StatusDraft {
		return nil, errs.New(400, 40020, "only draft contracts can be patched")
	}
	updates := map[string]any{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = *req.ExpiresAt
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}
	after, err := s.repo.Patch(ctx, req.ID, updates)
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "contract.patched", "contract", req.ID,
		map[string]any{"title": before.Title},
		map[string]any{"title": after.Title})
	return after, nil
}

func (s *service) Delete(ctx context.Context, req ByIDRequest) error {
	contract, err := s.Detail(ctx, req)
	if err != nil {
		return err
	}
	if contract.Status != StatusDraft {
		return errs.New(400, 40020, "only draft contracts can be deleted")
	}
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return err
	}
	s.audit.Record(ctx, "contract.deleted", "contract", req.ID, nil, map[string]any{"deleted": true})
	return nil
}

func (s *service) Detail(ctx context.Context, req ByIDRequest) (*Contract, error) {
	return s.repo.Get(ctx, req.ID)
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Contract], error) {
	return s.repo.List(ctx, q)
}

// SendSignatureRequest generates per-signer signing tokens and moves the
// contract to awaiting_signature.
func (s *service) SendSignatureRequest(ctx context.Context, req ByIDRequest) (*Contract, error) {
	contract, err := s.Detail(ctx, req)
	if err != nil {
		return nil, err
	}
	if contract.Status != StatusDraft && contract.Status != StatusPartiallySigned {
		return nil, errs.New(400, 40021, "signature request already sent")
	}
	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range contract.Signers {
			tok, _ := model.NewID()
			sum := sha256.Sum256([]byte(tok))
			if err := tx.Model(&Signer{}).Where("id = ?", contract.Signers[i].ID).
				Update("sign_url_token", fmt.Sprintf("%x", sum)).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&Contract{}).Where("id = ?", req.ID).
			Updates(map[string]any{"status": StatusAwaitingSignature, "sent_at": now}).Error; err != nil {
			return err
		}
		return outbox.Enqueue(ctx, tx, "contract", req.ID, "signature_requested", map[string]any{
			"contract_id": req.ID, "sent_at": now,
		})
	})
	if err != nil {
		return nil, err
	}
	s.bus.Publish(events.Event{EventType: "signature_requested", Data: json.RawMessage(fmt.Sprintf(`{"contract_id":%q}`, req.ID))})
	s.audit.Record(ctx, "contract.signature_requested", "contract", req.ID,
		map[string]any{"status": contract.Status}, map[string]any{"status": StatusAwaitingSignature})
	return s.Detail(ctx, req)
}

// Execute finalises an executed contract behind a distributed lock.
func (s *service) Execute(ctx context.Context, req ByIDRequest) (*Contract, error) {
	distLock, err := lock.Acquire(ctx, s.cache, "contract:execute:"+req.ID, 10*time.Second)
	if err != nil {
		return nil, err
	}
	defer distLock.Release(ctx)
	contract, err := s.Detail(ctx, req)
	if err != nil {
		return nil, err
	}
	if contract.Status != StatusSigned {
		return nil, errs.New(400, 40022, "contract must be fully signed before execution")
	}
	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Contract{}).Where("id = ?", req.ID).
			Updates(map[string]any{"status": StatusExecuted, "executed_at": now}).Error; err != nil {
			return err
		}
		return outbox.Enqueue(ctx, tx, "contract", req.ID, "contract_executed", map[string]any{
			"contract_id": req.ID, "executed_at": now,
		})
	})
	if err != nil {
		return nil, err
	}
	s.bus.Publish(events.Event{EventType: "contract_executed", Data: json.RawMessage(fmt.Sprintf(`{"contract_id":%q}`, req.ID))})
	s.audit.Record(ctx, "contract.executed", "contract", req.ID,
		map[string]any{"status": contract.Status}, map[string]any{"status": StatusExecuted})
	return s.Detail(ctx, req)
}

// Cancel moves a pending contract to cancelled.
func (s *service) Cancel(ctx context.Context, req ByIDRequest) (*Contract, error) {
	contract, err := s.Detail(ctx, req)
	if err != nil {
		return nil, err
	}
	if contract.Status != StatusDraft && contract.Status != StatusAwaitingSignature {
		return nil, errs.New(400, 40023, "contract cannot be cancelled in current state")
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Contract{}).Where("id = ?", req.ID).Update("status", StatusCancelled).Error; err != nil {
			return err
		}
		return outbox.Enqueue(ctx, tx, "contract", req.ID, "contract_cancelled", map[string]any{"contract_id": req.ID})
	})
	if err != nil {
		return nil, err
	}
	s.audit.Record(ctx, "contract.cancelled", "contract", req.ID,
		map[string]any{"status": contract.Status}, map[string]any{"status": StatusCancelled})
	return s.Detail(ctx, req)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func orDefaultI(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
