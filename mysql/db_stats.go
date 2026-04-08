package mysql

import (
	"time"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
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
	MinPercentage    float64                `gorm:"column:min_percentage"`
	MaxPercentage    float64                `gorm:"column:max_percentage"`
	MinTimeSpent     int                    `gorm:"column:min_time_spent"`
	MaxTimeSpent     int                    `gorm:"column:max_time_spent"`
}

// toEntityStats конвертирует DB-структуру в entity.TestStats.
func (db *dbTestStats) toEntityStats() *entities.TestStats {
	return &entities.TestStats{
		TestName:       db.TestName,
		UpdatedAt:      db.UpdatedAt,
		Attempts:       db.Attempts,
		PercentDistrib: db.PercentDistribDB.toEntity(),
		TimeDistrib:    db.TimeDistribDB.toEntity(),
		AvgPercentage:  db.AvgPercentage,
		AvgTimeSpent:   db.AvgTimeSpent,
		MinPercentage:  db.MinPercentage,
		MaxPercentage:  db.MaxPercentage,
		MinTimeSpent:   db.MinTimeSpent,
		MaxTimeSpent:   db.MaxTimeSpent,
	}
}

// fromEntityStats конвертирует entity.TestStats в DB-структуру.
func fromEntityStats(e *entities.TestStats) *dbTestStats {
	if e == nil {
		return nil
	}
	var percentDistribDB *dbPercentDistribution
	var timeDistribDB *dbTimeDistribution

	if e.PercentDistrib != nil {
		percentDistribDB = (&dbPercentDistribution{}).fromEntity(e.PercentDistrib)
	}
	if e.TimeDistrib != nil {
		timeDistribDB = (&dbTimeDistribution{}).fromEntity(e.TimeDistrib)
	}

	return &dbTestStats{
		TestName:         e.TestName,
		UpdatedAt:        e.UpdatedAt,
		Attempts:         e.Attempts,
		PercentDistribDB: percentDistribDB,
		TimeDistribDB:    timeDistribDB,
		AvgPercentage:    e.AvgPercentage,
		AvgTimeSpent:     e.AvgTimeSpent,
		MinPercentage:    e.MinPercentage,
		MaxPercentage:    e.MaxPercentage,
		MinTimeSpent:     e.MinTimeSpent,
		MaxTimeSpent:     e.MaxTimeSpent,
	}
}
