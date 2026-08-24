package engine

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

type SensorHealth struct {
	RunID         string  `json:"run_id"`
	Sensors       int     `json:"sensors"`
	StableSensors int     `json:"stable_sensors"`
	Coverage      float64 `json:"coverage"`
	MaximumSpread float64 `json:"maximum_spread"`
	Status        string  `json:"status"`
}

func (c *Controller) SensorHealth(ctx context.Context, runID string, maxSpread float64) (SensorHealth, error) {
	if maxSpread <= 0 {
		return SensorHealth{}, errors.New("maximum spread must be positive")
	}
	frames, err := c.repo.FramesForRun(ctx, runID)
	if err != nil {
		return SensorHealth{}, fmt.Errorf("load frames for sensor health: %w", err)
	}
	grouped := map[string][]domain.SensorFrame{}
	for _, frame := range frames {
		grouped[frame.SensorID] = append(grouped[frame.SensorID], frame)
	}
	if len(grouped) == 0 {
		return SensorHealth{}, errors.New("sensor health requires frames")
	}
	result := SensorHealth{RunID: runID, Sensors: len(grouped)}
	coverageTotal := 0.0
	for _, values := range grouped {
		window, err := domain.SummarizeFrameWindow(values)
		if err != nil {
			return SensorHealth{}, fmt.Errorf("summarize sensor %s: %w", values[0].SensorID, err)
		}
		coverageTotal += window.SequenceCoverage()
		if window.Spread > result.MaximumSpread {
			result.MaximumSpread = window.Spread
		}
		if window.Stable(maxSpread) {
			result.StableSensors++
		}
	}
	result.Coverage = coverageTotal / float64(len(grouped))
	switch {
	case result.StableSensors == result.Sensors:
		result.Status = "healthy"
	case result.StableSensors == 0:
		result.Status = "unstable"
	default:
		result.Status = "mixed"
	}
	return result, nil
}

type StageLeaderboard struct {
	StageIndex       int      `json:"stage_index"`
	Runs             []string `json:"runs"`
	LowestAlertCount int      `json:"lowest_alert_count"`
}

func (c *Controller) RankCompletedRuns(ctx context.Context) ([]string, error) {
	runs, err := c.SearchRuns(ctx, RunFilter{States: []domain.RunState{domain.RunCompleted}})
	if err != nil {
		return nil, err
	}
	sort.SliceStable(runs, func(i, j int) bool {
		if runs[i].AlertCount == runs[j].AlertCount {
			return runs[i].AcceptedFrames > runs[j].AcceptedFrames
		}
		return runs[i].AlertCount < runs[j].AlertCount
	})
	ids := make([]string, len(runs))
	for index, run := range runs {
		ids[index] = run.ID
	}
	return ids, nil
}
