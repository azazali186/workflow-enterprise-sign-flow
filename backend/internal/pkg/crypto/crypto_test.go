package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	require.NoError(t, Init("test-secret-for-encryption-32b!!"))
	enc, err := Encrypt("sensitive payload")
	require.NoError(t, err)
	assert.NotEqual(t, "sensitive payload", enc)
	dec, err := Decrypt(enc)
	require.NoError(t, err)
	assert.Equal(t, "sensitive payload", dec)
}

func TestEmptyAndNilCipher(t *testing.T) {
	gcm = nil
	out, err := Encrypt("x")
	require.NoError(t, err)
	assert.Equal(t, "x", out)
	out, err = Encrypt("")
	require.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestTamperedCiphertextFails(t *testing.T) {
	require.NoError(t, Init("another-secret-key-1234567890"))
	enc, err := Encrypt("value")
	require.NoError(t, err)
	_, err = Decrypt(enc + "tampered")
	assert.Error(t, err)
}
