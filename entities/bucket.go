package entities

import "time"

// Здесь структуры, связанные с хранением данных.

type TestBucket struct {
	ID        string    `json:"id" gorm:"primaryKey;column:id"`
	TestID    string    `json:"test_id" gorm:"column:test_id;index;unique"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	
	TotalAttempts   uint64 `json:"total_attempts" gorm:"column:total_attempts"`
	ValidAttempts   uint64 `json:"valid_attempts" gorm:"column:valid_attempts"`
	InvalidAttempts uint64 `json:"invalid_attempts" gorm:"column:invalid_attempts"`

	Pct0_5   uint64 `json:"pct_0_5" gorm:"column:pct_0_5"`
	Pct5_10  uint64 `json:"pct_5_10" gorm:"column:pct_5_10"`
	Pct10_15 uint64 `json:"pct_10_15" gorm:"column:pct_10_15"`
	Pct15_20 uint64 `json:"pct_15_20" gorm:"column:pct_15_20"`
	Pct20_25 uint64 `json:"pct_20_25" gorm:"column:pct_20_25"`
	Pct25_30 uint64 `json:"pct_25_30" gorm:"column:pct_25_30"`
	Pct30_35 uint64 `json:"pct_30_35" gorm:"column:pct_30_35"`
	Pct35_40 uint64 `json:"pct_35_40" gorm:"column:pct_35_40"`
	Pct40_45 uint64 `json:"pct_40_45" gorm:"column:pct_40_45"`
	Pct45_50 uint64 `json:"pct_45_50" gorm:"column:pct_45_50"`
	Pct50_55 uint64 `json:"pct_50_55" gorm:"column:pct_50_55"`
	Pct55_60 uint64 `json:"pct_55_60" gorm:"column:pct_55_60"`
	Pct60_65 uint64 `json:"pct_60_65" gorm:"column:pct_60_65"`
	Pct65_70 uint64 `json:"pct_65_70" gorm:"column:pct_65_70"`
	Pct70_75 uint64 `json:"pct_70_75" gorm:"column:pct_70_75"`
	Pct75_80 uint64 `json:"pct_75_80" gorm:"column:pct_75_80"`
	Pct80_85 uint64 `json:"pct_80_85" gorm:"column:pct_80_85"`
	Pct85_90 uint64 `json:"pct_85_90" gorm:"column:pct_85_90"`
	Pct90_95 uint64 `json:"pct_90_95" gorm:"column:pct_90_95"`
	Pct95_100 uint64 `json:"pct_95_100" gorm:"column:pct_95_100"`
	
	// Распределение времени.
	Time0_60    uint64 `json:"time_0_60" gorm:"column:time_0_60"`
	Time60_120  uint64 `json:"time_60_120" gorm:"column:time_60_120"`
	Time120_180 uint64 `json:"time_120_180" gorm:"column:time_120_180"`
	Time180_240 uint64 `json:"time_180_240" gorm:"column:time_180_240"`
	Time240_300 uint64 `json:"time_240_300" gorm:"column:time_240_300"`
	Time300_360 uint64 `json:"time_300_360" gorm:"column:time_300_360"`
	Time360_    uint64 `json:"time_360_" gorm:"column:time_360_"`
	
	AvgPercentage float64 `json:"avg_percentage" gorm:"column:avg_percentage"`
	AvgTimeSpent  float64 `json:"avg_time_spent" gorm:"column:avg_time_spent"`
	
	MinPercentage float64 `json:"min_percentage" gorm:"column:min_percentage"`
	MaxPercentage float64 `json:"max_percentage" gorm:"column:max_percentage"`
	MinTimeSpent  int     `json:"min_time_spent" gorm:"column:min_time_spent"`
	MaxTimeSpent  int     `json:"max_time_spent" gorm:"column:max_time_spent"`
}

type TestStats struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	TestID        string    `json:"test_id" gorm:"index"`
	Period        string    `json:"period" gorm:"column:period"`
	TotalAttempts int       `json:"total_attempts"`
	ValidAttempts int       `json:"valid_attempts"`
	AvgPercentage float64   `json:"avg_percentage"`
	AvgTimeSpent  float64   `json:"avg_time_spent"`
	UpdatedAt     time.Time `json:"updated_at"`
}