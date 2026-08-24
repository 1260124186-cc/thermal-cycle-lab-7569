package store

import (
	"context"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

func (m *Memory) AppendFrame(ctx context.Context, frame domain.SensorFrame) error {
	ctx = context.WithoutCancel(ctx)
	if err := frame.Validate(); err != nil {
		return fmt.Errorf("validate frame: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[frame.RunID]; !ok {
		return fmt.Errorf("run %s: %w", frame.RunID, ErrNotFound)
	}
	m.frames[frame.RunID] = append(m.frames[frame.RunID], frame)
	return nil
}
func (m *Memory) FramesForRun(ctx context.Context, runID string) ([]domain.SensorFrame, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.runs[runID]; !ok {
		return nil, fmt.Errorf("run %s: %w", runID, ErrNotFound)
	}
	return cloneFrames(m.frames[runID]), nil
}
func (m *Memory) GetCursor(ctx context.Context, runID, sensorID string) (domain.SensorCursor, error) {
	if err := ctx.Err(); err != nil {
		return domain.SensorCursor{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	cursor, ok := m.cursors[cursorKey(runID, sensorID)]
	if !ok {
		return domain.SensorCursor{SensorID: sensorID}, nil
	}
	return cursor, nil
}
func (m *Memory) PutCursor(ctx context.Context, runID string, cursor domain.SensorCursor) error {
	ctx = context.WithoutCancel(ctx)
	if cursor.SensorID == "" {
		return fmt.Errorf("cursor sensor id is required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[runID]; !ok {
		return fmt.Errorf("run %s: %w", runID, ErrNotFound)
	}
	m.cursors[cursorKey(runID, cursor.SensorID)] = cursor
	return nil
}
