package mysql

import (
	"time"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetBucket получает бакет по ID теста.
func (c *Client) GetBucket(testID string) (*entities.TestBucket, error) {
	var bucket entities.TestBucket
	err := c.db.Where("test_id = ?", testID).First(&bucket).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &bucket, err
}

// AddAttemptToBucket добавляет попытку в бакет.
func (c *Client) AddAttemptToBucket(testID, userHash string, percentage float64, timeSpent int, isValid bool) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		// Получаем или создаем бакет для теста.
		var bucket entities.TestBucket
		err := tx.Where("test_id = ?", testID).First(&bucket).Error

		if err == gorm.ErrRecordNotFound {
			// Создаем новый бакет.
			bucket = entities.TestBucket{
				ID:        uuid.New().String(),
				TestID:    testID,
				UpdatedAt: time.Now(),
			}
			if err := tx.Create(&bucket).Error; err != nil {
				return err
			}
			// Получаем созданный бакет.
			tx.Where("test_id = ?", testID).First(&bucket)
		} else if err != nil {
			return err
		}

		// Обновляем счетчики.
		bucket.TotalAttempts++
		if isValid {
			bucket.ValidAttempts++

			// Распределение процентов.
			switch {
			case percentage < 5:
				bucket.Pct0_5++
			case percentage < 10:
				bucket.Pct5_10++
			case percentage < 15:
				bucket.Pct10_15++
			case percentage < 20:
				bucket.Pct15_20++
			case percentage < 25:
				bucket.Pct20_25++
			case percentage < 30:
				bucket.Pct25_30++
			case percentage < 35:
				bucket.Pct30_35++
			case percentage < 40:
				bucket.Pct35_40++
			case percentage < 45:
				bucket.Pct40_45++
			case percentage < 50:
				bucket.Pct45_50++
			case percentage < 55:
				bucket.Pct50_55++
			case percentage < 60:
				bucket.Pct55_60++
			case percentage < 65:
				bucket.Pct60_65++
			case percentage < 70:
				bucket.Pct65_70++
			case percentage < 75:
				bucket.Pct70_75++
			case percentage < 80:
				bucket.Pct75_80++
			case percentage < 85:
				bucket.Pct80_85++
			case percentage < 90:
				bucket.Pct85_90++
			case percentage < 95:
				bucket.Pct90_95++
			default:
				bucket.Pct95_100++
			}

			// Распределение времени.
			switch {
			case timeSpent < 60:
				bucket.Time0_60++
			case timeSpent < 120:
				bucket.Time60_120++
			case timeSpent < 180:
				bucket.Time120_180++
			case timeSpent < 240:
				bucket.Time180_240++
			case timeSpent < 300:
				bucket.Time240_300++
			case timeSpent < 360:
				bucket.Time300_360++
			default:
				bucket.Time360_++
			}

			// Обновляем средние значения.
			oldTotal := float64(bucket.ValidAttempts - 1)
			if bucket.ValidAttempts == 1 {
				bucket.AvgPercentage = percentage
				bucket.AvgTimeSpent = float64(timeSpent)
				bucket.MinPercentage = percentage
				bucket.MaxPercentage = percentage
				bucket.MinTimeSpent = timeSpent
				bucket.MaxTimeSpent = timeSpent
			} else {
				bucket.AvgPercentage = (bucket.AvgPercentage*oldTotal + percentage) / float64(bucket.ValidAttempts)
				bucket.AvgTimeSpent = (bucket.AvgTimeSpent*oldTotal + float64(timeSpent)) / float64(bucket.ValidAttempts)

				if percentage < bucket.MinPercentage {
					bucket.MinPercentage = percentage
				}
				if percentage > bucket.MaxPercentage {
					bucket.MaxPercentage = percentage
				}
				if timeSpent < bucket.MinTimeSpent {
					bucket.MinTimeSpent = timeSpent
				}
				if timeSpent > bucket.MaxTimeSpent {
					bucket.MaxTimeSpent = timeSpent
				}
			}
		} else {
			bucket.InvalidAttempts++
		}

		bucket.UpdatedAt = time.Now()
		return tx.Save(&bucket).Error
	})
}

// GetTestStats получает статистику теста из бакета.
func (c *Client) GetTestStats(testID string) (*entities.TestStats, error) {
	bucket, err := c.GetBucket(testID)
	if err != nil || bucket == nil {
		return &entities.TestStats{
			TestID:        testID,
			Period:        "total",
			TotalAttempts: 0,
			ValidAttempts: 0,
			AvgPercentage: 0,
			AvgTimeSpent:  0,
			UpdatedAt:     time.Now(),
		}, err
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
	}, nil
}