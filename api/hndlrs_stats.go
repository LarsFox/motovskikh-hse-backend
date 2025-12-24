package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
)

// SaveAttemptRequest -  Cохранение попытки.
type SaveAttemptRequest struct {
	TestID    string `json:"test_id"`
	VersionID string `json:"version_id"`
	UserHash  string `json:"user_hash"`
	Score     int    `json:"score"`
	MaxScore  int    `json:"max_score"`
	TimeSpent int    `json:"time_spent"`
	Answers   string `json:"answers"`
}

// GetAnalysisRequest - Запрос для анализа.
type GetAnalysisRequest struct {
	TestID     string  `json:"test_id"`
	Percentage float64 `json:"percentage"`
	TimeSpent  int     `json:"time_spent"`
}

// GetDetailedAnalysisRequest - Запрос для подробного анализа.
type GetDetailedAnalysisRequest struct {
    TestID     string  `json:"test_id"`
    VersionID  string  `json:"version_id"`
    Percentage float64 `json:"percentage"`
    TimeSpent  int     `json:"time_spent"`
}

// Метод hndlrSaveAttempt - сохраняет результат теста.
func (m *Manager) hndlrSaveAttempt(w http.ResponseWriter, r *http.Request) {
	var req SaveAttemptRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("SaveAttempt JSON decode error: %v", err)
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	log.Printf("SaveAttempt request: %+v", req)
	
	// Базовая валидация.
	if req.TestID == "" {
		log.Printf("SaveAttempt validation: empty test_id")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if req.Score < 0 {
		log.Printf("SaveAttempt validation: score < 0")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if req.MaxScore <= 0 {
		log.Printf("SaveAttempt validation: max_score <= 0")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if req.TimeSpent <= 0 {
		log.Printf("SaveAttempt validation: time_spent <= 0")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	// Рассчитываем процент и создаем сущность.
	percentage := float64(req.Score) / float64(req.MaxScore) * 100
	
	attempt := &entities.Attempt{
		ID:         uuid.New().String(),
		TestID:     req.TestID,
		VersionID:  req.VersionID,
		UserHash:   req.UserHash,
		Score:      req.Score,
		MaxScore:   req.MaxScore,
		Percentage: percentage,
		TimeSpent:  req.TimeSpent,
		Answers:    req.Answers,
		CreatedAt:  time.Now(),
	}
	
	log.Printf("Saving attempt: %+v", attempt)
	// Сохраняем.
	if err := m.manager.SaveTestAttempt(attempt); err != nil {
		log.Printf("SaveAttempt DB error: %v", err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	
	log.Printf("Attempt saved: %s", attempt.ID)
	// Возвращаем удачный результат.
	m.send(w, map[string]interface{}{
		"saved":      true,
		"attempt_id": attempt.ID,
	})
}

// Метод hndlrGetAnalysis - возвращает анализ результатов.
func (m *Manager) hndlrGetAnalysis(w http.ResponseWriter, r *http.Request) {
	var req GetAnalysisRequest
	
	// Парсинг.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("GetAnalysis JSON decode error: %v", err)
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	log.Printf("GetAnalysis request: %+v", req)
	
	// Валидация.
	if req.TestID == "" {
		log.Printf("GetAnalysis validation: empty test_id")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if req.Percentage < 0 || req.Percentage > 100 {
		log.Printf("GetAnalysis validation: percentage out of range: %f", req.Percentage)
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	if req.TimeSpent <= 0 {
		log.Printf("GetAnalysis validation: time_spent <= 0")
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	// Получаем анализ.
	stats, percentile, err := m.manager.GetTestAnalysis(req.TestID, req.Percentage, req.TimeSpent)
	if err != nil {
		log.Printf("GetAnalysis manager error: %v", err)
		m.sendErrorPage(w, http.StatusInternalServerError)
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
	
	log.Printf("Analysis response: %+v", response)
	
	m.send(w, response)
}

// Метод hndlrGetDetailedAnalysis - возвращает подробный анализ результатов.
func (m *Manager) hndlrGetDetailedAnalysis(w http.ResponseWriter, r *http.Request) {
    var req GetDetailedAnalysisRequest
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        log.Printf("GetDetailedAnalysis JSON decode error: %v", err)
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    log.Printf("GetDetailedAnalysis request: %+v", req)
    
    // Валидация.
    if req.TestID == "" {
        log.Printf("GetDetailedAnalysis validation: empty test_id")
        m.sendErrorPage(w, http.StatusBadRequest)
        return
    }
    
    // Получаем детальный анализ.
    analysis, err := m.manager.GetDetailedAnalysis(req.TestID, req.VersionID, req.Percentage, req.TimeSpent)
    if err != nil {
        log.Printf("GetDetailedAnalysis manager error: %v", err)
        m.sendErrorPage(w, http.StatusInternalServerError)
        return
    }
    
    log.Printf("Детальный анализ сделан: %s", req.TestID)
    
    m.send(w, analysis)
}

// Для создания тестовых данных.
func (m *Manager) hndlrCreateTestData(w http.ResponseWriter, r *http.Request) {
	log.Println("Запрос на создание тестовых данных")
	
	if err := m.manager.CreateTestData(); err != nil {
		log.Printf("Ошибка создания тестовых данных: %v", err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	
	m.send(w, map[string]interface{}{
		"test_data_created": true,
		"message": "Тестовые данные созданы",
	})
}
