package repository

import (
	"context"
	"coursework/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, email string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
}
