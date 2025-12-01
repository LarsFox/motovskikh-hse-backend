package entities

import (
	"time"
)

// Attempt — попытка прохождения теста.
type Attempt struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	TestID     string    `json:"test_id"`
	VersionID  string    `json:"version_id"`
	UserHash   string    `json:"user_hash"` // Хэш пользователя для анонимного трекинга.
	Score      int       `json:"score"`
	MaxScore   int       `json:"max_score"`
	Percentage float64   `json:"percentage"`
	TimeSpent  int       `json:"time_spent"` // в секундах.
	Answers    string    `json:"answers"`    // JSON с ответами.
	CreatedAt  time.Time `json:"created_at"`
}

// TestStats — статистика по тесту.
type TestStats struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	TestID        string    `json:"test_id"`
	VersionID     string    `json:"version_id"`
	Date          time.Time `json:"date"`
	TotalAttempts int       `json:"total_attempts"`
	ValidAttempts int       `json:"valid_attempts"`
	AvgPercentage float64   `json:"avg_percentage"`
	AvgTimeSpent  float64   `json:"avg_time_spent"`
	Percentile50  float64   `json:"percentile_50"`
	Percentile80  float64   `json:"percentile_80"`
	Percentile95  float64   `json:"percentile_95"`
}