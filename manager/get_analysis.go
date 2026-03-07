package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetTestAnalysis возвращает анализ.
func (m *Manager) GetTestAnalysis(testID string, userPercentage float64, userTimeSpent int) (*entities.TestStats, float64, error) {
	// Получаем бакет из БД.
	bucket, err := m.db.GetBucket(testID)
	if err != nil {
		return createEmptyStats(testID), 50.0, err
	}

	stats := convertBucketToStats(bucket)
	if stats == nil {
		stats = createEmptyStats(testID)
	}

	// Получаем перцентиль через менеджер.
	percentile := m.CalculatePercentile(bucket, userPercentage)

	return stats, percentile, nil
}