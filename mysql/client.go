package mysql

import (
	"fmt"
	"strings"
	"time"
	"gorm.io/driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// Config — конфигурация клиента.
type Config struct {
	Host    string `envconfig:"optional"`
	Pass    string
	MaxConn int `envconfig:"default=0"`
	Name    string
	User    string
}

type Client struct {
	db *gorm.DB // GORM ORM для работы с БД.
}

// Подключение.
func (c *Config) connection() string {
	sqlHost := c.Host
	if !strings.Contains(sqlHost, "tcp") {
		sqlHost = fmt.Sprintf("tcp(%s)", sqlHost)
	}
	return fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4&parseTime=True&loc=Local", c.User, c.Pass, sqlHost, c.Name)
}

func NewClient(cfg *Config) (*Client, error) {
	// Через gorm подключаемся.
	db, err := gorm.Open(mysql.Open(cfg.connection()))
	if err != nil {
		return nil, fmt.Errorf("dbs new client err: %w", err)
	}

	// Автоматическое создание таблиц.
	err = db.AutoMigrate(
		&entities.Attempt{},
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

// Метод SaveAttempt - сохраняем попытку теста.
func (c *Client) SaveAttempt(attempt *entities.Attempt) error {
	err := c.db.Create(attempt).Error
	if err != nil {
		return err
	}
	
	// Пересчет статистики через 10 попыток.
	var count int64
	c.db.Model(&entities.Attempt{}).Where("test_id = ?", attempt.TestID).Count(&count)
	if count % 10 == 0 {
		go c.CalculateTestStats(attempt.TestID)
	}
	
	return nil
}

// Метод GetTestStats - возвращает статистику по тесту.
func (c *Client) GetTestStats(testID string) (*entities.TestStats, error) {
	var stats entities.TestStats
	err := c.db.Where("test_id = ?", testID).
		Order("date DESC").
		First(&stats).Error
		
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &stats, err
}

// GetUserPercentile возвращает перцентиль пользователя.
func (c *Client) GetUserPercentile(testID string, percentage float64, timeSpent int) (float64, error) {
	// Кол-во попыток хуже.
	var worseAttempts int64
	err := c.db.Model(&entities.Attempt{}).
		Where("test_id = ? AND (percentage < ? OR (percentage = ? AND time_spent > ?))", 
			testID, percentage, percentage, timeSpent).
		Count(&worseAttempts).Error
		
	if err != nil {
		return 0, err
	}
	
	var totalAttempts int64
	err = c.db.Model(&entities.Attempt{}).
		Where("test_id = ?", testID).
		Count(&totalAttempts).Error
		
	if err != nil || totalAttempts == 0 {
		return 0, err
	}
	
	// Перцентиль.
	percentile := (float64(worseAttempts) + 0.5) / float64(totalAttempts) * 100
	
	if percentile > 100 {
		percentile = 100
	}
	
	return percentile, nil
}

// CreateTestData создает тестовые данные, чтобы тестить.
func (c *Client) CreateTestData() error {
	c.db.Exec("DELETE FROM attempts")
	
	testAttempts := []*entities.Attempt{
		{ID: "test1", TestID: "europe_test", Score: 2, MaxScore: 10, Percentage: 20, TimeSpent: 600, CreatedAt: time.Now()},
		{ID: "test2", TestID: "europe_test", Score: 3, MaxScore: 10, Percentage: 30, TimeSpent: 550, CreatedAt: time.Now()},
		{ID: "test3", TestID: "europe_test", Score: 4, MaxScore: 10, Percentage: 40, TimeSpent: 500, CreatedAt: time.Now()},
		{ID: "test4", TestID: "europe_test", Score: 5, MaxScore: 10, Percentage: 50, TimeSpent: 400, CreatedAt: time.Now()},
		{ID: "test5", TestID: "europe_test", Score: 6, MaxScore: 10, Percentage: 60, TimeSpent: 350, CreatedAt: time.Now()},
		{ID: "test6", TestID: "europe_test", Score: 7, MaxScore: 10, Percentage: 70, TimeSpent: 300, CreatedAt: time.Now()},
		{ID: "test7", TestID: "europe_test", Score: 8, MaxScore: 10, Percentage: 80, TimeSpent: 250, CreatedAt: time.Now()},
		{ID: "test8", TestID: "europe_test", Score: 9, MaxScore: 10, Percentage: 90, TimeSpent: 200, CreatedAt: time.Now()},
		{ID: "test9", TestID: "europe_test", Score: 10, MaxScore: 10, Percentage: 100, TimeSpent: 150, CreatedAt: time.Now()},
	}
	
	for _, attempt := range testAttempts {
		if err := c.db.Create(attempt).Error; err != nil {
			return err
		}
	}
	return nil
}

// CalculateTestStats вычисляет статистику по тесту.
func (c *Client) CalculateTestStats(testID string) error {
	var avgPercentage float64
	var avgTimeSpent float64
	var totalAttempts int64
	
	// Средний процент.
	err := c.db.Model(&entities.Attempt{}).
		Where("test_id = ?", testID).
		Select("AVG(percentage)").
		Scan(&avgPercentage).Error
	if err != nil {
		return err
	}
	
	// Среднее время.
	err = c.db.Model(&entities.Attempt{}).
		Where("test_id = ?", testID).
		Select("AVG(time_spent)").
		Scan(&avgTimeSpent).Error
	if err != nil {
		return err
	}
	
	// Общее количество.
	err = c.db.Model(&entities.Attempt{}).
		Where("test_id = ?", testID).
		Count(&totalAttempts).Error
	if err != nil {
		return err
	}
	
	// Для простоты.
	percentile50 := avgPercentage * 0.9
	percentile80 := avgPercentage * 1.1
	percentile95 := avgPercentage * 1.2
	
	stats := &entities.TestStats{
		ID:            uuid.New().String(),
		TestID:        testID,
		Date:          time.Now(),
		TotalAttempts: int(totalAttempts),
		ValidAttempts: int(totalAttempts),
		AvgPercentage: avgPercentage,
		AvgTimeSpent:  avgTimeSpent,
		Percentile50:  percentile50,
		Percentile80:  percentile80,
		Percentile95:  percentile95,
	}
	
	return c.db.Create(stats).Error
}