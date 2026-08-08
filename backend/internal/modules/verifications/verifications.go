// Package verifications implements identity verification (OTP) with hashed
// codes, expiry and attempt limits — no codes are ever stored in plaintext.
package verifications

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"

	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/cache"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signatures"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/ctxval"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/pagination"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/repo"
)

// Verification statuses and methods.
const (
	StatusPending  = "pending"
	StatusVerified = "verified"
	StatusFailed   = "failed"
	MethodOTP      = "otp"
	maxAttempts    = 5
	otpTTL         = 10 * time.Minute
)

// Verification tracks an identity check for a signature.
type Verification struct {
	model.Base
	SignatureID  string     `gorm:"size:60;index" json:"signature_id"`
	ContractID   string     `gorm:"size:60;index" json:"contract_id"`
	Method       string     `gorm:"size:30" json:"method"`
	Status       string     `gorm:"size:20;index;default:pending" json:"status"`
	Attempts     int        `gorm:"default:0" json:"attempts"`
	OTPHash      string     `gorm:"size:255" json:"-"` // sha256 of the code
	OTPExpiresAt *time.Time `json:"otp_expires_at"`
	VerifiedBy   string     `gorm:"size:60" json:"verified_by"`
	VerifiedAt   *time.Time `json:"verified_at"`
}

// Service is the verification use-case boundary.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*Verification, error)
	Verify(ctx context.Context, req VerifyRequest) (*Verification, error)
	List(ctx context.Context, q pagination.Query) (*repo.Page[Verification], error)
	Detail(ctx context.Context, req DetailRequest) (*Verification, error)
}

type service struct {
	db    *gorm.DB
	repo  *repo.Repo[Verification]
	audit *audit.Service
}

// NewService wires the verification repository.
func NewService(db *gorm.DB, c cache.Cache, a *audit.Service) Service {
	return &service{
		db: db,
		repo: repo.New(db, c, repo.Options[Verification]{
			Slug:       "verification",
			Searchable: []string{},
			Filterable: map[string]string{"status": "status", "method": "method", "signature_id": "signature_id"},
			Sortable:   map[string]string{},
			DateFields: []string{"created_at", "verified_at"},
			Summary: func(tx *gorm.DB) (any, error) {
				var rows []struct {
					Status string `gorm:"column:status" json:"status"`
					Count  int64  `gorm:"column:count" json:"count"`
				}
				err := tx.Model(&Verification{}).Select("status, count(*) as count").Group("status").Scan(&rows).Error
				return rows, err
			},
		}),
		audit: a,
	}
}

// CreateRequest starts an OTP verification for a signature.
type CreateRequest struct {
	SignatureID string `json:"signature_id"`
	Method      string `json:"method"`
}

// VerifyRequest submits the OTP code.
type VerifyRequest struct {
	VerificationID string `json:"verification_id"`
	Code           string `json:"code"`
}

// DetailRequest targets a verification via the body.
type DetailRequest struct {
	ID string `json:"id"`
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*Verification, error) {
	if req.SignatureID == "" {
		return nil, errs.ErrValidation
	}
	var sig signatures.Signature
	if err := s.db.WithContext(ctx).First(&sig, "id = ?", req.SignatureID).Error; err != nil {
		return nil, errs.ErrNotFound
	}
	code, err := randomCode()
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256([]byte(code))
	expires := time.Now().Add(otpTTL)
	v := Verification{
		SignatureID: req.SignatureID, ContractID: sig.ContractID,
		Method: orDefault(req.Method, MethodOTP), Status: StatusPending,
		OTPHash: hex.EncodeToString(hash[:]), OTPExpiresAt: &expires,
	}
	if err := s.repo.Create(ctx, &v); err != nil {
		return nil, err
	}
	// In production the code is delivered out-of-band (email/SMS). It is never
	// stored in plaintext and never written to logs or audit records.
	s.audit.Record(ctx, "verification.created", "verification", v.ID, nil, map[string]any{
		"signature_id": req.SignatureID, "method": v.Method, "expires_at": expires,
	})
	return &v, nil
}

func (s *service) Verify(ctx context.Context, req VerifyRequest) (*Verification, error) {
	if req.VerificationID == "" || req.Code == "" {
		return nil, errs.ErrValidation
	}
	v, err := s.repo.Get(ctx, req.VerificationID)
	if err != nil {
		return nil, err
	}
	if v.Status == StatusVerified {
		return v, nil
	}
	if v.Attempts >= maxAttempts || v.OTPExpiresAt == nil || time.Now().After(*v.OTPExpiresAt) {
		if _, err := s.repo.Patch(ctx, v.ID, map[string]any{"status": StatusFailed}); err != nil {
			return nil, err
		}
		return nil, errs.New(400, 40040, "verification expired")
	}
	hash := sha256.Sum256([]byte(req.Code))
	if hex.EncodeToString(hash[:]) != v.OTPHash {
		_, _ = s.repo.Patch(ctx, v.ID, map[string]any{"attempts": v.Attempts + 1})
		s.audit.Record(ctx, "verification.failed", "verification", v.ID, nil, map[string]any{"attempts": v.Attempts + 1})
		return nil, errs.New(400, 40041, "invalid verification code")
	}
	now := time.Now()
	updated, err := s.repo.Patch(ctx, v.ID, map[string]any{
		"status": StatusVerified, "verified_at": now, "verified_by": ctxval.UserID(ctx),
	})
	if err != nil {
		return nil, err
	}
	if v.SignatureID != "" {
		_ = s.db.WithContext(ctx).Model(&signatures.Signature{}).Where("id = ?", v.SignatureID).
			Update("status", signatures.StatusVerified)
	}
	s.audit.Record(ctx, "verification.verified", "verification", v.ID,
		map[string]any{"status": StatusPending}, map[string]any{"status": StatusVerified})
	return updated, nil
}

func (s *service) List(ctx context.Context, q pagination.Query) (*repo.Page[Verification], error) {
	return s.repo.List(ctx, q)
}

func (s *service) Detail(ctx context.Context, req DetailRequest) (*Verification, error) {
	return s.repo.Get(ctx, req.ID)
}

func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// Handler exposes HTTP routes. Implemented in handler.go.
