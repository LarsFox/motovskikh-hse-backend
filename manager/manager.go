package manager

import (
	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// Для БД абстрацкия.
type db interface {
	Stub() bool
	SaveAttempt(attempt *entities.Attempt) error
	GetTestStats(testID string) (*entities.TestStats, error)
	GetUserPercentile(testID string, percentage float64, timeSpent int) (float64, error)
	CreateTestData() error
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

// SaveTestAttempt сохраняет результат теста.
func (m *Manager) SaveTestAttempt(attempt *entities.Attempt) error {
	// Валидация (простая)
	if attempt.TimeSpent < 10 { // Минимум 10 секунд.
		return entities.ErrInvalidInput
	}
	if attempt.Percentage < 5 { // Минимум 5% правильных ответов.
		return entities.ErrInvalidInput
	}
	
	return m.db.SaveAttempt(attempt)
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

// Создание тестовых данных.
func (m *Manager) CreateTestData() error {
	// Приведение типа к нашему клиенту MySQL
	if client, ok := m.db.(interface{ CreateTestData() error }); ok {
		return client.CreateTestData()
	}
	return nil
}