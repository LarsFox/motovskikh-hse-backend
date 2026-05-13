package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
)

func (m *Manager) Enjoy(ctx context.Context, email, password string) (*entities.TokenPair, error) {
	user, err := m.db.GetUserByEmail(ctx, email)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return nil, entities.ErrInvalidInput
	default:
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return nil, entities.ErrInvalidInput
	}

	tokens, err := auth.GeneratePair(user.ID, m.secretKey)
	if err != nil {
		return nil, fmt.Errorf("generate pair: %w", err)
	}
	return tokens, nil
}
