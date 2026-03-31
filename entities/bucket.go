package entities

import (
	"time"
	"errors"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

const (
    secondsPerQuestionMin = 3
    secondsPerQuestionMax = 30
		step = 5
)

var (
    errInvalidTimeDistribution = errors.New("failed to unmarshal TimeDistribution value")
    errInvalidPercentDistribution = errors.New("failed to unmarshal PercentDistribution value")
)

// TestBucket - структура бакета.
type TestBucket struct {
	TestID          string                `gorm:"column:test_id;primaryKey"        json:"test_id"`
	UpdatedAt       time.Time             `gorm:"column:updated_at"                json:"updated_at"`
	Attempts        uint64                `gorm:"column:attempts"                  json:"attempts"`
	PercentDistrib  *PercentDistribution  `gorm:"column:percent_distrib;type:json" json:"percent_distrib"`
	TimeDistrib     *TimeDistribution     `gorm:"column:time_distrib;type:json"    json:"time_distrib"`
	AvgPercentage   float64               `gorm:"column:avg_percentage"            json:"avg_percentage"`
	AvgTimeSpent    float64               `gorm:"column:avg_time_spent"            json:"avg_time_spent"`
	MinPercentage   float64               `gorm:"column:min_percentage"            json:"min_percentage"`
	MaxPercentage   float64               `gorm:"column:max_percentage"            json:"max_percentage"`
	MinTimeSpent    int                   `gorm:"column:min_time_spent"            json:"min_time_spent"`
	MaxTimeSpent    int                   `gorm:"column:max_time_spent"            json:"max_time_spent"`
}


// TimeDistribution - распределение времени.
type TimeDistribution struct {
	Buckets []TimeBucket `json:"buckets"`
}

type TimeBucket struct {
	MinSeconds int    `json:"min_seconds"`
	MaxSeconds int    `json:"max_seconds"`
	Label      string `json:"label"`
	Count      uint64 `json:"count"`
}

// PercentDistribution - распределение процентов.
type PercentDistribution struct {
	Buckets []PercentBucket `json:"buckets"`
}

type PercentBucket struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"`
	Count uint64  `json:"count"`
}

// TestStats представляет статистику теста.
type TestStats struct {
	ID            string    `gorm:"primaryKey"     json:"id"`
	TestID        string    `gorm:"index"          json:"test_id"`
	TotalAttempts int       `json:"total_attempts"`
	ValidAttempts int       `json:"valid_attempts"`
	AvgPercentage float64   `json:"avg_percentage"`
	AvgTimeSpent  float64   `json:"avg_time_spent"`
	UpdatedAt     time.Time `json:"updated_at"`
}
// Scan для TimeDistribution.
func (td *TimeDistribution) Scan(value any) error {
	if value == nil {
		*td = TimeDistribution{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errInvalidTimeDistribution
	}
	return json.Unmarshal(bytes, td)
}

// Value для TimeDistribution.
func (td *TimeDistribution) Value() (driver.Value, error) {
	if len(td.Buckets) == 0 {
		return nil, nil //nolint:nilnil
	}
	return json.Marshal(td)
}

func (td *TimeDistribution) IsEmpty() bool {
    return td == nil || len(td.Buckets) == 0
}

// Scan для PercentDistribution.
func (pd *PercentDistribution) Scan(value any) error {
	if value == nil {
		*pd = PercentDistribution{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errInvalidPercentDistribution
	}
	return json.Unmarshal(bytes, pd)
}

// Value для PercentDistribution.
func (pd *PercentDistribution) Value() (driver.Value, error) {
	if len(pd.Buckets) == 0 {
		return nil, nil //nolint:nilnil
	}
	return json.Marshal(pd)
}

func (pd *PercentDistribution) IsEmpty() bool {
    return pd == nil || len(pd.Buckets) == 0
}

// InitializeBuckets создает стандартные бакеты для теста.
func (b *TestBucket) InitializeBuckets(questionCount int) {
	b.initializePercentBuckets()
	b.initializeTimeBuckets(questionCount)
}

// initializePercentBuckets инициализирует процентные интервалы.
func (b *TestBucket) initializePercentBuckets() {
	if b.PercentDistrib == nil {
		b.PercentDistrib = &PercentDistribution{
			Buckets: make([]PercentBucket, 0),
		}
		// 20 интервалов по 5%.
		for i := range 20 {
			minVal := float64(i * step)
			maxVal := minVal + step
			b.PercentDistrib.Buckets = append(b.PercentDistrib.Buckets, PercentBucket{
					Min:   minVal,
					Max:   maxVal,
					Label: fmt.Sprintf("%.0f-%.0f%%", minVal, maxVal),
					Count: 0,
			})
		}
	}
}

// initializeTimeBuckets создает интервалы на основе количества вопросов.
func (b *TestBucket) initializeTimeBuckets(questionCount int) {
	if b.TimeDistrib == nil {
		b.TimeDistrib = &TimeDistribution{
			Buckets: make([]TimeBucket, 0),
		}
		// Минимальное время: 3 секунды на вопрос.
		minTime := questionCount * secondsPerQuestionMin
		// Максимальное время: 30 секунд на вопрос.
		maxTime := questionCount * secondsPerQuestionMax
		steps := []float64{1.0, 1.3, 1.7, 2.2, 3.0}
		
		prevMax := 0
		for i, mult := range steps {
			// Вычисляем верхнюю границу интервала.
			boundary := int(float64(minTime) * mult)
			// Ограничиваем максимальным временем.
			boundary = min(boundary, maxTime)
			// Пропускаем интервалы, которые не увеличивают границу, чтобы не дублировать интервалы.
			if boundary <= prevMax {
				continue
			}
			
			// Для последнего интервала делаем открытую верхнюю границу.
			if i == len(steps)-1 {
				b.TimeDistrib.Buckets = append(b.TimeDistrib.Buckets, TimeBucket{
					MinSeconds: prevMax,
					MaxSeconds: -1,
					Label:      fmt.Sprintf("> %d", prevMax),
					Count:      0,
				})
			} else {
				b.TimeDistrib.Buckets = append(b.TimeDistrib.Buckets, TimeBucket{
					MinSeconds: prevMax,
					MaxSeconds: boundary,
					Label:      fmt.Sprintf("%d-%d", prevMax, boundary),
					Count:      0,
				})
				prevMax = boundary
			}
		}
	}
}

// GetPercentBucketIndex возвращает индекс бакета для заданного процента.
func (b *TestBucket) GetPercentBucketIndex(percentage float64) int {
	if b.PercentDistrib == nil || len(b.PercentDistrib.Buckets) == 0 {
		return 0
	}
	for i, bucket := range b.PercentDistrib.Buckets {
		if percentage <= bucket.Max {
			return i
		}
	}
	return len(b.PercentDistrib.Buckets) - 1
}

// GetTimeBucketIndex возвращает индекс бакета для заданного времени.
func (b *TestBucket) GetTimeBucketIndex(timeSpent int) int {
	if b.TimeDistrib == nil || len(b.TimeDistrib.Buckets) == 0 {
		return 0
	}
	
	for i, bucket := range b.TimeDistrib.Buckets {
		if bucket.MaxSeconds == -1 {
			return i
		}
		if timeSpent <= bucket.MaxSeconds {
			return i
		}
	}
	return len(b.TimeDistrib.Buckets) - 1
}
