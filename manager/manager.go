package manager

import (
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"math"
	"strings"
)

type db interface {
	GetTest(testID string) (*entities.Test, error)
	// Работа с бакетами.
	AddAttemptToBucket(testID, userHash string, percentage float64, timeSpent int, isValid bool) error
	GetTestStats(testID string) (*entities.TestStats, error)
	GetPercentileFromBucket(testID string, percentage float64, timeSpent int) (float64, error)
	GetTimePercentile(testID string, timeSpent int) (float64, error)
	CreateTestData() error
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{db: db}
}

// GetTest возвращает тест.
func (m *Manager) GetTest(testID string) (*entities.Test, error) {
	return m.db.GetTest(testID)
}

// GetTestAnalysis возвращает анализ.
func (m *Manager) GetTestAnalysis(testID string, userPercentage float64, userTimeSpent int) (*entities.TestStats, float64, error) {
	// Получаем статистику из бакета.
	stats, err := m.db.GetTestStats(testID)
	if err != nil {
		stats = &entities.TestStats{
			ID:            uuid.New().String(),
			TestID:        testID,
			Period:        "total",
			TotalAttempts: 0,
			ValidAttempts: 0,
			AvgPercentage: 0,
			AvgTimeSpent:  0,
			UpdatedAt:     time.Now(),
		}
	}
	
	// Получаем перцентиль из бакета.
	percentile, err := m.db.GetPercentileFromBucket(testID, userPercentage, userTimeSpent)
	if err != nil {
		// Возвращаем примерный перцентиль на основе статистики.
		if stats.AvgPercentage > 0 {
			if userPercentage > stats.AvgPercentage {
				percentile = 75.0
			} else {
				percentile = 25.0
			}
		} else {
			percentile = 50.0
		}
	}
	
	return stats, percentile, nil
}

// ValidateAttempt проверяет валидность попытки.
func (m *Manager) ValidateAttempt(testID, userHash string, percentage float64, timeSpent int) (bool, string) {
	// Проверка минимального времени (60 секунд)
	if timeSpent < 60 {
		return false, "Time spent is too short (minimum 60 seconds)"
	}
	
	// Проверка минимального процента (5%)
	if percentage < 5 {
		return false, "Score is too low (minimum 5%)"
	}
	
	return true, ""
}

// SubmitTestResult сохраняет результат теста и возвращает расширенный анализ.
func (m *Manager) SubmitTestResult(testName, userHash string, percentage float64, timeSpent int) (string, map[string]interface{}, error) {
	// Валидация.
	isValid, validationMessage := m.ValidateAttempt(testName, userHash, percentage, timeSpent)
	
	// Генерируем ID попытки.
	attemptID := uuid.New().String()
	
	// Сохраняем в бакет.
	err := m.db.AddAttemptToBucket(testName, userHash, percentage, timeSpent, isValid)
	if err != nil {
		return "", nil, fmt.Errorf("failed to save attempt: %w", err)
	}
	
	// Получаем статистику из бакета.
	stats, _ := m.db.GetTestStats(testName)
	
	// Получаем перцентили.
	percentileRank, _ := m.db.GetPercentileFromBucket(testName, percentage, timeSpent)
	timePercentile, _ := m.db.GetTimePercentile(testName, timeSpent)
	
	// Определяем категорию и сообщение.
	category := getCategoryFromPercentile(percentileRank)
	timeCategory := getTimeCategory(timePercentile)
	
	// Формируем расширенный анализ.
	analysis := map[string]interface{}{
		// Основные показатели.
		"percentage":      percentage,
		"time_spent":      timeSpent,
		"is_valid":        isValid,
		
		// Сравнение с другими.
		"percentile_rank":    math.Round(percentileRank*10)/10,
		"time_percentile":    math.Round(timePercentile*10)/10,
		"better_than":        int(percentileRank),
		"faster_than":        int(timePercentile),
		
		// Категории.
		"category":           category,
		"time_category":      timeCategory,
		
		// Средние значения.
		"average_percentage": math.Round(stats.AvgPercentage*10)/10,
		"average_time":       math.Round(stats.AvgTimeSpent*10)/10,
		
		// Сравнение со средним.
		"vs_average": map[string]interface{}{
			"percentage_diff": math.Round((percentage - stats.AvgPercentage)*10)/10,
			"time_diff":       math.Round((float64(timeSpent) - stats.AvgTimeSpent)*10)/10,
			"percentage_status": getComparisonStatus(percentage, stats.AvgPercentage),
			"time_status":       getTimeComparisonStatus(float64(timeSpent), stats.AvgTimeSpent),
		},
		
		// Квадрант эффективности.
		"quadrant": getPerformanceQuadrant(percentage, timePercentile),
		
		// Сообщение пользователю.
		"message": generateResultMessage(percentage, timeSpent, percentileRank, timePercentile, stats),
	}
	
	// Формируем итоговый результат.
	result := map[string]interface{}{
		"attempt_id":         attemptID,
		"submitted":          true,
		"validation_message": validationMessage,
		"analysis":           analysis,
	}
	
	return attemptID, result, nil
}

// Вспомогательные функции.
func getTimeCategory(timePercentile float64) string {
	switch {
	case timePercentile >= 90:
		return "very_fast"
	case timePercentile >= 70:
		return "fast"
	case timePercentile >= 40:
		return "average_speed"
	case timePercentile >= 20:
		return "slow"
	default:
		return "very_slow"
	}
}

func getComparisonStatus(userValue, avgValue float64) string {
	diff := userValue - avgValue
	switch {
	case diff > 15:
		return "significantly_higher"
	case diff > 5:
		return "higher"
	case diff < -15:
		return "significantly_lower"
	case diff < -5:
		return "lower"
	default:
		return "similar"
	}
}

func getTimeComparisonStatus(userTime, avgTime float64) string {
	diff := userTime - avgTime
	switch {
	case diff < -60:
		return "much_faster"
	case diff < -20:
		return "faster"
	case diff > 60:
		return "much_slower"
	case diff > 20:
		return "slower"
	default:
		return "similar"
	}
}

func getPerformanceQuadrant(percentage float64, timePercentile float64) map[string]interface{} {
	quadrant := ""
	description := ""
	
	if percentage >= 80 && timePercentile >= 80 {
		quadrant = "expert"
		description = "Эксперт: быстро и правильно!"
	} else if percentage >= 80 && timePercentile < 80 {
		quadrant = "precise_but_slow"
		description = "Точно, но можно быстрее"
	} else if percentage < 80 && timePercentile >= 80 {
		quadrant = "fast_but_inaccurate"
		description = "Быстро, но есть ошибки"
	} else if percentage >= 50 && timePercentile >= 50 {
		quadrant = "solid"
		description = "Хороший результат, продолжай в том же духе"
	} else if percentage < 50 && timePercentile < 50 {
		quadrant = "needs_practice"
		description = "Нужно больше практики"
	} else {
		quadrant = "mixed"
		description = "Противоречивые результаты"
	}
	
	return map[string]interface{}{
		"name":        quadrant,
		"description": description,
		"x":           percentage,
		"y":           timePercentile,
	}
}

func generateResultMessage(percentage float64, timeSpent int, percentileRank, timePercentile float64, stats *entities.TestStats) string {
	var parts []string
	switch {
	case percentileRank >= 90:
		parts = append(parts, "Превосходный результат!")
	case percentileRank >= 75:
		parts = append(parts, "Отличный результат!")
	case percentileRank >= 50:
		parts = append(parts, "Хороший результат.")
	case percentileRank >= 25:
		parts = append(parts, "Средний результат.")
	default:
		parts = append(parts, "Есть над чем поработать.")
	}
	
	// Сообщение о времени.
	switch {
	case timePercentile >= 90:
		parts = append(parts, "Ты справился очень быстро!")
	case timePercentile >= 75:
		parts = append(parts, "Хорошая скорость.")
	case timePercentile <= 25:
		parts = append(parts, "Медленно.")
	}
	
	// Сравнение со средним.
	if stats.ValidAttempts > 0 {
		if percentage > stats.AvgPercentage+10 {
			parts = append(parts, fmt.Sprintf("На %.1f%% выше среднего.", percentage-stats.AvgPercentage))
		} else if percentage < stats.AvgPercentage-10 {
			parts = append(parts, fmt.Sprintf("На %.1f%% ниже среднего.", stats.AvgPercentage-percentage))
		}
		
		timeDiff := float64(timeSpent) - stats.AvgTimeSpent
		if timeDiff < -30 {
			parts = append(parts, "Быстрее среднего на полминуты.")
		} else if timeDiff > 30 {
			parts = append(parts, "Медленнее среднего.")
		}
	}
	
	return strings.Join(parts, " ")
}

func (m *Manager) getCategoryMessage(category string, percentile float64) string {
	messages := map[string]string{
		"excellent":        fmt.Sprintf("Превосходно! Вы лучше, чем %.0f%% участников!", percentile),
		"good":             fmt.Sprintf("Хороший результат! Вы лучше %.0f%% участников.", percentile),
		"average":          "Средний результат. Вы в середине распределения.",
		"below_average":    fmt.Sprintf("Ниже среднего. Вы лучше %.0f%% участников.", percentile),
		"needs_improvement": "Есть над чем поработать.",
	}
	
	if msg, ok := messages[category]; ok {
		return msg
	}
	return "Спасибо за прохождение теста!"
}

func getCategoryFromPercentile(percentile float64) string {
	switch {
	case percentile >= 90:
		return "excellent"
	case percentile >= 70:
		return "good"
	case percentile >= 40:
		return "average"
	case percentile >= 20:
		return "below_average"
	default:
		return "needs_improvement"
	}
}

// CreateTestData создает тестовые данные.
func (m *Manager) CreateTestData() error {
	return m.db.CreateTestData()
}