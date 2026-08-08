package contracts

import (
	"encoding/json"
	"time"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/model"
)

// Contract statuses.
const (
	StatusDraft             = "draft"
	StatusAwaitingSignature = "awaiting_signature"
	StatusPartiallySigned   = "partially_signed"
	StatusSigned            = "signed"
	StatusExecuted          = "executed"
	StatusCancelled         = "cancelled"
	StatusExpired           = "expired"
)

// Signer statuses.
const (
	SignerPending  = "pending"
	SignerSigned   = "signed"
	SignerDeclined = "declined"
)

// Contract is the central document entity.
type Contract struct {
	model.Base
	Title             string          `gorm:"size:200" json:"title"`
	ReferenceNo       string          `gorm:"size:80;uniqueIndex" json:"reference_no"`
	Description       string          `gorm:"type:text" json:"description"`
	Status            string          `gorm:"size:30;index;default:draft" json:"status"`
	TemplateID        string          `gorm:"size:40;index" json:"template_id"`
	CreatedBy         string          `gorm:"size:60;index" json:"created_by"`
	DocumentStorageID string          `gorm:"size:60" json:"document_storage_id"`
	SentAt            *time.Time      `json:"sent_at"`
	ExecutedAt        *time.Time      `json:"executed_at"`
	ExpiresAt         *time.Time      `json:"expires_at"`
	Metadata          json.RawMessage `gorm:"type:jsonb" json:"metadata,omitempty"`
	Signers           []Signer        `gorm:"foreignKey:ContractID" json:"signers,omitempty"`
}

// Signer is a party required to sign a contract.
type Signer struct {
	model.Base
	ContractID   string     `gorm:"size:60;index" json:"contract_id"`
	Name         string     `gorm:"size:120" json:"name"`
	Email        string     `gorm:"size:160;index" json:"email"`
	Phone        string     `gorm:"size:255;serializer:enc" json:"phone"`
	Role         string     `gorm:"size:20;default:signer" json:"role"`
	Status       string     `gorm:"size:20;index;default:pending" json:"status"`
	Order        int        `gorm:"default:1" json:"order"`
	SignURLToken string     `gorm:"size:255" json:"-"` // sha256 of signing token
	SignedAt     *time.Time `json:"signed_at"`
	SignatureID  string     `gorm:"size:60" json:"signature_id"`
}
