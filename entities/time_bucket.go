package entities

// TimeDistribution - распределение времени.
type TimeDistribution struct {
	Buckets []TimeBucket `json:"buckets"`
}

type TimeBucket struct {
	MinSeconds int64  `json:"min_seconds"`
	MaxSeconds int64  `json:"max_seconds"`
	Count      uint64 `json:"count"`
}

// InitTimeBuckets создает интервалы на основе количества вопросов.
func (s *TestStats) InitTimeBuckets(questionCount int64) {
	s.TimeDistrib = &TimeDistribution{
		Buckets: make([]TimeBucket, 0),
	}

	// Минимальное время: 3 секунды на вопрос.
	minTime := questionCount * secondsPerQuestionMin
	// Максимальное время: 30 секунд на вопрос.
	maxTime := questionCount * secondsPerQuestionMax
	steps := []float64{1.0, 1.3, 1.7, 2.2, 3.0}

	prevMax := 0
	for i, mult := range steps {
		boundary := int(float64(minTime) * mult)
		boundary = min(boundary, int(maxTime))

		if boundary <= prevMax {
			continue
		}

		if i == len(steps)-1 {
			s.TimeDistrib.Buckets = append(s.TimeDistrib.Buckets, TimeBucket{
				MinSeconds: int64(prevMax),
				MaxSeconds: -1,
				Count:      0,
			})
		} else {
			s.TimeDistrib.Buckets = append(s.TimeDistrib.Buckets, TimeBucket{
				MinSeconds: int64(prevMax),
				MaxSeconds: int64(boundary),
				Count:      0,
			})
			prevMax = boundary
		}
	}
}

// GetTimeBucketIndex возвращает индекс бакета для заданного времени.
func (s *TestStats) GetTimeBucketIndex(timeSpent int64) int {
	if s.TimeDistrib == nil || len(s.TimeDistrib.Buckets) == 0 {
		return 0
	}

	for i, bucket := range s.TimeDistrib.Buckets {
		if bucket.MaxSeconds == -1 {
			return i
		}
		if timeSpent <= bucket.MaxSeconds {
			return i
		}
	}
	return len(s.TimeDistrib.Buckets) - 1
}

// UpdateTimeDistribution обновляет временное распределение.
func (s *TestStats) UpdateTimeDistribution(timeSpent int64) {
	idx := s.GetTimeBucketIndex(timeSpent)
	if idx >= 0 && idx < len(s.TimeDistrib.Buckets) {
		s.TimeDistrib.Buckets[idx].Count++
	}
}
