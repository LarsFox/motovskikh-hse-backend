package manager

const (
	secondsPerQuestionMin = 2
	smallTestThreshold1   = 5
	smallTestThreshold2   = 10
	minPercentageDefault  = 5.0
	minPercentageSmall    = 10.0
)

// ValidateAttempt проверяет валидность попытки.
func (m *Manager) validateAttempt(_ string, percentage float64, timeSpent int, questionCount int) bool {
	// Минимальное время: 2 секунды на вопрос.
	minTime := questionCount * secondsPerQuestionMin
	// Для маленьких тестов.
	if questionCount < smallTestThreshold1 {
		minTime = 1.0
	}
	if timeSpent < minTime {
		return false
	}
	// Минимальный процент для теста.
	minPercentage := minPercentageDefault
	// Для маленьких тестов.
	if questionCount < smallTestThreshold2 {
		minPercentage = minPercentageSmall
	}

	if percentage < minPercentage {
		return false
	}
	return true
}
