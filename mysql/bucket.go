package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetStats получает статистику по имени теста.
func (c *Client) GetStats(ctx context.Context, testName string) (*entities.TestStats, error) {
	dbStats := &testStats{}
	err := c.db.WithContext(ctx).Where("test_name = ?", testName).First(dbStats).Error
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, entities.ErrNotFound
	default:
		return nil, fmt.Errorf("get stats err: %w", err)
	}

	stats := &entities.TestStats{
		Name:          dbStats.Name,
		UpdatedAt:     dbStats.UpdatedAt,
		Attempts:      dbStats.Attempts,
		AvgPercentage: dbStats.AvgPercentage,
		AvgTimeSpent:  dbStats.AvgTimeSpent,
		MinTimeSpent:  dbStats.MinTimeSpent,
		MaxTimeSpent:  dbStats.MaxTimeSpent,
	}

	return stats, nil
}

// SaveStats сохраняет статистику.
func (c *Client) SaveStats(ctx context.Context, stats *entities.TestStats) error {
	dbStats := &testStats{
		Name:          stats.Name,
		UpdatedAt:     time.Now(),
		Attempts:      stats.Attempts,
		AvgPercentage: stats.AvgPercentage,
		AvgTimeSpent:  stats.AvgTimeSpent,
		MinTimeSpent:  stats.MinTimeSpent,
		MaxTimeSpent:  stats.MaxTimeSpent,
	}

	if err := c.db.WithContext(ctx).Save(dbStats).Error; err != nil {
		return fmt.Errorf("save test stats err: %w", err)
	}

	return nil
}
