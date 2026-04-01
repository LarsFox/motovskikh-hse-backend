package models

import "time"

type VerificationCodeType string

const (
	EmailVerification VerificationCodeType = "email_verification"
	PasswordReset     VerificationCodeType = "password_reset"
)

type VerificationCode struct {
	ID        int                  `json:"id"`
	UserID    int                  `json:"user_id"`
	Code      string               `json:"code"`
	Type      VerificationCodeType `json:"type"`
	ExpiresAt time.Time            `json:"expires_at"`
	Used      bool                 `json:"used"`
	CreatedAt time.Time            `json:"created_at"`
}
