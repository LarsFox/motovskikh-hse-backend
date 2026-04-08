package entities

// PercentDistribution - распределение процентов.
type PercentDistribution struct {
	Buckets map[float64]uint64 `json:"buckets"`
}

// InitPercentBuckets инициализирует процентные интервалы.
func (s *TestStats) InitPercentBuckets() {
	s.PercentDistrib = &PercentDistribution{
		Buckets: make(map[float64]uint64),
	}

	// Инициализируем интервалы от 0 до 100 с шагом 5.
	for i := 0; i <= interval; i++ {
		minVal := float64(i * step)
		s.PercentDistrib.Buckets[minVal] = 0
	}
}

// UpdatePercentDistribution обновляет процентное распределение.
func (s *TestStats) UpdatePercentDistribution(percentage float64) {
	key := float64(int(percentage/step) * step)
	if _, ok := s.PercentDistrib.Buckets[key]; ok {
		s.PercentDistrib.Buckets[key]++
	}
}
