package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// Generate returns a URL-safe random token with n raw bytes (encoded ~4/3*n).
func GenerateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// RawURLEncoding removes +/ and = padding; safe for links
	return base64.RawURLEncoding.EncodeToString(b), nil
}
