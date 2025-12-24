package entities

import "time"

// Test - структура теста.
type Test struct {
	TestID      string    `json:"test_id" gorm:"primaryKey"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Difficulty  int       `json:"difficulty"`
	CreatedBy   string    `json:"created_by"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

// Question - вопрос в тесте.
type Question struct {
	QuestionID   string    `json:"question_id" gorm:"primaryKey"`
	TestID       string    `json:"test_id"`
	Text         string    `json:"text"`
	QuestionType string    `json:"question_type"`
	Options      string    `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
	Points       int       `json:"points"`
	OrderIndex   int       `json:"order_index"`
	CreatedAt    time.Time `json:"created_at"`
}

// Статистика по вопросу.
type QuestionStats struct {
    QuestionID    string    `json:"question_id" gorm:"primaryKey"`
    TestID        string    `json:"test_id"`
    VersionID     string    `json:"version_id"`
    Date          time.Time `json:"date"`
    TotalAnswers  int       `json:"total_answers"`     // Всего ответов.
    CorrectAnswers int      `json:"correct_answers"`   // Правильных ответов.
    SuccessRate   float64   `json:"success_rate"`      // Процент правильных.
    AverageTime   float64   `json:"average_time"`      // Среднее время на вопрос.
    CommonMistakes string   `json:"common_mistakes"`   // Частые ошибки.
}

// UserAnswer - ответ пользователя.
type UserAnswer struct {
	AnswerID   string    `json:"answer_id" gorm:"primaryKey"`
	AttemptID  string    `json:"attempt_id"`
	QuestionID string    `json:"question_id"`
	UserHash   string    `json:"user_hash"`
	Answer     string    `json:"answer"`
	IsCorrect  bool      `json:"is_correct"`
	Score      int       `json:"score"`
	CreatedAt  time.Time `json:"created_at"`
}

// TestConfig - конфигурация теста.
type TestConfig struct {
	ConfigID           string    `json:"config_id" gorm:"primaryKey"`
	TestID             string    `json:"test_id"`
	MinTimeSpent       int       `json:"min_time_spent"`
	MaxTimeSpent       int       `json:"max_time_spent"`
	MinPercentage      int       `json:"min_percentage"`
	MaxAttemptsPerUser int       `json:"max_attempts_per_user"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TestVersion - версия теста.
type TestVersion struct {
    VersionID     string    `json:"version_id" gorm:"primaryKey"`
    TestID        string    `json:"test_id"`
    VersionNumber int       `json:"version_number"`
    Description   string    `json:"description"`
    CreatedAt     time.Time `json:"created_at"`
    IsActive      bool      `json:"is_active"`
}