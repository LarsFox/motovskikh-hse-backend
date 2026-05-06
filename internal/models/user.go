package models

import "time"

type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null"     json:"email"`
	PasswordHash  string    `gorm:"not null"                 json:"-"`
	EmailVerified bool      `gorm:"default:false"            json:"email_verified"`
	CreatedAt     time.Time `gorm:"autoCreateTime"           json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"           json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
