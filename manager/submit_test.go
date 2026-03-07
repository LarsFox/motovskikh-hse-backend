package manager

import (
	"errors"
	"testing"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/stretchr/testify/assert"
)

func TestSubmitTestResult(t *testing.T) {
	mockDB := new(MockDB)
	manager := New(mockDB)

	// Тестовые.
	testName := "europe"
	userHash := "user123"
	percentage := 75.0
	timeSpent := 180

	// Мокаем успешное сохранение.
	mockDB.On("AddAttemptToBucket", testName, userHash, percentage, timeSpent, true).
		Return(nil)

	// Мокаем получение бакета.
	mockDB.On("GetBucket", testName).Return(&entities.TestBucket{
		ValidAttempts: 100,
		Pct70_75:     10,
		Pct75_80:     8,
		AvgPercentage: 65.0,
		AvgTimeSpent:  200,
	}, nil)

	// Вызываем тестируемый метод.
	attemptID, result, err := manager.SubmitTestResult(testName, userHash, percentage, timeSpent)

	// Проверяем результаты.
	assert.NoError(t, err)
	assert.NotEmpty(t, attemptID)
	assert.True(t, result["submitted"].(bool))
	
	analysis := result["analysis"].(map[string]interface{})
	assert.Equal(t, 75.0, analysis["percentage"])
	assert.Equal(t, "excellent", analysis["category"])
	
	// Проверяем что моки были вызваны.
	mockDB.AssertExpectations(t)
}

func TestSubmitTestResult_DBError(t *testing.T) {
	mockDB := new(MockDB)
	manager := New(mockDB)

	testName := "europe"
	userHash := "user123"
	percentage := 75.0
	timeSpent := 180

	// Мокаем ошибку сохранения.
	mockDB.On("AddAttemptToBucket", testName, userHash, percentage, timeSpent, true).
		Return(errors.New("db error"))

	// Вызываем тестируемый метод.
	attemptID, result, err := manager.SubmitTestResult(testName, userHash, percentage, timeSpent)

	// Проверяем что вернулась ошибка
	assert.Error(t, err)
	assert.Empty(t, attemptID)
	assert.Nil(t, result)
	
	mockDB.AssertExpectations(t)
}