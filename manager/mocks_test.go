package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/stretchr/testify/mock"
)

// MockDB - мок для базы данных.
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetTest(testID string) (*entities.Test, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Test), args.Error(1)
}

func (m *MockDB) AddAttemptToBucket(testID, userHash string, percentage float64, timeSpent int, isValid bool) error {
	args := m.Called(testID, userHash, percentage, timeSpent, isValid)
	return args.Error(0)
}

func (m *MockDB) GetBucket(testID string) (*entities.TestBucket, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TestBucket), args.Error(1)
}

func (m *MockDB) GetTestStats(testID string) (*entities.TestStats, error) {
	args := m.Called(testID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TestStats), args.Error(1)
}

func (m *MockDB) CreateTestData() error {
	args := m.Called()
	return args.Error(0)
}