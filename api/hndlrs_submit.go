package api

import (
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
)

type SubmitTestRequest struct {
	TestName    string  `json:"test_name"`
	Percentage  float64 `json:"percentage"`
	TimeSpent   int     `json:"time_spent"`
}

// hndlrSubmitTest - хэндлер отправки теста.
func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
	var req SubmitTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	
	// Валидация.
	if req.TestName == "" || req.TimeSpent <= 0 || req.Percentage < 0 || req.Percentage > 100 {
		m.sendError(w, http.StatusBadRequest, "Missing or invalid required fields: test_name, time_spent > 0, percentage 0-100")
		return
	}
	
	// Получаем или генерируем userHash.
	userHash := r.Header.Get("X-User-Hash")
	if userHash == "" {
		userHash = "anonymous_" + uuid.New().String()[:8]
	}
	
	// Передаем процент и время.
	attemptID, result, err := m.manager.SubmitTestResult(req.TestName, userHash, req.Percentage, req.TimeSpent)
	if err != nil {
		m.sendError(w, http.StatusInternalServerError, "Failed to submit test: "+err.Error())
		return
	}

	result["attempt_id"] = attemptID
	
	m.send(w, result)
}