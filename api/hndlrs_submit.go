package api

import (
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitTestRequest

	if err := unmarshalParams(r, &req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}


	result, err := m.manager.SubmitTestResult(
		*req.TestName,
		*req.Percentage,
		*req.TimeSpent,
		*req.QuestionCount,
	)
	if err != nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, result)
}
