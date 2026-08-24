package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

func call(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func requireStatus(t *testing.T, r *httptest.ResponseRecorder, want int, label string) {
	t.Helper()
	if r.Code != want {
		t.Fatalf("%s status=%d want=%d body=%s", label, r.Code, want, r.Body.String())
	}
}
func provisionSequentialRuns(t *testing.T, h http.Handler) (string, string) {
	t.Helper()
	for i := 1; i <= 2; i++ {
		requireStatus(t, call(t, h, http.MethodPost, "/api/specimens", fmt.Sprintf(`{"label":"coupon %d sample","material":"ceramic","min_celsius":-40,"max_celsius":180}`, i)), 201, "specimen")
	}
	requireStatus(t, call(t, h, http.MethodPost, "/api/profiles", `{"name":"sequential cycle","stages":[{"name":"hold","target_celsius":30,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`), 201, "profile")
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`), 201, "run one")
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs/run-001/frames", `{"sensor_id":"probe-one","sequence":1,"stage_index":0,"celsius":30,"humidity":40}`), 202, "frame one")
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs", `{"specimen_id":"sp-002","profile_id":"pr-001"}`), 201, "run two")
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs/run-002/frames", `{"sensor_id":"probe-two","sequence":1,"stage_index":0,"celsius":30,"humidity":40}`), 202, "frame two")
	return "run-001", "run-002"
}

func TestSequentialRunsOwnIndependentCompletionResources(t *testing.T) {
	previousLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(previousLogWriter) })
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	first, second := provisionSequentialRuns(t, h)
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs/"+first+"/finalize", `{}`), 200, "first finalize")
	secondFinalize := call(t, h, http.MethodPost, "/api/runs/"+second+"/finalize", `{}`)
	if secondFinalize.Code != http.StatusOK {
		t.Errorf("second finalize status=%d body=%s", secondFinalize.Code, secondFinalize.Body.String())
	}
	for _, id := range []string{first, second} {
		r := call(t, h, http.MethodGet, "/api/runs/"+id+"/report", "")
		if r.Code != 200 {
			t.Errorf("report %s status=%d body=%s", id, r.Code, r.Body.String())
			continue
		}
		var v struct {
			Data struct {
				RunID string `json:"run_id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(r.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if v.Data.RunID != id {
			t.Errorf("report path=%s returned=%s", id, v.Data.RunID)
		}
	}
}

func TestSingleRunCompletionStillProducesReport(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	first, _ := provisionSequentialRuns(t, h)
	requireStatus(t, call(t, h, http.MethodPost, "/api/runs/"+first+"/finalize", `{}`), 200, "single finalize")
	requireStatus(t, call(t, h, http.MethodGet, "/api/runs/"+first+"/report", ""), 200, "single report")
}
