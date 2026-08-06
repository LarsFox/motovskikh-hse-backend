package api

import (
	"net/http"
	"strings"
)

func (m *Manager) hndlrAchievements(w http.ResponseWriter, r *http.Request) {
	userID, err := m.userIDs.UserID(r)
	if err != nil || userID <= 0 {
		m.sendErrorPage(w, http.StatusUnauthorized)
		return
	}
	items, err := m.achievements.List(r.Context(), userID, requestLanguage(r))
	if err != nil {
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}
	m.send(w, map[string]any{"achievements": items})
}

func requestLanguage(r *http.Request) string {
	if language := strings.TrimSpace(r.URL.Query().Get("language")); language != "" {
		return language
	}
	return r.Header.Get("Accept-Language")
}
