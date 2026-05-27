package manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
)

const verificationCodeTTL = time.Hour

func (m *Manager) Register(ctx context.Context, email, password string) error {
	if !auth.ValidateEmail(email) {
		return entities.ErrInvalidInput
	}

	// Checking password complexity
	if err := auth.ValidatePassword(password); err != nil {
		return entities.ErrInvalidInput
	}

	// If there is no the email, it is a new user. We should save him in DB and send a code.
	// If the email exists but is not verified, send a code
	existing, err := m.userDB.CheckExistingEmail(ctx, email)
	switch {
	case errors.Is(err, entities.ErrNotFound):
	case errors.Is(err, nil):
		return m.sendVerificationCode(ctx, existing.ID, existing.Email)
	default:
		return fmt.Errorf("check existing email: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Create a user in DB
	user := &entities.User{
		Email:        email,
		PasswordHash: hash,
	}
	if err := m.userDB.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return m.sendVerificationCode(ctx, user.ID, user.Email)
}

func (m *Manager) VerifyEmail(ctx context.Context, email, code string) error {
	err := m.userDB.VerifyEmail(ctx, email, code)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return entities.ErrInvalidInput
	default:
		return fmt.Errorf("verify email: %w", err)
	}

	return nil
}

// sendVerificationCode (private) generates and sendes 6 signed code.
func (m *Manager) sendVerificationCode(ctx context.Context, userID int64, email string) error {
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	// Save the code, it will be valid for an hour
	vc := &entities.VerificationCode{
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}
	if err := m.userDB.SaveVerificationCode(ctx, vc); err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}

	// Send by email
	return m.emailer.SendEmail(ctx, email,
		"Подтверждение email в Мастерской Мотовских",
		fmt.Sprintf(
			"Ваш код подтверждения:\n%s\n\nКод действует 1 час.", code),
	)
}

func (m *Manager) ResendCode(ctx context.Context, email string) error {
	existing, err := m.userDB.CheckExistingEmail(ctx, email)
	switch {
	case errors.Is(err, entities.ErrNotFound):
		return entities.ErrInvalidInput // user does not exist
	case errors.Is(err, nil):
		if existing.EmailVerified {
			return entities.ErrInvalidInput // already confirmed
		}
		return m.sendVerificationCode(ctx, existing.ID, existing.Email)
	default:
		return fmt.Errorf("check existing email: %w", err)
	}
}
