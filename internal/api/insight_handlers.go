package api

import (
	"net/http"
	"strconv"
)

func (s *Server) handleProfileInsights(w http.ResponseWriter, r *http.Request) {
	insights, err := s.catalog.ProfileInsights(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"profiles": insights, "count": len(insights), "trace_id": traceFrom(r.Context())})
}
func (s *Server) handleCatalogHealth(w http.ResponseWriter, r *http.Request) {
	health, err := s.catalog.Health(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	status := http.StatusOK
	if len(health.Warnings) > 0 {
		w.Header().Set("X-Catalog-Warnings", strconv.Itoa(len(health.Warnings)))
	}
	writeJSON(w, status, health)
}
func (s *Server) handleRunSummary(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "run id is required")
		return
	}
	summary, err := s.engine.SummarizeRun(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"summary": summary})
}
func (s *Server) handleSensorHealth(w http.ResponseWriter, r *http.Request) {
	spread := 2.0
	if raw := r.URL.Query().Get("max_spread"); raw != "" {
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "max_spread must be numeric")
			return
		}
		spread = value
	}
	health, err := s.engine.SensorHealth(r.Context(), r.PathValue("id"), spread)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, health)
}
func (s *Server) handleRankedRuns(w http.ResponseWriter, r *http.Request) {
	ids, err := s.engine.RankCompletedRuns(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	w.Header().Set("X-Run-Count", strconv.Itoa(len(ids)))
	writeJSON(w, http.StatusOK, map[string]any{"run_ids": ids, "count": len(ids)})
}
