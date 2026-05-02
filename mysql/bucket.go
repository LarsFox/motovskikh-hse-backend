package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetStats получает статистику по имени теста.
func (c *Client) GetStats(ctx context.Context, testName string) (*entities.TestStats, error) {
	dbStats := &testStats{}
	err := c.db.WithContext(ctx).Where("name = ?", testName).First(dbStats).Error
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

	// Бакеты процентов.
	if len(dbStats.PercentDistrib) > 0 {
		if err := json.Unmarshal(dbStats.PercentDistrib, &stats.PercentBuckets); err != nil {
			return nil, fmt.Errorf("unmarshal percent buckets err: %w", err)
		}
	}

	// Бакеты времени.
	if len(dbStats.TimeDistrib) > 0 {
		if err := json.Unmarshal(dbStats.TimeDistrib, &stats.TimeBuckets); err != nil {
			return nil, fmt.Errorf("unmarshal time buckets err: %w", err)
		}
	}

	return stats, nil
}

// SaveStats сохраняет статистику.
func (c *Client) SaveStats(ctx context.Context, stats *entities.TestStats) error {
	percentDistrib, err := json.Marshal(stats.PercentBuckets)
	if err != nil {
		return fmt.Errorf("marshal percent buckets err: %w", err)
	}

	timeDistrib, err := json.Marshal(stats.TimeBuckets)
	if err != nil {
		return fmt.Errorf("marshal time buckets err: %w", err)
	}

	dbStats := &testStats{
		Name:          stats.Name,
		UpdatedAt:     time.Now(),
		Attempts:      stats.Attempts,
		PercentDistrib: percentDistrib,
		TimeDistrib:    timeDistrib,
		AvgPercentage: stats.AvgPercentage,
		AvgTimeSpent:  stats.AvgTimeSpent,
		MinTimeSpent:  stats.MinTimeSpent,
		MaxTimeSpent:  stats.MaxTimeSpent,
	}

	err = c.db.WithContext(ctx).Save(dbStats).Error
	if err != nil {
		return fmt.Errorf("save test stats err: %w", err)
	}

	return nil
}
