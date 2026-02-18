package api

import (
	"fmt"
	"net/http"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
)

// sendErrorPage возвращает страницу ошибки.
func (m *Manager) sendErrorPage(w http.ResponseWriter, code int) {
	w.WriteHeader(code)

	_, err := fmt.Fprintf(w, "nope, %d", code)
	if err != nil {
		entities.Notify(err)
	}
}
