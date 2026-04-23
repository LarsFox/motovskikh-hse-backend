package manager

import (
	"fmt"
	"math"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

const roundMultiplier = 10

// SubmitTestResult сохраняет результат теста и возвращает анализ.
func (m *Manager) SubmitTestResult(testName string, percentage float64, timeSpent int64, questionCount int64) (*models.SubmitTestResponse, error) {
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
		updatePercentDistribution(stats, percentage)
		updateTimeDistribution(stats, timeSpent)
		updateAverages(stats, percentage, float64(timeSpent))
		updateMinMax(stats, timeSpent)

		// Сохраняем бакет.
		if err := m.db.SaveStats(stats); err != nil {
			return nil, fmt.Errorf("failed to save stats: %w", err)
		}
	}

	percentageDiff := percentage - stats.AvgPercentage
	timeDiff := float64(timeSpent) - stats.AvgTimeSpent

	// Формируем ответ.
	return &models.SubmitTestResponse{
        Submitted: isValid,
        Analysis: &models.TestAnalysis{
					Percentage:        percentage,
					TimeSpent:         timeSpent,
					IsValid:           isValid,
					PercentileRank:    math.Round(percentileRank*roundMultiplier) / roundMultiplier,
					TimePercentile:    math.Round(timePercentile*roundMultiplier) / roundMultiplier,
					BetterThan:        math.Round(percentileRank*roundMultiplier) / roundMultiplier,
       		FasterThan:        math.Round(timePercentile*roundMultiplier) / roundMultiplier,
					AveragePercentage: math.Round(stats.AvgPercentage*roundMultiplier) / roundMultiplier,
					AverageTime:       math.Round(stats.AvgTimeSpent*roundMultiplier) / roundMultiplier,
					VsAverage: 				 &models.TestAnalysisVsAverage{
															PercentageDiff: math.Round(percentageDiff*roundMultiplier) / roundMultiplier,
															TimeDiff:       math.Round(timeDiff*roundMultiplier) / roundMultiplier,
														 },
        	},
  }, nil
}
