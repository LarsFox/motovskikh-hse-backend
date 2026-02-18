package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateVerificationCode генерирует 6-значный код
func GenerateVerificationCode() (string, error) {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2]))
	return code[:6], nil
}
