package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"thermal-cycle-lab/internal/api"
	"thermal-cycle-lab/internal/catalog"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/domain"
	"thermal-cycle-lab/internal/engine"
	"thermal-cycle-lab/internal/store"
)

func invoke(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
func prepareTwoRuns(t *testing.T, h http.Handler) {
	t.Helper()
	for i := 1; i <= 2; i++ {
		r := invoke(t, h, http.MethodPost, "/api/specimens", fmt.Sprintf(`{"label":"parallel coupon %d","material":"ceramic","min_celsius":-40,"max_celsius":180}`, i))
		if r.Code != 201 {
			t.Fatalf("specimen %d status=%d body=%s", i, r.Code, r.Body.String())
		}
	}
	r := invoke(t, h, http.MethodPost, "/api/profiles", `{"name":"parallel read profile","stages":[{"name":"hold","target_celsius":30,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`)
	if r.Code != 201 {
		t.Fatalf("profile status=%d body=%s", r.Code, r.Body.String())
	}
	for i := 1; i <= 2; i++ {
		r = invoke(t, h, http.MethodPost, "/api/runs", fmt.Sprintf(`{"specimen_id":"sp-%03d","profile_id":"pr-001"}`, i))
		if r.Code != 201 {
			t.Fatalf("run %d status=%d body=%s", i, r.Code, r.Body.String())
		}
	}
}

func TestConcurrentRunReadsRemainIsolated(t *testing.T) {
	previousLogWriter := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(previousLogWriter) })
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	prepareTwoRuns(t, h)
	start := make(chan struct{})
	var wg sync.WaitGroup
	errCh := make(chan string, 256)
	for worker := 0; worker < 48; worker++ {
		want := fmt.Sprintf("run-%03d", worker%2+1)
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			<-start
			for attempt := 0; attempt < 40; attempt++ {
				r := invoke(t, h, http.MethodGet, "/api/runs/"+id, "")
				if r.Code != 200 {
					errCh <- fmt.Sprintf("path %s status=%d", id, r.Code)
					continue
				}
				var payload struct {
					Data domain.ExperimentRun `json:"data"`
				}
				if err := json.Unmarshal(r.Body.Bytes(), &payload); err != nil {
					errCh <- err.Error()
					continue
				}
				if payload.Data.ID != id {
					errCh <- fmt.Sprintf("path %s returned %s", id, payload.Data.ID)
					return
				}
			}
		}(want)
	}
	close(start)
	wg.Wait()
	close(errCh)
	for message := range errCh {
		t.Errorf("%s", message)
	}
}

func TestSerialRunReadsUseRequestedIdentifier(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	prepareTwoRuns(t, h)
	for _, id := range []string{"run-001", "run-002"} {
		r := invoke(t, h, http.MethodGet, "/api/runs/"+id, "")
		if r.Code != 200 || !bytes.Contains(r.Body.Bytes(), []byte(`"id":"`+id+`"`)) {
			t.Fatalf("path=%s status=%d body=%s", id, r.Code, r.Body.String())
		}
	}
}
