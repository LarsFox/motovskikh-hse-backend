package manager

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

const (
	roundMultiplier = 10

	bucketsCount = 20
	bucketsStep  = 5

	timeLimitTiny  = 7
	timeLimitSmall = 15

	secondsPerQuestionMin = 2
	secondsPerQuestionMax = 30
)

// SubmitTestResult сохраняет результат теста и возвращает анализ.
// Если попытка не валидная, результат посчитаем, но сохранять попытку не будем.
// TODO questionCount с фронта??
func (m *Manager) SubmitTestResult(testName string, percentage, timeSpent float64, questionCount int64) (*entities.TestStatsAnalysis, error) {
	// Валидация.
	isValid := m.validateAttempt(testName, percentage, timeSpent, questionCount)
	if isValid {
		return nil, entities.ErrInvalidInput
	}

	// Получаем текущий бакет.
	stats, err := m.db.GetStats(testName)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		stats = &entities.TestStats{
			Name:           testName,
			PercentBuckets: make([]*entities.TestStatsBucket, 0, bucketsCount),
			UpdatedAt:      time.Now(),
		}

		// Инициализируем интервалы от 0 до 100 с шагом.
		for i := range bucketsCount {
			stats.PercentBuckets = append(stats.PercentBuckets, &entities.TestStatsBucket{
				Value: float64((i + 1) * bucketsStep),
				Count: 0,
			})

			stats.TimeBuckets = makeTimeBuckets(questionCount)
		}

	default:
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Обновляем бакет.
	stats.Attempts++
	incrementBucket(stats.PercentBuckets, percentage)
	incrementBucket(stats.TimeBuckets, timeSpent)
	updateAverages(stats, percentage, float64(timeSpent))
	updateMinMax(stats, timeSpent)

	// Сохраняем бакет.
	if err := m.db.SaveStats(stats); err != nil {
		return nil, fmt.Errorf("failed to save stats: %w", err)
	}

	betterThan := countPercentile(stats.PercentBuckets, percentage, stats.Attempts)
	fasterThan := stats.Attempts - countPercentile(stats.TimeBuckets, timeSpent, stats.Attempts)
	percentile := float64(stats.Attempts)

	percentageDiff := percentage - stats.AvgPercentage
	timeDiff := float64(timeSpent) - stats.AvgTimeSpent

	// Формируем ответ.
	return &entities.TestStatsAnalysis{
		ScorePercentile:   float64(betterThan) / percentile,
		TimePercentile:    float64(fasterThan) / percentile,
		BetterThan:        betterThan,
		FasterThan:        fasterThan,
		AveragePercentage: math.Round(stats.AvgPercentage*roundMultiplier) / roundMultiplier,
		AverageTime:       math.Round(stats.AvgTimeSpent*roundMultiplier) / roundMultiplier,
		VsAverage: &models.TestAnalysisVsAverage{
			PercentageDiff: math.Round(percentageDiff*roundMultiplier) / roundMultiplier,
			TimeDiff:       math.Round(timeDiff*roundMultiplier) / roundMultiplier,
		},
	}, nil
}

func getBucketMinTime(questions int64) float64 {
	if questions < timeLimitTiny {
		return 0.5 // TODO consts
	}

	if questions < timeLimitSmall {
		return 1
	}

	return secondsPerQuestionMin
}

// makeTimeBuckets создает интервалы на основе количества вопросов.
func makeTimeBuckets(questionCount int64) []*entities.TestStatsBucket {
	var minTime, maxTime float64
	minTime = getBucketMinTime(questionCount) * secondsPerQuestionMin
	maxTime = float64(questionCount) * secondsPerQuestionMax

	step := (maxTime - minTime) / bucketsCount

	result := make([]*entities.TestStatsBucket, 0, bucketsCount)
	result = append(result, &entities.TestStatsBucket{
		Value: 0,
		Count: 0,
	})

	for i := range bucketsCount - 1 {
		minSeconds := minTime + float64(i)*step
		result = append(result, &entities.TestStatsBucket{
			Value: minSeconds,
			Count: 0,
		})
	}

	return result
}

func countPercentile(buckets []*entities.TestStatsBucket, value float64, max int64) int64 {
	if value >= buckets[len(buckets)-1].Value {
		return max
	}

	var total int64
	for i, bucket := range buckets {
		if bucket.Value < value {
			total += bucket.Count
			continue
		}

		extra := value - bucket.Value
		size := buckets[i+1].Value - bucket.Value
		total += int64(extra / size)
		break
	}

	return total
}

func incrementBucket(buckets []*entities.TestStatsBucket, value float64) {
	// TODO
}
