package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

type responseEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func sendJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func createRunningExperiment(t *testing.T, h http.Handler) string {
	t.Helper()
	specimen := sendJSON(t, h, http.MethodPost, "/api/specimens", `{"label":"ceramic coupon","material":"ceramic","min_celsius":-40,"max_celsius":180}`)
	if specimen.Code != http.StatusCreated {
		t.Fatalf("create specimen: status=%d body=%s", specimen.Code, specimen.Body.String())
	}
	profile := sendJSON(t, h, http.MethodPost, "/api/profiles", `{"name":"steady qualification","stages":[{"name":"hold","target_celsius":30,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`)
	if profile.Code != http.StatusCreated {
		t.Fatalf("create profile: status=%d body=%s", profile.Code, profile.Body.String())
	}
	run := sendJSON(t, h, http.MethodPost, "/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`)
	if run.Code != http.StatusCreated {
		t.Fatalf("start run: status=%d body=%s", run.Code, run.Body.String())
	}
	return "run-001"
}

func TestDuplicateFrameDoesNotPoisonRunningState(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	runID := createRunningExperiment(t, h)
	first := sendJSON(t, h, http.MethodPost, "/api/runs/"+runID+"/frames", `{"sensor_id":"probe-a","sequence":1,"stage_index":0,"celsius":30,"humidity":40}`)
	if first.Code != http.StatusAccepted {
		t.Fatalf("first frame: status=%d body=%s", first.Code, first.Body.String())
	}

	duplicate := sendJSON(t, h, http.MethodPost, "/api/runs/"+runID+"/frames", `{"sensor_id":"probe-a","sequence":1,"stage_index":0,"celsius":30.2,"humidity":41}`)
	if duplicate.Code != http.StatusUnprocessableEntity {
		t.Errorf("duplicate status=%d, want 422; body=%s", duplicate.Code, duplicate.Body.String())
	}

	detail := sendJSON(t, h, http.MethodGet, "/api/runs/"+runID, "")
	if detail.Code != http.StatusOK {
		t.Fatalf("get run: status=%d body=%s", detail.Code, detail.Body.String())
	}
	var outer struct {
		Data struct {
			State    string `json:"state"`
			Accepted int    `json:"accepted_frames"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &outer); err != nil {
		t.Fatal(err)
	}
	if outer.Data.State != "running" || outer.Data.Accepted != 1 {
		t.Errorf("run after duplicate: state=%s accepted=%d, want running/1", outer.Data.State, outer.Data.Accepted)
	}

	next := sendJSON(t, h, http.MethodPost, "/api/runs/"+runID+"/frames", `{"sensor_id":"probe-a","sequence":2,"stage_index":0,"celsius":29.9,"humidity":39}`)
	if next.Code != http.StatusAccepted {
		t.Errorf("next frame: status=%d body=%s", next.Code, next.Body.String())
	}
	if t.Failed() {
		fmt.Println("duplicate frame rejection changed the public error contract or persisted run state")
	}
}

func TestRunningExperimentDetailRemainsReadable(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	runID := createRunningExperiment(t, h)
	detail := sendJSON(t, h, http.MethodGet, "/api/runs/"+runID, "")
	if detail.Code != http.StatusOK || !bytes.Contains(detail.Body.Bytes(), []byte(`"state":"running"`)) {
		t.Fatalf("running detail: status=%d body=%s", detail.Code, detail.Body.String())
	}
}
