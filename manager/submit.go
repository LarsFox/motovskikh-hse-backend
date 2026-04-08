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
	isValid := m.validateAttempt(testName, percentage, timeSpent, questionCount)

	// Получаем текущий бакет.
	stats, err := m.db.GetOrCreateStats(testName, questionCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Сохраняем старое количество попыток для расчета перцентилей.
	oldAttempts := stats.Attempts

	// Рассчитываем перцентили на основе текущих данных. Если пустой бакет, то понятно, что будет 100%.
	percentileRank := 100.0
	timePercentile := 100.0

	if oldAttempts > 0 {
		percentileRank = m.calculatePercentile(stats, percentage)
		timePercentile = m.calculateTimePercentile(stats, timeSpent)
	}

	// Обновляем бакет.
	if isValid {
		stats.Attempts++
		stats.UpdatePercentDistribution(percentage)
		stats.UpdateTimeDistribution(timeSpent)
		stats.UpdateAverages(percentage, float64(timeSpent))
		stats.UpdateMinMax(percentage, timeSpent)
	}

	// Сохраняем бакет.
	if err := m.db.SaveStats(stats); err != nil {
		return nil, fmt.Errorf("failed to save stats: %w", err)
	}

	// Формируем статистику для ответа.
	responseStats := &entities.TestStatsResponse{
		TestName:      testName,
		TotalAttempts: int(stats.Attempts), //nolint:gosec
		AvgPercentage: stats.AvgPercentage,
		AvgTimeSpent:  stats.AvgTimeSpent,
		UpdatedAt:     stats.UpdatedAt,
	}

	percentageDiff := percentage - stats.AvgPercentage
	timeDiff := float64(timeSpent) - stats.AvgTimeSpent

	// Формируем анализ.
	analysis := m.buildAnalysis(percentage, timeSpent, percentileRank, timePercentile, responseStats, isValid, percentageDiff, timeDiff)

	// Возвращаем результат.
	result := map[string]any{
		"submitted": true,
		"analysis":  analysis,
	}

	return result, nil
}

// buildAnalysis формирует анализ.
func (m *Manager) buildAnalysis(percentage float64, timeSpent int, percentileRank, timePercentile float64, stats *entities.TestStatsResponse, isValid bool, percentageDiff, timeDiff float64) map[string]any {
	category := m.determineDistributionCategory(percentage)

	return map[string]any{
		"percentage": percentage,
		"time_spent": timeSpent,
		"is_valid":   isValid,

		"percentile_rank": math.Round(percentileRank*roundMultiplier) / roundMultiplier,
		"time_percentile": math.Round(timePercentile*roundMultiplier) / roundMultiplier,
		"better_than":     int(percentileRank),
		"faster_than":     int(timePercentile),

		"category": category.Name,

		"average_percentage": math.Round(stats.AvgPercentage*roundMultiplier) / roundMultiplier,
		"average_time":       math.Round(stats.AvgTimeSpent*roundMultiplier) / roundMultiplier,

		"vs_average": map[string]any{
			"percentage_diff":   math.Round(percentageDiff*roundMultiplier) / roundMultiplier,
			"time_diff":         math.Round(timeDiff*roundMultiplier) / roundMultiplier,
			"percentage_status": m.getComparisonStatus(percentage, stats.AvgPercentage),
			"time_status":       m.getTimeComparisonStatus(float64(timeSpent), stats.AvgTimeSpent),
		},

		"quadrant": m.getPerformanceQuadrant(percentage, timePercentile),
	}
}
