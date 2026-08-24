package domain

import (
	"errors"
	"strings"
	"time"
)

type RunState string

const (
	RunRunning   RunState = "running"
	RunPaused    RunState = "paused"
	RunCompleted RunState = "completed"
	RunAborted   RunState = "aborted"
)

type ExperimentRun struct {
	ID             string     `json:"id"`
	SpecimenID     string     `json:"specimen_id"`
	ProfileID      string     `json:"profile_id"`
	State          RunState   `json:"state"`
	StageIndex     int        `json:"stage_index"`
	AcceptedFrames int        `json:"accepted_frames"`
	AlertCount     int        `json:"alert_count"`
	StartedAt      time.Time  `json:"started_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	PauseReason    string     `json:"pause_reason,omitempty"`
}

func NewExperimentRun(id, specimenID, profileID string, now time.Time) (ExperimentRun, error) {
	run := ExperimentRun{ID: strings.TrimSpace(id), SpecimenID: strings.TrimSpace(specimenID), ProfileID: strings.TrimSpace(profileID), State: RunRunning, StartedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := run.Validate(); err != nil {
		return ExperimentRun{}, err
	}
	return run, nil
}

func (r ExperimentRun) Validate() error {
	if r.ID == "" || r.SpecimenID == "" || r.ProfileID == "" {
		return errors.New("run identity fields are required")
	}
	if r.StageIndex < 0 || r.AcceptedFrames < 0 || r.AlertCount < 0 {
		return errors.New("run counters cannot be negative")
	}
	switch r.State {
	case RunRunning, RunPaused, RunCompleted, RunAborted:
	default:
		return errors.New("run state is invalid")
	}
	if r.State == RunPaused && strings.TrimSpace(r.PauseReason) == "" {
		return errors.New("paused run requires a reason")
	}
	if r.State != RunPaused && r.PauseReason != "" {
		return errors.New("only paused run can keep a pause reason")
	}
	if r.State == RunCompleted && r.CompletedAt == nil {
		return errors.New("completed run requires completion time")
	}
	return nil
}

func (r ExperimentRun) Pause(reason string, now time.Time) (ExperimentRun, error) {
	if r.State != RunRunning {
		return ExperimentRun{}, errors.New("only running experiment can be paused")
	}
	reason = strings.TrimSpace(reason)
	if len(reason) < 4 {
		return ExperimentRun{}, errors.New("pause reason is too short")
	}
	r.State, r.PauseReason, r.UpdatedAt = RunPaused, reason, now.UTC()
	return r, r.Validate()
}

func (r ExperimentRun) Resume(now time.Time) (ExperimentRun, error) {
	if r.State != RunPaused {
		return ExperimentRun{}, errors.New("only paused experiment can resume")
	}
	r.State, r.PauseReason, r.UpdatedAt = RunRunning, "", now.UTC()
	return r, r.Validate()
}

func (r ExperimentRun) AcceptFrame(stage int, alert bool, now time.Time) (ExperimentRun, error) {
	if r.State != RunRunning {
		return ExperimentRun{}, errors.New("run does not accept sensor frames in current state")
	}
	if stage < r.StageIndex {
		return ExperimentRun{}, errors.New("sensor frame belongs to an earlier stage")
	}
	r.StageIndex, r.AcceptedFrames, r.UpdatedAt = stage, r.AcceptedFrames+1, now.UTC()
	if alert {
		r.AlertCount++
	}
	return r, r.Validate()
}

func (r ExperimentRun) Complete(now time.Time) (ExperimentRun, error) {
	if r.State != RunRunning && r.State != RunPaused {
		return ExperimentRun{}, errors.New("run cannot complete in current state")
	}
	if r.AcceptedFrames == 0 {
		return ExperimentRun{}, errors.New("run cannot complete without sensor evidence")
	}
	completed := now.UTC()
	r.State, r.PauseReason, r.CompletedAt, r.UpdatedAt = RunCompleted, "", &completed, completed
	return r, r.Validate()
}

func (r ExperimentRun) IsTerminal() bool { return r.State == RunCompleted || r.State == RunAborted }
