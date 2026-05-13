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

// Register регистрирует нового пользователя.
func (m *Manager) Register(ctx context.Context, email, password string) error {
	// Валидация email
	if !auth.ValidateEmail(email) {
		return entities.ErrInvalidInput
	}

	// Если имейла нет, это новый пользователь, записываем его имейл, высылаем код.
	// Если есть проверенный имейл, возвращаем entities.ErrInvalidInput, имейл уже занят.
	// Если есть непроверенный имейл, перезаписываем и высылаем новый код.
	existing, err := m.db.CheckExistingEmail(ctx, email)
	switch {
	case errors.Is(err, entities.ErrNotFound):
	case errors.Is(err, nil):
		if existing.EmailVerified {
			return entities.ErrInvalidInput
		}
		return m.sendVerificationCode(ctx, existing.ID, existing.Email)
	default:
		return fmt.Errorf("check existing email: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Создаём пользователя в БД.
	user := &entities.User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if err := m.db.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	// Генерация кода подтверждения.
	return m.sendVerificationCode(ctx, user.ID, user.Email)
}

// VerifyEmail подтверждает email пользователя.
func (m *Manager) VerifyEmail(ctx context.Context, email, code string) error {
	err := m.db.VerifyEmail(ctx, email, code)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		return entities.ErrInvalidInput
	default:
		return fmt.Errorf("verify email: %w", err)
	}

	return nil
}

// sendVerificationCode (приватный метод) генерирует и отправляет ссылку с вшитым кодом подтверждения.
func (m *Manager) sendVerificationCode(ctx context.Context, userID int64, email string) error {
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	// Сохраняем код в БД, действует 1 час.
	vc := &entities.VerificationCode{
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}
	if err := m.db.SaveVerificationCode(ctx, vc); err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}

	// Отправка кода на email.
	link := fmt.Sprintf("http://localhost:8080/verify?email=%s&code=%s", email, code)
	return m.emailer.SendEmail(ctx, email,
		"Подтверждение email — Мастерская Мотовских",
		fmt.Sprintf(
			"Подтвердите email, перейдя по ссылке:\n%s\n\nСсылка действует 1 час.", link,
		),
	)
}
