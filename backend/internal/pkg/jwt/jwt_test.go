package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueAndParse(t *testing.T) {
	m := New("test-secret-key-32-bytes-long!!", time.Hour)
	token, ttl, err := m.Issue("user-123")
	require.NoError(t, err)
	assert.Equal(t, time.Hour, ttl)
	assert.NotEmpty(t, token)

	claims, err := m.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "access", claims.Type)
	assert.Equal(t, "sign-flow", claims.Issuer)
}

func TestParseRejectsWrongSecret(t *testing.T) {
	m := New("test-secret-key-32-bytes-long!!", time.Hour)
	other := New("a-different-secret-key-32-bytes!!", time.Hour)
	token, _, err := m.Issue("user-123")
	require.NoError(t, err)

	_, err = other.Parse(token)
	assert.Error(t, err)
}

func TestParseRejectsTamperedToken(t *testing.T) {
	m := New("test-secret-key-32-bytes-long!!", time.Hour)
	token, _, err := m.Issue("user-123")
	require.NoError(t, err)

	_, err = m.Parse(token + "x")
	assert.Error(t, err)
}

func TestParseRejectsExpiredToken(t *testing.T) {
	// A negative TTL produces a token that is already expired.
	m := New("test-secret-key-32-bytes-long!!", -time.Second)
	token, _, err := m.Issue("user-123")
	require.NoError(t, err)

	_, err = m.Parse(token)
	assert.Error(t, err)
}

func TestParseRejectsGarbage(t *testing.T) {
	m := New("test-secret-key-32-bytes-long!!", time.Hour)
	_, err := m.Parse("not-a-jwt")
	assert.Error(t, err)
}
