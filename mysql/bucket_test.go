package mysql

import (
	"testing"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testEuropeID = "europe"

func setupTestDB(t *testing.T) *Client {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.TestBucket{})
	require.NoError(t, err)

	return &Client{db: db}
}

func cleanupTestBucket(t *testing.T, client *Client, testID string) {
	t.Helper()
	err := client.db.Exec("DELETE FROM test_buckets WHERE test_id = ?", testID).Error
	require.NoError(t, err)
}

func TestGetOrCreateBucket_CreateNewBucket(t *testing.T) {
	client := setupTestDB(t)
	testID := testEuropeID
	defer cleanupTestBucket(t, client, testID)

	bucket, err := client.GetOrCreateBucket(testID, 15)
	require.NoError(t, err)
	
	assert.NotNil(t, bucket)
	assert.Equal(t, testID, bucket.TestID)
	assert.Equal(t, uint64(0), bucket.Attempts)
	assert.NotNil(t, bucket.PercentDistrib)
	assert.NotNil(t, bucket.TimeDistrib)
}

func TestGetOrCreateBucket_GetExisting(t *testing.T) {
	client := setupTestDB(t)
	testID := testEuropeID
	defer cleanupTestBucket(t, client, testID)

	// Создаем первый бакет
	bucket1, err := client.GetOrCreateBucket(testID, 15)
	require.NoError(t, err)
	
	// Обновляем данные
	bucket1.Attempts = 5
	err = client.SaveBucket(bucket1)
	require.NoError(t, err)

	// Получаем существующий бакет
	bucket2, err := client.GetOrCreateBucket(testID, 15)
	require.NoError(t, err)
	
	assert.Equal(t, uint64(5), bucket2.Attempts)
	assert.Equal(t, testID, bucket2.TestID)
}

func TestGetBucket_NotFound(t *testing.T) {
	client := setupTestDB(t)
	testID := "nonexistent"
	defer cleanupTestBucket(t, client, testID)

	bucket, err := client.GetBucket(testID)
	require.Error(t, err)
	assert.Equal(t, entities.ErrNotFound, err)
	assert.Nil(t, bucket)
}

func TestSaveBucket_Update(t *testing.T) {
	client := setupTestDB(t)
	testID := testEuropeID
	defer cleanupTestBucket(t, client, testID)

	// Создаем бакет
	bucket, err := client.GetOrCreateBucket(testID, 15)
	require.NoError(t, err)

	// Обновляем значения
	bucket.Attempts = 10
	bucket.AvgPercentage = 75.5
	bucket.AvgTimeSpent = 180.0

	// Сохраняем
	err = client.SaveBucket(bucket)
	require.NoError(t, err)

	// Получаем и проверяем
	saved, err := client.GetBucket(testID)
	require.NoError(t, err)
	
	assert.Equal(t, uint64(10), saved.Attempts)
	assert.InDelta(t, 75.5, saved.AvgPercentage, 0.001)
	assert.InDelta(t, 180.0, saved.AvgTimeSpent, 0.001)
}

func TestCreateBucket(t *testing.T) {
	client := setupTestDB(t)
	testID := "asia"
	defer cleanupTestBucket(t, client, testID)

	bucket := &entities.TestBucket{
		TestID: testID,
	}
	bucket.InitializeBuckets(20)

	err := client.CreateBucket(bucket)
	require.NoError(t, err)

	// Проверяем, что создался
	saved, err := client.GetBucket(testID)
	require.NoError(t, err)
	
	assert.Equal(t, testID, saved.TestID)
	assert.Equal(t, uint64(0), saved.Attempts)
}

func TestBucketWithDistributions(t *testing.T) {
	client := setupTestDB(t)
	testID := "test-dist"
	defer cleanupTestBucket(t, client, testID)

	bucket, err := client.GetOrCreateBucket(testID, 10)
	require.NoError(t, err)
	
	// Проверяем процентные бакеты
	assert.NotNil(t, bucket.PercentDistrib)
	assert.Len(t, bucket.PercentDistrib.Buckets, 20)
	
	// Проверяем временные бакеты
	assert.NotNil(t, bucket.TimeDistrib)
	assert.NotEmpty(t, bucket.TimeDistrib.Buckets)
	
	// Проверяем первый процентный бакет
	firstPercentBucket := bucket.PercentDistrib.Buckets[0]
	assert.InDelta(t, 0.0, firstPercentBucket.Min, 0.001)
	assert.InDelta(t, 5.0, firstPercentBucket.Max, 0.001)
	assert.Equal(t, "0-5%", firstPercentBucket.Label)
}

func TestSaveAndRetrieveBucket(t *testing.T) {
	client := setupTestDB(t)
	testID := "test-save"
	defer cleanupTestBucket(t, client, testID)

	// Создаем бакет
	bucket := &entities.TestBucket{
		TestID:        testID,
		Attempts:      100,
		AvgPercentage: 68.5,
		AvgTimeSpent:  150.5,
		MinPercentage: 20.0,
		MaxPercentage: 95.0,
		MinTimeSpent:  60,
		MaxTimeSpent:  300,
	}
	bucket.InitializeBuckets(15)

	// Сохраняем
	err := client.CreateBucket(bucket)
	require.NoError(t, err)

	// Получаем и проверяем
	retrieved, err := client.GetBucket(testID)
	require.NoError(t, err)
	
	assert.Equal(t, bucket.TestID, retrieved.TestID)
	assert.Equal(t, bucket.Attempts, retrieved.Attempts)
	assert.InDelta(t, bucket.AvgPercentage, retrieved.AvgPercentage, 0.001)
	assert.InDelta(t, bucket.AvgTimeSpent, retrieved.AvgTimeSpent, 0.001)
	assert.InDelta(t, bucket.MinPercentage, retrieved.MinPercentage, 0.001)
	assert.InDelta(t, bucket.MaxPercentage, retrieved.MaxPercentage, 0.001)
	assert.Equal(t, bucket.MinTimeSpent, retrieved.MinTimeSpent)
	assert.Equal(t, bucket.MaxTimeSpent, retrieved.MaxTimeSpent)
}

func TestMultipleDifferentBuckets(t *testing.T) {
	client := setupTestDB(t)
	
	testIDs := []string{"test1", "test2", "test3"}
	
	// Создаем бакеты
	for _, id := range testIDs {
		bucket, err := client.GetOrCreateBucket(id, 20)
		require.NoError(t, err)
		assert.NotNil(t, bucket)
	}
	
	// Проверяем, что все создались
	for _, id := range testIDs {
		bucket, err := client.GetBucket(id)
		require.NoError(t, err)
		assert.NotNil(t, bucket)
		assert.Equal(t, id, bucket.TestID)
		
		// Очищаем
		cleanupTestBucket(t, client, id)
	}
}

func TestBucketAttemptsIncrement(t *testing.T) {
	client := setupTestDB(t)
	testID := "test-inc"
	defer cleanupTestBucket(t, client, testID)

	// Создаем бакет
	bucket, err := client.GetOrCreateBucket(testID, 10)
	require.NoError(t, err)
	
	initialAttempts := bucket.Attempts
	
	// Увеличиваем попытки
	bucket.Attempts++
	err = client.SaveBucket(bucket)
	require.NoError(t, err)
	
	// Получаем и проверяем
	updated, err := client.GetBucket(testID)
	require.NoError(t, err)
	assert.Equal(t, initialAttempts+1, updated.Attempts)
}