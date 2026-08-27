package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

func TestCanceledFinalizeLeavesRunAndSpecimenActive(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	send := func(method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, x := range []struct{ p, b string }{{"/api/specimens", `{"label":"copper plate","material":"copper","min_celsius":-50,"max_celsius":160}`}, {"/api/profiles", `{"name":"short hold","stages":[{"name":"hold","target_celsius":35,"ramp_seconds":2,"hold_seconds":7,"tolerance":2}]}`}, {"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`}, {"/api/runs/run-001/frames", `{"sensor_id":"p1","sequence":1,"stage_index":0,"celsius":35,"humidity":30}`}} {
		if w := send(http.MethodPost, x.p, x.b); w.Code != http.StatusCreated && w.Code != http.StatusAccepted {
			t.Fatalf("setup %s: %d %s", x.p, w.Code, w.Body.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest(http.MethodPost, "/api/runs/run-001/finalize", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 499 {
		t.Fatalf("canceled finalize=%d %s", w.Code, w.Body.String())
	}
	detail := send(http.MethodGet, "/api/runs/run-001", "")
	if !bytes.Contains(detail.Body.Bytes(), []byte(`"state":"running"`)) {
		t.Fatalf("run transitioned after canceled finalize: %s", detail.Body.String())
	}
}

func TestUncanceledFinalizeCompletesAndReleasesSpecimen(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	send := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
		return w
	}
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"control copper","material":"copper","min_celsius":-50,"max_celsius":160}`},
		{"/api/profiles", `{"name":"control finalize","stages":[{"name":"hold","target_celsius":35,"ramp_seconds":2,"hold_seconds":7,"tolerance":2}]}`},
		{"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`},
		{"/api/runs/run-001/frames", `{"sensor_id":"p1","sequence":1,"stage_index":0,"celsius":35,"humidity":30}`},
	} {
		if w := send(http.MethodPost, item.path, item.body); w.Code != http.StatusCreated && w.Code != http.StatusAccepted {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	if w := send(http.MethodPost, "/api/runs/run-001/finalize", ""); w.Code != http.StatusOK {
		t.Fatalf("normal finalize=%d body=%s, want 200", w.Code, w.Body.String())
	}
	inventory := send(http.MethodGet, "/api/inventory", "")
	if !bytes.Contains(inventory.Body.Bytes(), []byte(`"state":"ready"`)) {
		t.Fatalf("specimen was not released: %s", inventory.Body.String())
	}
}
