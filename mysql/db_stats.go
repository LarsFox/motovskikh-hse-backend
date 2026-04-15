package mysql

import (
	"time"
)

// dbTestStats - структура для работы с БД.
type dbTestStats struct {
	TestName         string                 `gorm:"column:test_name;primaryKey"`
	UpdatedAt        time.Time              `gorm:"column:updated_at"`
	Attempts         uint64                 `gorm:"column:attempts"`
	PercentDistribDB *dbPercentDistribution `gorm:"column:percent_distrib;type:json"`
	TimeDistribDB    *dbTimeDistribution    `gorm:"column:time_distrib;type:json"`
	AvgPercentage    float64                `gorm:"column:avg_percentage"`
	AvgTimeSpent     float64                `gorm:"column:avg_time_spent"`
	MinTimeSpent     int64                  `gorm:"column:min_time_spent"`
	MaxTimeSpent     int64                  `gorm:"column:max_time_spent"`
}

func (dbTestStats) TableName() string {
    return "test_stats"
}
