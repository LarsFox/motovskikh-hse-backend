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

	scoreBelowAverageMin = 20
	scoreAverageMin      = 40
	scoreGoodMin         = 60
	scoreExcellentMin    = 75
	scoreEliteMin        = 90
	scoreMax             = 100

	diffSignificantlyHigher = 15
	diffHigher              = 5

	timeMuchSlower = 60
	timeSlower     = 20

	quadHigh = 80
	quadMid  = 50
	quadLow  = 30

	quadTimeLow = 20
	quadTimeMid = 60
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
func (m *Manager) calculateTimePercentile(stats *entities.TestStats, timeSpent int) float64 {
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

// DetermineDistributionCategory определяет категорию распределения.
func (m *Manager) determineDistributionCategory(percentage float64) *entities.DistributionCategory {
	categories := []entities.DistributionCategory{
		{Name: "elite", MinScore: scoreEliteMin, MaxScore: scoreMax},
		{Name: "excellent", MinScore: scoreExcellentMin, MaxScore: scoreEliteMin},
		{Name: "good", MinScore: scoreGoodMin, MaxScore: scoreExcellentMin},
		{Name: "average", MinScore: scoreAverageMin, MaxScore: scoreGoodMin},
		{Name: "below_average", MinScore: scoreBelowAverageMin, MaxScore: scoreAverageMin},
		{Name: "needs_improvement", MinScore: 0, MaxScore: scoreBelowAverageMin},
	}

	for _, cat := range categories {
		if percentage >= cat.MinScore && percentage < cat.MaxScore {
			return &cat
		}
	}
	return &categories[3]
}

// GetPerformanceQuadrant определяет квадрант производительности.
func (m *Manager) getPerformanceQuadrant(percentage float64, timePercentile float64) map[string]any {
	var quadrant string

	switch {
	case percentage >= quadHigh && timePercentile <= quadTimeLow:
		quadrant = "expert"
	case percentage >= quadHigh && timePercentile > quadTimeLow:
		quadrant = "slow_expert"
	case percentage < quadLow && timePercentile <= quadTimeLow:
		quadrant = "fast_but_inaccurate"
	case percentage >= quadMid && timePercentile <= quadTimeMid:
		quadrant = "solid"
	case percentage < quadMid && timePercentile > quadTimeMid:
		quadrant = "needs_practice"
	default:
		quadrant = "mixed"
	}

	return map[string]any{
		"name": quadrant,
		"x":    percentage,
		"y":    timePercentile,
	}
}

// GetComparisonStatus возвращает статус сравнения с средним.
func (m *Manager) getComparisonStatus(userValue, avgValue float64) string {
	diff := userValue - avgValue
	switch {
	case diff > diffSignificantlyHigher:
		return "significantly_higher"
	case diff > diffHigher:
		return "higher"
	case diff < -diffSignificantlyHigher:
		return "significantly_lower"
	case diff < -diffHigher:
		return "lower"
	default:
		return "similar"
	}
}

// GetTimeComparisonStatus возвращает статус сравнения времени.
func (m *Manager) getTimeComparisonStatus(userTime, avgTime float64) string {
	diff := userTime - avgTime
	switch {
	case diff < -timeMuchSlower:
		return "much_faster"
	case diff < -timeSlower:
		return "faster"
	case diff > timeMuchSlower:
		return "much_slower"
	case diff > timeSlower:
		return "slower"
	default:
		return "similar"
	}
}
