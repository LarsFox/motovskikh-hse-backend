package entities

import (
	"time"
)

const (
	secondsPerQuestionMin = 3
	secondsPerQuestionMax = 30
	step                  = 5
	interval              = 20
)

// TestStats - статистика по тесту.
type TestStats struct {
	TestName       string               `json:"test_name"`
	UpdatedAt      time.Time            `json:"updated_at"`
	Attempts       uint64               `json:"attempts"`
	PercentDistrib *PercentDistribution `json:"percent_distrib"`
	TimeDistrib    *TimeDistribution    `json:"time_distrib"`
	AvgPercentage  float64              `json:"avg_percentage"`
	AvgTimeSpent   float64              `json:"avg_time_spent"`
	MinPercentage  float64              `json:"min_percentage"`
	MaxPercentage  float64              `json:"max_percentage"`
	MinTimeSpent   int                  `json:"min_time_spent"`
	MaxTimeSpent   int                  `json:"max_time_spent"`
}

// TestStatsResponse - статистика для ответа клиенту.
type TestStatsResponse struct {
	TestName      string    `json:"test_id"`
	TotalAttempts int       `json:"total_attempts"`
	AvgPercentage float64   `json:"avg_percentage"`
	AvgTimeSpent  float64   `json:"avg_time_spent"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// NewTestStats создает новую статистику теста с инициализированными бакетами.
func NewTestStats(testName string, questionCount int) *TestStats {
	stats := &TestStats{
		TestName:  testName,
		UpdatedAt: time.Now(),
	}
	stats.InitPercentBuckets()
	stats.InitTimeBuckets(questionCount)
	return stats
}

// UpdateAverages обновляет средние значения.
func (s *TestStats) UpdateAverages(percentage, timeSpent float64) {
	oldTotal := float64(s.Attempts - 1)

	if s.Attempts == 1 {
		s.AvgPercentage = percentage
		s.AvgTimeSpent = timeSpent
	} else {
		s.AvgPercentage = (s.AvgPercentage*oldTotal + percentage) / float64(s.Attempts)
		s.AvgTimeSpent = (s.AvgTimeSpent*oldTotal + timeSpent) / float64(s.Attempts)
	}
}

// UpdateMinMax обновляет минимальные и максимальные значения.
func (s *TestStats) UpdateMinMax(percentage float64, timeSpent int) {
	if s.Attempts == 1 {
		s.MinPercentage = percentage
		s.MaxPercentage = percentage
		s.MinTimeSpent = timeSpent
		s.MaxTimeSpent = timeSpent
		return
	}

	// Если не первая попытка
	if percentage < s.MinPercentage {
		s.MinPercentage = percentage
	}
	if percentage > s.MaxPercentage {
		s.MaxPercentage = percentage
	}
	if timeSpent < s.MinTimeSpent {
		s.MinTimeSpent = timeSpent
	}
	if timeSpent > s.MaxTimeSpent {
		s.MaxTimeSpent = timeSpent
	}
}
