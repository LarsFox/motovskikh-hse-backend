package manager

import (
	"testing"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/stretchr/testify/assert"
)

func TestCalculatePercentile(t *testing.T) {
	m := &Manager{}
	
	tests := []struct {
		name       string
		bucket     *entities.TestBucket
		percentage float64
		expected   float64
	}{
		{
			name:       "nil bucket returns 50",
			bucket:     nil,
			percentage: 70,
			expected:   50,
		},
		{
			name: "empty bucket returns 50",
			bucket: &entities.TestBucket{
				ValidAttempts: 0,
			},
			percentage: 70,
			expected:   50,
		},
		{
			name: "perfect score returns 100",
			bucket: &entities.TestBucket{
				ValidAttempts: 100,
				Pct0_5:  1, Pct5_10: 2, Pct10_15: 2, Pct15_20: 3,
				Pct20_25: 3, Pct25_30: 4, Pct30_35: 4, Pct35_40: 5,
				Pct40_45: 5, Pct45_50: 6, Pct50_55: 6, Pct55_60: 7,
				Pct60_65: 7, Pct65_70: 8, Pct70_75: 8, Pct75_80: 7,
				Pct80_85: 6, Pct85_90: 5, Pct90_95: 4, Pct95_100: 2,
			},
			percentage: 100,
			expected:   94,
		},
		{
			name: "average score calculates correctly",
			bucket: &entities.TestBucket{
				ValidAttempts: 100,
				Pct0_5:  1, Pct5_10: 2, Pct10_15: 2, Pct15_20: 3,
				Pct20_25: 3, Pct25_30: 4, Pct30_35: 4, Pct35_40: 5,
				Pct40_45: 5, Pct45_50: 6, Pct50_55: 6, Pct55_60: 7,
				Pct60_65: 7, Pct65_70: 8, Pct70_75: 8, Pct75_80: 7,
				Pct80_85: 6, Pct85_90: 5, Pct90_95: 4, Pct95_100: 2,
			},
			percentage: 72,
			expected:   67,
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
	m := &Manager{}
	
	tests := []struct {
		name       string
		bucket     *entities.TestBucket
		timeSpent  int
		expected   float64
	}{
		{
			name: "fast time returns high percentile",
			bucket: &entities.TestBucket{
				ValidAttempts: 100,
				Time0_60:   30, Time60_120: 30, Time120_180: 20,
				Time180_240: 10, Time240_300: 5, Time300_360: 3,
				Time360_:    2,
			},
			timeSpent: 45,
			expected:  15,
		},
		{
			name: "slow time returns low percentile",
			bucket: &entities.TestBucket{
				ValidAttempts: 100,
				Time0_60:   30, Time60_120: 30, Time120_180: 20,
				Time180_240: 10, Time240_300: 5, Time300_360: 3,
				Time360_:    2,
			},
			timeSpent: 400,
			expected:  99,
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
	m := &Manager{}
	
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
	m := &Manager{}
	
	tests := []struct {
		name          string
		percentage    float64
		timePercentile float64
		expected      string
	}{
		{"expert", 85, 90, "expert"},
		{"precise but slow", 85, 30, "precise_but_slow"},
		{"fast but inaccurate", 60, 85, "fast_but_inaccurate"},
		{"solid", 65, 60, "solid"},
		{"needs practice", 40, 30, "needs_practice"},
		{"mixed", 85, 40, "precise_but_slow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.GetPerformanceQuadrant(tt.percentage, tt.timePercentile)
			assert.Equal(t, tt.expected, result["name"])
		})
	}
}