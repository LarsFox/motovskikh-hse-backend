package manager

import (
	"math"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	step 							= 5
	diff 							= 0.01
	defaultPercentile = 100.0
	scoreMax          = 100
)

// CalculatePercentile рассчитывает перцентиль.
func (m *Manager) calculatePercentile(stats *entities.TestStats, percentage float64) float64 {
	if stats == nil || stats.Attempts == 0 || stats.PercentDistrib == nil {
		return defaultPercentile
	}

	var worseAttempts float64

	// Находим ключ текущего процента.
	currentKey := int(percentage/step) * step
	// Внутри бакета попытка не самая лучшая может быть. Поэтому там добавляем пропорционально.
	offset := (percentage - float64(currentKey)) / step

	// Проходим по всем бакетам.
	for key, count := range stats.PercentDistrib.Buckets {
		if key < currentKey {
			// Попытки, которые хуже.
			worseAttempts += float64(count)
		} else if key == currentKey {
			if (currentKey == 100) {
				worseAttempts += float64(count)
			} else {
				worseAttempts += float64(float64(count) * offset)
			}
		}
	}

	percentile := (worseAttempts / float64(stats.Attempts)) * defaultPercentile
	return math.Min(percentile, defaultPercentile)
}

// CalculateTimePercentile рассчитывает перцентиль по времени.
func (m *Manager) calculateTimePercentile(stats *entities.TestStats, timeSpent int64) float64 {
	if stats == nil || stats.Attempts == 0 || stats.TimeDistrib == nil {
		return defaultPercentile
	}

	var fasterAttempts float64
	idx := getTimeBucketIndex(stats, timeSpent)

	if idx >= 0 && idx < len(stats.TimeDistrib.Buckets) {
		bucket := stats.TimeDistrib.Buckets[idx]
		// Считаем offset внутри бакета.
		var offset float64
		if idx < len(stats.TimeDistrib.Buckets)-1 {
			nextMin := stats.TimeDistrib.Buckets[idx+1].MinSeconds
			step := float64(nextMin - bucket.MinSeconds)
			if step > 0 {
				offset = float64(timeSpent-bucket.MinSeconds) / step
			}
		} else {
			// Последний бакет — все попытки считаем быстрее или равными.
			offset = 1.0
		}
		
		// Бакеты с меньшим индексом — полностью быстрее.
		for i := range idx {
			fasterAttempts += float64(stats.TimeDistrib.Buckets[i].Count)
		}
		// Текущий бакет — пропорционально.
		fasterAttempts += float64(float64(bucket.Count) * offset)
	}

	timePercentile := scoreMax - (fasterAttempts/float64(stats.Attempts))*defaultPercentile

	if timePercentile < 0 {
		timePercentile = 0
	}
	if timePercentile > 100 {
		timePercentile = 100
	}
	return math.Min(timePercentile, defaultPercentile)
}
