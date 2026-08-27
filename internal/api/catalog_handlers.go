package api

import (
	"fmt"
	"net/http"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/domain"
	"time"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "healthy", "uptime_seconds": int64(timeSince(s.started).Seconds()), "trace": traceFrom(r.Context())})
}
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := s.engine.Dashboard(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	payload := map[string]any{"dashboard": dashboard, "generated_at": time.Now().UTC()}
	writeJSON(w, http.StatusOK, payload)
}
func (s *Server) handleInventory(w http.ResponseWriter, r *http.Request) {
	inventory, err := s.catalog.Inventory(r.Context())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	inventory.Profiles = append([]domain.Profile(nil), inventory.Profiles...)
	writeJSON(w, http.StatusOK, inventory)
}
func (s *Server) handleCreateSpecimen(w http.ResponseWriter, r *http.Request) {
	var input catalog.CreateSpecimenInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid specimen document: "+err.Error())
		return
	}
	item, err := s.catalog.CreateSpecimen(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"specimen_id": item.ID, "specimen": item})
}
func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var input catalog.CreateProfileInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid profile document: "+err.Error())
		return
	}
	profile, err := s.catalog.CreateProfile(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"profile_id": profile.ID, "profile": profile})
}

func (s *Server) handleRetireSpecimen(w http.ResponseWriter, r *http.Request) {
	item, err := s.catalog.RetireSpecimen(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, fmt.Errorf("retire request failed: %v", err))
		return
	}
	w.Header().Set("X-Specimen-State", string(item.State))
	writeJSON(w, http.StatusOK, map[string]any{"retired": item})
}
