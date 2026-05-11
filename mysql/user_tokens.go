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
		"hash = ? = ? AND expires_at > ?",
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
	// TODO: tx.
	token := &refreshToken{}
	err := c.db.Where(
		"hash = ? = ? AND expires_at > ?",
		hash, time.Now(),
	).First(token).Error

	// TODO: switch
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get refresh token: %w", err)
	}

	// TODO: delete used token
	// TODO: save fresh token

	return nil
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
