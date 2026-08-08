// Package cryptoser registers a GORM serializer that transparently encrypts
// fields marked with `gorm:"serializer:enc"` at rest.
package cryptoser

import (
	"context"
	"reflect"

	"gorm.io/gorm/schema"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/crypto"
)

// Name is the serializer name registered on the GORM connection.
const Name = "enc"

// Enc implements schema.SerializerInterface with AES-256-GCM.
type Enc struct{}

// Scan decrypts the stored value into the destination field.
func (Enc) Scan(_ context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}
	plain, err := crypto.Decrypt(asString(dbValue))
	if err != nil {
		return err
	}
	field.ReflectValueOf(context.Background(), dst).SetString(plain)
	return nil
}

// Value encrypts the field before it is persisted.
func (Enc) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}
	s, ok := fieldValue.(string)
	if !ok {
		return fieldValue, nil
	}
	enc, err := crypto.Encrypt(s)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

func asString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return ""
	}
}
