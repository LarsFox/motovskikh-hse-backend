package manager

import (
	"context"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type db interface {
	CheckExistingEmail(ctx context.Context, email string) (*entities.User, error)
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetRefreshToken(ctx context.Context, hash string) (int64, error)
	RefreshToken(ctx context.Context, hash, fresh string, expiresAt time.Time) error
	SaveVerificationCode(ctx context.Context, code *entities.VerificationCode) error
	VerifyEmail(ctx context.Context, email, code string) error
}

type emailer interface {
	SendEmail(ctx context.Context, to, subj, msg string) error
}

type Manager struct {
	db        db
	emailer   emailer
	secretKey []byte
}

func New(
	db db,
	emailer emailer,
	jwtSecret string,
) *Manager {
	return &Manager{
		db:        db,
		emailer:   emailer,
		secretKey: []byte(jwtSecret),
	}
}
