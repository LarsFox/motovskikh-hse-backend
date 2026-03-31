package api

import (
    "net/http"
    
    "github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
    var req models.SubmitTestRequest
    
    if err := unmarshalParams(r, &req); err != nil {
        m.sendError(w, http.StatusBadRequest, "Invalid request: "+err.Error())
        return
    }
    
    if req.TestName == nil || req.Percentage == nil || req.TimeSpent == nil || req.QuestionCount == nil {
        m.sendError(w, http.StatusBadRequest, "Missing required fields")
        return
    }
    
    // Разыменование указателей.
    testName := *req.TestName
    percentage := float64(*req.Percentage)
    timeSpent := int(*req.TimeSpent)
    questionCount := int(*req.QuestionCount)

    result, err := m.manager.SubmitTestResult(
        testName,
        percentage,
        timeSpent,
        questionCount,
    )
    if err != nil {
        m.sendError(w, http.StatusInternalServerError, "Failed to submit test: "+err.Error())
        return
    }
    
    m.send(w, result)
}