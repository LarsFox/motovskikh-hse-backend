package auth

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 100
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// TODO: написать тесты на каждую экспортируемую функцию этого пакета.

func ValidateEmail(email string) bool {
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	return emailRegex.MatchString(email)
}

func ValidatePassword(password string) error {
	password = strings.TrimSpace(password)
	if len(password) < MinPasswordLength {
		return entities.ErrInvalidInput
	}

	if len(password) > MaxPasswordLength {
		return entities.ErrInvalidInput
	}

	if !isPasswordStrong(password) {
		return entities.ErrInvalidInput
	}
	return nil
}

func isPasswordStrong(password string) bool {
	hasLetter := false
	hasDigit := false
	hasSpecial := false

	for _, c := range password {
		switch {
		case unicode.IsLetter(c):
			hasLetter = true
		case unicode.IsDigit(c):
			hasDigit = true
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			hasSpecial = true
		}
		if hasLetter && hasDigit && hasSpecial {
			return true
		}
	}

	return false
}
