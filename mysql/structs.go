package mysql

import (
	"encoding/json"
	"time"
)

type testStats struct {
	Name           string 		`gorm:"primaryKey"`
	UpdatedAt      time.Time
	Attempts       int64
	PercentDistrib json.RawMessage
	TimeDistrib    json.RawMessage
	AvgPercentage  float64
	AvgTimeSpent   float64
	MinTimeSpent   float64
	MaxTimeSpent   float64
}
