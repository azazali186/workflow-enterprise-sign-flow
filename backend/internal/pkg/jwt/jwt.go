// Package jwt provides signed access tokens with single-login validation.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/errs"
)

// Claims is the JWT payload.
type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

// Manager issues and verifies tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// New builds a token manager.
func New(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// TTL returns the configured token lifetime.
func (m *Manager) TTL() time.Duration { return m.ttl }

// Issue creates a signed token for a user.
func (m *Manager) Issue(userID string) (string, time.Duration, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			Issuer:    Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, m.ttl, err
}

// Issuer is the single trusted issuer for issued tokens.
const Issuer = "sign-flow"

// Parse verifies signature and expiry, returning claims.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errs.ErrUnauthorized
	}
	if claims.Type != "access" || claims.UserID == "" {
		return nil, errs.ErrUnauthorized
	}
	if claims.Issuer != Issuer {
		return nil, errs.ErrUnauthorized
	}
	return claims, nil
}
