package api

import (
	"errors"
	"net/http"
)

// hndlrGetTests возвращает список тестов пользователя.
func (m *Manager) hndlrListUserTests(w http.ResponseWriter, r *http.Request) {
	userID := getFromCtxTester(r).ID

	tests, err := m.manager.ListUserTests(r.Context(), userID)
	switch {
	case errors.Is(err, nil):
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, tests)
}
