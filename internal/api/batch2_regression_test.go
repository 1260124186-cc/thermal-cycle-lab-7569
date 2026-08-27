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

func TestRetiringActiveSpecimenKeepsValidationResponse(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, b string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(b))
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	for _, x := range []struct{ p, b string }{{"/api/specimens", `{"label":"aluminum plate","material":"aluminum","min_celsius":-20,"max_celsius":190}`}, {"/api/profiles", `{"name":"retire guard","stages":[{"name":"hold","target_celsius":45,"ramp_seconds":2,"hold_seconds":9,"tolerance":2}]}`}, {"/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`}} {
		if w := post(x.p, x.b); w.Code != 201 {
			t.Fatal(w.Body.String())
		}
	}
	w := post("/api/specimens/sp-001/retire", "")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("retire active specimen=%d body=%s, want 422", w.Code, w.Body.String())
	}
}

func TestReadySpecimenCanRetireNormally(t *testing.T) {
	repo := store.NewMemory()
	h := api.New(catalog.New(repo, clock.System{}), engine.New(repo, clock.System{})).Handler()
	post := func(path, b string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(b)))
		return w
	}
	if w := post("/api/specimens", `{"label":"retire control","material":"aluminum","min_celsius":-20,"max_celsius":190}`); w.Code != http.StatusCreated {
		t.Fatalf("create specimen=%d %s", w.Code, w.Body.String())
	}
	if w := post("/api/specimens/sp-001/retire", ""); w.Code != http.StatusOK {
		t.Fatalf("ready retire=%d body=%s, want 200", w.Code, w.Body.String())
	}
}
