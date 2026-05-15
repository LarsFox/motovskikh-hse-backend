package api

import (
	"errors"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

// hndlrEnjoy запускает в личный кабинет.
func (m *Manager) hndlrEnjoy(w http.ResponseWriter, r *http.Request) {
	prms := &models.EnjoyRequest{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	tokens, err := m.manager.Enjoy(r.Context(), *prms.Name, *prms.Pass)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.sendTokens(w, tokens)
}

// hndlrRefreshToken обновляет аксес-токен по рефреш-токену.
func (m *Manager) hndlrRefreshToken(w http.ResponseWriter, r *http.Request) {
	prms := &models.RefreshTokenV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	tokens, err := m.manager.RefreshToken(r.Context(), *prms.RefreshToken)
	if err != nil {
		m.sendErrorPage(w, http.StatusUnauthorized)
		return
	}

	m.sendTokens(w, tokens)
}
