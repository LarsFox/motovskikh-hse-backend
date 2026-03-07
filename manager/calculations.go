package manager

import (
	"math"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// CalculatePercentile рассчитывает перцентиль на основе данных бакета.
func (m *Manager) CalculatePercentile(bucket *entities.TestBucket, percentage float64) float64 {
	if bucket == nil || bucket.ValidAttempts == 0 {
		return 50.0
	}

	// Считаем сколько попыток хуже.
	worseAttempts := uint64(0)

	// Суммируем все корзины с меньшим процентом.
	if percentage > 5 {
		worseAttempts += bucket.Pct0_5
	}
	if percentage > 10 {
		worseAttempts += bucket.Pct5_10
	}
	if percentage > 15 {
		worseAttempts += bucket.Pct10_15
	}
	if percentage > 20 {
		worseAttempts += bucket.Pct15_20
	}
	if percentage > 25 {
		worseAttempts += bucket.Pct20_25
	}
	if percentage > 30 {
		worseAttempts += bucket.Pct25_30
	}
	if percentage > 35 {
		worseAttempts += bucket.Pct30_35
	}
	if percentage > 40 {
		worseAttempts += bucket.Pct35_40
	}
	if percentage > 45 {
		worseAttempts += bucket.Pct40_45
	}
	if percentage > 50 {
		worseAttempts += bucket.Pct45_50
	}
	if percentage > 55 {
		worseAttempts += bucket.Pct50_55
	}
	if percentage > 60 {
		worseAttempts += bucket.Pct55_60
	}
	if percentage > 65 {
		worseAttempts += bucket.Pct60_65
	}
	if percentage > 70 {
		worseAttempts += bucket.Pct65_70
	}
	if percentage > 75 {
		worseAttempts += bucket.Pct70_75
	}
	if percentage > 80 {
		worseAttempts += bucket.Pct75_80
	}
	if percentage > 85 {
		worseAttempts += bucket.Pct80_85
	}
	if percentage > 90 {
		worseAttempts += bucket.Pct85_90
	}
	if percentage > 95 {
		worseAttempts += bucket.Pct90_95
	}

	// Добавляем половину из текущей корзины.
	currentBucketAttempts := m.getCurrentBucketAttempts(bucket, percentage)
	worseAttempts += uint64(float64(currentBucketAttempts) * 0.5)

	percentile := (float64(worseAttempts) / float64(bucket.ValidAttempts)) * 100.0
	return math.Min(percentile, 100)
}

// CalculateTimePercentile рассчитывает перцентиль по времени.
func (m *Manager) CalculateTimePercentile(bucket *entities.TestBucket, timeSpent int) float64 {
	if bucket == nil || bucket.ValidAttempts == 0 {
		return 50.0
	}

	// Считаем сколько попыток быстрее.
	fasterAttempts := uint64(0)

	switch {
	case timeSpent < 60:
		// Nothing.
	case timeSpent < 120:
		fasterAttempts += bucket.Time0_60
	case timeSpent < 180:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120
	case timeSpent < 240:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180
	case timeSpent < 300:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240
	case timeSpent < 360:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240 + bucket.Time240_300
	default:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240 + bucket.Time240_300 + bucket.Time300_360
	}

	// Добавляем половину из текущей категории.
	currentCategoryAttempts := m.getCurrentTimeCategoryAttempts(bucket, timeSpent)
	fasterAttempts += uint64(float64(currentCategoryAttempts) * 0.5)

	timePercentile := (float64(fasterAttempts) / float64(bucket.ValidAttempts)) * 100.0
	return math.Min(timePercentile, 100)
}

// getCurrentBucketAttempts возвращает количество попыток в текущей корзине процентов.
func (m *Manager) getCurrentBucketAttempts(bucket *entities.TestBucket, percentage float64) uint64 {
	switch {
	case percentage <= 5:
		return bucket.Pct0_5
	case percentage <= 10:
		return bucket.Pct5_10
	case percentage <= 15:
		return bucket.Pct10_15
	case percentage <= 20:
		return bucket.Pct15_20
	case percentage <= 25:
		return bucket.Pct20_25
	case percentage <= 30:
		return bucket.Pct25_30
	case percentage <= 35:
		return bucket.Pct30_35
	case percentage <= 40:
		return bucket.Pct35_40
	case percentage <= 45:
		return bucket.Pct40_45
	case percentage <= 50:
		return bucket.Pct45_50
	case percentage <= 55:
		return bucket.Pct50_55
	case percentage <= 60:
		return bucket.Pct55_60
	case percentage <= 65:
		return bucket.Pct60_65
	case percentage <= 70:
		return bucket.Pct65_70
	case percentage <= 75:
		return bucket.Pct70_75
	case percentage <= 80:
		return bucket.Pct75_80
	case percentage <= 85:
		return bucket.Pct80_85
	case percentage <= 90:
		return bucket.Pct85_90
	case percentage <= 95:
		return bucket.Pct90_95
	default:
		return bucket.Pct95_100
	}
}

// getCurrentTimeCategoryAttempts возвращает количество попыток в текущей категории времени.
func (m *Manager) getCurrentTimeCategoryAttempts(bucket *entities.TestBucket, timeSpent int) uint64 {
	switch {
	case timeSpent < 60:
		return bucket.Time0_60
	case timeSpent < 120:
		return bucket.Time60_120
	case timeSpent < 180:
		return bucket.Time120_180
	case timeSpent < 240:
		return bucket.Time180_240
	case timeSpent < 300:
		return bucket.Time240_300
	case timeSpent < 360:
		return bucket.Time300_360
	default:
		return bucket.Time360_
	}
}

// DetermineDistributionCategory определяет категорию распределения.
func (m *Manager) DetermineDistributionCategory(percentage float64) *entities.DistributionCategory {
	categories := []entities.DistributionCategory{
		{Name: "elite", MinScore: 90, MaxScore: 100},
		{Name: "excellent", MinScore: 75, MaxScore: 90},
		{Name: "good", MinScore: 60, MaxScore: 75},
		{Name: "average", MinScore: 40, MaxScore: 60},
		{Name: "below_average", MinScore: 20, MaxScore: 40},
		{Name: "needs_improvement", MinScore: 0, MaxScore: 20},
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
	case percentile >= 90:
		return "expert"
	case percentile >= 70:
		return "advanced"
	case percentile >= 40:
		return "intermediate"
	default:
		return "beginner"
	}
}

// GetTimeCategory определяет категорию скорости.
func (m *Manager) GetTimeCategory(timePercentile float64) string {
	switch {
	case timePercentile >= 90:
		return "very_fast"
	case timePercentile >= 70:
		return "fast"
	case timePercentile >= 40:
		return "average_speed"
	case timePercentile >= 20:
		return "slow"
	default:
		return "very_slow"
	}
}

// GetPerformanceQuadrant определяет квадрант производительности.
func (m *Manager) GetPerformanceQuadrant(percentage float64, timePercentile float64) map[string]interface{} {
	quadrant := ""

	if percentage >= 80 && timePercentile >= 80 {
		quadrant = "expert"
	} else if percentage >= 80 && timePercentile < 80 {
		quadrant = "precise_but_slow"
	} else if percentage < 80 && timePercentile >= 80 {
		quadrant = "fast_but_inaccurate"
	} else if percentage >= 50 && timePercentile >= 50 {
		quadrant = "solid"
	} else if percentage < 50 && timePercentile < 50 {
		quadrant = "needs_practice"
	} else {
		quadrant = "mixed"
	}

	return map[string]interface{}{
		"name": quadrant,
		"x":    percentage,
		"y":    timePercentile,
	}
}

// GetComparisonStatus возвращает статус сравнения с средним.
func (m *Manager) GetComparisonStatus(userValue, avgValue float64) string {
	diff := userValue - avgValue
	switch {
	case diff > 15:
		return "significantly_higher"
	case diff > 5:
		return "higher"
	case diff < -15:
		return "significantly_lower"
	case diff < -5:
		return "lower"
	default:
		return "similar"
	}
}

// GetTimeComparisonStatus возвращает статус сравнения времени.
func (m *Manager) GetTimeComparisonStatus(userTime, avgTime float64) string {
	diff := userTime - avgTime
	switch {
	case diff < -60:
		return "much_faster"
	case diff < -20:
		return "faster"
	case diff > 60:
		return "much_slower"
	case diff > 20:
		return "slower"
	default:
		return "similar"
	}
}
