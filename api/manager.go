package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/LarsFox/motovskikh-hse-backend/manager"
)

const (
	defaultReadTimeout  = time.Second * 15
	defaultWriteTimeout = time.Second * 30
	defaultIdleTimeout  = time.Second * 30
)

type Manager struct {
	manager *manager.Manager
	router  *mux.Router
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
		manager: manager,
		router:  mux.NewRouter().StrictSlash(true),
	}

	m.addRoutes()
	
	// Swagger
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)
	
	m.router.PathPrefix("/swagger/").Handler(swaggerHandler)
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
		routePost("/tests/submit/", m.hndlrSubmitTest, m.wrapContentTypeJSON), // После окончания теста получить анализ.
	})
	m.router.HandleFunc("/doc.json", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./doc.json")
    })
    
    // Swagger UI
    m.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
        httpSwagger.URL("/doc.json"),
    ))
	log.Println("Routes registered")
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
