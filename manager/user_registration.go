package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager/auth"
	"golang.org/x/crypto/bcrypt"
)

const verificationCodeTTL = time.Hour

// Register регистрирует нового пользователя.
func (m *Manager) Register(ctx context.Context, email, password string) error {
	// Валидация email
	if !auth.ValidateEmail(email) {
		return entities.ErrInvalidInput
	}

	// Проверка, что email не занят.
	// TODO:
	// Если имейла нет, это новый пользователь, записываем его имейл, высылаем код.
	// Если есть проверенный имейл, возвращаем entities.ErrInvalidInput, имейл уже занят.
	// Если есть непроверенный имейл, перезаписываем и высылаем новый код.
	existing, err := m.db.CheckExistingEmail(ctx, email)
	if existing != nil {
		return entities.ErrInvalidInput
	}
	// TODO: switch err
	if err != nil {
		return err
	}

	// TODO: унести это в manager/auth
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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
	// TODO: switch-case nil, entities.NotFound, default
	if err != nil {
		return fmt.Errorf("verify email: %w", err)
	}

	return nil
}

// sendVerificationCode (приватный метод) генерирует и отправляет код подтверждения.
func (m *Manager) sendVerificationCode(ctx context.Context, userID int64, email string) error {
	code, err := auth.GenerateVerificationCode()
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	// Сохраняем код в БД, действует 15 минут.
	vc := &entities.VerificationCode{
		UserID:    userID,
		Code:      code,
		ExpiresAt: time.Now().Add(verificationCodeTTL),
	}
	if err := m.db.SaveVerificationCode(ctx, vc); err != nil {
		return fmt.Errorf("save verification code: %w", err)
	}

	// Отправка кода на email.
	// TODO: правильное формирование текста и темы.
	return m.emailer.SendEmail(ctx, email, "", fmt.Sprintf("code: %s", code))
}
