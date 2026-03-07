package mysql

import (
	"testing"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB создает тестовую БД в памяти.
func setupTestDB(t *testing.T) *Client {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	
	err = db.AutoMigrate(&entities.TestBucket{})
	require.NoError(t, err)
	
	return &Client{db: db}
}

// cleanupDB очищает таблицу между тестами.
func cleanupDB(t *testing.T, client *Client) {
	err := client.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entities.TestBucket{}).Error
	require.NoError(t, err)
}

func TestAddAttemptToBucket_CreateNewBucket(t *testing.T) {
	client := setupTestDB(t)
	cleanupDB(t, client)
	
	// Добавляем первую попытку.
	err := client.AddAttemptToBucket("europe", "user1", 75.0, 180, true)
	assert.NoError(t, err)
	
	// Проверяем что бакет создался.
	bucket, err := client.GetBucket("europe")
	assert.NoError(t, err)
	assert.NotNil(t, bucket)
	assert.Equal(t, uint64(1), bucket.TotalAttempts)
	assert.Equal(t, uint64(1), bucket.ValidAttempts)
	assert.Equal(t, 75.0, bucket.AvgPercentage)
	assert.Equal(t, 180.0, bucket.AvgTimeSpent)
}

func TestAddAttemptToBucket_UpdateExisting(t *testing.T) {
	client := setupTestDB(t)
	cleanupDB(t, client)
	
	// Добавляем первую попытку.
	client.AddAttemptToBucket("europe", "user1", 70.0, 180, true)
	
	// Добавляем вторую попытку.
	err := client.AddAttemptToBucket("europe", "user2", 80.0, 150, true)
	assert.NoError(t, err)
	
	// Проверяем обновленные значения.
	bucket, _ := client.GetBucket("europe")
	assert.Equal(t, uint64(2), bucket.TotalAttempts)
	assert.Equal(t, uint64(2), bucket.ValidAttempts)
	assert.InDelta(t, 75.0, bucket.AvgPercentage, 0.1)
	assert.InDelta(t, 165.0, bucket.AvgTimeSpent, 0.1)
}

func TestAddAttemptToBucket_InvalidAttempt(t *testing.T) {
	client := setupTestDB(t)
	cleanupDB(t, client)
	
	// Добавляем невалидную попытку.
	err := client.AddAttemptToBucket("europe", "user1", 3.0, 30, false)
	assert.NoError(t, err)
	
	bucket, _ := client.GetBucket("europe")
	assert.Equal(t, uint64(1), bucket.TotalAttempts)
	assert.Equal(t, uint64(0), bucket.ValidAttempts)
	assert.Equal(t, uint64(1), bucket.InvalidAttempts)
}

func TestGetBucket_NotFound(t *testing.T) {
	client := setupTestDB(t)
	cleanupDB(t, client)
	
	bucket, err := client.GetBucket("nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, bucket)
}

func TestGetTestStats_EmptyBucket(t *testing.T) {
	client := setupTestDB(t)
	cleanupDB(t, client)
	
	stats, err := client.GetTestStats("europe")
	assert.NoError(t, err)
	assert.Equal(t, 0, stats.ValidAttempts)
	assert.Equal(t, 0.0, stats.AvgPercentage)
}