package api

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/google/uuid"
)

// Создание теста.
type CreateTestRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Difficulty  int                    `json:"difficulty"`
	IsPublic    bool                   `json:"is_public"`
	Questions   []CreateQuestionRequest `json:"questions"`
}

// Создание вопроса.
type CreateQuestionRequest struct {
	Text         string          `json:"text"`
	QuestionType string          `json:"question_type"`
	Options      json.RawMessage `json:"options"`
	CorrectAnswer json.RawMessage `json:"correct_answer"`
	Points       int             `json:"points"`
	OrderIndex   int             `json:"order_index"`
}

// Отправка теста.
type SubmitTestRequest struct {
	TestID  string            `json:"test_id"`
	Answers map[string]string `json:"answers"`
	TimeSpent int            `json:"time_spent"`
}

// Метод hndlrCreateTest - создает новый тест.
func (m *Manager) hndlrCreateTest(w http.ResponseWriter, r *http.Request) {
	var req CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	// Валидация.
	if req.Title == "" || len(req.Questions) == 0 {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	// Создаем тест.
	testID := uuid.New().String()
	test := &entities.Test{
		TestID:      testID,
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Difficulty:  req.Difficulty,
		CreatedBy:   r.Header.Get("X-User-Hash"),
		IsPublic:    req.IsPublic,
		CreatedAt:   time.Now(),
	}
	
	// Создаем вопросы.
	questions := make([]*entities.Question, len(req.Questions))
	for i, qReq := range req.Questions {
		questions[i] = &entities.Question{
			QuestionID:   uuid.New().String(),
			TestID:       testID,
			Text:         qReq.Text,
			QuestionType: qReq.QuestionType,
			Options:      string(qReq.Options),
			CorrectAnswer: string(qReq.CorrectAnswer),
			Points:       qReq.Points,
			OrderIndex:   qReq.OrderIndex,
			CreatedAt:    time.Now(),
		}
	}
	
	// Сохраняем в БД.
	if err := m.manager.CreateTest(test, questions); err != nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	
	m.send(w, map[string]interface{}{
		"test_id":    testID,
		"created":    true,
		"questions_count": len(questions),
	})
}

// Метод hndlrGetTest - возвращает тест для прохождения.
func (m *Manager) hndlrGetTest(w http.ResponseWriter, r *http.Request) {
	testID := r.URL.Query().Get("test_id")
	if testID == "" {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	test, questions, err := m.manager.GetTest(testID)
	if err != nil {
		m.sendErrorPage(w, http.StatusNotFound)
		return
	}
	
	// Преобразуем вопросы.
	var responseQuestions []map[string]interface{}
	for _, q := range questions {
		var options map[string]interface{}
		json.Unmarshal([]byte(q.Options), &options)
		
		responseQuestions = append(responseQuestions, map[string]interface{}{
			"question_id":   q.QuestionID,
			"text":          q.Text,
			"question_type": q.QuestionType,
			"options":       options,
			"points":        q.Points,
			"order_index":   q.OrderIndex,
		})
	}
	
	m.send(w, map[string]interface{}{
		"test": map[string]interface{}{
			"test_id":     test.TestID,
			"title":       test.Title,
			"description": test.Description,
			"category":    test.Category,
			"difficulty":  test.Difficulty,
			"created_by":  test.CreatedBy,
			"is_public":   test.IsPublic,
		},
		"questions": responseQuestions,
		"questions_count": len(responseQuestions),
	})
}

// Метод hndlrSubmitTest - отправляет ответы на тест.
func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
	var req SubmitTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}
	
	userHash := r.Header.Get("X-User-Hash")
	if userHash == "" {
		userHash = uuid.New().String()
	}
	
	// Получаем активную версию теста.
	versionID, err := m.manager.GetActiveVersionID(req.TestID)
	if err != nil {
		log.Printf("Failed to get active version: %v", err)
		versionID = ""
	}
	
	// Проверяем ответы и получаем результат.
	result, err := m.manager.CheckAndSaveAttempt(req.TestID, userHash, req.Answers, req.TimeSpent)
	if err != nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	
	// Получаем базовый анализ.
	stats, percentile, _ := m.manager.GetTestAnalysis(req.TestID, result.Percentage, req.TimeSpent)
	
	// Проверяем валидность попытки.
	isValid, _ := m.manager.ValidateAttempt(req.TestID, userHash, result.Percentage, req.TimeSpent)
	
	response := map[string]interface{}{
		"attempt_id": result.AttemptID,
		"version_id": versionID,
		"score":      result.Score,
		"max_score":  result.MaxScore,
		"percentage": result.Percentage,
		"time_spent": req.TimeSpent,
		"is_valid":   isValid,
		"details":    result.Details,
		"analysis": map[string]interface{}{
			"percentile": percentile,
			"better_than": int(percentile),
		},
	}
	
	if stats != nil {
		response["analysis"].(map[string]interface{})["average_percentage"] = stats.AvgPercentage
		response["analysis"].(map[string]interface{})["average_time"] = stats.AvgTimeSpent
		response["analysis"].(map[string]interface{})["total_attempts"] = stats.TotalAttempts
	}
	
	m.send(w, response)
}