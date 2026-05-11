package utils

import (
	"crypto/rand"
	"fmt"
)

const codeBytes = 3

// GenerateVerificationCode генерирует 6-значный код.
func GenerateVerificationCode() (string, error) {
	b := make([]byte, codeBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2]))
	return code[:6], nil
}
