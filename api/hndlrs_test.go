package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDBForAPI struct {
	mock.Mock
}

func (m *MockDBForAPI) GetTest(testID string) (*entities.Test, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Test), args.Error(1)
}

func (m *MockDBForAPI) AddAttemptToBucket(testID, userHash string, percentage float64, timeSpent int, isValid bool) error {
	args := m.Called(testID, userHash, percentage, timeSpent, isValid)
	return args.Error(0)
}

func (m *MockDBForAPI) GetBucket(testID string) (*entities.TestBucket, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TestBucket), args.Error(1)
}

func (m *MockDBForAPI) GetTestStats(testID string) (*entities.TestStats, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TestStats), args.Error(1)
}

func (m *MockDBForAPI) CreateTestData() error {
	args := m.Called()
	return args.Error(0)
}

func TestHndlrSubmitTest(t *testing.T) {
	// Создаем мок БД.
	mockDB := new(MockDBForAPI)
	// Менеджер с замоканной БД.
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
		router:  mux.NewRouter(),
	}

	// Тестовый запрос.
	reqBody := SubmitTestRequest{
		TestName:   "europe",
		Percentage: 75.0,
		TimeSpent:  180,
	}
	body, _ := json.Marshal(reqBody)

	// Мокаем методы БД, которые вызовет реальный менеджер.
	mockDB.On("AddAttemptToBucket", "europe", mock.Anything, 75.0, 180, true).
		Return(nil)
	
	mockDB.On("GetBucket", "europe").Return(&entities.TestBucket{
		ValidAttempts: 100,
		Pct70_75:     10,
		Pct75_80:     8,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
		Time120_180:   20,
		Time180_240:   10,
	}, nil)

	// Создаем запрос.
	req := httptest.NewRequest("POST", "/tests/submit/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Hash", "test-user-123")
	
	w := httptest.NewRecorder()

	// Вызываем хендлер.
	apiMgr.hndlrSubmitTest(w, req)

	// Проверяем ответ.
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["ok"].(bool))
	
	result := response["result"].(map[string]interface{})
	assert.NotEmpty(t, result["attempt_id"])
	assert.True(t, result["submitted"].(bool))
	
	analysis, ok := result["analysis"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 75.0, analysis["percentage"])
	assert.Equal(t, "excellent", analysis["category"])
	
	mockDB.AssertExpectations(t)
}

func TestHndlrSubmitTest_InvalidJSON(t *testing.T) {
	// Создаем мок БД.
	mockDB := new(MockDBForAPI)
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
	}

	// Невалидный JSON.
	req := httptest.NewRequest("POST", "/tests/submit/", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiMgr.hndlrSubmitTest(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["ok"].(bool))
	assert.Contains(t, response["error"], "Invalid JSON")
	mockDB.AssertNotCalled(t, "AddAttemptToBucket")
}

func TestHndlrSubmitTest_MissingFields(t *testing.T) {
	mockDB := new(MockDBForAPI)
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
	}

	tests := []struct {
		name       string
		testName   string
		percentage float64
		timeSpent  int
	}{
		{"empty test name", "", 75, 180},
		{"negative percentage", "europe", -10, 180},
		{"percentage > 100", "europe", 150, 180},
		{"zero time", "europe", 75, 0},
		{"negative time", "europe", 75, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := SubmitTestRequest{
				TestName:   tt.testName,
				Percentage: tt.percentage,
				TimeSpent:  tt.timeSpent,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/tests/submit/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiMgr.hndlrSubmitTest(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
	
	mockDB.AssertNotCalled(t, "AddAttemptToBucket")
}

func TestHndlrSubmitTest_AnonymousUser(t *testing.T) {
	mockDB := new(MockDBForAPI)
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
		router:  mux.NewRouter(),
	}

	reqBody := SubmitTestRequest{
		TestName:   "europe",
		Percentage: 75.0,
		TimeSpent:  180,
	}
	body, _ := json.Marshal(reqBody)

	mockDB.On("AddAttemptToBucket", "europe", mock.MatchedBy(func(hash string) bool {
		return len(hash) > 10 && hash[:10] == "anonymous_"
	}), 75.0, 180, true).Return(nil)
	
	mockDB.On("GetBucket", "europe").Return(&entities.TestBucket{
		ValidAttempts: 100,
		Pct70_75:     10,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
		Time120_180:   20,
	}, nil)

	// Запрос без X-User-Hash.
	req := httptest.NewRequest("POST", "/tests/submit/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiMgr.hndlrSubmitTest(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}

func TestHndlrGetAnalysis(t *testing.T) {
	mockDB := new(MockDBForAPI)
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
		router:  mux.NewRouter(),
	}

	reqBody := GetAnalysisRequest{
		TestID:     "europe",
		Percentage: 75.0,
		TimeSpent:  180,
	}
	body, _ := json.Marshal(reqBody)

	// Мокаем получение бакета для расчета перцентиля.
	mockDB.On("GetBucket", "europe").Return(&entities.TestBucket{
		ValidAttempts: 100,
		Pct70_75:     10,
		Pct75_80:     8,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
	}, nil)

	req := httptest.NewRequest("POST", "/stats/analysis/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	apiMgr.hndlrGetAnalysis(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["ok"].(bool))
	
	result := response["result"].(map[string]interface{})
	assert.Contains(t, result, "percentile")
	assert.Contains(t, result, "average_percentage")
	
	mockDB.AssertExpectations(t)
}

func TestHndlrGetAnalysis_InvalidParams(t *testing.T) {
	mockDB := new(MockDBForAPI)
	realManager := manager.New(mockDB)
	
	apiMgr := &Manager{
		manager: realManager,
	}

	tests := []struct {
		name       string
		testID     string
		percentage float64
		timeSpent  int
	}{
		{"empty test id", "", 75, 180},
		{"negative percentage", "europe", -10, 180},
		{"percentage > 100", "europe", 150, 180},
		{"zero time", "europe", 75, 0},
		{"negative time", "europe", 75, -10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := GetAnalysisRequest{
				TestID:     tt.testID,
				Percentage: tt.percentage,
				TimeSpent:  tt.timeSpent,
			}
			body, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/stats/analysis/", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiMgr.hndlrGetAnalysis(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
	
	mockDB.AssertNotCalled(t, "GetBucket")
}
