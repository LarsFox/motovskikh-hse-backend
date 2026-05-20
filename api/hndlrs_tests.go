package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

// hndlrGetTests возвращает список тестов пользователя.
func (m *Manager) hndlrGetTests(w http.ResponseWriter, r *http.Request) {
	userID := getFromCtxTester(r).ID

	tests, err := m.manager.GetTests(r.Context(), userID)
	switch {
	case errors.Is(err, nil):
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, tests)
}

// hndlrCreateTest создаёт новый тест.
func (m *Manager) hndlrCreateTest(w http.ResponseWriter, r *http.Request) {
	userID := getFromCtxTester(r).ID

	prms := &models.CreateTestV1Request{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	test, err := m.manager.CreateTest(r.Context(), userID, *prms.Title)
	switch {
	case errors.Is(err, nil):
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, test)
}

// hndlrDeleteTest удаляет тест.
func (m *Manager) hndlrDeleteTest(w http.ResponseWriter, r *http.Request) {
	userID := getFromCtxTester(r).ID

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	err = m.manager.DeleteTest(r.Context(), id, userID)
	switch {
	case errors.Is(err, nil):
	case errors.Is(err, entities.ErrNotFound):
		m.sendErrorPage(w, http.StatusNotFound)
		return
	default:
		notify(err)
		m.sendErrorPage(w, http.StatusInternalServerError)
		return
	}

	m.send(w, nil)
}
