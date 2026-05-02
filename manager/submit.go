package manager

import (
	"errors"
	"fmt"
	"math"
	"time"
	"context"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
	roundMultiplier = 10

	bucketsCount = 20
	bucketsStep  = 5

	timeLimitTiny  = 7
	timeLimitSmall = 15

	secondsPerQuestionMin = 2
	secondsPerQuestionMax = 30

	verySmallTestMinTime = 0.5
	SmallTestMinTime 		 = 1

	defaultPercentile = 100.0
)

// SubmitTestResult сохраняет результат теста и возвращает анализ.
// Если попытка не валидная, результат посчитаем, но сохранять попытку не будем.
// TODO questionCount с фронта??
func (m *Manager) SubmitTestResult(testName string, percentage, timeSpent float64, questionCount int64) (*entities.TestStatsAnalysis, error) {
	// Валидация.
	if !m.validateAttempt(testName, percentage, timeSpent, questionCount) {
		return nil, entities.ErrInvalidInput
	}
	ctx := context.Background()
	// Получаем текущий бакет.
	stats, err := m.db.GetStats(ctx, testName)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		stats = &entities.TestStats{
			Name:           testName,
			UpdatedAt:      time.Now(),
			PercentBuckets: make([]*entities.TestStatsBucket, 0, bucketsCount),
		}

		// Инициализируем интервалы от 0 до 100 с шагом.
		for i := 0; i < bucketsCount; i++ {
			stats.PercentBuckets = append(stats.PercentBuckets, &entities.TestStatsBucket{
				Value: float64((i + 1) * bucketsStep),
				Count: 0,
			})
		}
		stats.TimeBuckets = makeTimeBuckets(questionCount)

	default:
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	oldAttempts := stats.Attempts

	// Кол-во попыток, которые обошел пользователь.
	betterThan := countPercentile(stats.PercentBuckets, percentage, stats.Attempts, false)
	fasterThan := stats.Attempts - countPercentile(stats.TimeBuckets, timeSpent, stats.Attempts, true)

	// Считаем перцентиль через это количество попыток и общее.
	var scorePercentile float64
	var timePercentile float64

	if oldAttempts > 0 {
		scorePercentile = float64(betterThan * 100) / float64(oldAttempts)
		timePercentile = float64(fasterThan * 100) / float64(oldAttempts)
	} else {
		scorePercentile = defaultPercentile
		timePercentile = defaultPercentile
	}


	// Разница между средними показателями.
	percentageDiff := percentage - stats.AvgPercentage
	timeDiff := timeSpent - stats.AvgTimeSpent

	// Обновляем бакет.
	stats.Attempts++
	incrementBucket(stats.PercentBuckets, percentage)
	incrementBucket(stats.TimeBuckets, timeSpent)
	updateAverages(stats, percentage, timeSpent)
	updateMinMax(stats, timeSpent)

	// Сохраняем бакет.
	if err := m.db.SaveStats(ctx, stats); err != nil {
		return nil, fmt.Errorf("failed to save stats: %w", err)
	}

	// Здесь округляем до знака после запятой.
	scorePercentile = math.Round(scorePercentile*roundMultiplier) / roundMultiplier
	timePercentile = math.Round(timePercentile*roundMultiplier) / roundMultiplier

	percentageDiff = math.Round(percentageDiff*roundMultiplier) / roundMultiplier
	timeDiff = math.Round(timeDiff*roundMultiplier) / roundMultiplier

	// Формируем ответ.
	return &entities.TestStatsAnalysis{
		ScorePercentile:   scorePercentile,
		TimePercentile:    timePercentile,
		BetterThan:        betterThan,
		FasterThan:        fasterThan,
		AveragePercentage: math.Round(stats.AvgPercentage*roundMultiplier) / roundMultiplier,
		AverageTime:       math.Round(stats.AvgTimeSpent*roundMultiplier) / roundMultiplier,
		PercentageDiff:    percentageDiff,
		TimeDiff:          timeDiff,
	}, nil
}

func getBucketMinTime(questions int64) float64 {
	if questions < timeLimitTiny {
		return verySmallTestMinTime
	}

	if questions < timeLimitSmall {
		return SmallTestMinTime
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

	for i := range bucketsCount {
		minSeconds := minTime + float64(i)*step
		result = append(result, &entities.TestStatsBucket{
			Value: math.Round(minSeconds*100) / 100,
			Count: 0,
		})
	}

	return result
}

func countPercentile(buckets []*entities.TestStatsBucket, value float64, max int64, isTime bool) int64 {
	if len(buckets) == 0 {
		return 0
	}

	if value <= buckets[0].Value {
		return 0
	}

	if value >= buckets[len(buckets)-1].Value {
		return max
	}

	var total int64
	for i, bucket := range buckets {
		if bucket.Value < value {
			total += bucket.Count
			continue
		}

		if i == len(buckets)-1 {
			total += bucket.Count
			break
		}

		extra := bucket.Value - value
		size := buckets[i+1].Value - bucket.Value

		if size <= 0 {
			break
		}

		fraction := extra / size
		if (isTime) {
			total += int64((1-fraction) * float64(bucket.Count))
		} else {
			total += int64(fraction * float64(bucket.Count))
		}
		break
	}

	return total
}

func incrementBucket(buckets []*entities.TestStatsBucket, value float64) {
	for i, bucket := range buckets {
		if value <= bucket.Value {
			buckets[i].Count++
			return
		}
	}
	// Увеличиваем последний.
	buckets[len(buckets)-1].Count++
}
