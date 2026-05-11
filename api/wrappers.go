package api

import (
	"net/http"
	"strings"
)

type wrapper func(http.Handler) http.Handler

func (m *Manager) wrapContentTypeJSON(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if strings.ToLower(ct) != "application/json; charset=utf-8" {
			m.sendErrorPage(w, http.StatusBadRequest)
			return
		}
		inner.ServeHTTP(w, r)
	})
}

func (m *Manager) wrapBodyMaxSize(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		inner.ServeHTTP(w, r)
	})
}

// wrapEasterEggHeader добавляет ржомбу в заголовки.
// nolint:canonicalheader
func (m *Manager) wrapEasterEggHeader(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("hey", "what are you trying to find here")
		w.Header().Set("Leon-Motovskikh", "is the best")
		w.Header().Set("x-files", "Scully approves")
		inner.ServeHTTP(w, r)
	})
}

// wrapRecover отправляет ошибку.
func wrapRecover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer notifyRecover(map[string]any{"uri": r.RequestURI})
		h.ServeHTTP(w, r)
	})
}

// wrapAuth проверяет JWT токен и стоит на страже маршрутов личного кабинета.
func (m *Manager) wrapAuth(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.sendError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			m.sendError(w, http.StatusUnauthorized, "invalid authorization format")
			return
		}

		userID, err := m.manager.Token().ValidateAccess(tokenString)
		if err != nil {
			m.sendError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		// Кладём user_id в контекст
		ctx := contextWithUserID(r.Context(), userID)
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
