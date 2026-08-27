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

func TestCanceledStartDoesNotConsumeSpecimen(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	for _, item := range []struct{ path, body string }{{"/api/specimens", `{"label":"nickel coupon","material":"nickel","min_celsius":-40,"max_celsius":180}`}, {"/api/profiles", `{"name":"thermal hold","stages":[{"name":"hold","target_celsius":40,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`}} {
		r := httptest.NewRequest(http.MethodPost, item.path, bytes.NewBufferString(item.body))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusCreated {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest(http.MethodPost, "/api/runs", bytes.NewBufferString(`{"specimen_id":"sp-001","profile_id":"pr-001"}`)).WithContext(ctx)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 499 {
		t.Fatalf("canceled start status=%d body=%s, want 499", w.Code, w.Body.String())
	}
	inventory := httptest.NewRecorder()
	h.ServeHTTP(inventory, httptest.NewRequest(http.MethodGet, "/api/inventory", nil))
	if !bytes.Contains(inventory.Body.Bytes(), []byte(`"state":"ready"`)) {
		t.Fatalf("specimen was consumed: %s", inventory.Body.String())
	}
}

func TestUncanceledStartStillCreatesRun(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"control nickel","material":"nickel","min_celsius":-40,"max_celsius":180}`},
		{"/api/profiles", `{"name":"control start","stages":[{"name":"hold","target_celsius":40,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`},
	} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, item.path, bytes.NewBufferString(item.body)))
		if w.Code != http.StatusCreated {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/runs", bytes.NewBufferString(`{"specimen_id":"sp-001","profile_id":"pr-001"}`)))
	if w.Code != http.StatusCreated {
		t.Fatalf("uncanceled start=%d body=%s, want 201", w.Code, w.Body.String())
	}
}
