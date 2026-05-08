package models

import "time"

// RefreshToken — токен для обновления access токена, храним хеш токена
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;index"           json:"user_id"`
	TokenHash string    `gorm:"not null"                 json:"-"`
	ExpiresAt time.Time `gorm:"not null"                 json:"expires_at"`
	Used      bool      `gorm:"default:false"            json:"used"`
	CreatedAt time.Time `gorm:"autoCreateTime"           json:"created_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
