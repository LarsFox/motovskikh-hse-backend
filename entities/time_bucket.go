package entities

// TimeDistribution - распределение времени.
type TimeDistribution struct {
	Buckets []TimeBucket `json:"buckets"`
}

type TimeBucket struct {
	MinSeconds int64  `json:"min_seconds"`
	Count      uint64 `json:"count"`
}

// initTimeBuckets создает интервалы на основе количества вопросов.
func (s *TestStats) initTimeBuckets(questionCount int64) {
	minPerQuestion := float64(secondsPerQuestionMin)
	if questionCount < smallTestThreshold1 {
		minPerQuestion = 0.5
	} else if questionCount < smallTestThreshold2 {
		minPerQuestion = 1
	}
	// Минимальное время: 3 секунды на вопрос.
	minTime := questionCount * int64(minPerQuestion)
	// Максимальное время: 30 секунд на вопрос.
	maxTime := questionCount * secondsPerQuestionMax
	step := (maxTime - minTime) / bucketsCount

	s.TimeDistrib = &TimeDistribution{
		Buckets: make([]TimeBucket, bucketsCount),
	}
	s.TimeDistrib.Buckets[0] = TimeBucket{
			MinSeconds: 0,
			Count:      0,
		}
	for i := range bucketsCount - 1 {
		minSeconds := minTime + int64(i)*step
		s.TimeDistrib.Buckets[i+1] = TimeBucket{
			MinSeconds: minSeconds,
			Count:      0,
		}
	}
}
