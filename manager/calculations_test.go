package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/mocks"
)

func TestCalculatePercentile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)

	tests := []struct {
		name       string
		stats      *entities.TestStats
		percentage float64
		expected   float64
	}{
		{
			name:       "nil stats returns 100",
			stats:      nil,
			percentage: 70,
			expected:   100,
		},
		{
			name: "empty stats returns 100",
			stats: &entities.TestStats{
				Attempts: 0,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: map[float64]uint64{
						0: 0, 5: 0, 10: 0, 15: 0, 20: 0, 25: 0, 30: 0, 35: 0, 40: 0, 45: 0,
						50: 0, 55: 0, 60: 0, 65: 0, 70: 0, 75: 0, 80: 0, 85: 0, 90: 0, 95: 0, 100: 0,
					},
				},
			},
			percentage: 70,
			expected:   100,
		},
		{
			name: "perfect score returns 100",
			stats: &entities.TestStats{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: map[float64]uint64{
						0: 10, 5: 5, 10: 5, 15: 5, 20: 5, 25: 5, 30: 5, 35: 5, 40: 5, 45: 5,
						50: 5, 55: 5, 60: 5, 65: 5, 70: 5, 75: 5, 80: 5, 85: 5, 90: 5, 95: 5, 100: 0,
					},
				},
			},
			percentage: 98,
			expected:   100,
		},
		{
			name: "average score calculates correctly",
			stats: &entities.TestStats{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: map[float64]uint64{
						0: 10, 5: 10, 10: 10, 15: 10, 20: 10, 25: 10, 30: 10, 35: 10, 40: 10, 45: 10,
						50: 0, 55: 0, 60: 0, 65: 0, 70: 0, 75: 0, 80: 0, 85: 0, 90: 0, 95: 0, 100: 0,
					},
				},
			},
			percentage: 48,
			expected:   95,
		},
		{
			name: "50th percentile calculation",
			stats: &entities.TestStats{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: map[float64]uint64{
						0: 5, 5: 5, 10: 5, 15: 5, 20: 5, 25: 5, 30: 5, 35: 5, 40: 5, 45: 5,
						50: 10, 55: 10, 60: 10, 65: 10, 70: 10, 75: 10, 80: 0, 85: 0, 90: 0, 95: 0, 100: 0,
					},
				},
			},
			percentage: 52,
			expected:   55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.calculatePercentile(tt.stats, tt.percentage)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestCalculateTimePercentile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)

	tests := []struct {
		name      string
		stats     *entities.TestStats
		timeSpent int
		expected  float64
	}{
		{
			name: "fast time returns low percentile",
			stats: &entities.TestStats{
				Attempts: 100,
				TimeDistrib: &entities.TimeDistribution{
					Buckets: []entities.TimeBucket{
						{MinSeconds: 0, MaxSeconds: 60, Count: 30},
						{MinSeconds: 60, MaxSeconds: 120, Count: 30},
						{MinSeconds: 120, MaxSeconds: 180, Count: 20},
						{MinSeconds: 180, MaxSeconds: 240, Count: 10},
						{MinSeconds: 240, MaxSeconds: 300, Count: 5},
						{MinSeconds: 300, MaxSeconds: 360, Count: 3},
						{MinSeconds: 360, MaxSeconds: -1, Count: 2},
					},
				},
			},
			timeSpent: 45,
			expected:  85,
		},
		{
			name: "slow time returns high percentile",
			stats: &entities.TestStats{
				Attempts: 100,
				TimeDistrib: &entities.TimeDistribution{
					Buckets: []entities.TimeBucket{
						{MinSeconds: 0, MaxSeconds: 60, Count: 30},
						{MinSeconds: 60, MaxSeconds: 120, Count: 30},
						{MinSeconds: 120, MaxSeconds: 180, Count: 20},
						{MinSeconds: 180, MaxSeconds: 240, Count: 10},
						{MinSeconds: 240, MaxSeconds: 300, Count: 5},
						{MinSeconds: 300, MaxSeconds: 360, Count: 3},
						{MinSeconds: 360, MaxSeconds: -1, Count: 2},
					},
				},
			},
			timeSpent: 400,
			expected:  1,
		},
		{
			name:      "nil stats returns 100",
			stats:     nil,
			timeSpent: 45,
			expected:  100,
		},
		{
			name: "empty stats returns 100",
			stats: &entities.TestStats{
				Attempts: 0,
				TimeDistrib: &entities.TimeDistribution{
					Buckets: []entities.TimeBucket{},
				},
			},
			timeSpent: 45,
			expected:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.calculateTimePercentile(tt.stats, int64(tt.timeSpent))
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}
