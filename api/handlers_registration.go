package api

import (
	"errors"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrRegister(w http.ResponseWriter, r *http.Request) {
	prms := &models.RegisterV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest) // 400
		return
	}

	err := m.manager.Register(r.Context(), *prms.Email, *prms.Password)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrInvalidInput):
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, nil)
}

// hndlrVerifyEmail в ответе за подтверждение email по ссылке-коду.
func (m *Manager) hndlrVerifyEmail(w http.ResponseWriter, r *http.Request) {
	prms := &models.VerifyEmailV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	err := m.manager.VerifyEmail(r.Context(), *prms.Email, *prms.Code)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrInvalidInput), errors.Is(err, entities.ErrNotFound):
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	// Email подтверждён — выдаём токены
	tokens, err := m.manager.GenerateTokensByEmail(r.Context(), *prms.Email)
	if err != nil {
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError) // 500
		return
	}

	m.sendTokens(w, tokens)
}
