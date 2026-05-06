package repository

import (
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"gorm.io/gorm"
)

// userRepository — реализация UserRepository через GORM
type userRepository struct {
	db *gorm.DB
}

// Create создаёт нового пользователя в БД
func (r *userRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetByEmail ищет пользователя по email
func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

// UpdateEmailVerified обновляет статус подтверждения email
func (r *userRepository) UpdateEmailVerified(userID uint, verified bool) error {
	err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("email_verified", verified).Error
	if err != nil {
		return fmt.Errorf("update email verified: %w", err)
	}
	return nil
}

// UpdatePassword обновляет хеш пароля пользователя
func (r *userRepository) UpdatePassword(userID uint, passwordHash string) error {
	err := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("password_hash", passwordHash).Error
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
