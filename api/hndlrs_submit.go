package api

import (
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

const roundMultiplier = 10

func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitTestRequest

	if err := unmarshalParams(r, &req); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	analysis, err := m.manager.SubmitTestResult(
		*req.TestName,
		*req.Percentage,
		*req.TimeSpent,
		*req.QuestionCount,
	)
	if err != nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, &models.SubmitTestResponse{
		ScorePercentile:   analysis.ScorePercentile,
		TimePercentile:    analysis.TimePercentile,
		BetterThan:        analysis.BetterThan,
		FasterThan:        analysis.FasterThan,
		AveragePercentage: analysis.AveragePercentage,
		AverageTime:       analysis.AverageTime,
		VsAverage: &models.SubmitTestResponseVsAverage{
			PercentageDiff: analysis.PercentageDiff,
			TimeDiff:       analysis.TimeDiff,
		},
	})
}
