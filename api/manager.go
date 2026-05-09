package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/LarsFox/motovskikh-hse-backend/manager"
)

const (
	defaultReadTimeout  = time.Second * 15
	defaultWriteTimeout = time.Second * 30
	defaultIdleTimeout  = time.Second * 30
)

type Manager struct {
	manager        *manager.Manager
	router         *mux.Router
	sessionManager *SessionManager
}

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
		manager:        manager,
		router:         mux.NewRouter().StrictSlash(true),
		sessionManager: NewSessionManager(),
	}

	m.addRoutes()

	return m
}

func (m *Manager) Listen(addr string) error {
	log.Println("API started on addr", addr)

	handler := wrapCORS(m.router)

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	return server.ListenAndServe()
}

func (m *Manager) addRoutes() {
	m.addHandlers([]route{
		routeGet("/api/v1/stub/get", m.hndlrStubGet),
		routePost("/api/v1/stub/post", m.hndlrStubPost, m.wrapContentTypeJSON),
		//routeGet("/api/v1/hello", m.hndlrHello),
		routeGet("/api/v1/graph", m.hndlrGraph),
		routeGet("/api/v1/isomorphism/round", m.hndlrIsomorphismRound),
		routePost("/api/v1/isomorphism/start", m.hndlrStartGame),
		//routePost("/api/v1/isomorphism/start", m.hndlrCheckIsomorphism, m.wrapContentTypeJSON), // новый маршрут
		//routePost("/api/v1/isomorphism/submit", m.hndlrSubmitAnswer, m.wrapContentTypeJSON),
		routePost("/api/v1/isomorphism/submit", m.hndlrSubmitAnswer),
		routeGet("/api/v1/debug/sessions", m.hndlrDebugSessions),
		routePost("/api/v1/isomorphism/end", m.hndlrEndGame),
		routePost("/api/v1/isomorphism/confirm", m.hndlrConfirm),
		routeGet("/api/v1/find_way/start", m.hndlrFindWayStart),
		routePost("/api/v1/find_way/start", m.hndlrFindWayStart),
		routePost("/api/v1/find_way/submit", m.hndlrFindWaySubmit),
		routePost("/api/v1/find_way/confirm", m.hndlrFindWayConfirm),
		routePost("/api/v1/find_way/end", m.hndlrFindWayEnd),
		routePost("/api/v1/escape/start", m.hndlrEscapeStart),
		routePost("/api/v1/escape/submit", m.hndlrEscapeSubmit),
		routePost("/api/v1/escape/confirm", m.hndlrEscapeConfirm),
		routePost("/api/v1/escape/end", m.hndlrEscapeEnd),
	
	})

	//m.router.HandleFunc("/api/v1/isomorphism/submit", m.hndlrSubmitAnswer).Methods("POST")
}

func (m *Manager) addHandlers(routes []route) {
	essentialWrappers := []wrapper{
		m.wrapBodyMaxSize,
		m.wrapEasterEggHeader,
		wrapRecover,
		wrapCORS,
	}
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

func (m *Manager) send(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := map[string]interface{}{
		"ok":     true,
		"result": data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		notify(err)
	}
}
