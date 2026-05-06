package models

import "time"

type VerificationCodeType string

const (
	EmailVerification VerificationCodeType = "email_verification"
	PasswordReset     VerificationCodeType = "password_reset"
)

type VerificationCode struct {
	ID        uint                 `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint                 `gorm:"not null;index"           json:"user_id"`
	Code      string               `gorm:"not null"                 json:"code"`
	Type      VerificationCodeType `gorm:"not null"                 json:"type"`
	ExpiresAt time.Time            `gorm:"not null"                 json:"expires_at"`
	Used      bool                 `gorm:"default:false"            json:"used"`
	CreatedAt time.Time            `gorm:"autoCreateTime"           json:"created_at"`

	// cвязь с таблицей users, а при удалении пользователя все его коды удаляются автоматически, дабы не создавать мусор в БД
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
