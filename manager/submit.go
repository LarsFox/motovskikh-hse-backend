package manager

import (
	"fmt"
	"math"
	"github.com/google/uuid"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// SubmitTestResult сохраняет результат теста и возвращает весь проведенный анализ.
func (m *Manager) SubmitTestResult(testName, userHash string, percentage float64, timeSpent int) (string, map[string]interface{}, error) {
	// Валидация.
	isValid, validationMessage := m.ValidateAttempt(testName, userHash, percentage, timeSpent)
	// Генерируем ID попытки.
	attemptID := uuid.New().String()
	// Сохраняем в бакет.
	err := m.db.AddAttemptToBucket(testName, userHash, percentage, timeSpent, isValid)
	if err != nil {
		return "", nil, fmt.Errorf("failed to save attempt: %w", err)
	}
	// Получаем бакет для расчетов.
	bucket, _ := m.db.GetBucket(testName)
	stats := convertBucketToStats(bucket)
	if stats == nil {
		stats = createEmptyStats(testName)
	}

	// Получаем перцентили через менеджер.
	percentileRank := m.CalculatePercentile(bucket, percentage)
	timePercentile := m.CalculateTimePercentile(bucket, timeSpent)

	// Формируем анализ.
	analysis := m.buildAnalysis(percentage, timeSpent, percentileRank, timePercentile, stats, isValid)

	// Формируем итоговый результат.
	result := map[string]interface{}{
		"attempt_id":         attemptID,
		"submitted":          true,
		"validation_message": validationMessage,
		"analysis":           analysis,
	}

	return attemptID, result, nil
}

// buildAnalysis формирует анализ.
func (m *Manager) buildAnalysis(percentage float64, timeSpent int, percentileRank, timePercentile float64, stats *entities.TestStats, isValid bool) map[string]interface{} {
	category := m.DetermineDistributionCategory(percentage)
	timeCategory := m.GetTimeCategory(timePercentile)

	return map[string]interface{}{
		// Основные показатели.
		"percentage": percentage,
		"time_spent": timeSpent,
		"is_valid":   isValid,

		// Сравнение с другими.
		"percentile_rank": math.Round(percentileRank*10) / 10,
		"time_percentile": math.Round(timePercentile*10) / 10,
		"better_than":     int(percentileRank),
		"faster_than":     int(timePercentile),

		// Категории.
		"category":      category.Name,
		"time_category": timeCategory,

		// Средние значения.
		"average_percentage": math.Round(stats.AvgPercentage*10) / 10,
		"average_time":       math.Round(stats.AvgTimeSpent*10) / 10,

		// Сравнение со средним.
		"vs_average": map[string]interface{}{
			"percentage_diff":  math.Round((percentage-stats.AvgPercentage)*10) / 10,
			"time_diff":        math.Round((float64(timeSpent)-stats.AvgTimeSpent)*10) / 10,
			"percentage_status": m.GetComparisonStatus(percentage, stats.AvgPercentage),
			"time_status":       m.GetTimeComparisonStatus(float64(timeSpent), stats.AvgTimeSpent),
		},

		// Квадрант эффективности.
		"quadrant": m.GetPerformanceQuadrant(percentage, timePercentile),
	}
}