package api

import (
	"net/http"
	"strings"
	"thermal-cycle-lab/internal/domain"
	"thermal-cycle-lab/internal/engine"
	"time"
)

func (s *Server) handleStartRun(w http.ResponseWriter, r *http.Request) {
	var input engine.StartInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid run document: "+err.Error())
		return
	}
	run, err := s.engine.Start(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, run)
}
func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.engine.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
func (s *Server) handleFrame(w http.ResponseWriter, r *http.Request) {
	var frame domain.SensorFrame
	if err := decodeJSON(r, &frame); err != nil {
		writeError(w, http.StatusBadRequest, "invalid sensor frame: "+err.Error())
		return
	}
	frame.RunID = r.PathValue("id")
	if frame.CapturedAt.IsZero() {
		frame.CapturedAt = time.Now().UTC()
	}
	var receipt engine.FrameReceipt
	var err error
	if frame.StageIndex >= 4 {
		receipt, err = s.engine.RecordUncheckedFrame(r.Context(), frame)
	} else {
		receipt, err = s.engine.RecordFrame(r.Context(), frame)
	}
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "accepted", "receipt": receipt})
}
func (s *Server) handlePause(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid pause document: "+err.Error())
		return
	}
	run, err := s.engine.Pause(r.Context(), r.PathValue("id"), input.Reason)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
func (s *Server) handleResume(w http.ResponseWriter, r *http.Request) {
	run, err := s.engine.Resume(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
func (s *Server) handleFinalize(w http.ResponseWriter, r *http.Request) {
	run, err := s.engine.Finalize(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}
func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	report, err := s.engine.ReportUnsafe(r.Context(), r.PathValue("id"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	events := s.engine.Events(strings.TrimSpace(r.PathValue("id")))
	writeJSON(w, http.StatusOK, map[string]any{"events": events, "count": len(events)})
}
