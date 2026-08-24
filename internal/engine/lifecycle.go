package engine

import (
	"context"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

func (c *Controller) Pause(ctx context.Context, runID, reason string) (domain.ExperimentRun, error) {
	run, err := c.repo.GetRun(ctx, runID)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load run to pause: %w", err)
	}
	updated, err := run.Pause(reason, c.clock.Now())
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("pause run: %w", err)
	}
	if err := c.repo.UpdateRun(ctx, updated); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("persist paused run: %w", err)
	}
	c.observers.Publish(Event{RunID: runID, Kind: "run_paused", Detail: updated.PauseReason, At: updated.UpdatedAt})
	return updated, nil
}
func (c *Controller) Resume(ctx context.Context, runID string) (domain.ExperimentRun, error) {
	run, err := c.repo.GetRun(ctx, runID)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load run to resume: %w", err)
	}
	updated, err := run.Resume(c.clock.Now())
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("resume run: %w", err)
	}
	if err := c.repo.UpdateRun(ctx, updated); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("persist resumed run: %w", err)
	}
	c.observers.Publish(Event{RunID: runID, Kind: "run_resumed", At: updated.UpdatedAt})
	return updated, nil
}
func (c *Controller) Finalize(ctx context.Context, runID string) (domain.ExperimentRun, error) {
	run, err := c.repo.GetRun(ctx, runID)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load run to finalize: %w", err)
	}
	now := c.clock.Now()
	completed, err := run.Complete(now)
	if err != nil {
		c.observers.Publish(Event{RunID: runID, Kind: "run_finalize_rejected", Detail: err.Error(), At: now})
		return domain.ExperimentRun{}, fmt.Errorf("finalize run: %w", err)
	}
	specimen, err := c.repo.GetSpecimen(ctx, run.SpecimenID)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("load specimen to release: %w", err)
	}
	released, err := specimen.Release(now)
	if err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("release specimen: %w", err)
	}
	if err := c.repo.UpdateRun(ctx, completed); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("persist completed run: %w", err)
	}
	if err := c.repo.UpdateSpecimen(ctx, released); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("persist released specimen: %w", err)
	}
	if _, err := c.GenerateReport(ctx, completed.ID); err != nil {
		return domain.ExperimentRun{}, fmt.Errorf("generate completion report: %w", err)
	}
	c.observers.Publish(Event{RunID: runID, Kind: "run_completed", At: now})
	return completed, nil
}
