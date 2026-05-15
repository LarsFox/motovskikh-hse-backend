package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/LarsFox/motovskikh-hse-backend/entities"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
)

const (
	defaultReadTimeout  = time.Second * 15
	defaultWriteTimeout = time.Second * 30
	defaultIdleTimeout  = time.Second * 30
)

// Manager is an API manager and listener.
type Manager struct {
	manager *manager.Manager
	router  *mux.Router
}

// route is a single path for a mux handler.
type route struct {
	Method   string
	Path     string
	Handler  http.HandlerFunc
	Wrappers []wrapper
}

func routeGet(path string, handler http.HandlerFunc, wrappers ...wrapper) route {
	return newRoute(http.MethodGet, path, handler, wrappers...)
}

func routePost(path string, handler http.HandlerFunc, wrappers ...wrapper) route {
	return newRoute(http.MethodPost, path, handler, wrappers...)
}

func newRoute(method, path string, handler http.HandlerFunc, wrappers ...wrapper) route {
	return route{
		method,
		path,
		handler,
		wrappers,
	}
}

func NewManager(manager *manager.Manager) *Manager {
	m := &Manager{
		manager: manager,
		router:  mux.NewRouter().StrictSlash(true),
	}

	m.addRoutes()

	return m
}

// Listen запускает сервер на указанном порту.
func (m *Manager) Listen(addr string) error {
	log.Println("API started on addr", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      m.router,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	return server.ListenAndServe()
}

func (m *Manager) addRoutes() {
	m.addHandlers([]route{
		routePost("/api/auth/register", m.hndlrRegister, m.wrapContentTypeJSON),
		routePost("/api/auth/verify-email", m.hndlrVerifyEmail, m.wrapContentTypeJSON),
		routePost("/api/auth/sign-in", m.hndlrEnjoy, m.wrapContentTypeJSON),
		routePost("/api/auth/refresh", m.hndlrRefreshToken, m.wrapContentTypeJSON),
	})
}

// addHandlers добавляет пути и обработчики запросов в мультиплексор (mux).
func (m *Manager) addHandlers(routes []route) {
	essentialWrappers := []wrapper{m.wrapBodyMaxSize, m.wrapEasterEggHeader, wrapRecover}
	for _, r := range routes {
		var wrapper http.Handler = r.Handler
		for _, w := range r.Wrappers {
			wrapper = w(wrapper)
		}
		for _, w := range essentialWrappers {
			wrapper = w(wrapper)
		}
		m.router.Methods(r.Method).Path(r.Path).Handler(wrapper)
	}
}

// send responds with a success.
func (m *Manager) send(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := map[string]any{
		"ok":     true,
		"result": data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		notify(err)
	}
}

// sendTokens for cookie
func (m *Manager) sendTokens(w http.ResponseWriter, tokens *entities.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   900,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Path:     "/api/auth/refresh",
		MaxAge:   365 * 24 * 60 * 60,
	})

	m.send(w, nil)
}
