package database

import (
	"github.com/aeroxe/sign-flow/backend/internal/audit"
	"github.com/aeroxe/sign-flow/backend/internal/models"
	"github.com/aeroxe/sign-flow/backend/internal/modules/certificates"
	"github.com/aeroxe/sign-flow/backend/internal/modules/compliances"
	"github.com/aeroxe/sign-flow/backend/internal/modules/contracts"
	"github.com/aeroxe/sign-flow/backend/internal/modules/signatures"
	"github.com/aeroxe/sign-flow/backend/internal/modules/storages"
	"github.com/aeroxe/sign-flow/backend/internal/modules/templates"
	"github.com/aeroxe/sign-flow/backend/internal/modules/verifications"
	"github.com/aeroxe/sign-flow/backend/internal/outbox"
)

// AllModels returns every entity table for migrations.
func AllModels() []any {
	return []any{
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&templates.Template{},
		&contracts.Contract{},
		&contracts.Signer{},
		&signatures.Signature{},
		&verifications.Verification{},
		&storages.Storage{},
		&compliances.Compliance{},
		&certificates.Certificate{},
		&audit.Log{},
		&audit.LoginLog{},
		&outbox.Event{},
	}
}
