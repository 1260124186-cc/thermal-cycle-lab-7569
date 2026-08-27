package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"thermal-cycle-lab/internal/domain"
	"time"
)

type RunFilter struct {
	States        []domain.RunState
	SpecimenID    string
	ProfileID     string
	StartedAfter  time.Time
	StartedBefore time.Time
	MinimumAlerts int
}

func (c *Controller) SearchRuns(ctx context.Context, filter RunFilter) ([]domain.ExperimentRun, error) {
	runs, err := c.repo.ListRuns(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runs for search: %w", err)
	}
	states := make(map[domain.RunState]struct{}, len(filter.States))
	for _, state := range filter.States {
		states[state] = struct{}{}
	}
	matches := make([]domain.ExperimentRun, 0, len(runs))
	for _, run := range runs {
		if len(states) > 0 {
			if _, ok := states[run.State]; !ok {
				continue
			}
		}
		if filter.SpecimenID != "" && run.SpecimenID != strings.TrimSpace(filter.SpecimenID) {
			continue
		}
		if filter.ProfileID != "" && run.ProfileID != strings.TrimSpace(filter.ProfileID) {
			continue
		}
		if !filter.StartedAfter.IsZero() && run.StartedAt.Before(filter.StartedAfter) {
			continue
		}
		if !filter.StartedBefore.IsZero() && run.StartedAt.After(filter.StartedBefore) {
			continue
		}
		if run.AlertCount < filter.MinimumAlerts {
			continue
		}
		matches = append(matches, run)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].StartedAt.Equal(matches[j].StartedAt) {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].StartedAt.After(matches[j].StartedAt)
	})
	return matches, nil
}

type RunSummary struct {
	RunID       string          `json:"run_id"`
	State       domain.RunState `json:"state"`
	AgeSeconds  int64           `json:"age_seconds"`
	FrameCount  int             `json:"frame_count"`
	AlertRatio  float64         `json:"alert_ratio"`
	LatestEvent string          `json:"latest_event,omitempty"`
}

func (c *Controller) SummarizeRun(ctx context.Context, runID string) (RunSummary, error) {
	run, err := c.repo.GetRun(ctx, runID)
	if err != nil {
		return RunSummary{}, fmt.Errorf("load run for summary: %w", err)
	}
	summary := RunSummary{RunID: run.ID, State: run.State, AgeSeconds: int64(c.clock.Now().Sub(run.StartedAt).Seconds()), FrameCount: run.AcceptedFrames}
	if run.AcceptedFrames > 0 {
		summary.AlertRatio = float64(run.AlertCount) / float64(run.AcceptedFrames)
	}
	events := c.observers.History(runID)
	if len(events) > 0 {
		summary.LatestEvent = events[len(events)-1].Kind
	}
	return summary, nil
}
