package repository

import (
	"coursework/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByEmail(email string) (*models.User, error)
	GetByLogin(login string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	UpdateEmailVerified(userID int, verified bool) error
	UpdatePassword(userID int, passwordHash string) error
}
