package manager

import (
	"context"
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetTests возвращает все тесты пользователя.
func (m *Manager) GetTests(ctx context.Context, userID int64) ([]*entities.Test, error) {
	tests, err := m.testDB.GetTests(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get tests: %w", err)
	}
	return tests, nil
}

// CreateTest создаёт новый тест.
func (m *Manager) CreateTest(ctx context.Context, userID int64, title string) (*entities.Test, error) {
	t := &entities.Test{
		UserID: userID,
		Title:  title,
	}
	if err := m.testDB.CreateTest(ctx, t); err != nil {
		return nil, fmt.Errorf("create test: %w", err)
	}
	return t, nil
}

// DeleteTest удаляет тест.
func (m *Manager) DeleteTest(ctx context.Context, id, userID int64) error {
	if err := m.testDB.DeleteTest(ctx, id, userID); err != nil {
		return fmt.Errorf("delete test: %w", err)
	}
	return nil
}
