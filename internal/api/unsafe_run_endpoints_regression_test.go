package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/domain"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

func perform(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func setupUnsafeRun(t *testing.T, h http.Handler) string {
	t.Helper()
	steps := []struct {
		path, body string
		want       int
	}{{"/api/specimens", `{"label":"boundary coupon","material":"ceramic","min_celsius":-40,"max_celsius":180}`, 201}, {"/api/profiles", `{"name":"single stage profile","stages":[{"name":"hold","target_celsius":30,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`, 201}, {"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`, 201}}
	for _, s := range steps {
		r := perform(t, h, http.MethodPost, s.path, s.body)
		if r.Code != s.want {
			t.Fatalf("setup %s status=%d body=%s", s.path, r.Code, r.Body.String())
		}
	}
	return "run-001"
}

func TestRunEndpointsRejectUnsafeStateWithoutPanic(t *testing.T) {
	previousLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(previousLogWriter) })
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	runID := setupUnsafeRun(t, h)
	early := perform(t, h, http.MethodGet, "/api/runs/"+runID+"/report", "")
	if early.Code != http.StatusNotFound {
		t.Errorf("early report status=%d want=404 body=%s", early.Code, early.Body.String())
	}
	frame := perform(t, h, http.MethodPost, "/api/runs/"+runID+"/frames", `{"sensor_id":"probe-unsafe","sequence":1,"stage_index":7,"celsius":30,"humidity":40}`)
	if frame.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid stage status=%d want=422 body=%s", frame.Code, frame.Body.String())
	}
	detail := perform(t, h, http.MethodGet, "/api/runs/"+runID, "")
	var payload struct {
		Data domain.ExperimentRun `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.AcceptedFrames != 0 {
		t.Errorf("unsafe frame changed accepted_frames=%d", payload.Data.AcceptedFrames)
	}
	finalized := perform(t, h, http.MethodPost, "/api/runs/"+runID+"/finalize", `{}`)
	if finalized.Code == http.StatusInternalServerError {
		t.Errorf("unsafe input escaped validation and panicked during finalize: %s", finalized.Body.String())
	}
}

func TestMissingReportUsesNotFoundContract(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	id := setupUnsafeRun(t, h)
	r := perform(t, h, http.MethodGet, "/api/runs/"+id+"/report", "")
	if r.Code != 404 {
		t.Fatalf("status=%d body=%s", r.Code, r.Body.String())
	}
}
