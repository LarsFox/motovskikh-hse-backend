package entities

import (
	"time"
)

const (
	secondsPerQuestionMin = 2
	secondsPerQuestionMax = 30
	bucketsCount          = 20
	percStep              = 5
	smallTestThreshold1   = 7
	smallTestThreshold2   = 15
)

// TestStats - статистика по тесту.
type TestStats struct {
	Name           string
	UpdatedAt      time.Time
	Attempts       int64
	PercentBuckets []*TestStatsBucket // TODO: implement
	TimeBuckets    []*TestStatsBucket
	AvgPercentage  float64
	AvgTimeSpent   float64
	MinTimeSpent   float64
	MaxTimeSpent   float64
}

type TestStatsAnalysis struct {
	ScorePercentile   float64
	TimePercentile    float64
	BetterThan        int64
	FasterThan        int64
	AveragePercentage float64
	AverageTime       float64
	PercentageDiff float64
	TimeDiff       float64
}

type TestStatsBucket struct {
	Value float64 `json:"value"`
	Count int64   `json:"count"`
}
