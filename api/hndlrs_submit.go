package api

import (
	"errors"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/achievements"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

const roundMultiplier = 10

func (m *Manager) hndlrSubmitTest(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitTestRequest
	var userID int64
	if m.achievements != nil && m.userIDs != nil {
		var err error
		userID, err = m.userIDs.UserID(r)
		if err != nil || userID <= 0 {
			m.sendErrorPage(w, http.StatusUnauthorized)
			return
		}
	}

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

	response := map[string]any{
		"score_percentile":   analysis.ScorePercentile,
		"time_percentile":    analysis.TimePercentile,
		"better_than":        analysis.BetterThan,
		"faster_than":        analysis.FasterThan,
		"average_percentage": analysis.AveragePercentage,
		"average_time":       analysis.AverageTime,
		"vs_average": map[string]float64{
			"percentage_diff": analysis.PercentageDiff,
			"time_diff":       analysis.TimeDiff,
		},
	}
	if m.achievements != nil && m.userIDs != nil {
		awarded, err := m.achievements.AwardForTestResult(r.Context(), userID, achievements.TestResult{
			TestSlug: *req.TestName, Percentage: *req.Percentage, TimeSpent: *req.TimeSpent,
		}, requestLanguage(r))
		if err != nil {
			if errors.Is(err, achievements.ErrInvalidResult) || errors.Is(err, achievements.ErrUnknownTest) {
				m.sendErrorPage(w, http.StatusBadRequest)
			} else {
				m.sendErrorPage(w, http.StatusInternalServerError)
			}
			return
		}
		response["new_achievements"] = awarded
	}
	m.send(w, response)
}
