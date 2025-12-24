package mysql

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"gorm.io/gorm"
	"time"
)

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
		// Если статистики нет, возвращаем пустую структуру.
		return &entities.TestStats{
			TestID:        testID,
			Date:          time.Now(),
			TotalAttempts: 0,
			ValidAttempts: 0,
			AvgPercentage: 0,
			AvgTimeSpent:  0,
		}, nil
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

// ValidateAttempt проверяет, является ли попытка валидной по правилам фильтрации.
func (c *Client) ValidateAttempt(testID, userHash string, percentage float64, timeSpent int) (bool, error) {
    // Получаем конфигурацию теста.
    var config entities.TestConfig
    err := c.db.Where("test_id = ?", testID).First(&config).Error
    if err != nil {
        return false, err
    }
    
    // Проверка по времени.
    if timeSpent < config.MinTimeSpent || timeSpent > config.MaxTimeSpent {
        return false, nil
    }
    
    // Проверка по проценту.
    if percentage < float64(config.MinPercentage) {
        return false, nil
    }
    
    // Проверка частоты попыток.
    var attemptsLast24h int64
    oneDayAgo := time.Now().Add(-24 * time.Hour)
    
    err = c.db.Model(&entities.Attempt{}).
        Where("test_id = ? AND user_hash = ? AND created_at > ?", 
            testID, userHash, oneDayAgo).
        Count(&attemptsLast24h).Error
    if err != nil {
        return false, err
    }
    
    if attemptsLast24h >= int64(config.MaxAttemptsPerUser) {
        return false, nil
    }
    
    return true, nil
}

// GetValidAttempts возвращает только валидные попытки для статистики.
func (c *Client) GetValidAttempts(testID, versionID string, startDate, endDate time.Time) ([]*entities.Attempt, error) {
    // Сначала получаем конфигурацию.
    var config entities.TestConfig
    err := c.db.Where("test_id = ?", testID).First(&config).Error
    if err != nil {
        config = entities.TestConfig{
            MinTimeSpent:       30,
            MaxTimeSpent:       3600,
            MinPercentage:      0,
            MaxAttemptsPerUser: 3,
        }
    }
    
    // Основной запрос с указанием таблицы для test_id.
    query := c.db.Model(&entities.Attempt{}).Where("attempts.test_id = ?", testID)
    
    if versionID != "" {
        query = query.Where("attempts.version_id = ?", versionID)
    }
    if !startDate.IsZero() {
        query = query.Where("attempts.created_at >= ?", startDate)
    }
    if !endDate.IsZero() {
        query = query.Where("attempts.created_at <= ?", endDate)
    }
    
    // Применяем фильтры из конфигурации.
    query = query.Where("attempts.time_spent BETWEEN ? AND ?", 
        config.MinTimeSpent, config.MaxTimeSpent)
    query = query.Where("attempts.percentage >= ?", config.MinPercentage)
    
    // Спам, группируем по пользователю и берем лучший результат за период.
    subQuery := c.db.Model(&entities.Attempt{}).
        Select("MIN(created_at) as min_date, user_hash, test_id").
        Where("test_id = ?", testID)
    
    // Добавляем условия даты только если они не нулевые.
    if !startDate.IsZero() && !endDate.IsZero() {
        subQuery = subQuery.Where("created_at BETWEEN ? AND ?", startDate, endDate)
    }
    
    subQuery = subQuery.Group("user_hash, test_id")
    
    // Основной запрос с JOIN.
    var attempts []*entities.Attempt
    err = query.
        Joins("INNER JOIN (?) as sub ON attempts.user_hash = sub.user_hash AND attempts.test_id = sub.test_id AND attempts.created_at = sub.min_date", subQuery).
        Order("attempts.percentage DESC, attempts.time_spent ASC").
        Find(&attempts).Error
    
    return attempts, err
}

// GetMedianTime возвращает медианное время прохождения.
func (c *Client) GetMedianTime(testID, versionID string) (float64, error) {
    attempts, err := c.GetValidAttempts(testID, versionID, time.Time{}, time.Time{})
    if err != nil {
        return 0, err
    }
    if len(attempts) == 0 {
        return 0, nil
    }
    totalTime := 0
    for _, attempt := range attempts {
        totalTime += attempt.TimeSpent
    }
    return float64(totalTime) / float64(len(attempts)), nil
}