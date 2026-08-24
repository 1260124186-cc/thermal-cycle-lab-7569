package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

func TestFinalizeWithoutFramesKeepsValidationContract(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		return w
	}
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"empty evidence alloy","material":"alloy","min_celsius":-30,"max_celsius":180}`},
		{"/api/profiles", `{"name":"empty evidence profile","stages":[{"name":"hold","target_celsius":45,"ramp_seconds":3,"hold_seconds":8,"tolerance":2}]}`},
		{"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`},
	} {
		if w := post(item.path, item.body); w.Code != http.StatusCreated {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	if w := post("/api/runs/run-001/finalize", ""); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty finalize=%d body=%s, want 422", w.Code, w.Body.String())
	}
	detail := httptest.NewRecorder()
	h.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/runs/run-001", nil))
	if !bytes.Contains(detail.Body.Bytes(), []byte(`"state":"running"`)) {
		t.Fatalf("failed finalization changed run: %s", detail.Body.String())
	}
}

func TestFinalizeWithFramesStillCompletes(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		return w
	}
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"evidence alloy","material":"alloy","min_celsius":-30,"max_celsius":180}`},
		{"/api/profiles", `{"name":"evidence profile","stages":[{"name":"hold","target_celsius":45,"ramp_seconds":3,"hold_seconds":8,"tolerance":2}]}`},
		{"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`},
		{"/api/runs/run-001/frames", `{"sensor_id":"evidence-1","sequence":1,"stage_index":0,"celsius":45,"humidity":30}`},
	} {
		if w := post(item.path, item.body); w.Code != http.StatusCreated && w.Code != http.StatusAccepted {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	if w := post("/api/runs/run-001/finalize", ""); w.Code != http.StatusOK {
		t.Fatalf("finalize with evidence=%d %s, want 200", w.Code, w.Body.String())
	}
}
