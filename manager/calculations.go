package manager

import (
	"math"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	step = 5
	diff = 0.01

	defaultPercentile = 100.0
	halfMultiplier    = 0.5

	scoreMax             = 100
)

// CalculatePercentile рассчитывает перцентиль.
func (m *Manager) calculatePercentile(stats *entities.TestStats, percentage float64) float64 {
	if stats == nil || stats.Attempts == 0 || stats.PercentDistrib == nil {
		return defaultPercentile
	}

	var worseAttempts uint64

	// Находим ключ текущего процента.
	currentKey := float64(int(percentage/step) * step)

	// Проходим по всем бакетам.
	for key, count := range stats.PercentDistrib.Buckets {
		if key < currentKey {
			// Попытки, которые хуже.
			worseAttempts += count
		} else if math.Abs(key-currentKey) < diff {
			// Здесь делим пополам, потому что внутри бакета попытка не самая лучшая может быть.
			worseAttempts += uint64(float64(count) * halfMultiplier)
		}
	}

	percentile := (float64(worseAttempts) / float64(stats.Attempts)) * defaultPercentile
	return math.Min(percentile, defaultPercentile)
}

// CalculateTimePercentile рассчитывает перцентиль по времени.
func (m *Manager) calculateTimePercentile(stats *entities.TestStats, timeSpent int64) float64 {
	if stats == nil || stats.Attempts == 0 || stats.TimeDistrib == nil {
		return defaultPercentile
	}

	var fasterAttempts uint64

	for _, b := range stats.TimeDistrib.Buckets {
		if timeSpent < b.MinSeconds {
			break
		}

		if b.MaxSeconds == -1 || timeSpent <= b.MaxSeconds {
			// Здесь делим пополам, потому что внутри бакета попытка не самая лучшая может быть.
			fasterAttempts += uint64(float64(b.Count) * halfMultiplier)
			break
		}
		fasterAttempts += b.Count
	}

	timePercentile := scoreMax - (float64(fasterAttempts)/float64(stats.Attempts))*defaultPercentile
	return math.Min(timePercentile, defaultPercentile)
}
