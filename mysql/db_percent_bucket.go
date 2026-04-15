package mysql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// dbPercentDistribution - для БД.
type dbPercentDistribution struct {
	Buckets map[string]uint64 `json:"buckets"`
}

// Scan для dbPercentDistribution.
func (pd *dbPercentDistribution) Scan(value any) error {
	if value == nil {
		*pd = dbPercentDistribution{Buckets: make(map[string]uint64)}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return entities.ErrInvalidInput
	}
	return json.Unmarshal(bytes, pd)
}

// Value для dbPercentDistribution.
func (pd *dbPercentDistribution) Value() (driver.Value, error) {
	if pd == nil || len(pd.Buckets) == 0 {
		return nil, nil
	}
	return json.Marshal(pd)
}

// ToEntity конвертирует DB-структуру в entity.
func (pd *dbPercentDistribution) toEntity() *entities.PercentDistribution {
	if pd == nil || len(pd.Buckets) == 0 {
		return &entities.PercentDistribution{Buckets: make(map[float64]uint64)}
	}
	buckets := make(map[float64]uint64)
	for k, v := range pd.Buckets {
		var key float64
		_, err := fmt.Sscanf(k, "%f", &key)
		if err != nil {
			continue
		}
		buckets[key] = v
	}
	return &entities.PercentDistribution{Buckets: buckets}
}

// FromEntity конвертирует entity в DB-структуру.
func fromEntityPercent(e *entities.PercentDistribution) *dbPercentDistribution {
    if e == nil || len(e.Buckets) == 0 {
        return &dbPercentDistribution{Buckets: make(map[string]uint64)}
    }
    buckets := make(map[string]uint64)
    for k, v := range e.Buckets {
        buckets[fmt.Sprintf("%f", k)] = v
    }
    return &dbPercentDistribution{Buckets: buckets}
}
