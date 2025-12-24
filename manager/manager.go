package manager

import (
	"encoding/json"
	"time"
	"github.com/google/uuid"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// Для БД абстракция.
type db interface {
	Stub() bool
	SaveAttempt(attempt *entities.Attempt) error
	GetTestStats(testID string) (*entities.TestStats, error)
	GetUserPercentile(testID string, percentage float64, timeSpent int) (float64, error)
	CreateTestData() error
	CalculateTestStats(testID string) error
	CreateTest(test *entities.Test, questions []*entities.Question) error
	GetTest(testID string) (*entities.Test, []*entities.Question, error)
	CheckAnswers(testID, userHash string, answers map[string]string) (*entities.AttemptResult, error)
	SaveFullAttempt(attempt *entities.Attempt, userAnswers []*entities.UserAnswer) error
	ValidateAttempt(testID, userHash string, percentage float64, timeSpent int) (bool, error)
	GetQuestionStats(testID, versionID string) ([]*entities.QuestionStats, error)
	GetMedianTime(testID, versionID string) (float64, error)
	GetActiveVersionID(testID string) (string, error)
}

type Manager struct {
	db db
}

func New(db db) *Manager {
	return &Manager{
		db: db,
	}
}

func (m *Manager) Stub() bool {
	return m.db.Stub()
}

// CalculateTestStats вычисляет статистику по тесту.
func (m *Manager) CalculateTestStats(testID string) error {
	return m.db.CalculateTestStats(testID)
}

// GetUserPercentile возвращает перцентиль пользователя.
func (m *Manager) GetUserPercentile(testID string, percentage float64, timeSpent int) (float64, error) {
	return m.db.GetUserPercentile(testID, percentage, timeSpent)
}

// CreateTest создает новый тест.
func (m *Manager) CreateTest(test *entities.Test, questions []*entities.Question) error {
	// Валидация.
	if test.Title == "" {
		return entities.ErrInvalidInput
	}
	if len(questions) == 0 {
		return entities.ErrInvalidInput
	}
	for _, q := range questions {
		if q.Text == "" || q.Points <= 0 {
			return entities.ErrInvalidInput
		}
	}
	
	return m.db.CreateTest(test, questions)
}

// GetTest возвращает тест для прохождения.
func (m *Manager) GetTest(testID string) (*entities.Test, []*entities.Question, error) {
	return m.db.GetTest(testID)
}

// CheckAndSaveAttempt проверяет ответы и сохраняет попытку.
func (m *Manager) CheckAndSaveAttempt(testID, userHash string, answers map[string]string, timeSpent int) (*entities.AttemptResult, error) {
	// Проверяем ответы
	result, err := m.db.CheckAnswers(testID, userHash, answers)
	if err != nil {
		return nil, err
	}
	// Сериализуем answers в JSON.
	answersJSON, _ := json.Marshal(answers)
	// Создаем запись попытки.
	attempt := &entities.Attempt{
		ID:         uuid.New().String(),
		TestID:     testID,
		UserHash:   userHash,
		Score:      result.Score,
		MaxScore:   result.MaxScore,
		Percentage: result.Percentage,
		TimeSpent:  timeSpent,
		Answers:    string(answersJSON),
		IsValid:    true,
		CreatedAt:  time.Now(),
	}
	// Сохраняем попытку и ответы.
	if err := m.db.SaveFullAttempt(attempt, result.UserAnswers); err != nil {
		return nil, err
	}
	result.AttemptID = attempt.ID
	return result, nil
}

// GetTestAnalysis возвращает анализ результатов теста.
func (m *Manager) GetTestAnalysis(testID string, userPercentage float64, userTimeSpent int) (*entities.TestStats, float64, error) {
	// Получаем общую статистику по тесту.
	stats, err := m.db.GetTestStats(testID)
	if err != nil {
		return nil, 0, err
	}
	// Вычисляем перцентиль пользователя.
	percentile, err := m.db.GetUserPercentile(testID, userPercentage, userTimeSpent)
	if err != nil {
		return stats, 0, err
	}
	
	return stats, percentile, nil
}

// SaveTestAttempt сохраняет результат теста.
func (m *Manager) SaveTestAttempt(attempt *entities.Attempt) error {
	// Валидация.
	if attempt.TimeSpent < 30 { // Минимум 30 секунд.
		return entities.ErrInvalidInput
	}
	if attempt.Percentage < 5 { // Минимум 5% правильных ответов.
		return entities.ErrInvalidInput
	}
	return m.db.SaveAttempt(attempt)
}

// CreateTestData создает тестовые данные.
func (m *Manager) CreateTestData() error {
	return m.db.CreateTestData()
}

// GetDetailedAnalysis возвращает детальный анализ теста.
func (m *Manager) GetDetailedAnalysis(testID, versionID string, userPercentage float64, userTimeSpent int) (map[string]interface{}, error) {
    // Общая статистика.
    stats, percentile, err := m.GetTestAnalysis(testID, userPercentage, userTimeSpent)
    if err != nil {
        return nil, err
    }
    
    // Статистика по вопросам.
    questionStats, err := m.db.GetQuestionStats(testID, versionID)
    if err != nil {
        return nil, err
    }
    
    // Медианное время.
    medianTime, err := m.db.GetMedianTime(testID, versionID)
    if err != nil {
        return nil, err
    }
    
    // Формируем ответ.
    result := map[string]interface{}{
        "user_performance": map[string]interface{}{
            "percentage": userPercentage,
            "time_spent": userTimeSpent,
            "percentile": percentile,
            "better_than": int(percentile),
        },
        "test_statistics": map[string]interface{}{
            "average_percentage": stats.AvgPercentage,
            "average_time":       stats.AvgTimeSpent,
            "median_time":        medianTime,
            "total_attempts":     stats.TotalAttempts,
            "valid_attempts":     stats.ValidAttempts,
        },
        "question_analysis": []map[string]interface{}{},
    }
    
    // Добавляем статистику по вопросам.
    for _, qStat := range questionStats {
        var commonMistakes []map[string]interface{}
        json.Unmarshal([]byte(qStat.CommonMistakes), &commonMistakes)
        
        result["question_analysis"] = append(result["question_analysis"].([]map[string]interface{}), 
            map[string]interface{}{
                "question_id":    qStat.QuestionID,
                "success_rate":   qStat.SuccessRate,
                "total_answers":  qStat.TotalAnswers,
                "correct_answers": qStat.CorrectAnswers,
                "average_time":   qStat.AverageTime,
                "difficulty":     getDifficultyLevel(qStat.SuccessRate),
                "common_mistakes": commonMistakes,
            })
    }
    
    return result, nil
}

// Сложность.
func getDifficultyLevel(successRate float64) string {
    switch {
    case successRate < 30:
        return "сложный"
    case successRate < 60:
        return "средний"
		default:
        return "легкий"
    }
}

// Заглушка.
func (m *Manager) GetActiveVersionID(testID string) (string, error) {
    return "", nil
}

// ValidateAttempt проверяет валидность попытки.
func (m *Manager) ValidateAttempt(testID, userHash string, percentage float64, timeSpent int) (bool, error) {
    return m.db.ValidateAttempt(testID, userHash, percentage, timeSpent)
}
