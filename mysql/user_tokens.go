package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type refreshToken struct {
	ID        int64
	UserID    int64
	Hash      string
	ExpiresAt time.Time
}

// GetRefreshToken ищет действующий refresh токен по хешу.
func (c *Client) GetRefreshToken(ctx context.Context, hash string) (int64, error) {
	token := &refreshToken{}
	err := c.db.Where(
		"hash = ? AND expires_at > ?",
		hash, time.Now(),
	).First(token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, entities.ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("get refresh token: %w", err)
	}
	return token.UserID, nil
}

func (c *Client) RefreshToken(ctx context.Context, hash, fresh string, expiresAt time.Time) error {
	// Транзакция, чтобы удаление старого и сохранение нового токена выполнялись атомарно.
	// Иначе при сбое между операциями пользователь потеряет доступ.
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		token := &refreshToken{}
		err := tx.Where(
			"hash = ? AND expires_at > ?",
			hash, time.Now(),
		).First(token).Error

		switch {
		case errors.Is(err, nil):
		case errors.Is(err, gorm.ErrRecordNotFound):
			return entities.ErrNotFound
		default:
			return fmt.Errorf("get refresh token: %w", err)
		}

		// Удаляем старый токен
		if err := tx.Delete(token).Error; err != nil {
			return fmt.Errorf("delete old token: %w", err)
		}

		// Сохраняем новый токен
		newToken := &refreshToken{
			UserID:    token.UserID,
			Hash:      fresh,
			ExpiresAt: expiresAt,
		}

		if err := tx.Create(newToken).Error; err != nil {
			return fmt.Errorf("save fresh token: %w", err)
		}

		return nil

	})
}

func (c *Client) DeleteExpired() error {
	// TODO: cron task
	err := c.db.Where("expires_at < ?", time.Now()).
		Delete(&refreshToken{}).Error
	if err != nil {
		return fmt.Errorf("delete expired tokens: %w", err)
	}
	return nil
}
