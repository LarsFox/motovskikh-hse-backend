package mysql

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// dbTimeDistribution - для БД.
type dbTimeDistribution struct {
	Buckets []dbTimeBucket `json:"buckets"`
}

type dbTimeBucket struct {
	MinSeconds int    `json:"min_seconds"`
	MaxSeconds int    `json:"max_seconds"`
	Count      uint64 `json:"count"`
}

// Scan для dbTimeDistribution.
func (td *dbTimeDistribution) Scan(value any) error {
	if value == nil {
		*td = dbTimeDistribution{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return entities.ErrInvalidInput
	}
	return json.Unmarshal(bytes, td)
}

// Value для dbTimeDistribution.
func (td *dbTimeDistribution) Value() (driver.Value, error) {
	if td == nil || len(td.Buckets) == 0 {
		return nil, nil
	}
	return json.Marshal(td)
}

// ToEntity конвертирует DB-структуру в entity.
func (td *dbTimeDistribution) toEntity() *entities.TimeDistribution {
	if td == nil || len(td.Buckets) == 0 {
		return nil
	}
	buckets := make([]entities.TimeBucket, len(td.Buckets))
	for i, b := range td.Buckets {
		buckets[i] = entities.TimeBucket{
			MinSeconds: b.MinSeconds,
			MaxSeconds: b.MaxSeconds,
			Count:      b.Count,
		}
	}
	return &entities.TimeDistribution{Buckets: buckets}
}

// FromEntity конвертирует entity в DB-структуру.
func (td *dbTimeDistribution) fromEntity(e *entities.TimeDistribution) *dbTimeDistribution {
	if e == nil || len(e.Buckets) == 0 {
		return nil
	}
	buckets := make([]dbTimeBucket, len(e.Buckets))
	for i, b := range e.Buckets {
		buckets[i] = dbTimeBucket{
			MinSeconds: b.MinSeconds,
			MaxSeconds: b.MaxSeconds,
			Count:      b.Count,
		}
	}
	return &dbTimeDistribution{Buckets: buckets}
}
