package entities

// PercentDistribution - распределение процентов.
type PercentDistribution struct {
	Buckets map[int]uint64 `json:"buckets"`
}

// initPercentBuckets инициализирует процентные интервалы.
func (s *TestStats) initPercentBuckets() {
	s.PercentDistrib = &PercentDistribution{
		Buckets: make(map[int]uint64),
	}

	// Инициализируем интервалы от 0 до 100 с шагом 5.
	for i := 1; i <= bucketsCount; i++ {
		minVal := i * percStep
		s.PercentDistrib.Buckets[minVal] = 0
	}
}
