package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux" // Роутер.
	httpSwagger "github.com/swaggo/http-swagger" // Сваггер.
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
	m.router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
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
		routeGet("/stub/get/", m.hndlrStubGet),
		routePost("/stub/post/", m.hndlrStubPost),
		routeGet("/doc.json", m.hndlrSwaggerJSON),
		// Новые маршруты для статистики.
		routePost("/stats/save/", m.hndlrSaveAttempt),
		routePost("/stats/analysis/", m.hndlrGetAnalysis),
		routePost("/dev/create-test-data/", m.hndlrCreateTestData),
	})
	log.Println("Routes registered: /stats/save/, /stats/analysis/")
	m.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
        httpSwagger.URL("/doc.json"),
    ))
	m.router.HandleFunc("/doc.json", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./doc.json")
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

func (m *Manager) hndlrSwaggerJSON(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./doc.json")
}