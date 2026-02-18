package models

import "time"

type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	Login         string    `json:"login"`
	PasswordHash  string    `json:"-"` // не отправляю в JSON (!)
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
