package api

import (
	"fmt"
	"net/http"
)

// sendErrorPage возвращает страницу ошибки.
func (m *Manager) sendErrorPage(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("nope, %d", code)))
}
