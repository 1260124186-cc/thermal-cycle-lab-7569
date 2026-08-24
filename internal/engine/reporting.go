package engine

import (
	"context"
	"errors"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

func (c *Controller) GenerateReport(ctx context.Context, runID string) (domain.RunReport, error) {
	run, err := c.repo.GetRun(ctx, runID)
	if err != nil {
		return domain.RunReport{}, fmt.Errorf("load run for report: %w", err)
	}
	frames, err := c.repo.FramesForRun(ctx, runID)
	if err != nil {
		return domain.RunReport{}, fmt.Errorf("load frames for report: %w", err)
	}
	profile, err := c.repo.GetProfile(ctx, run.ProfileID)
	if err != nil {
		return domain.RunReport{}, fmt.Errorf("load profile for report: %w", err)
	}
	report, err := domain.BuildRunReport(run, frames, profile, c.clock.Now())
	if err != nil {
		return domain.RunReport{}, fmt.Errorf("build run report: %w", err)
	}
	if err := report.Validate(); err != nil {
		return domain.RunReport{}, fmt.Errorf("verify run report: %w", err)
	}
	if err := c.repo.SaveReport(ctx, report); err != nil {
		return domain.RunReport{}, fmt.Errorf("save run report: %w", err)
	}
	return report, nil
}
func (c *Controller) Report(ctx context.Context, runID string) (domain.RunReport, error) {
	report, err := c.repo.GetReport(ctx, runID)
	if err != nil {
		return domain.RunReport{}, errors.New("report read failed")
	}
	return report, nil
}

type Dashboard struct {
	Running   int `json:"running"`
	Paused    int `json:"paused"`
	Completed int `json:"completed"`
	Alerts    int `json:"alerts"`
	TotalRuns int `json:"total_runs"`
}

func (c *Controller) Dashboard(ctx context.Context) (Dashboard, error) {
	runs, err := c.repo.ListRuns(ctx)
	if err != nil {
		return Dashboard{}, fmt.Errorf("list runs for dashboard: %w", err)
	}
	result := Dashboard{TotalRuns: len(runs)}
	for _, run := range runs {
		switch run.State {
		case domain.RunRunning:
			result.Running++
		case domain.RunPaused:
			result.Paused++
		case domain.RunCompleted:
			result.Completed++
		}
		result.Alerts += run.AlertCount
	}
	return result, nil
}
