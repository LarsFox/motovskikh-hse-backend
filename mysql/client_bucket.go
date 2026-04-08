package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetOrCreateStats получает статистику или создает новую.
func (c *Client) GetOrCreateStats(testName string, questionCount int) (*entities.TestStats, error) {
	stats, err := c.GetStats(testName)
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	if stats != nil {
		return stats, nil
	}

	// Создаем новую статистику.
	newStats := entities.NewTestStats(testName, questionCount)

	dbStats := fromEntityStats(newStats)
	if err := c.db.Create(dbStats).Error; err != nil {
		return nil, fmt.Errorf("failed to create stats: %w", err)
	}

	return c.GetStats(testName)
}

// GetStats получает статистику по имени теста.
func (c *Client) GetStats(testName string) (*entities.TestStats, error) {
	var dbStats dbTestStats
	err := c.db.Where("test_name = ?", testName).First(&dbStats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return dbStats.toEntityStats(), nil
}

// SaveStats сохраняет статистику.
func (c *Client) SaveStats(stats *entities.TestStats) error {
	dbStats := fromEntityStats(stats)
	dbStats.UpdatedAt = time.Now()
	return c.db.Save(dbStats).Error
}
