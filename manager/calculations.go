package manager

import (
	"math"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const (
		defaultPercentile = 100.0
    halfMultiplier = 0.5

		scoreBelowAverageMin = 20
		scoreAverageMin     = 40
		scoreGoodMin       = 60
		scoreExcellentMin  = 75
    scoreEliteMin      = 90
    scoreMax           = 100

		percentileBeginner   = 20
		percentileIntermediate = 40
		percentileAdvanced   = 70
    percentileExpert     = 90

		diffSignificantlyHigher = 15
    diffHigher              = 5

		timeMuchSlower = 60
    timeSlower     = 20

		quadHigh = 80
		quadMid = 50
)

// CalculatePercentile рассчитывает перцентиль.
func (m *Manager) CalculatePercentile(bucket *entities.TestBucket, percentage float64) float64 {
	if bucket == nil || bucket.Attempts == 0 || bucket.PercentDistrib == nil {
		return defaultPercentile
	}
	
	var worseAttempts uint64
	
	for _, b := range bucket.PercentDistrib.Buckets {
		switch {
		case percentage > b.Max:
    	worseAttempts += b.Count
		case percentage >= b.Min:
			// Здесь делим пополам, потому что внутри бакета попытка не самая лучшая может быть.
    	worseAttempts += uint64(float64(b.Count) * halfMultiplier)
		default:
    	break
}
	}
	
	percentile := (float64(worseAttempts) / float64(bucket.Attempts)) * defaultPercentile
	return math.Min(percentile, defaultPercentile)
}

// CalculateTimePercentile рассчитывает перцентиль по времени.
func (m *Manager) CalculateTimePercentile(bucket *entities.TestBucket, timeSpent int) float64 {
	if bucket == nil || bucket.Attempts == 0 || bucket.TimeDistrib == nil {
		return defaultPercentile
	}
	
	fasterAttempts := uint64(0)
	
	for _, b := range bucket.TimeDistrib.Buckets {
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
	
	timePercentile := (float64(fasterAttempts) / float64(bucket.Attempts)) * defaultPercentile
	return math.Min(timePercentile, defaultPercentile)
}

// DetermineDistributionCategory определяет категорию распределения.
func (m *Manager) DetermineDistributionCategory(percentage float64) *entities.DistributionCategory {
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

// DetermineSkillLevel определяет уровень навыка.
func (m *Manager) DetermineSkillLevel(percentile float64) string {
	switch {
	case percentile >= percentileExpert:
		return "expert"
	case percentile >= percentileAdvanced:
		return "advanced"
	case percentile >= percentileIntermediate:
		return "intermediate"
	default:
		return "beginner"
	}
}

// GetTimeCategory определяет категорию скорости.
func (m *Manager) GetTimeCategory(timePercentile float64) string {
	switch {
		case timePercentile >= percentileExpert:
			return "very_fast"
		case timePercentile >= percentileAdvanced:
			return "fast"
		case timePercentile >= percentileIntermediate:
			return "average_speed"
		case timePercentile >= percentileBeginner:
			return "slow"
		default:
			return "very_slow"
	}
}

// GetPerformanceQuadrant определяет квадрант производительности.
func (m *Manager) GetPerformanceQuadrant(percentage float64, timePercentile float64) map[string]any {
	var quadrant string

	switch {
		case percentage >= quadHigh && timePercentile >= quadHigh:
			quadrant = "expert"
		case percentage >= quadHigh && timePercentile < quadHigh:
			quadrant = "slow_expert"
		case percentage < quadHigh && timePercentile >= quadHigh:
			quadrant = "fast_but_inaccurate"
		case percentage >= quadMid && timePercentile >= quadMid:
			quadrant = "solid"
		case percentage < quadMid && timePercentile < quadMid:
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
func (m *Manager) GetComparisonStatus(userValue, avgValue float64) string {
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
func (m *Manager) GetTimeComparisonStatus(userTime, avgTime float64) string {
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