package manager

import (
	"context"
	"log"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const cleanInterval = 24 * time.Hour

type userDB interface {
	CheckExistingEmail(ctx context.Context, email string) (*entities.User, error)
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetRefreshToken(ctx context.Context, hash string) (int64, error)
	RefreshToken(ctx context.Context, hash, fresh string, expiresAt time.Time) error
	SaveVerificationCode(ctx context.Context, code *entities.VerificationCode) error
	VerifyEmail(ctx context.Context, email, code string) error
}

type testDB interface {
	WorkshopListUserTests(ctx context.Context, userID int64) ([]*entities.WorkshopTest, error)
	DeleteExpired() error
}

type emailer interface {
	SendEmail(ctx context.Context, to, subj, msg string) error
}

type Manager struct {
	userDB    userDB
	testDB    testDB
	emailer   emailer
	secretKey []byte
}

func New(
	userDB userDB,
	testDB testDB,
	emailer emailer,
	jwtSecret string,
) *Manager {
	m := &Manager{
		userDB:    userDB,
		testDB:    testDB,
		emailer:   emailer,
		secretKey: []byte(jwtSecret),
	}

	go func() {
		ticker := time.NewTicker(cleanInterval)
		for range ticker.C {
			if err := testDB.DeleteExpired(); err != nil {
				log.Println("delete expired tokens:", err)
			}
		}
	}()

	return m
}
