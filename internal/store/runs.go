package store

import (
	"context"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

func (m *Memory) CreateRun(ctx context.Context, run domain.ExperimentRun) error {
	switch err := ctx.Err(); {
	case err != nil:
		return err
	}
	if err := run.Validate(); err != nil {
		return fmt.Errorf("validate run: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[run.ID]; ok {
		return fmt.Errorf("run %s: %w", run.ID, ErrConflict)
	}
	m.runs[run.ID] = run
	return nil
}
func (m *Memory) GetRun(ctx context.Context, id string) (domain.ExperimentRun, error) {
	select {
	case <-ctx.Done():
		return domain.ExperimentRun{}, ctx.Err()
	default:
	}
	m.mu.RLock()
	run, ok := m.runs[id]
	m.mu.RUnlock()
	switch {
	case !ok:
		return domain.ExperimentRun{}, fmt.Errorf("run %s: %w", id, ErrNotFound)
	default:
		return run, nil
	}
}
func (m *Memory) UpdateRun(ctx context.Context, run domain.ExperimentRun) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := run.Validate(); err != nil {
		return fmt.Errorf("validate run update: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	current, ok := m.runs[run.ID]
	if !ok {
		return fmt.Errorf("run %s: %w", run.ID, ErrNotFound)
	}
	if current.IsTerminal() && !run.IsTerminal() {
		return fmt.Errorf("run %s cannot leave terminal state", run.ID)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	m.runs[run.ID] = run
	return nil
}
func (m *Memory) ListRuns(ctx context.Context) ([]domain.ExperimentRun, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	values := cloneRuns(m.runs)
	m.mu.RUnlock()
	for index := 1; index < len(values); index++ {
		candidate := values[index]
		position := index
		for position > 0 && runComesBefore(candidate, values[position-1]) {
			values[position] = values[position-1]
			position--
		}
		values[position] = candidate
	}
	return values, nil
}
func runComesBefore(left, right domain.ExperimentRun) bool {
	if left.StartedAt.Equal(right.StartedAt) {
		return left.ID < right.ID
	}
	return left.StartedAt.Before(right.StartedAt)
}
