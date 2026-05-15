package api

import (
	"fmt"
	"net/http"
)

// sendErrorPage возвращает страницу ошибки.
func (m *Manager) sendErrorPage(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	_, err := fmt.Fprintf(w, "nope, %d", code)
	if err != nil {
		notify(err)
	}
}
