package api_test

import (
	"bytes"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

var verifyPublicEntryPoint = flag.String(
	"verify-public-entry-point",
	"",
	"public endpoint explicitly covered by the focused verification command",
)

func TestMissingReportKeepsNotFoundContract(t *testing.T) {
	const entryPoint = "GET /api/runs/{id}/report"
	if got := *verifyPublicEntryPoint; got != entryPoint {
		t.Fatalf("verify public entry point = %q, want %q", got, entryPoint)
	}
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		return w
	}
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"missing report alloy","material":"alloy","min_celsius":-20,"max_celsius":180}`},
		{"/api/profiles", `{"name":"missing report profile","stages":[{"name":"hold","target_celsius":50,"ramp_seconds":3,"hold_seconds":8,"tolerance":2}]}`},
		{"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`},
	} {
		if w := post(item.path, item.body); w.Code != http.StatusCreated {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runs/run-001/report", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing report=%d body=%s, want 404", w.Code, w.Body.String())
	}
}

func TestCompletedRunReportRemainsAvailable(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body)))
		return w
	}
	for _, item := range []struct{ path, body string }{
		{"/api/specimens", `{"label":"report ready alloy","material":"alloy","min_celsius":-20,"max_celsius":180}`},
		{"/api/profiles", `{"name":"report ready profile","stages":[{"name":"hold","target_celsius":50,"ramp_seconds":3,"hold_seconds":8,"tolerance":2}]}`},
		{"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`},
		{"/api/runs/run-001/frames", `{"sensor_id":"report-1","sequence":1,"stage_index":0,"celsius":50,"humidity":30}`},
		{"/api/runs/run-001/finalize", ``},
	} {
		if w := post(item.path, item.body); w.Code != http.StatusCreated && w.Code != http.StatusAccepted && w.Code != http.StatusOK {
			t.Fatalf("setup %s: %d %s", item.path, w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runs/run-001/report", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("available report=%d body=%s, want 200", w.Code, w.Body.String())
	}
}
