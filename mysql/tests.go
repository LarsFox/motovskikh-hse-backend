package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type test struct {
	ID             int64
	UserID         int64
	Title          string
	QuestionsCount int
	Rating         float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// GetTests возвращает все тесты пользователя.
func (c *Client) GetTests(ctx context.Context, userID int64) ([]*entities.Test, error) {
	var tests []test
	err := c.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tests).Error
	if err != nil {
		return nil, fmt.Errorf("get tests: %w", err)
	}

	result := make([]*entities.Test, len(tests))
	for i, t := range tests {
		result[i] = &entities.Test{
			ID:             t.ID,
			UserID:         t.UserID,
			Title:          t.Title,
			QuestionsCount: t.QuestionsCount,
			Rating:         t.Rating,
		}
	}
	return result, nil
}

// CreateTest создаёт новый тест.
func (c *Client) CreateTest(ctx context.Context, t *entities.Test) error {
	internal := &test{
		UserID: t.UserID,
		Title:  t.Title,
	}
	if err := c.db.WithContext(ctx).Create(internal).Error; err != nil {
		return fmt.Errorf("create test: %w", err)
	}
	t.ID = internal.ID
	return nil
}

// DeleteTest удаляет тест по ID и userID.
func (c *Client) DeleteTest(ctx context.Context, id, userID int64) error {
	err := c.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&test{}).Error
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, gorm.ErrRecordNotFound):
		return entities.ErrNotFound
	default:
		return fmt.Errorf("delete test: %w", err)
	}
	return nil
}
