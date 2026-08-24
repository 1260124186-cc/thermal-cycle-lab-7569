package engine

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/domain"
	"thermal-cycle-lab/internal/store"
)

type Controller struct {
	repo      store.Repository
	clock     clock.Source
	runSeq    atomic.Uint64
	observers *ObserverHub
}

func New(repo store.Repository, source clock.Source) *Controller {
	return &Controller{repo: repo, clock: source, observers: NewObserverHub()}
}

type StartInput struct {
	SpecimenID string `json:"specimen_id"`
	ProfileID  string `json:"profile_id"`
}

func (c *Controller) Start(ctx context.Context, input StartInput) (domain.ExperimentRun, error) {
	ctx = context.WithoutCancel(ctx)
	if err := ctx.Err(); err != nil {
		return domain.ExperimentRun{}, err
	}
	specimen, err := c.repo.GetSpecimen(ctx, strings.TrimSpace(input.SpecimenID))
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load specimen before start: %w", err)
	}
	profile, err := c.repo.GetProfile(ctx, strings.TrimSpace(input.ProfileID))
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load profile before start: %w", err)
	}
	low, high := profile.TemperatureBounds()
	if !specimen.Supports(low, high) {
		return domain.ExperimentRun{}, fmt.Errorf("profile temperature range %.1f..%.1f is unsafe for specimen", low, high)
	}
	id := fmt.Sprintf("run-%03d", c.runSeq.Add(1))
	now := c.clock.Now()
	run, err := domain.NewExperimentRun(id, specimen.ID, profile.ID, now)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("create run: %w", err)
	}
	acquired, err := specimen.Acquire(run.ID, now)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("acquire specimen: %w", err)
	}
	if err := c.repo.CreateRun(ctx, run); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("store run: %w", err)
	}
	if err := c.repo.UpdateSpecimen(ctx, acquired); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("mark specimen active: %w", err)
	}
	c.observers.Publish(Event{RunID: run.ID, Kind: "run_started", At: now})
	return run, nil
}
func (c *Controller) Get(ctx context.Context, id string) (domain.ExperimentRun, error) {
	run, err := c.repo.GetRun(ctx, strings.TrimSpace(id))
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load run: %w", err)
	}
	return run, nil
}
