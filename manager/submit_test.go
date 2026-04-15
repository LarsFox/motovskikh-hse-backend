package manager

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/mocks"
)

func TestSubmitTestResult_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 75.0
	timeSpent := int64(180)
	questionCount := int64(30)

	testStats := &entities.TestStats{
		TestName:      testName,
		Attempts:      100,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
		PercentDistrib: &entities.PercentDistribution{
			Buckets: map[float64]uint64{
				0:  10,
				20: 20,
				40: 30,
				60: 25,
				80: 15,
			},
		},
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
	}

	mockDB.EXPECT().
		GetOrCreateStats(testName, questionCount).
		Return(testStats, nil)

	mockDB.EXPECT().
		SaveStats(gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result["submitted"].(bool))

	analysis := result["analysis"].(map[string]any)
	assert.InDelta(t, 75.0, analysis["percentage"], 0.001)
	assert.Equal(t, int64(180), analysis["time_spent"])
}

func TestSubmitTestResult_InvalidAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 3.0
	timeSpent := int64(30)
	questionCount := int64(30)

	testStats := &entities.TestStats{
		TestName: testName,
		Attempts: 0,
	}

	mockDB.EXPECT().
		GetOrCreateStats(testName, questionCount).
		Return(testStats, nil)

	mockDB.EXPECT().
		SaveStats(gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	assert.NotNil(t, result)

	analysis := result["analysis"].(map[string]any)
	assert.Equal(t, false, analysis["is_valid"])
}
