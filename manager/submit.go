package manager

import (
	"fmt"
	"math"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const roundMultiplier = 10

// SubmitTestResult сохраняет результат теста и возвращает анализ.
func (m *Manager) SubmitTestResult(testName string, percentage float64, timeSpent int, questionCount int) (map[string]any, error) {
	// Валидация.
	isValid := m.ValidateAttempt(testName, percentage, timeSpent, questionCount)
	
	// Получаем текущий бакет.
	bucket, err := m.db.GetOrCreateBucket(testName, questionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	
	// Сохраняем старое количество попыток для расчета перцентилей.
	oldAttempts := bucket.Attempts
	
	// Рассчитываем перцентили на основе текущих данных. Если пустой бакет, то понятно, что будет 100%.
	percentileRank := 100.0
	timePercentile := 100.0
	
	if oldAttempts > 0 {
		percentileRank = m.CalculatePercentile(bucket, percentage)
		timePercentile = m.CalculateTimePercentile(bucket, timeSpent)
	}
	
	// Обновляем бакет.
	if isValid {
		bucket.Attempts++
		
		// Обновляем распределения.
		m.updatePercentDistribution(bucket, percentage, questionCount)
		m.updateTimeDistribution(bucket, timeSpent, questionCount)
		
		// Обновляем средние значения.
		m.updateAverages(bucket, percentage, float64(timeSpent))
		
		// Обновляем min/max.
		m.updateMinMax(bucket, percentage, timeSpent)
	}
	
	// Сохраняем бакет.
	if err := m.db.SaveBucket(bucket); err != nil {
		return nil, fmt.Errorf("failed to save bucket: %w", err)
	}
	
	// Формируем статистику для ответа.
	stats := &entities.TestStats{
		TestID:        testName,
		TotalAttempts: int(bucket.Attempts),  //nolint:gosec
		AvgPercentage: bucket.AvgPercentage,
		AvgTimeSpent:  bucket.AvgTimeSpent,
		UpdatedAt:     bucket.UpdatedAt,
	}
	
	percentageDiff := percentage - stats.AvgPercentage
  timeDiff := float64(timeSpent) - stats.AvgTimeSpent

	// Формируем анализ.
	analysis := m.buildAnalysis(percentage, timeSpent, percentileRank, timePercentile, stats, isValid, percentageDiff, timeDiff)
	
	// Возвращаем результат.
	result := map[string]any{
		"submitted":          true,
		"analysis":           analysis,
	}
	
	return result, nil
}

// updatePercentDistribution обновляет процентное распределение.
func (m *Manager) updatePercentDistribution(bucket *entities.TestBucket, percentage float64, questionCount int) {
	if bucket.PercentDistrib == nil {
		bucket.PercentDistrib = &entities.PercentDistribution{}
		bucket.InitializeBuckets(questionCount)
	}
	
	idx := bucket.GetPercentBucketIndex(percentage)
	if idx >= 0 && idx < len(bucket.PercentDistrib.Buckets) {
		bucket.PercentDistrib.Buckets[idx].Count++
	}
}

// updateTimeDistribution обновляет временное распределение.
func (m *Manager) updateTimeDistribution(bucket *entities.TestBucket, timeSpent, questionCount int) {
	if bucket.TimeDistrib == nil {
		bucket.TimeDistrib = &entities.TimeDistribution{}
		bucket.InitializeBuckets(questionCount)
	}
	
	idx := bucket.GetTimeBucketIndex(timeSpent)
	if idx >= 0 && idx < len(bucket.TimeDistrib.Buckets) {
		bucket.TimeDistrib.Buckets[idx].Count++
	}
}

// updateAverages обновляет средние значения.
func (m *Manager) updateAverages(bucket *entities.TestBucket, percentage, timeSpent float64) {
	oldTotal := float64(bucket.Attempts - 1)
	
	if bucket.Attempts == 1 {
		bucket.AvgPercentage = percentage
		bucket.AvgTimeSpent = timeSpent
	} else {
		bucket.AvgPercentage = (bucket.AvgPercentage*oldTotal + percentage) / float64(bucket.Attempts)
		bucket.AvgTimeSpent = (bucket.AvgTimeSpent*oldTotal + timeSpent) / float64(bucket.Attempts)
	}
}

// updateMinMax обновляет минимальные и максимальные значения.
func (m *Manager) updateMinMax(bucket *entities.TestBucket, percentage float64, timeSpent int) {
	if bucket.Attempts == 1 { //nolint:nestif
		bucket.MinPercentage = percentage
		bucket.MaxPercentage = percentage
		bucket.MinTimeSpent = timeSpent
		bucket.MaxTimeSpent = timeSpent
	} else {
		if percentage < bucket.MinPercentage {
			bucket.MinPercentage = percentage
		}
		if percentage > bucket.MaxPercentage {
			bucket.MaxPercentage = percentage
		}
		if timeSpent < bucket.MinTimeSpent {
			bucket.MinTimeSpent = timeSpent
		}
		if timeSpent > bucket.MaxTimeSpent {
			bucket.MaxTimeSpent = timeSpent
		}
	}
}

// buildAnalysis формирует анализ.
func (m *Manager) buildAnalysis(percentage float64, timeSpent int, percentileRank, timePercentile float64, stats *entities.TestStats, isValid bool,  percentageDiff, timeDiff float64) map[string]any {
	category := m.DetermineDistributionCategory(percentage)
	timeCategory := m.GetTimeCategory(timePercentile)

	return map[string]any{
		"percentage": percentage,
		"time_spent": timeSpent,
		"is_valid":   isValid,

		"percentile_rank": math.Round(percentileRank * roundMultiplier) / roundMultiplier,
		"time_percentile": math.Round(timePercentile * roundMultiplier) / roundMultiplier,
		"better_than":     int(percentileRank),
		"faster_than":     int(timePercentile),

		"category":      category.Name,
		"time_category": timeCategory,

		"average_percentage": math.Round(stats.AvgPercentage * roundMultiplier) / roundMultiplier,
		"average_time":       math.Round(stats.AvgTimeSpent * roundMultiplier) / roundMultiplier,

		"vs_average": map[string]any{
			"percentage_diff":   math.Round(percentageDiff * roundMultiplier) / roundMultiplier,
			"time_diff":         math.Round(timeDiff * roundMultiplier) / roundMultiplier,
			"percentage_status": m.GetComparisonStatus(percentage, stats.AvgPercentage),
			"time_status":       m.GetTimeComparisonStatus(float64(timeSpent), stats.AvgTimeSpent),
		},

		"quadrant": m.GetPerformanceQuadrant(percentage, timePercentile),
	}
}