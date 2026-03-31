package manager

import (
	"testing"
	
	"github.com/stretchr/testify/require"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

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
	timeSpent := 180
	questionCount := 30


	testBucket := &entities.TestBucket{
		TestID:        testName,
		Attempts:      100,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
		PercentDistrib: &entities.PercentDistribution{
			Buckets: []entities.PercentBucket{
				{Min: 0, Max: 20, Count: 10},
				{Min: 20, Max: 40, Count: 20},
				{Min: 40, Max: 60, Count: 30},
				{Min: 60, Max: 80, Count: 25},
				{Min: 80, Max: 100, Count: 15},
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
		GetOrCreateBucket(testName, questionCount).
		Return(testBucket, nil)

	mockDB.EXPECT().
		SaveBucket(gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result["submitted"].(bool))

	analysis := result["analysis"].(map[string]any)
	assert.InDelta(t, 75.0, analysis["percentage"], 0.001)
	assert.Equal(t, 180, analysis["time_spent"])
}

func TestSubmitTestResult_InvalidAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 3.0
	timeSpent := 30
	questionCount := 30


	testBucket := &entities.TestBucket{
		TestID:   testName,
		Attempts: 0,
	}


	mockDB.EXPECT().
		GetOrCreateBucket(testName, questionCount).
		Return(testBucket, nil)

	mockDB.EXPECT().
		SaveBucket(gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	assert.NotNil(t, result)

	analysis := result["analysis"].(map[string]any)
	assert.Equal(t, false, analysis["is_valid"])
}