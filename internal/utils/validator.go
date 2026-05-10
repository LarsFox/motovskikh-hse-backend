package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

const (
	MinPasswordLength = 8
	MaxPasswordLength = 100
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	ErrEmailEmpty   = errors.New("email cannot be empty")
	ErrEmailInvalid = errors.New("invalid email format")

	ErrPasswordTooShort = errors.New("password must be at lest 8 characters")
	ErrPasswordTooLong  = errors.New("password is too long")
	ErrPasswordTooWeak  = errors.New("password must contain at least one letter, one digit and one special character")
)

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmailEmpty
	}
	if !emailRegex.MatchString(email) {
		return ErrEmailInvalid
	}
	return nil
}

func ValidatePassword(password string) error {
	password = strings.TrimSpace(password)
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	if !isPasswordStrong(password) {
		return ErrPasswordTooWeak
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
