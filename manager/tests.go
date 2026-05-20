package manager

import (
	"context"
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// ListUserTests возвращает все тесты пользователя.
func (m *Manager) ListUserTests(ctx context.Context, userID int64) ([]*entities.WorkshopTest, error) {
	tests, err := m.testDB.WorkshopListUserTests(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user tests: %w", err)
	}

	if len(tests) == 0 {
		return nil, entities.ErrNotFound
	}

	return tests, nil
}
