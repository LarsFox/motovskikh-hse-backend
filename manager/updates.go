package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// GetTimeBucketIndex возвращает индекс бакета для заданного времени.
func getTimeBucketIndex(s *entities.TestStats, timeSpent int64) int {
	if s.TimeDistrib == nil || len(s.TimeDistrib.Buckets) == 0 {
		return 0
	}

	for i := len(s.TimeDistrib.Buckets) - 1; i >= 0; i-- {
		if timeSpent >= s.TimeDistrib.Buckets[i].MinSeconds {
			return i
		}
	}
	return 0
}

// UpdateTimeDistribution обновляет временное распределение.
func updateTimeDistribution(s *entities.TestStats, timeSpent int64) {
	if s.TimeDistrib == nil || len(s.TimeDistrib.Buckets) == 0 {
		return
	}
	idx := getTimeBucketIndex(s, timeSpent)
	if idx >= 0 && idx < len(s.TimeDistrib.Buckets) {
		s.TimeDistrib.Buckets[idx].Count++
	}
}

// UpdatePercentDistribution обновляет процентное распределение.
func updatePercentDistribution(s *entities.TestStats, percentage float64) {
	if s.PercentDistrib == nil || s.PercentDistrib.Buckets == nil {
		return
	}
	key := int(percentage/step) * step
	if _, ok := s.PercentDistrib.Buckets[key]; ok {
		s.PercentDistrib.Buckets[key]++
	}
}

// UpdateAverages обновляет средние значения.
func updateAverages(s *entities.TestStats, percentage, timeSpent float64) {
	oldTotal := float64(s.Attempts - 1)
	if s.Attempts == 1 {
		s.AvgPercentage = percentage
		s.AvgTimeSpent = timeSpent
	} else {
		s.AvgPercentage = (s.AvgPercentage*oldTotal + percentage) / float64(s.Attempts)
		s.AvgTimeSpent = (s.AvgTimeSpent*oldTotal + timeSpent) / float64(s.Attempts)
	}
}

// UpdateMinMax обновляет минимальные и максимальные значения.
func updateMinMax(s *entities.TestStats, timeSpent int64) {
	if s.Attempts == 1 {
		s.MinTimeSpent = timeSpent
		s.MaxTimeSpent = timeSpent
		return
	}
	if timeSpent < s.MinTimeSpent {
		s.MinTimeSpent = timeSpent
	}
	if timeSpent > s.MaxTimeSpent {
		s.MaxTimeSpent = timeSpent
	}
}