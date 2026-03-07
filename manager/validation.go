package manager

// ValidateAttempt проверяет валидность попытки.
func (m *Manager) ValidateAttempt(testID, userHash string, percentage float64, timeSpent int) (bool, string) {
	// Проверка минимального времени (60 секунд).
	if timeSpent < 60 {
		return false, "Time spent is too short (minimum 60 seconds)"
	}
	
	// Проверка минимального процента (5%).
	if percentage < 5 {
		return false, "Score is too low (minimum 5%)"
	}
	
	return true, ""
}