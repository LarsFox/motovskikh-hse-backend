package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

type db interface {
	GetTest(testID string) (*entities.Test, error)
	AddAttemptToBucket(testID, userHash string, percentage float64, timeSpent int, isValid bool) error
	GetBucket(testID string) (*entities.TestBucket, error)
	GetTestStats(testID string) (*entities.TestStats, error)
	CreateTestData() error
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{db: db}
}

func (m *Manager) GetTest(testID string) (*entities.Test, error) {
	return m.db.GetTest(testID)
}

func (m *Manager) CreateTestData() error {
	return m.db.CreateTestData()
}

// Вспомогательная функция для конвертации.
func convertBucketToStats(bucket *entities.TestBucket) *entities.TestStats {
	if bucket == nil {
		return nil
	}
	return &entities.TestStats{
		ID:            bucket.ID,
		TestID:        bucket.TestID,
		Period:        "total",
		TotalAttempts: int(bucket.TotalAttempts),
		ValidAttempts: int(bucket.ValidAttempts),
		AvgPercentage: bucket.AvgPercentage,
		AvgTimeSpent:  bucket.AvgTimeSpent,
		UpdatedAt:     bucket.UpdatedAt,
	}
}

func createEmptyStats(testID string) *entities.TestStats {
	return &entities.TestStats{
		TestID:        testID,
		Period:        "total",
		TotalAttempts: 0,
		ValidAttempts: 0,
		AvgPercentage: 0,
		AvgTimeSpent:  0,
	}
}