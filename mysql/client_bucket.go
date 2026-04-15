package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetOrCreateStats получает статистику или создает новую.
func (c *Client) GetOrCreateStats(testName string, questionCount int64) (*entities.TestStats, error) {
	stats, err := c.GetStats(testName)
	
	switch {
	case err == nil:
		return stats, nil
	case errors.Is(err, entities.ErrNotFound):
		// Создаем новую статистику.
		newStats := entities.NewTestStats(testName, questionCount)

		dbStats := &dbTestStats{
			TestName:      newStats.TestName,
			UpdatedAt:     newStats.UpdatedAt,
			Attempts:      newStats.Attempts,
			AvgPercentage: newStats.AvgPercentage,
			AvgTimeSpent:  newStats.AvgTimeSpent,
			MinTimeSpent:  newStats.MinTimeSpent,
			MaxTimeSpent:  newStats.MaxTimeSpent,
		}

		if newStats.PercentDistrib != nil {
			dbStats.PercentDistribDB = fromEntityPercent(newStats.PercentDistrib)
		}
		if newStats.TimeDistrib != nil {
			dbStats.TimeDistribDB = fromEntityTime(newStats.TimeDistrib)
		}

		if err := c.db.Create(dbStats).Error; err != nil {
			return nil, fmt.Errorf("failed to create stats: %w", err)
		}
		return c.GetStats(testName)
	default:
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
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

	stats := &entities.TestStats{
		TestName:      dbStats.TestName,
		UpdatedAt:     dbStats.UpdatedAt,
		Attempts:      dbStats.Attempts,
		AvgPercentage: dbStats.AvgPercentage,
		AvgTimeSpent:  dbStats.AvgTimeSpent,
		MinTimeSpent:  dbStats.MinTimeSpent,
		MaxTimeSpent:  dbStats.MaxTimeSpent,
	}

	if dbStats.PercentDistribDB != nil {
		stats.PercentDistrib = dbStats.PercentDistribDB.toEntity()
	}
	if dbStats.TimeDistribDB != nil {
		stats.TimeDistrib = dbStats.TimeDistribDB.toEntity()
	}

	return stats, nil
}

// SaveStats сохраняет статистику.
func (c *Client) SaveStats(stats *entities.TestStats) error {
	dbStats := &dbTestStats{
		TestName:      stats.TestName,
		UpdatedAt:     time.Now(),
		Attempts:      stats.Attempts,
		AvgPercentage: stats.AvgPercentage,
		AvgTimeSpent:  stats.AvgTimeSpent,
		MinTimeSpent:  stats.MinTimeSpent,
		MaxTimeSpent:  stats.MaxTimeSpent,
	}

	if stats.PercentDistrib != nil {
		dbStats.PercentDistribDB = fromEntityPercent(stats.PercentDistrib)
	}
	if stats.TimeDistrib != nil {
		dbStats.TimeDistribDB = fromEntityTime(stats.TimeDistrib)
	}

	return c.db.Save(dbStats).Error
}
