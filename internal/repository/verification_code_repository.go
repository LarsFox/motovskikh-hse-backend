package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"gorm.io/gorm"
)

var ErrCodeNotFound = errors.New("code not found or expired")

// VerificationCodeRepository — интерфейс для работы с кодами подтверждения.
type VerificationCodeRepository interface {
	Create(code *models.VerificationCode) error
	GetValidCode(userID uint, codeType models.VerificationCodeType, code string) (*models.VerificationCode, error)
	MarkAsUsed(id uint) error
	DeleteExpired() error
}

// verificationCodeRepository — реализация через GORM.
type verificationCodeRepository struct {
	db *gorm.DB
}

// NewVerificationCodeRepository создаёт новый репозиторий кодов.
func NewVerificationCodeRepository(db *gorm.DB) VerificationCodeRepository {
	return &verificationCodeRepository{db: db}
}

// Create сохраняет новый код подтверждения в БД.
func (r *verificationCodeRepository) Create(code *models.VerificationCode) error {
	if err := r.db.Create(code).Error; err != nil {
		return fmt.Errorf("create verification code: %w", err)
	}
	return nil
}

// GetValidCode ищет действующий (не использованный, не истёкший) код.
func (r *verificationCodeRepository) GetValidCode(userID uint, codeType models.VerificationCodeType, code string) (*models.VerificationCode, error) {
	var vc models.VerificationCode
	err := r.db.Where(
		"user_id = ? AND type = ? AND code = ? AND used = ? AND expires_at > ?",
		userID, codeType, code, false, time.Now(),
	).First(&vc).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get valid code: %w", err)
	}
	return &vc, nil
}

// MarkAsUsed помечает код как использованный.
func (r *verificationCodeRepository) MarkAsUsed(id uint) error {
	err := r.db.Model(&models.VerificationCode{}).
		Where("id = ?", id).
		Update("used", true).Error
	if err != nil {
		return fmt.Errorf("mark code as used: %w", err)
	}
	return nil
}

// DeleteExpired удаляет все истёкшие коды из БД.
func (r *verificationCodeRepository) DeleteExpired() error {
	err := r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.VerificationCode{}).Error
	if err != nil {
		return fmt.Errorf("delete expired codes: %w", err)
	}
	return nil
}
