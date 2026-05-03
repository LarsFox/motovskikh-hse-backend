package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// UpdateAverages обновляет средние значения.
func updateAverages(s *entities.TestStats, percentage, timeSpent float64) {
	n := float64(s.Attempts)
	if n == 0 {
		return
	}
	s.AvgPercentage = ((s.AvgPercentage * (n - 1)) + percentage) / n
	s.AvgTimeSpent = ((s.AvgTimeSpent * (n - 1)) + timeSpent) / n
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
