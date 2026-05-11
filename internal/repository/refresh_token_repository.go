package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"gorm.io/gorm"
)

var ErrTokenNotFound = errors.New("token not found or expired")

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	GetValid(tokenHash string) (*models.RefreshToken, error)
	MarkAsUsed(id uint) error
	DeleteExpired() error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

//nolint:ireturn
func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create в ответе за сохранение нового refresh токена в БД.
func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	if err := r.db.Create(token).Error; err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}
	return nil
}

// GetValid ищет действующий refresh токен по хешу.
func (r *refreshTokenRepository) GetValid(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.Where(
		"token_hash = ? AND used = ? AND expires_at > ?",
		tokenHash, false, time.Now(),
	).First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return &token, nil
}

func (r *refreshTokenRepository) MarkAsUsed(id uint) error {
	err := r.db.Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("used", true).Error
	if err != nil {
		return fmt.Errorf("mark token as used: %w", err)
	}
	return nil
}

func (r *refreshTokenRepository) DeleteExpired() error {
	err := r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.RefreshToken{}).Error
	if err != nil {
		return fmt.Errorf("delete expired tokens: %w", err)
	}
	return nil
}
