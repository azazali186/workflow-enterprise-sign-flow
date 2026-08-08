// Package model defines shared persistence base types.
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base is embedded in every entity: UUID v7 primary key + timestamps + soft delete.
// The PK is stored as varchar (not native uuid) so every foreign key column
// inherits a compatible type and empty string FKs remain valid.
type Base struct {
	ID        string         `gorm:"type:varchar(40);primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate assigns a UUID v7 when not already set.
func (b *Base) BeforeCreate(_ *gorm.DB) error {
	if b.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		b.ID = id.String()
	}
	return nil
}

// NewID returns a fresh UUID v7 string.
func NewID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
