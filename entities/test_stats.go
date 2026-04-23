package entities

import (
	"time"
)

const (
	secondsPerQuestionMin = 2
	secondsPerQuestionMax = 30
	bucketsCount          = 20
	percStep							= 5
	smallTestThreshold1 = 7
	smallTestThreshold2 = 15
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
	MinTimeSpent   int64                `json:"min_time_spent"`
	MaxTimeSpent   int64                `json:"max_time_spent"`
}

// NewTestStats создает новую статистику теста с инициализированными бакетами.
func NewTestStats(testName string, questionCount int64) *TestStats {
	stats := &TestStats{
		TestName:  testName,
		UpdatedAt: time.Now(),
	}
	stats.initPercentBuckets()
	stats.initTimeBuckets(questionCount)
	return stats
}
