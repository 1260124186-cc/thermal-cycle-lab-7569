package api

import (
	"net/http"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/engine"
	"time"
)

type Server struct {
	catalog   *catalog.Service
	engine    *engine.Controller
	mux       *http.ServeMux
	started   time.Time
	lastRunID string
}

func New(catalogService *catalog.Service, controller *engine.Controller) *Server {
	s := &Server{catalog: catalogService, engine: controller, mux: http.NewServeMux(), started: time.Now().UTC()}
	s.registerRoutes()
	return s
}
func (s *Server) Handler() http.Handler { return requestTrace(recoverPanic(s.mux)) }
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/dashboard", s.handleDashboard)
	s.mux.HandleFunc("GET /api/inventory", s.handleInventory)
	s.mux.HandleFunc("GET /api/profile-insights", s.handleProfileInsights)
	s.mux.HandleFunc("GET /api/catalog-health", s.handleCatalogHealth)
	s.mux.HandleFunc("GET /api/ranked-runs", s.handleRankedRuns)
	s.mux.HandleFunc("POST /api/specimens", s.handleCreateSpecimen)
	s.mux.HandleFunc("POST /api/specimens/{id}/retire", s.handleRetireSpecimen)
	s.mux.HandleFunc("POST /api/profiles", s.handleCreateProfile)
	s.mux.HandleFunc("POST /api/runs", s.handleStartRun)
	s.mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
	s.mux.HandleFunc("GET /api/runs/{id}/summary", s.handleRunSummary)
	s.mux.HandleFunc("GET /api/runs/{id}/sensor-health", s.handleSensorHealth)
	s.mux.HandleFunc("POST /api/runs/{id}/frames", s.handleFrame)
	s.mux.HandleFunc("POST /api/runs/{id}/pause", s.handlePause)
	s.mux.HandleFunc("POST /api/runs/{id}/resume", s.handleResume)
	s.mux.HandleFunc("POST /api/runs/{id}/finalize", s.handleFinalize)
	s.mux.HandleFunc("GET /api/runs/{id}/report", s.handleReport)
	s.mux.HandleFunc("GET /api/runs/{id}/events", s.handleEvents)
}
