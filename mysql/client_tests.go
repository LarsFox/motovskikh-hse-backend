package mysql

import (
	"fmt"
	"log"
	"time"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetTest возвращает информацию о тесте.
func (c *Client) GetTest(testID string) (*entities.Test, error) {
	var test entities.Test
	err := c.db.Where("name = ?", testID).First(&test).Error
	if err == nil {
		return &test, nil
	}
	
	var id int
	fmt.Sscanf(testID, "%d", &id)
	if id > 0 {
		err = c.db.Where("id = ?", id).First(&test).Error
		if err == nil {
			return &test, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// CreateTestData создает тестовые данные (для разработки).
func (c *Client) CreateTestData() error {
	log.Println("Creating test data...")
	
	// Создаем тесты.
	tests := []entities.Test{
		{ID: 1, Type: 1, Name: "europe", 
		 I18n: entities.JSON{"ru": map[string]interface{}{"title": "Европа", "desc": "Тест по Европе"}},
		 Settings: entities.JSON{"modes": []string{"flags", "capitals"}}},
		{ID: 2, Type: 1, Name: "asia",
		 I18n: entities.JSON{"ru": map[string]interface{}{"title": "Азия", "desc": "Тест по Азии"}},
		 Settings: entities.JSON{"modes": []string{"flags", "capitals"}}},
	}
	
	for _, test := range tests {
		c.db.Save(&test)
	}
	
	// Создаем бакеты с тестовыми данными.
	for _, test := range tests {
		bucket := entities.TestBucket{
			ID:        uuid.New().String(),
			TestID:    test.Name,
			UpdatedAt: time.Now(),
			
			TotalAttempts: 100,
			ValidAttempts: 95,
			InvalidAttempts: 5,

			Pct0_5:   1, Pct5_10: 2, Pct10_15: 2, Pct15_20: 3,
			Pct20_25: 3, Pct25_30: 4, Pct30_35: 4, Pct35_40: 5,
			Pct40_45: 5, Pct45_50: 6, Pct50_55: 6, Pct55_60: 7,
			Pct60_65: 7, Pct65_70: 8, Pct70_75: 8, Pct75_80: 7,
			Pct80_85: 6, Pct85_90: 5, Pct90_95: 4, Pct95_100: 2,
			
			Time0_60:    5, Time60_120: 15, Time120_180: 30,
			Time180_240: 25, Time240_300: 10, Time300_360: 5,
			Time360_:    5,
			
			AvgPercentage: 65.5,
			AvgTimeSpent:  180.0,
			MinPercentage: 10,
			MaxPercentage: 98,
			MinTimeSpent:  45,
			MaxTimeSpent:  420,
		}
		c.db.Save(&bucket)
	}
	
	log.Println("Test data created")
	return nil
}