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
	timeSpent := float64(180)
	questionCount := int64(30)

	stats := &entities.TestStats{
	Name:          testName,
	Attempts:      23,
	AvgPercentage: 71.0,
	AvgTimeSpent:  130.0,
	PercentBuckets: []*entities.TestStatsBucket{
		{Value: 70, Count: 10},
		{Value: 75, Count: 8},
		{Value: 80, Count: 5},
	},
	TimeBuckets: []*entities.TestStatsBucket{
		{Value: 60, Count: 5},
		{Value: 120, Count: 10},
		{Value: 180, Count: 5},
		{Value: 240, Count: 3},
	},
}

	mockDB.EXPECT().
		GetStats(gomock.Any(), testName).
		Return(stats, nil)

	mockDB.EXPECT().
		SaveStats(gomock.Any(), gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.InDelta(t, 43.5, result.ScorePercentile, 1)
	assert.Equal(t, int64(8), result.FasterThan)
	assert.NotZero(t, result.AveragePercentage)
	assert.NotZero(t, result.AverageTime)
}

func TestSubmitTestResult_InvalidAttempt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"
	percentage := 3.0
	timeSpent := float64(30)
	questionCount := int64(30)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSubmitTestResult_CreateNewStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "new-test"
	percentage := 60.0
	timeSpent := float64(150)
	questionCount := int64(20)

	mockDB.EXPECT().
		GetStats(gomock.Any(), testName).
		Return(nil, entities.ErrNotFound)

	mockDB.EXPECT().
		SaveStats(gomock.Any(), gomock.Any()).
		Return(nil)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotZero(t, result.AveragePercentage)
	assert.NotZero(t, result.AverageTime)
}

func TestSubmitTestResult_InvalidAttempt_NewTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "new-test"
	percentage := 1.0
	timeSpent := float64(10)
	questionCount := int64(20)

	result, err := mgr.SubmitTestResult(testName, percentage, timeSpent, questionCount)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestSubmitTestResult_AttemptsIncrement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	mgr := New(mockDB)

	testName := "europe"

	stats := &entities.TestStats{
		Name:     testName,
		Attempts: 5,
		PercentBuckets: []*entities.TestStatsBucket{
			{Value: 50, Count: 1},
		},
		TimeBuckets: []*entities.TestStatsBucket{
			{Value: 100, Count: 1},
		},
	}

	mockDB.EXPECT().
		GetStats(gomock.Any(), testName).
		Return(stats, nil)

	mockDB.EXPECT().
		SaveStats(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ interface{}, s *entities.TestStats) error {
			assert.Equal(t, int64(6), s.Attempts)
			return nil
		})

	_, err := mgr.SubmitTestResult(testName, 60, 120, 10)
	require.NoError(t, err)
}