package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

const testEuropeID = "europe"

func setupTestDB(t *testing.T) *Client {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&dbTestStats{})
	require.NoError(t, err)

	return &Client{db: db}
}

func cleanupTestStats(t *testing.T, client *Client, testName string) {
	t.Helper()
	err := client.db.Exec("DELETE FROM test_stats WHERE test_name = ?", testName).Error
	require.NoError(t, err)
}

func TestGetOrCreateStats_CreateNew(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	stats, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	assert.NotNil(t, stats)
	assert.Equal(t, testName, stats.TestName)
	assert.Equal(t, uint64(0), stats.Attempts)
	assert.NotNil(t, stats.PercentDistrib)
	assert.NotNil(t, stats.TimeDistrib)

	assert.Len(t, stats.PercentDistrib.Buckets, 21)
}

func TestGetOrCreateStats_GetExisting(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	// Создаем первую статистику
	stats1, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	// Обновляем данные
	stats1.Attempts = 5
	err = client.SaveStats(stats1)
	require.NoError(t, err)

	// Получаем существующую статистику
	stats2, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	assert.Equal(t, uint64(5), stats2.Attempts)
	assert.Equal(t, testName, stats2.TestName)
}

func TestGetStats_NotFound(t *testing.T) {
	client := setupTestDB(t)
	testName := "nonexistent"

	stats, err := client.GetStats(testName)
	require.Error(t, err)
	assert.Equal(t, entities.ErrNotFound, err)
	assert.Nil(t, stats)
}

func TestSaveStats_Update(t *testing.T) {
	client := setupTestDB(t)
	testName := testEuropeID
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats, err := client.GetOrCreateStats(testName, 15)
	require.NoError(t, err)

	// Обновляем значения
	stats.Attempts = 10
	stats.AvgPercentage = 75.5
	stats.AvgTimeSpent = 180.0

	// Сохраняем
	err = client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	saved, err := client.GetStats(testName)
	require.NoError(t, err)

	assert.Equal(t, uint64(10), saved.Attempts)
	assert.InDelta(t, 75.5, saved.AvgPercentage, 0.001)
	assert.InDelta(t, 180.0, saved.AvgTimeSpent, 0.001)
}

func TestStatsWithDistributions(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-dist"
	defer cleanupTestStats(t, client, testName)

	stats, err := client.GetOrCreateStats(testName, 10)
	require.NoError(t, err)

	// Проверяем процентные бакеты
	assert.NotNil(t, stats.PercentDistrib)
	assert.Len(t, stats.PercentDistrib.Buckets, 21)

	// Проверяем наличие ключей 0,5,10,...,95,100
	for i := range 21 {
		key := float64(i * 5)
		_, ok := stats.PercentDistrib.Buckets[key]
		assert.True(t, ok, "missing key %f", key)
	}

	// Проверяем временные бакеты
	assert.NotNil(t, stats.TimeDistrib)
	assert.NotEmpty(t, stats.TimeDistrib.Buckets)
}

func TestSaveAndRetrieveStats(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-save"
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats := &entities.TestStats{
		TestName:      testName,
		Attempts:      100,
		AvgPercentage: 68.5,
		AvgTimeSpent:  150.5,
		MinPercentage: 20.0,
		MaxPercentage: 95.0,
		MinTimeSpent:  60,
		MaxTimeSpent:  300,
	}
	stats.InitPercentBuckets()
	stats.InitTimeBuckets(15)

	// Сохраняем
	err := client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	retrieved, err := client.GetStats(testName)
	require.NoError(t, err)

	assert.Equal(t, stats.TestName, retrieved.TestName)
	assert.Equal(t, stats.Attempts, retrieved.Attempts)
	assert.InDelta(t, stats.AvgPercentage, retrieved.AvgPercentage, 0.001)
	assert.InDelta(t, stats.AvgTimeSpent, retrieved.AvgTimeSpent, 0.001)
	assert.InDelta(t, stats.MinPercentage, retrieved.MinPercentage, 0.001)
	assert.InDelta(t, stats.MaxPercentage, retrieved.MaxPercentage, 0.001)
	assert.Equal(t, stats.MinTimeSpent, retrieved.MinTimeSpent)
	assert.Equal(t, stats.MaxTimeSpent, retrieved.MaxTimeSpent)
}

func TestMultipleDifferentStats(t *testing.T) {
	client := setupTestDB(t)

	testNames := []string{"test1", "test2", "test3"}

	// Создаем статистику
	for _, id := range testNames {
		stats, err := client.GetOrCreateStats(id, 20)
		require.NoError(t, err)
		assert.NotNil(t, stats)
	}

	// Проверяем, что все создались
	for _, id := range testNames {
		stats, err := client.GetStats(id)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, id, stats.TestName)

		// Очищаем
		cleanupTestStats(t, client, id)
	}
}

func TestStatsAttemptsIncrement(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-inc"
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats, err := client.GetOrCreateStats(testName, 10)
	require.NoError(t, err)

	initialAttempts := stats.Attempts

	// Увеличиваем попытки
	stats.Attempts++
	err = client.SaveStats(stats)
	require.NoError(t, err)

	// Получаем и проверяем
	updated, err := client.GetStats(testName)
	require.NoError(t, err)
	assert.Equal(t, initialAttempts+1, updated.Attempts)
}

func TestPercentDistributionSaveLoad(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-percent"
	defer cleanupTestStats(t, client, testName)

	// Создаем статистику
	stats, err := client.GetOrCreateStats(testName, 10)
	require.NoError(t, err)

	// Обновляем несколько бакетов
	stats.UpdatePercentDistribution(73.5)
	stats.UpdatePercentDistribution(68.0)
	stats.UpdatePercentDistribution(82.0)

	err = client.SaveStats(stats)
	require.NoError(t, err)

	// Загружаем и проверяем
	loaded, err := client.GetStats(testName)
	require.NoError(t, err)

	// Проверяем счетчики
	assert.Equal(t, uint64(1), loaded.PercentDistrib.Buckets[70.0])
	assert.Equal(t, uint64(1), loaded.PercentDistrib.Buckets[65.0])
	assert.Equal(t, uint64(1), loaded.PercentDistrib.Buckets[80.0])
}
