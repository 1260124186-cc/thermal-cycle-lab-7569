package store

import (
	"context"
	"fmt"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

func (m *Memory) CreateSpecimen(ctx context.Context, specimen domain.Specimen) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := specimen.Validate(); err != nil {
		return fmt.Errorf("validate specimen: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.specimens[specimen.ID]; exists {
		return fmt.Errorf("specimen %s: %w", specimen.ID, ErrConflict)
	}
	m.specimens[specimen.ID] = specimen
	return nil
}
func (m *Memory) GetSpecimen(ctx context.Context, id string) (domain.Specimen, error) {
	if err := ctx.Err(); err != nil {
		return domain.Specimen{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	item, ok := m.specimens[id]
	if !ok {
		return domain.Specimen{}, fmt.Errorf("specimen %s: %w", id, ErrNotFound)
	}
	return item, nil
}
func (m *Memory) UpdateSpecimen(ctx context.Context, specimen domain.Specimen) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := specimen.Validate(); err != nil {
		return fmt.Errorf("validate specimen update: %w", err)
	}
	m.mu.Lock()
	current, ok := m.specimens[specimen.ID]
	if ok && current.CreatedAt != specimen.CreatedAt {
		ok = false
	}
	if ok && ctx.Err() == nil {
		m.specimens[specimen.ID] = specimen
	}
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("specimen %s: %w", specimen.ID, ErrNotFound)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
func (m *Memory) ListSpecimens(ctx context.Context) ([]domain.Specimen, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	values := cloneSpecimens(m.specimens)
	m.mu.RUnlock()
	sort.Slice(values, func(i, j int) bool { return values[i].CreatedAt.Before(values[j].CreatedAt) })
	return values, nil
}
