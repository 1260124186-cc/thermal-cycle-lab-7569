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

func TestCreateSpecimenEndpoint(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	server := api.New(catalog.New(repo, source), engine.New(repo, source))
	request := httptest.NewRequest(http.MethodPost, "/api/specimens", bytes.NewBufferString(`{"label":"ceramic tile","material":"ceramic","min_celsius":-30,"max_celsius":180}`))
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if !bytes.Contains(response.Body.Bytes(), []byte("specimen_id")) {
		t.Fatalf("missing specimen id: %s", response.Body.String())
	}
}
