package mysql

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
	"time"
	"encoding/json"
    "strings"
    "log"
)

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
	
	stats := &entities.TestStats{
		ID:            uuid.New().String(),
		TestID:        testID,
		Date:          time.Now(),
		TotalAttempts: int(totalAttempts),
		ValidAttempts: int(totalAttempts),
		AvgPercentage: avgPercentage,
		AvgTimeSpent:  avgTimeSpent,
	}
	
	return c.db.Create(stats).Error
}

// CalculateQuestionStats рассчитывает статистику по каждому вопросу.
func (c *Client) CalculateQuestionStats(testID, versionID string) error {
    // Получаем все валидные попытки.
    startDate := time.Now().AddDate(0, -1, 0)
    attempts, err := c.GetValidAttempts(testID, versionID, startDate, time.Now())
    if err != nil {
        return err
    }
    
    // Используем map для агрегации статистики по вопросам.
    questionStatsMap := make(map[string]*entities.QuestionStats)
    
    for _, attempt := range attempts {
        var userAnswers []*entities.UserAnswer
        err := c.db.Where("attempt_id = ?", attempt.ID).Find(&userAnswers).Error
        if err != nil {
            continue
        }
        
        for _, answer := range userAnswers {
            // Проверяем, есть ли уже статистика для этого вопроса.
            key := answer.QuestionID + "_" + versionID
            stat, exists := questionStatsMap[key]
            
            if !exists {
                // Создаем новую статистику.
                stat = &entities.QuestionStats{
                    QuestionID:    answer.QuestionID,
                    TestID:        testID,
                    VersionID:     versionID,
                    Date:          time.Now(),
                    TotalAnswers:  0,
                    CorrectAnswers: 0,
                    SuccessRate:   0,
                    AverageTime:   0,
                    CommonMistakes: "[]",
                }
                questionStatsMap[key] = stat
            }
            
            // Обновляем статистику.
            stat.TotalAnswers++
            if answer.IsCorrect {
                stat.CorrectAnswers++
            }
            stat.SuccessRate = float64(stat.CorrectAnswers) / float64(stat.TotalAnswers) * 100
            stat.AverageTime = (stat.AverageTime*float64(stat.TotalAnswers-1) + float64(attempt.TimeSpent)/float64(len(userAnswers))) / float64(stat.TotalAnswers)
            
            // Обновляем частые ошибки.
            if !answer.IsCorrect && answer.Answer != "" {
                var mistakes []map[string]interface{}
                json.Unmarshal([]byte(stat.CommonMistakes), &mistakes)
                
                // Добавляем ошибку.
                mistakes = append(mistakes, map[string]interface{}{
                    "answer": answer.Answer,
                    "count":  1,
                })
                
                // Ограничиваем количество ошибок.
                if len(mistakes) > 5 {
                    mistakes = mistakes[:5]
                }
                
                mistakesJSON, _ := json.Marshal(mistakes)
                stat.CommonMistakes = string(mistakesJSON)
            }
        }
    }
    
    // Сохраняем или обновляем статистику.
    for _, stat := range questionStatsMap {
        if err := c.db.Save(stat).Error; err != nil {
            // Если ошибка дублирования, обновляем существующую запись.
            if strings.Contains(err.Error(), "Duplicate entry") {
                // Сначала пытаемся найти существующую.
                var existing entities.QuestionStats
                if err := c.db.Where("question_id = ? AND version_id = ?", 
                    stat.QuestionID, stat.VersionID).First(&existing).Error; err == nil {
                    // Обновляем существующую.
                    existing.TotalAnswers = stat.TotalAnswers
                    existing.CorrectAnswers = stat.CorrectAnswers
                    existing.SuccessRate = stat.SuccessRate
                    existing.AverageTime = stat.AverageTime
                    existing.CommonMistakes = stat.CommonMistakes
                    existing.Date = time.Now()
                    
                    if err := c.db.Save(&existing).Error; err != nil {
                        log.Printf("Failed to update question stats: %v", err)
                    }
                }
            } else {
                log.Printf("Failed to save question stats: %v", err)
            }
        }
    }
    
    return nil
}

// GetQuestionStats возвращает статистику по вопросам теста.
func (c *Client) GetQuestionStats(testID, versionID string) ([]*entities.QuestionStats, error) {
    var stats []*entities.QuestionStats
    query := c.db.Where("test_id = ?", testID)
    
    if versionID != "" {
        query = query.Where("version_id = ?", versionID)
    }
    
    err := query.Find(&stats).Error
    return stats, err
}