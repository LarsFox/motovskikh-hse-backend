package mysql

import (
	"errors"
	"fmt"
	
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"gorm.io/gorm"
)

// GetOrCreateBucket получает бакет или создает новый с дефолтными значениями.
func (c *Client) GetOrCreateBucket(testID string, questionCount int) (*entities.TestBucket, error) {
	bucket, err := c.GetBucket(testID)
	if err != nil && !errors.Is(err, entities.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	if bucket != nil {
		return bucket, nil
	}
	// Создаем новый бакет, если не нашли.
	newBucket := &entities.TestBucket{
		TestID: testID,
	}
	newBucket.InitializeBuckets(questionCount)
	
	if err := c.CreateBucket(newBucket); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}
	// Получаем созданный бакет, чтобы убедиться, что всё сохранилось корректно.
	createdBucket, err := c.GetBucket(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created bucket: %w", err)
	}
	
	return createdBucket, nil
}

// GetBucket получает бакет по ID теста.
func (c *Client) GetBucket(testID string) (*entities.TestBucket, error) {
	var bucket entities.TestBucket
	err := c.db.Where("test_id = ?", testID).First(&bucket).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entities.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}
	return &bucket, nil
}

// SaveBucket сохраняет бакет.
func (c *Client) SaveBucket(bucket *entities.TestBucket) error {
	return c.db.Save(bucket).Error
}

// CreateBucket создает новый бакет.
func (c *Client) CreateBucket(bucket *entities.TestBucket) error {
	return c.db.Create(bucket).Error
}
