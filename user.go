package models

import "time"

type User struct {
	ID            int64
	Email         string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
