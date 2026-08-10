// Package signatures captures electronic signatures behind a distributed lock,
// advances signer/contract state and emits outbox + WebSocket events.
package signatures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/events"
	"github.com/aeroxe/sign-flow/backend/internal/lock"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/outbox"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
)

// Signature statuses.
const (
	StatusPending  = "pending"
	StatusCaptured = "captured"
	StatusVerified = "verified"
	StatusInvalid  = "invalid"
	StatusDeclined = "declined"
)

// Signature is one captured signature.
type Signature struct {
	model.Base
	ContractID     string     `gorm:"size:60;index" json:"contract_id"`
	SignerID       string     `gorm:"size:60;index" json:"signer_id"`
	Status         string     `gorm:"size:20;index;default:pending" json:"status"`
	Type           string     `gorm:"size:20" json:"type"` // draw | type | upload
	Data           string     `gorm:"type:text;serializer:enc" json:"-"`
	IPAddress      string     `gorm:"size:64" json:"ip_address"`
	UserAgent      string     `gorm:"size:255" json:"user_agent"`
	SignedAt       *time.Time `json:"signed_at"`
	VerificationID string     `gorm:"size:60" json:"verification_id"`
}

// Service is the signature use-case boundary.
type Service interface {
	Capture(ctx context.Context, req CaptureRequest) (*Signature, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Signature], error)
	Detail(ctx context.Context, req DetailRequest) (*Signature, error)
}

type service struct {
	db    *gorm.DB
	repo  *repo.Repo[Signature]
	audit *audit.Service
	bus   *events.Bus
	cache cache.Cache
}

// NewService wires the signature repository, outbox and bus.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service, bus *events.Bus) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[Signature]{
			Slug:       "signature",
			Searchable: []string{},
			Filterable: map[string]string{"contract_id": "contract_id", "signer_id": "signer_id", "status": "status"},
			Sortable:   map[string]string{},
			DateFields: []string{"created_at", "signed_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&Signature{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a, bus: bus, cache: c,
	}
}

// CaptureRequest captures a signature for a signer.
type CaptureRequest struct {
	ContractID     string `json:"contract_id"`
	SignerID       string `json:"signer_id"`
	SignatureType  string `json:"signature_type"`
	Data           string `json:"data"`
	VerificationID string `json:"verification_id"`
}

// DetailRequest targets a signature via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

// Capture records the signature and advances contract state atomically.
func (s *service) Capture(ctx context.Context, req CaptureRequest) (*Signature, error) {
	if req.ContractID == "" || req.SignerID == "" || req.Data == "" {
		return nil, errs.ErrValidation
	}
	distLock, err := lock.Acquire(ctx, s.cache, "contract:sign:"+req.ContractID, 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer distLock.Release(ctx)

	var signer contracts.Signer
	if err := s.db.WithContext(ctx).First(&signer, "id = ? AND contract_id = ?", req.SignerID, req.ContractID).Error; err != nil {
		return nil, errs.ErrNotFound
	}
	if signer.Status != contracts.SignerPending {
		return nil, errs.New(409, 40920, "signer already signed")
	}
	var contract contracts.Contract
	if err := s.db.WithContext(ctx).First(&contract, "id = ?", req.ContractID).Error; err != nil {
		return nil, errs.ErrNotFound
	}
	if contract.Status != contracts.StatusAwaitingSignature && contract.Status != contracts.StatusPartiallySigned {
		return nil, errs.New(400, 40030, "contract is not awaiting signatures")
	}
	now := time.Now()
	sig := Signature{
		ContractID: req.ContractID, SignerID: req.SignerID, Status: StatusCaptured,
		Type: req.SignatureType, Data: req.Data,
		IPAddress: ctxval.IP(ctx), UserAgent: ctxval.UserAgent(ctx),
		SignedAt: &now, VerificationID: req.VerificationID,
	}
	var newContractStatus string
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&sig).Error; err != nil {
			return err
		}
		if err := tx.Model(&contracts.Signer{}).Where("id = ?", req.SignerID).Updates(map[string]any{
			"status": contracts.SignerSigned, "signed_at": now, "signature_id": sig.ID,
		}).Error; err != nil {
			return err
		}
		var pending int64
		if err := tx.Model(&contracts.Signer{}).Where("contract_id = ? AND status != ?", req.ContractID, contracts.SignerSigned).Count(&pending).Error; err != nil {
			return err
		}
		if pending == 0 {
			newContractStatus = contracts.StatusSigned
		} else {
			newContractStatus = contracts.StatusPartiallySigned
		}
		if err := tx.Model(&contracts.Contract{}).Where("id = ?", req.ContractID).Update("status", newContractStatus).Error; err != nil {
			return err
		}
		return outbox.Enqueue(ctx, tx, "signature", sig.ID, "signed_update", map[string]any{
			"signature_id": sig.ID, "contract_id": req.ContractID, "signer_id": req.SignerID,
			"contract_status": newContractStatus,
		})
	})
	if err != nil {
		return nil, err
	}
	s.bus.Publish(events.Event{EventType: "signed_update", Data: json.RawMessage(fmt.Sprintf(
		`{"signature_id":%q,"contract_id":%q,"contract_status":%q}`, sig.ID, req.ContractID, newContractStatus))})
	s.audit.Record(ctx, "signature.captured", "signature", sig.ID,
		map[string]any{"signer_status": signer.Status},
		map[string]any{"contract_status": newContractStatus, "signature_type": req.SignatureType})
	return s.Detail(ctx, DetailRequest{ID: sig.ID})
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Signature], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*Signature, error) {
	return s.repo.Get(ctx, req.ID)
}
