package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

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
func updateMinMax(s *entities.TestStats, timeSpent float64) {
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
