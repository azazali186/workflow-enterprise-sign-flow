package contracts

import (
	"encoding/json"
	"time"
)

// CreateRequest creates a contract with its signers.
type CreateRequest struct {
	Title       string          `json:"title"`
	ReferenceNo string          `json:"reference_no"`
	Description string          `json:"description"`
	TemplateID  string          `json:"template_id"`
	ExpiresAt   *time.Time      `json:"expires_at"`
	Metadata    json.RawMessage `json:"metadata"`
	Signers     []SignerInput   `json:"signers"`
}

// SignerInput is a signer to attach on creation.
type SignerInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
	Order int    `json:"order"`
}

// PatchRequest updates contract fields (draft only).
type PatchRequest struct {
	ID          string          `json:"id"`
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	ExpiresAt   *time.Time      `json:"expires_at"`
	Metadata    json.RawMessage `json:"metadata"`
}

// ByIDRequest targets a contract via the body.
type ByIDRequest struct {
	ID string `json:"id"`
}
