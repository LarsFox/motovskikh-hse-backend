package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type user struct {
	ID            int64
	Email         string
	PasswordHash  string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (c *Client) CheckExistingEmail(ctx context.Context, email string) (*entities.User, error) {
	u := &user{}
	err := c.db.Where("email = ?", email).First(&u).Error
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, entities.ErrNotFound
	default:
		return nil, fmt.Errorf("check existing email: %w", err)
	}
	return &entities.User{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
	}, nil
}

func (c *Client) CreateUser(ctx context.Context, user *entities.User) error {
	if err := c.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

// GetUserByEmail ищет пользователя по email.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	u := &user{}
	err := c.db.Where("email = ? AND email_verified = true", email).First(&u).Error
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, entities.ErrNotFound
	default:
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &entities.User{
		ID:            u.ID,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		EmailVerified: u.EmailVerified,
	}, nil
}
