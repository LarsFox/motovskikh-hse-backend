package repository

import (
	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	UpdateEmailVerified(userID uint, verified bool) error
	UpdatePassword(userID uint, passwordHash string) error
}

//nolint:ireturn
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}
