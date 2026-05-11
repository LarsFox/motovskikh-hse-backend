package manager

import (
	"context"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
	"golang.org/x/crypto/bcrypt"
)

func (m *Manager) Enjoy(ctx context.Context, email, password string) (*entities.TokenPair, error) {
	user, err := m.db.GetUserByEmail(ctx, email)
	// TODO: switch err
	if err != nil {
		return nil, err
	}

	// TODO: заложить на уровне WHERE в базу данных.
	if !user.EmailVerified {
		return nil, err
	}

	// TODO: положить в manager/auth
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, err
	}

	tokens, err := auth.GeneratePair(user.ID, m.secretKey)
	// TODO: fmt.Errorf
	return tokens, err
}
