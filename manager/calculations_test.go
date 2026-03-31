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
		bucket     *entities.TestBucket
		percentage float64
		expected   float64
	}{
		{
			name:       "nil bucket returns 100",
			bucket:     nil,
			percentage: 70,
			expected:   100,
		},
		{
			name: "empty bucket returns 100",
			bucket: &entities.TestBucket{
				Attempts: 0,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: []entities.PercentBucket{
						{Min: 0, Max: 20, Count: 0},
						{Min: 20, Max: 40, Count: 0},
						{Min: 40, Max: 60, Count: 0},
						{Min: 60, Max: 80, Count: 0},
						{Min: 80, Max: 100, Count: 0},
					},
				},
			},
			percentage: 70,
			expected:   100,
		},
		{
			name: "perfect score returns high percentile",
			bucket: &entities.TestBucket{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: []entities.PercentBucket{
						{Min: 0, Max: 20, Count: 10},
						{Min: 20, Max: 40, Count: 20},
						{Min: 40, Max: 60, Count: 30},
						{Min: 60, Max: 80, Count: 25},
						{Min: 80, Max: 100, Count: 15},
					},
				},
			},
			percentage: 95,
			expected:   92,
		},
		{
			name: "average score calculates correctly",
			bucket: &entities.TestBucket{
				Attempts: 100,
				PercentDistrib: &entities.PercentDistribution{
					Buckets: []entities.PercentBucket{
						{Min: 0, Max: 20, Count: 10},
						{Min: 20, Max: 40, Count: 20},
						{Min: 40, Max: 60, Count: 30},
						{Min: 60, Max: 80, Count: 25},
						{Min: 80, Max: 100, Count: 15},
					},
				},
			},
			percentage: 50,
			expected:   45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.CalculatePercentile(tt.bucket, tt.percentage)
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
		name       string
		bucket     *entities.TestBucket
		timeSpent  int
		expected   float64
	}{
		{
			name: "fast time returns low percentile",
			bucket: &entities.TestBucket{
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
			expected:  15, // в первом бакете, половина от 30 = 15
		},
		{
			name: "slow time returns high percentile",
			bucket: &entities.TestBucket{
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
			expected:  99, // 30+30+20+10+5+3 = 98 + половина последнего (1) = 99
		},
		{
			name: "nil bucket returns 100",
			bucket:     nil,
			timeSpent:  45,
			expected:   100,
		},
		{
			name: "empty bucket returns 100",
			bucket: &entities.TestBucket{
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
			result := m.CalculateTimePercentile(tt.bucket, tt.timeSpent)
			assert.InDelta(t, tt.expected, result, 0.1)
		})
	}
}

func TestDetermineDistributionCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)
	
	tests := []struct {
		percentage float64
		expected   string
	}{
		{95, "elite"},
		{82, "excellent"},
		{68, "good"},
		{50, "average"},
		{30, "below_average"},
		{10, "needs_improvement"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := m.DetermineDistributionCategory(tt.percentage)
			assert.Equal(t, tt.expected, result.Name)
		})
	}
}

func TestGetPerformanceQuadrant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockDB := mocks.NewMockdb(ctrl)
	m := New(mockDB)
	
	tests := []struct {
		name          string
		percentage    float64
		timePercentile float64
		expected      string
	}{
		{"expert", 85, 90, "expert"},
		{"slow expert", 85, 30, "slow_expert"},
		{"fast but inaccurate", 60, 85, "fast_but_inaccurate"},
		{"solid", 65, 60, "solid"},
		{"needs practice", 40, 30, "needs_practice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.GetPerformanceQuadrant(tt.percentage, tt.timePercentile)
			assert.Equal(t, tt.expected, result["name"])
			assert.InDelta(t, tt.percentage, result["x"], 0.001)
			assert.InDelta(t, tt.timePercentile, result["y"], 0.001)
		})
	}
}