package entities

import "time"

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type User struct {
	ID            int64
	Email         string
	PasswordHash  string
	EmailVerified bool
	Redirect      string
}

type VerificationCode struct {
	UserID    int64
	Code      string
	ExpiresAt time.Time
}
