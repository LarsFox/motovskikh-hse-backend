package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/mocks"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
)

func TestHndlrSubmitTest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockdb(ctrl)
	realManager := manager.New(mockDB)

	apiMgr := &Manager{
		manager: realManager,
		router:  mux.NewRouter(),
	}
	testName := "europe"
	percentage := 75.0
	timeSpent := float64(180)
	questionCount := int64(30)

	reqBody := models.SubmitTestRequest{
		TestName:      &testName,
		Percentage:    &percentage,
		TimeSpent:     &timeSpent,
		QuestionCount: &questionCount,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	stats := &entities.TestStats{
		Name:          "europe",
		Attempts:      100,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200.0,
		MinTimeSpent:  60.0,
		MaxTimeSpent:  400.0,
		PercentBuckets: []*entities.TestStatsBucket{
			{Value: 70, Count: 10},
			{Value: 75, Count: 8},
		},
		TimeBuckets: []*entities.TestStatsBucket{
			{Value: 60, Count: 30},
			{Value: 120, Count: 30},
			{Value: 180, Count: 20},
			{Value: 240, Count: 10},
			{Value: 300, Count: 5},
		},
	}

	statsCopy := *stats
	mockDB.EXPECT().
		GetStats(gomock.Any(), "europe").
		Return(&statsCopy, nil)

	// Expect SaveStats call
	mockDB.EXPECT().
		SaveStats(gomock.Any(), gomock.Any()).
		Return(nil)

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/tests/submit/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	apiMgr.hndlrSubmitTest(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response["ok"].(bool))

	result := response["result"].(map[string]any)

	assert.Contains(t, result, "score_percentile")
	assert.Contains(t, result, "time_percentile")
	assert.Contains(t, result, "better_than")
	assert.Contains(t, result, "faster_than")
	assert.Contains(t, result, "average_percentage")
	assert.Contains(t, result, "average_time")
	assert.Contains(t, result, "vs_average")
	
	if vsAverage, ok := result["vs_average"].(map[string]any); ok {
		assert.Contains(t, vsAverage, "percentage_diff")
		assert.Contains(t, vsAverage, "time_diff")
	}
}