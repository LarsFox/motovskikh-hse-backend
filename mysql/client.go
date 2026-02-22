package mysql

import (
	"fmt"
	"strings"
	"time"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
)

type Config struct {
	Host    string `envconfig:"optional"`
	Pass    string
	MaxConn int `envconfig:"default=0"`
	Name    string
	User    string
}

type Client struct {
	db *gorm.DB
}

func (c *Config) connection() string {
	sqlHost := c.Host
	if !strings.Contains(sqlHost, "tcp") {
		sqlHost = fmt.Sprintf("tcp(%s)", sqlHost)
	}
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, sqlHost, c.Name)
}

func NewClient(cfg *Config) (*Client, error) {
	db, err := gorm.Open(mysql.Open(cfg.connection()))
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	// Автомиграция для всех таблиц.
	err = db.AutoMigrate(
		&entities.Test{},
		&entities.Question{},
		&entities.UserAnswer{},
		&entities.TestBucket{},
		&entities.TestStats{},
	)
	if err != nil {
		return nil, fmt.Errorf("dbs migrate err: %w", err)
	}

	d, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	d.SetMaxOpenConns(cfg.MaxConn)
	return &Client{db: db}, nil
}

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

// AddAttemptToBucket - добавляет попытку в бакет теста.
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
			
			// Распределение процентов (20 корзин по 5%).
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
		
		// Сохраняем.
		return tx.Save(&bucket).Error
	})
}

// GetTestStats получает статистику теста из бакета.
func (c *Client) GetTestStats(testID string) (*entities.TestStats, error) {
	var bucket entities.TestBucket
	err := c.db.Where("test_id = ?", testID).First(&bucket).Error
	
	if err == gorm.ErrRecordNotFound {
		// Если бакета нет, возвращаем пустую статистику.
		return &entities.TestStats{
			ID:            uuid.New().String(),
			TestID:        testID,
			Period:        "total",
			TotalAttempts: 0,
			ValidAttempts: 0,
			AvgPercentage: 0,
			AvgTimeSpent:  0,
			UpdatedAt:     time.Now(),
		}, nil
	} else if err != nil {
		return nil, err
	}
	
	// Конвертируем в TestStats для обратной совместимости.
	stats := &entities.TestStats{
		ID:            bucket.ID,
		TestID:        bucket.TestID,
		Period:        "total",
		TotalAttempts: int(bucket.TotalAttempts),
		ValidAttempts: int(bucket.ValidAttempts),
		AvgPercentage: bucket.AvgPercentage,
		AvgTimeSpent:  bucket.AvgTimeSpent,
		UpdatedAt:     bucket.UpdatedAt,
	}
	
	return stats, nil
}

// GetPercentileFromBucket - получает перцентиль из данных бакета.
func (c *Client) GetPercentileFromBucket(testID string, percentage float64, timeSpent int) (float64, error) {
	var bucket entities.TestBucket
	err := c.db.Where("test_id = ?", testID).First(&bucket).Error
	
	if err == gorm.ErrRecordNotFound || bucket.ValidAttempts == 0 {
		return 50.0, nil
	} else if err != nil {
		return 0, err
	}
	
	// Считаем сколько попыток хуже.
	worseAttempts := uint64(0)
	
	// Суммируем все корзины с меньшим процентом.
	if percentage > 5 {
		worseAttempts += bucket.Pct0_5
	}
	if percentage > 10 {
		worseAttempts += bucket.Pct5_10
	}
	if percentage > 15 {
		worseAttempts += bucket.Pct10_15
	}
	if percentage > 20 {
		worseAttempts += bucket.Pct15_20
	}
	if percentage > 25 {
		worseAttempts += bucket.Pct20_25
	}
	if percentage > 30 {
		worseAttempts += bucket.Pct25_30
	}
	if percentage > 35 {
		worseAttempts += bucket.Pct30_35
	}
	if percentage > 40 {
		worseAttempts += bucket.Pct35_40
	}
	if percentage > 45 {
		worseAttempts += bucket.Pct40_45
	}
	if percentage > 50 {
		worseAttempts += bucket.Pct45_50
	}
	if percentage > 55 {
		worseAttempts += bucket.Pct50_55
	}
	if percentage > 60 {
		worseAttempts += bucket.Pct55_60
	}
	if percentage > 65 {
		worseAttempts += bucket.Pct60_65
	}
	if percentage > 70 {
		worseAttempts += bucket.Pct65_70
	}
	if percentage > 75 {
		worseAttempts += bucket.Pct70_75
	}
	if percentage > 80 {
		worseAttempts += bucket.Pct75_80
	}
	if percentage > 85 {
		worseAttempts += bucket.Pct80_85
	}
	if percentage > 90 {
		worseAttempts += bucket.Pct85_90
	}
	if percentage > 95 {
		worseAttempts += bucket.Pct90_95
	}
	
	// Добавляем половину из текущей корзины для точности.
	currentBucketAttempts := uint64(0)
	switch {
	case percentage <= 5:
		currentBucketAttempts = bucket.Pct0_5
	case percentage <= 10:
		currentBucketAttempts = bucket.Pct5_10
	case percentage <= 15:
		currentBucketAttempts = bucket.Pct10_15
	case percentage <= 20:
		currentBucketAttempts = bucket.Pct15_20
	case percentage <= 25:
		currentBucketAttempts = bucket.Pct20_25
	case percentage <= 30:
		currentBucketAttempts = bucket.Pct25_30
	case percentage <= 35:
		currentBucketAttempts = bucket.Pct30_35
	case percentage <= 40:
		currentBucketAttempts = bucket.Pct35_40
	case percentage <= 45:
		currentBucketAttempts = bucket.Pct40_45
	case percentage <= 50:
		currentBucketAttempts = bucket.Pct45_50
	case percentage <= 55:
		currentBucketAttempts = bucket.Pct50_55
	case percentage <= 60:
		currentBucketAttempts = bucket.Pct55_60
	case percentage <= 65:
		currentBucketAttempts = bucket.Pct60_65
	case percentage <= 70:
		currentBucketAttempts = bucket.Pct65_70
	case percentage <= 75:
		currentBucketAttempts = bucket.Pct70_75
	case percentage <= 80:
		currentBucketAttempts = bucket.Pct75_80
	case percentage <= 85:
		currentBucketAttempts = bucket.Pct80_85
	case percentage <= 90:
		currentBucketAttempts = bucket.Pct85_90
	case percentage <= 95:
		currentBucketAttempts = bucket.Pct90_95
	default:
		currentBucketAttempts = bucket.Pct95_100
	}
	
	// Добавляем половину из текущей корзины.
	worseAttempts += uint64(float64(currentBucketAttempts) * 0.5)
	
	percentile := (float64(worseAttempts) / float64(bucket.ValidAttempts)) * 100.0
	if percentile > 100 {
		percentile = 100
	}
	
	return percentile, nil
}

// GetTimePercentile возвращает перцентиль по времени.
func (c *Client) GetTimePercentile(testID string, timeSpent int) (float64, error) {
	var bucket entities.TestBucket
	err := c.db.Where("test_id = ?", testID).First(&bucket).Error
	
	if err == gorm.ErrRecordNotFound || bucket.ValidAttempts == 0 {
		return 50.0, nil
	} else if err != nil {
		return 0, err
	}
	
	// Считаем сколько попыток быстрее.
	fasterAttempts := uint64(0)
	
	switch {
	case timeSpent < 60:
	case timeSpent < 120:
		fasterAttempts += bucket.Time0_60
	case timeSpent < 180:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120
	case timeSpent < 240:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180
	case timeSpent < 300:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240
	case timeSpent < 360:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240 + bucket.Time240_300
	default:
		fasterAttempts += bucket.Time0_60 + bucket.Time60_120 + bucket.Time120_180 + bucket.Time180_240 + bucket.Time240_300 + bucket.Time300_360
	}
	
	// Добавляем половину из текущей категории.
	currentCategoryAttempts := uint64(0)
	switch {
	case timeSpent < 60:
		currentCategoryAttempts = bucket.Time0_60
	case timeSpent < 120:
		currentCategoryAttempts = bucket.Time60_120
	case timeSpent < 180:
		currentCategoryAttempts = bucket.Time120_180
	case timeSpent < 240:
		currentCategoryAttempts = bucket.Time180_240
	case timeSpent < 300:
		currentCategoryAttempts = bucket.Time240_300
	case timeSpent < 360:
		currentCategoryAttempts = bucket.Time300_360
	default:
		currentCategoryAttempts = bucket.Time360_
	}
	
	fasterAttempts += uint64(float64(currentCategoryAttempts) * 0.5)
	
	// Чем меньше время, тем лучше, поэтому возвращаем процент тех, кто быстрее.
	timePercentile := (float64(fasterAttempts) / float64(bucket.ValidAttempts)) * 100.0
	if timePercentile > 100 {
		timePercentile = 100
	}
	
	return timePercentile, nil
}

// Вспомогательные функции.
func determineDistributionCategory(percentage float64, percentile float64) *entities.DistributionCategory {
	categories := []entities.DistributionCategory{
		{Name: "elite", Description: "Элита", MinScore: 90, MaxScore: 100},
		{Name: "excellent", Description: "Отлично", MinScore: 75, MaxScore: 90},
		{Name: "good", Description: "Хорошо", MinScore: 60, MaxScore: 75},
		{Name: "average", Description: "Средне", MinScore: 40, MaxScore: 60},
		{Name: "below_average", Description: "Ниже среднего", MinScore: 20, MaxScore: 40},
		{Name: "needs_improvement", Description: "Плохо", MinScore: 0, MaxScore: 20},
	}
	
	for _, cat := range categories {
		if percentage >= cat.MinScore && percentage < cat.MaxScore {
			return &cat
		}
	}
	return &categories[3] 
}

func determineSkillLevel(percentage float64, percentile float64) string {
	if percentile >= 90 {
		return "Эксперт"
	}
	if percentile >= 70 {
		return "Продвинутый"
	}
	if percentile >= 40 {
		return "Средний"
	}
	return "Начинающий"
}

func generateRecommendations(byQuestionType map[string]entities.QuestionTypeStats, overallPercentage float64) []entities.Recommendation {
	var recommendations []entities.Recommendation
	
	if overallPercentage < 50 {
		recommendations = append(recommendations, entities.Recommendation{
			ID:          "general_low_score",
			Title:       "Повторите основы",
			Description: "Ваш результат ниже 50%. Рекомендуем повторить материал.",
			Priority:    "high",
		})
	}
	
	for qType, stats := range byQuestionType {
		if stats.Percentage < 60 && stats.TotalQuestions >= 2 {
			priority := "medium"
			if stats.Percentage < 40 {
				priority = "high"
			}
			
			recommendations = append(recommendations, entities.Recommendation{
				ID:          fmt.Sprintf("improve_%s", qType),
				Title:       fmt.Sprintf("Улучшите знание %s", qType),
				Description: fmt.Sprintf("Точность: %.1f%%", stats.Percentage),
				Priority:    priority,
				QuestionTypes: []string{qType},
			})
		}
	}
	
	return recommendations
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