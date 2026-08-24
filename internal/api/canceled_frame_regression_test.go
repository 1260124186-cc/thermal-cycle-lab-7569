package api_test

import (
	"bytes"
	"context"
	"encoding/json"
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

func requestJSON(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func prepareRun(t *testing.T, h http.Handler) string {
	t.Helper()
	if r := requestJSON(t, h, http.MethodPost, "/api/specimens", `{"label":"alloy coupon","material":"alloy","min_celsius":-40,"max_celsius":180}`); r.Code != 201 {
		t.Fatalf("specimen status=%d body=%s", r.Code, r.Body.String())
	}
	if r := requestJSON(t, h, http.MethodPost, "/api/profiles", `{"name":"cancellation profile","stages":[{"name":"hold","target_celsius":25,"ramp_seconds":5,"hold_seconds":20,"tolerance":2}]}`); r.Code != 201 {
		t.Fatalf("profile status=%d body=%s", r.Code, r.Body.String())
	}
	if r := requestJSON(t, h, http.MethodPost, "/api/runs", `{"specimen_id":"sp-001","profile_id":"pr-001"}`); r.Code != 201 {
		t.Fatalf("run status=%d body=%s", r.Code, r.Body.String())
	}
	return "run-001"
}

func TestCanceledFrameRequestLeavesNoPersistentEffects(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	runID := prepareRun(t, h)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/api/runs/"+runID+"/frames", bytes.NewBufferString(`{"sensor_id":"probe-cancel","sequence":1,"stage_index":0,"celsius":25,"humidity":42}`)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 499 {
		t.Errorf("canceled frame status=%d want=499 body=%s", rec.Code, rec.Body.String())
	}

	detail := requestJSON(t, h, http.MethodGet, "/api/runs/"+runID, "")
	var payload struct {
		Data domain.ExperimentRun `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.AcceptedFrames != 0 || payload.Data.State != domain.RunRunning {
		t.Errorf("canceled request persisted run effects: %+v", payload.Data)
	}
	frames, err := repo.FramesForRun(context.Background(), runID)
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := repo.GetCursor(context.Background(), runID, "probe-cancel")
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 0 || cursor.LastSequence != 0 {
		t.Errorf("canceled request persisted frames=%d cursor=%d", len(frames), cursor.LastSequence)
	}
}

func TestUncanceledFrameStillAdvancesWorkflow(t *testing.T) {
	repo := store.NewMemory()
	source := clock.System{}
	h := api.New(catalog.New(repo, source), engine.New(repo, source)).Handler()
	runID := prepareRun(t, h)
	rec := requestJSON(t, h, http.MethodPost, "/api/runs/"+runID+"/frames", `{"sensor_id":"probe-live","sequence":1,"stage_index":0,"celsius":25,"humidity":42}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("live frame status=%d body=%s", rec.Code, rec.Body.String())
	}
}
