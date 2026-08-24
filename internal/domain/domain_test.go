package domain_test

import (
	"testing"
	"thermal-cycle-lab/internal/domain"
	"time"
)

func TestRunLifecycleAndReport(t *testing.T) {
	now := time.Date(2026, 8, 24, 10, 0, 0, 0, time.UTC)
	specimen, err := domain.NewSpecimen("sp-1", "alloy coupon", "alloy", -40, 160, now)
	if err != nil {
		t.Fatal(err)
	}
	profile, err := domain.NewProfile("pr-1", "thermal cycle", []domain.Stage{{Name: "heat", TargetCelsius: 80, RampSeconds: 10, HoldSeconds: 20, Tolerance: 2}}, now)
	if err != nil {
		t.Fatal(err)
	}
	run, err := domain.NewExperimentRun("run-1", specimen.ID, profile.ID, now)
	if err != nil {
		t.Fatal(err)
	}
	frame := domain.SensorFrame{RunID: run.ID, SensorID: "sensor-a", Sequence: 1, StageIndex: 0, Celsius: 81, Humidity: 20, CapturedAt: now.Add(time.Second)}
	run, err = run.AcceptFrame(0, false, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	run, err = run.Complete(now.Add(2 * time.Second))
	if err != nil {
		t.Fatal(err)
	}
	report, err := domain.BuildRunReport(run, []domain.SensorFrame{frame}, profile, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if report.Result != "stable" {
		t.Fatalf("unexpected result %s", report.Result)
	}
}
