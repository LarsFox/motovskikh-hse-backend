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
	timeSpent := int64(180)
	questionCount := int64(30)

	reqBody := models.SubmitTestRequest{
		TestName:      &testName,
		Percentage:    &percentage,
		TimeSpent:     &timeSpent,
		QuestionCount: &questionCount,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	// Ожидаем вызов GetOrCreateStats.
	mockDB.EXPECT().
		GetOrCreateStats("europe", int64(30)).
		Return(&entities.TestStats{
			TestName:      "europe",
			Attempts:      100,
			AvgPercentage: 65.0,
			AvgTimeSpent:  200,
			PercentDistrib: &entities.PercentDistribution{
				Buckets: map[float64]uint64{
					70: 10,
					75: 8,
				},
			},
			TimeDistrib: &entities.TimeDistribution{
				Buckets: []entities.TimeBucket{
					{MinSeconds: 120, MaxSeconds: 180, Count: 20},
					{MinSeconds: 180, MaxSeconds: 240, Count: 10},
				},
			},
		}, nil)

	// Ожидаем вызов SaveStats.
	mockDB.EXPECT().
		SaveStats(gomock.Any()).
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
	analysis, ok := result["analysis"].(map[string]any)
	assert.True(t, ok)
	assert.InDelta(t, 75.0, analysis["percentage"], 0.001)
}
