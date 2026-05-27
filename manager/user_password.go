package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
)

// ForgotPassword sends verification code to email.
func (m *Manager) ForgotPassword(ctx context.Context, email string) error {
	user, err := m.userDB.GetUserByEmail(ctx, email)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return entities.ErrInvalidInput
	default:
		return fmt.Errorf("get user by email: %w", err)
	}
	return m.sendVerificationCode(ctx, user.ID, user.Email)
}

func (m *Manager) ResetPassword(ctx context.Context, email, code, password string) error {
	if err := auth.ValidatePassword(password); err != nil {
		return entities.ErrInvalidInput
	}

	user, err := m.userDB.GetUserByEmail(ctx, email)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return entities.ErrInvalidInput
	default:
		return fmt.Errorf("get user by email: %w", err)
	}

	if err := m.userDB.CheckAndDeleteCode(ctx, email, code); err != nil {
		return fmt.Errorf("verify code: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := m.userDB.UpdatePassword(ctx, user.ID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
