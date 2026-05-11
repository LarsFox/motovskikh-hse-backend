package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type verificationCode struct {
	UserID    int64
	Code      string
	ExpiresAt time.Time
}

// SaveVerificationCode сохраняет новый код подтверждения в БД.
func (c *Client) SaveVerificationCode(ctx context.Context, code *entities.VerificationCode) error {
	vc := &verificationCode{
		UserID:    code.UserID,
		Code:      code.Code,
		ExpiresAt: code.ExpiresAt,
	}
	if err := c.db.Save(vc).Error; err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

// GetValidCode ищет действующий код.
func (c *Client) GetValidCode(ctx context.Context, userID int64, code string) (*entities.VerificationCode, error) {
	vc := &verificationCode{}
	err := c.db.Where(
		"user_id = ? AND code = ? AND expires_at > ?",
		userID, code, time.Now(),
	).First(&vc).Error

	// TODO: переделать на свич
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, entities.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get valid code: %w", err)
	}

	// TODO: заполнить.
	return &entities.VerificationCode{}, nil
}

func (c *Client) VerifyEmail(ctx context.Context, email, code string) error {
	// TODO: c.db.Begin() // defer tx.Rollback() // tx.Commit()
	// В одной тразнакции ищем пользователя с кодом.
	// Если есть непроверенный имейл, верифицируем.
	// Если есть проверенный имейл, возвращаем ErrInvalidInput.
	// Если нет ничего, возвращаем ErrNotFound.
	// В одной транзакции обновлять флаг верификации и удалять код.
	userID := 0

	err := c.db.Model(&user{}).
		Where("id = ?", userID).
		Update("email_verified", true).Error
	if err != nil {
		return fmt.Errorf("update email verified: %w", err)
	}
	return nil
}
