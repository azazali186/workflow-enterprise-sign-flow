package ctxval

import (
	"github.com/google/uuid"
)

// NewRequestID returns a fresh UUID v7 string.
func NewRequestID() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
