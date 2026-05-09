package api

import (
	"encoding/json"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/internal/models"
)

func (m *Manager) hndlrRegister(w http.ResponseWriter, r *http.Request) {

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := m.manager.Auth().Register(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	m.send(w, map[string]string{"message": "verification code sent to email"})
}

// hndlrVerifyEmail в ответе за подтверждение email по коду
func (m *Manager) hndlrVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := m.manager.Auth().VerifyEmail(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	m.send(w, map[string]string{"message": "email verified successfully"})
}

// hndlrSignIn запускает в личный кабинет
func (m *Manager) hndlrSignIn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := m.manager.Auth().SignIn(req.Email, req.Password)
	if err != nil {
		m.sendError(w, http.StatusUnauthorized, err.Error())
		return
	}

	tokens, err := m.manager.Token().GeneratePair(user.ID)
	if err != nil {
		m.sendError(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	m.send(w, tokens)
}

// hndlrRefresh обновляет access токен по refresh токену
func (m *Manager) hndlrRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokens, err := m.manager.Token().Refresh(req.RefreshToken)
	if err != nil {
		m.sendError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	m.send(w, tokens)
}

// hndlrResendCode в ответе за повторную отправку кода подтверждения
func (m *Manager) hndlrResendCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		m.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := m.manager.Auth().ResendCode(req.Email); err != nil {
		m.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	m.send(w, map[string]string{"message": "verification code resent"})
}
