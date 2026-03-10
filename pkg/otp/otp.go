package otp

import (
	"crypto/rand"
	"fmt"
)

// GenerateNumeric returns a numeric OTP of n digits (e.g., n=6 -> "483920").
func GenerateNumeric(n int) (string, error) {
	if n <= 0 || n > 12 { // sane bound
		n = 6
	}
	const base = 10
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	digits := make([]byte, n)
	for i := 0; i < n; i++ {
		digits[i] = '0' + (bytes[i] % base)
	}
	// avoid leading zeros becoming shorter when casted later; return as string
	return string(digits), nil
}

func Must6() string {
	code, err := GenerateNumeric(6)
	if err != nil {
		panic(fmt.Errorf("otp.Must6: %w", err))
	}
	return code
}
