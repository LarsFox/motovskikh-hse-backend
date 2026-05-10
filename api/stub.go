package api

import (
	"fmt"
	"html"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrStubGet(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(
		w,
		"sup, db %v, user-agent %s",
		m.manager.Stub(),
		html.EscapeString(r.Header.Get("User-Agent")),
	)
	if err != nil {
		entities.Notify(err)
	}
}

func (m *Manager) hndlrStubPost(w http.ResponseWriter, r *http.Request) {
	prms := &models.GetTestRequest{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	m.send(w, &models.GetTestResponse{Settings: &models.Settings{CanvasPath: r.Header.Get("User-Agent"), Hymn: true}})
}
