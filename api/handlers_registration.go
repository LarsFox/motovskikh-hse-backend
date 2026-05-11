package api

import (
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrRegister(w http.ResponseWriter, r *http.Request) {
	prms := &models.RegisterV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	// TODO: switch err
	if err := m.manager.Register(r.Context(), *prms.Email, *prms.Password); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	m.send(w, nil)
}

// hndlrVerifyEmail в ответе за подтверждение email по коду.
func (m *Manager) hndlrVerifyEmail(w http.ResponseWriter, r *http.Request) {
	prms := &models.VerifyEmailV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	// TODO: наполнить из параметров
	// TODO: switch err
	if err := m.manager.VerifyEmail(r.Context(), prms.Email, prms.Code); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	// TODO: имейл успешный, выдаем сразу ему пару токенов, чтобы сразу был залогинен.
	m.send(w, nil)
}
