package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
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
//
//nolint:unused
func (m *Manager) wrapAuth(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		switch {
		case errors.Is(err, nil):
		default:
			m.sendErrorPage(w, http.StatusUnauthorized)
			return
		}

		userID, err := m.manager.ValidateAccess(cookie.Value)
		switch {
		case errors.Is(err, nil):
		default:
			m.sendErrorPage(w, http.StatusUnauthorized)
			return
		}

		// Кладём user_id в контекст
		ctx := context.WithValue(r.Context(), ctxUser, &entities.User{ID: userID})
		inner.ServeHTTP(w, r.WithContext(ctx))
	})
}
