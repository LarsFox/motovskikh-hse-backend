package manager

const (
	smallTestQuestions = 2

	smallTestThresholdTime       = 5
	smallTestThresholdPercentage = 10
	minPercentageDefault         = 5.0
	minPercentageSmall           = 10.0
)

// ValidateAttempt проверяет валидность попытки.
// TODO: закодить на фронте тоже.
func (m *Manager) validateAttempt(_ string, percentage, timeSpent float64, questionCount int64) bool {
	if questionCount < smallTestQuestions {
		return false
	}

	// Минимальное время: 2 секунды на вопрос.
	minTime := float64(questionCount * secondsPerQuestionMin)

	// Для маленьких тестов.
	if questionCount < smallTestThresholdTime {
		minTime = 0.5
	}
	if timeSpent < minTime {
		return false
	}

	// Минимальный процент для теста.
	minPercentage := minPercentageDefault
	// Для маленьких тестов.
	if questionCount < smallTestThresholdPercentage {
		minPercentage = minPercentageSmall
	}

	return percentage >= minPercentage
}
