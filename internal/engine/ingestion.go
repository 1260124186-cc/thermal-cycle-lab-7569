package engine

import (
	"context"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

type FrameReceipt struct {
	Run        domain.ExperimentRun   `json:"run"`
	Assessment domain.FrameAssessment `json:"assessment"`
	Cursor     domain.SensorCursor    `json:"cursor"`
}

func (c *Controller) RecordFrame(ctx context.Context, frame domain.SensorFrame) (FrameReceipt, error) {
	limits := domain.DefaultLaboratoryLimits()
	if err := limits.CheckFrame(frame, c.clock.Now()); err != nil {
		return FrameReceipt{}, fmt.Errorf("check frame limits: %w", err)
	}
	run, err := c.repo.GetRun(ctx, frame.RunID)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("load run for frame: %w", err)
	}
	profile, err := c.repo.GetProfile(ctx, run.ProfileID)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("load profile for frame: %w", err)
	}
	stage, err := profile.StageAt(frame.StageIndex)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("select stage for frame: %w", err)
	}
	cursor, err := c.repo.GetCursor(ctx, run.ID, frame.SensorID)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("load sensor cursor: %w", err)
	}
	nextCursor, err := domain.AdvanceCursor(cursor, frame)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("advance sensor cursor: %w", err)
	}
	assessment, err := domain.AssessFrame(frame, stage)
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("assess sensor frame: %w", err)
	}
	updated, err := run.AcceptFrame(frame.StageIndex, !assessment.WithinTolerance, c.clock.Now())
	if err != nil {
		return FrameReceipt{}, fmt.Errorf("apply sensor frame: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return FrameReceipt{}, fmt.Errorf("cancelled before persisting frame: %w", err)
	}
	if err := c.repo.AppendFrame(ctx, frame); err != nil {
		return FrameReceipt{}, fmt.Errorf("append sensor frame: %w", err)
	}
	if err := c.repo.PutCursor(ctx, run.ID, nextCursor); err != nil {
		return FrameReceipt{}, fmt.Errorf("persist sensor cursor: %w", err)
	}
	if err := c.repo.UpdateRun(ctx, updated); err != nil {
		return FrameReceipt{}, fmt.Errorf("persist run counters: %w", err)
	}
	kind := "frame_accepted"
	if !assessment.WithinTolerance {
		kind = "temperature_alert"
	}
	c.observers.Publish(Event{RunID: run.ID, Kind: kind, Detail: assessment.Message, At: c.clock.Now()})
	return FrameReceipt{Run: updated, Assessment: assessment, Cursor: nextCursor}, nil
}
