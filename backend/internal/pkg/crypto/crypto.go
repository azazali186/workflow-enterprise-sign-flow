// Package crypto provides AES-256-GCM encryption for sensitive data at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var gcm cipher.AEAD

// Init derives an AEAD cipher from the configured secret. Secrets longer than
// 32 bytes are hashed down to a 256-bit key.
func Init(secret string) error {
	sum := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return err
	}
	gcm, err = cipher.NewGCM(block)
	return err
}

// Encrypt seals plaintext into base64 ciphertext with a random nonce.
func Encrypt(plain string) (string, error) {
	if plain == "" || gcm == nil {
		return plain, nil
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt opens base64 ciphertext produced by Encrypt.
func Decrypt(enc string) (string, error) {
	if enc == "" || gcm == nil {
		return enc, nil
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, body := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
