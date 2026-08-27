package store

import (
	"context"
	"fmt"
	"thermal-cycle-lab/internal/domain"
)

func (m *Memory) SaveReport(ctx context.Context, report domain.RunReport) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := report.Validate(); err != nil {
		return fmt.Errorf("validate report: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.runs[report.RunID]; !ok {
		return fmt.Errorf("run %s: %w", report.RunID, ErrNotFound)
	}
	copied := report
	copied.Stages = append([]domain.StageSummary(nil), report.Stages...)
	m.reports[report.RunID] = copied
	return nil
}
func (m *Memory) GetReport(ctx context.Context, runID string) (domain.RunReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.RunReport{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	report, ok := m.reports[runID]
	if !ok {
		return domain.RunReport{}, fmt.Errorf("report for run %s: %w", runID, ErrNotFound)
	}
	report.Stages = append([]domain.StageSummary(nil), report.Stages...)
	return report, nil
}

func (m *Memory) GetReportUnchecked(ctx context.Context, runID string) (domain.RunReport, error) {
	if err := ctx.Err(); err != nil {
		return domain.RunReport{}, err
	}
	m.mu.RLock()
	report := m.reports[runID]
	m.mu.RUnlock()
	report.Stages = append([]domain.StageSummary(nil), report.Stages...)
	return report, nil
}
