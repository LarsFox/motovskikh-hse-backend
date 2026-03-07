package api

import (
	"encoding/json"
	"net/http"
)

type GetAnalysisRequest struct {
	TestID     string  `json:"test_id"`
	Percentage float64 `json:"percentage"`
	TimeSpent  int     `json:"time_spent"`
}

// hndlrGetAnalysis возвращает анализ результатов.
func (m *Manager) hndlrGetAnalysis(w http.ResponseWriter, r *http.Request) {
	var req GetAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}
	
	// Валидация.
	if req.TestID == "" || req.Percentage < 0 || req.Percentage > 100 || req.TimeSpent <= 0 {
		m.sendError(w, http.StatusBadRequest, "Invalid parameters")
		return
	}
	
	// Получаем анализ.
	stats, percentile, err := m.manager.GetTestAnalysis(req.TestID, req.Percentage, req.TimeSpent)
	if err != nil {
		m.sendError(w, http.StatusInternalServerError, "Failed to get analysis")
		return
	}
	
	response := map[string]interface{}{
		"your_percentage": req.Percentage,
		"your_time":       req.TimeSpent,
		"percentile":      percentile,
		"better_than":     int(percentile),
	}
	
	if stats != nil {
		response["average_percentage"] = stats.AvgPercentage
		response["average_time"] = stats.AvgTimeSpent
		response["total_attempts"] = stats.TotalAttempts
	}
	
	m.send(w, response)
}
