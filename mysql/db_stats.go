package mysql

import (
	"time"
	"encoding/json"
)

// dbTestStats - структура для работы с БД.
type dbTestStats struct {
	TestName       string          `gorm:"primaryKey"`
	UpdatedAt      time.Time
	Attempts       uint64
	PercentDistrib json.RawMessage `gorm:"type:json"`
	TimeDistrib    json.RawMessage `gorm:"type:json"`
	AvgPercentage  float64
	AvgTimeSpent   float64
	MinTimeSpent   int64
	MaxTimeSpent   int64
}
