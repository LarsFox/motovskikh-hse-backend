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
	if err := c.db.WithContext(ctx).Create(vc).Error; err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

// GetValidCode ищет действующий код.
func (c *Client) GetValidCode(ctx context.Context, userID int64, code string) (*entities.VerificationCode, error) {
	vc := &verificationCode{}
	err := c.db.WithContext(ctx).Where(
		"user_id = ? AND code = ? AND expires_at > ?",
		userID, code, time.Now(),
	).First(&vc).Error

	switch {
	case errors.Is(err, nil):
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, entities.ErrNotFound
	default:
		return nil, fmt.Errorf("get valid code: %w", err)
	}

	return &entities.VerificationCode{
		UserID:    vc.UserID,
		Code:      vc.Code,
		ExpiresAt: vc.ExpiresAt,
	}, nil
}

func (c *Client) VerifyEmail(ctx context.Context, email, code string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// В одной тразнакции ищем пользователя с кодом.
		u := &user{}
		err := tx.Where("email = ?", email).First(u).Error

		switch {
		case errors.Is(err, nil):
		case errors.Is(err, gorm.ErrRecordNotFound):
			return entities.ErrNotFound
		default:
			return fmt.Errorf("get user: %w", err)
		}

		// Проверяем что email не подтверждён
		// Если есть проверенный имейл, возвращаем ErrInvalidInput
		if u.EmailVerified {
			return entities.ErrInvalidInput
		}

		// Проверяем код
		vc := &verificationCode{}
		err = tx.Where(
			"user_id = ? AND code = ? AND expires_at > ?",
			u.ID, code, time.Now(),
		).First(vc).Error

		// Если есть непроверенный имейл, верифицируем.
		// Если нет ничего, возвращаем ErrNotFound.
		// В одной транзакции обновлять флаг верификации и удалять код.
		switch {
		case errors.Is(err, nil):
		case errors.Is(err, gorm.ErrRecordNotFound):
			return entities.ErrNotFound
		default:
			return fmt.Errorf("get verification code: %w", err)
		}

		// Подтверждаем email
		if err := tx.Model(u).Update("email_verified", true).Error; err != nil {
			return fmt.Errorf("update email verified: %w", err)
		}

		// Удаляем использованный код
		if err := tx.Where("user_id = ? AND code = ?", u.ID, code).Delete(&verificationCode{}).Error; err != nil {
			return fmt.Errorf("delete verification code: %w", err)
		}

		return nil
	})
}
