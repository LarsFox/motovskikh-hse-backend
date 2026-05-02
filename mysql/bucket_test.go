package mysql

import (
	"testing"
	"context"
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

	err = db.AutoMigrate(&testStats{})
	require.NoError(t, err)

	return &Client{db: db}
}

func cleanupTestStats(t *testing.T, client *Client, testName string) {
	t.Helper()
	err := client.db.Exec("DELETE FROM test_stats WHERE name = ?", testName).Error
	require.NoError(t, err)
}

func TestGetStats_NotFound(t *testing.T) {
	client := setupTestDB(t)

	stats, err := client.GetStats(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Equal(t, entities.ErrNotFound, err)
	assert.Nil(t, stats)
}

func TestUpdateStats(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-update"
	defer cleanupTestStats(t, client, testName)

	stats := &entities.TestStats{
		Name:     testName,
		Attempts: 1,
	}

	err := client.SaveStats(context.Background(), stats)
	require.NoError(t, err)

	stats.Attempts = 10
	stats.AvgPercentage = 80

	err = client.SaveStats(context.Background(), stats)
	require.NoError(t, err)

	loaded, err := client.GetStats(context.Background(), testName)
	require.NoError(t, err)

	assert.Equal(t, int64(10), loaded.Attempts)
	assert.InDelta(t, 80, loaded.AvgPercentage, 0.001)
}

func TestSaveAndGetStats(t *testing.T) {
	client := setupTestDB(t)
	testName := "test-save"
	defer cleanupTestStats(t, client, testName)

	stats := &entities.TestStats{
		Name:          testName,
		Attempts:      5,
		AvgPercentage: 70.5,
		AvgTimeSpent:  150.0,
		MinTimeSpent:  60,
		MaxTimeSpent:  300,
		PercentBuckets: []*entities.TestStatsBucket{
			{Value: 50, Count: 2},
			{Value: 75, Count: 3},
		},
		TimeBuckets: []*entities.TestStatsBucket{
			{Value: 100, Count: 2},
			{Value: 200, Count: 3},
		},
	}

	err := client.SaveStats(context.Background(), stats)
	require.NoError(t, err)

	loaded, err := client.GetStats(context.Background(), testName)
	require.NoError(t, err)

	assert.Equal(t, stats.Name, loaded.Name)
	assert.Equal(t, stats.Attempts, loaded.Attempts)
	assert.InDelta(t, stats.AvgPercentage, loaded.AvgPercentage, 0.001)
	assert.InDelta(t, stats.AvgTimeSpent, loaded.AvgTimeSpent, 0.001)

	assert.Len(t, loaded.PercentBuckets, 2)
	assert.Len(t, loaded.TimeBuckets, 2)
}

func TestSaveStats_EmptyBuckets(t *testing.T) {
	client := setupTestDB(t)
	testName := "empty-buckets"
	defer cleanupTestStats(t, client, testName)

	stats := &entities.TestStats{
		Name:           testName,
		Attempts:       1,
		PercentBuckets: []*entities.TestStatsBucket{},
		TimeBuckets:    []*entities.TestStatsBucket{},
	}

	err := client.SaveStats(context.Background(), stats)
	require.NoError(t, err)

	loaded, err := client.GetStats(context.Background(), testName)
	require.NoError(t, err)

	assert.NotNil(t, loaded.PercentBuckets)
	assert.NotNil(t, loaded.TimeBuckets)
	assert.Len(t, loaded.PercentBuckets, 0)
	assert.Len(t, loaded.TimeBuckets, 0)
}

func TestBuckets_PersistCorrectly(t *testing.T) {
	client := setupTestDB(t)
	testName := "bucket-persist"
	defer cleanupTestStats(t, client, testName)

	stats := &entities.TestStats{
		Name: testName,
		PercentBuckets: []*entities.TestStatsBucket{
			{Value: 10, Count: 1},
			{Value: 20, Count: 2},
		},
		TimeBuckets: []*entities.TestStatsBucket{
			{Value: 100, Count: 3},
		},
	}

	err := client.SaveStats(context.Background(), stats)
	require.NoError(t, err)

	loaded, err := client.GetStats(context.Background(), testName)
	require.NoError(t, err)

	assert.Equal(t, float64(10), loaded.PercentBuckets[0].Value)
	assert.Equal(t, int64(1), loaded.PercentBuckets[0].Count)

	assert.Equal(t, float64(100), loaded.TimeBuckets[0].Value)
	assert.Equal(t, int64(3), loaded.TimeBuckets[0].Count)
}

func TestBuckets_Overwrite(t *testing.T) {
	client := setupTestDB(t)
	testName := "bucket-overwrite"
	defer cleanupTestStats(t, client, testName)

	stats := &entities.TestStats{
		Name: testName,
		PercentBuckets: []*entities.TestStatsBucket{
			{Value: 50, Count: 5},
		},
	}

	require.NoError(t, client.SaveStats(context.Background(), stats))

	stats.PercentBuckets = []*entities.TestStatsBucket{
		{Value: 90, Count: 1},
	}

	require.NoError(t, client.SaveStats(context.Background(), stats))

	loaded, err := client.GetStats(context.Background(), testName)
	require.NoError(t, err)

	assert.Len(t, loaded.PercentBuckets, 1)
	assert.Equal(t, float64(90), loaded.PercentBuckets[0].Value)
}

func TestMultipleStatsIsolation(t *testing.T) {
	client := setupTestDB(t)

	tests := []string{"t1", "t2", "t3"}

	for _, name := range tests {
		stats := &entities.TestStats{Name: name, Attempts: 1}
		require.NoError(t, client.SaveStats(context.Background(), stats))
	}

	for _, name := range tests {
		stats, err := client.GetStats(context.Background(), name)
		require.NoError(t, err)
		assert.Equal(t, name, stats.Name)

		cleanupTestStats(t, client, name)
	}
}
