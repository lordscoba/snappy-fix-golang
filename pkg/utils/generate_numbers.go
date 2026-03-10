package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// RandomDigits generates a string with n random digits (0-9).
func RandomDigits(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("number of digits must be > 0")
	}

	digits := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10)) // secure random 0–9
		if err != nil {
			return "", err
		}
		digits[i] = byte('0') + byte(num.Int64())
	}

	return string(digits), nil
}
