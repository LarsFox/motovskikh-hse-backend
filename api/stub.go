package api

import (
	"fmt"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/generated/models"
)

func (m *Manager) hndlrStubGet(w http.ResponseWriter, r *http.Request) {
	sup := fmt.Sprintf("sup, db %v, user-agent %s", m.manager.Stub(), r.Header.Get("User-Agent"))

	w.Write([]byte(sup))
}

func (m *Manager) hndlrStubPost(w http.ResponseWriter, r *http.Request) {
	prms := &models.GetTestRequest{}
	if err := unmarshalParams(r, prms); err != nil {
		m.sendErrorPage(w, http.StatusBadRequest)
		return
	}

	m.send(w, &models.GetTestResponse{Settings: &models.Settings{CanvasPath: r.Header.Get("User-Agent"), Hymn: true}})
}
