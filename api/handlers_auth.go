package api

import (
	"errors"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrEnjoy(w http.ResponseWriter, r *http.Request) {
	prms := &models.EnjoyRequest{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	tokens, err := m.manager.Enjoy(r.Context(), *prms.Name, *prms.Pass)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrEmailNotVerified):
		w.WriteHeader(http.StatusNoContent)
		return
	case errors.Is(err, entities.ErrInvalidInput):
		m.sendErrorPage(w, http.StatusUnauthorized)
		return
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.sendTokens(w, tokens)
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
	case errors.Is(err, entities.ErrInvalidInput):
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	case errors.Is(err, entities.ErrNotFound):
		m.sendErrorPage(w, http.StatusNotFound)
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

func (m *Manager) hndlrResendCode(w http.ResponseWriter, r *http.Request) {
	prms := &models.ResendCodeV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	err := m.manager.ResendCode(r.Context(), *prms.Email)
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

// hndlrRefreshToken refreshes the access token with the refresh token.
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

func (m *Manager) hndlrForgotPassword(w http.ResponseWriter, r *http.Request) {
	prms := &models.ForgotPasswordV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	err := m.manager.ForgotPassword(r.Context(), *prms.Email)
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

func (m *Manager) hndlrResetPassword(w http.ResponseWriter, r *http.Request) {
	prms := &models.ResetPasswordV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	err := m.manager.ResetPassword(r.Context(), *prms.Email, *prms.Code, *prms.Password)
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

func (m *Manager) hndlrFarewell(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
	})

	m.send(w, nil)
}
