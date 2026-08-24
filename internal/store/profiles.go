package store

import (
	"context"
	"fmt"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

func (m *Memory) CreateProfile(ctx context.Context, profile domain.Profile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("validate profile: %w", err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.profiles[profile.ID]; ok {
		return fmt.Errorf("profile %s: %w", profile.ID, ErrConflict)
	}
	profile.Stages = profile.CloneStages()
	m.profiles[profile.ID] = profile
	return nil
}
func (m *Memory) GetProfile(ctx context.Context, id string) (domain.Profile, error) {
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	profile, ok := m.profiles[id]
	if !ok {
		return domain.Profile{}, fmt.Errorf("profile %s: %w", id, ErrNotFound)
	}
	profile.Stages = profile.CloneStages()
	return profile, nil
}
func (m *Memory) ListProfiles(ctx context.Context) ([]domain.Profile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	values := cloneProfiles(m.profiles)
	m.mu.RUnlock()
	sort.Slice(values, func(i, j int) bool {
		if values[i].CreatedAt.Equal(values[j].CreatedAt) {
			return values[i].ID < values[j].ID
		}
		return values[i].CreatedAt.Before(values[j].CreatedAt)
	})
	return values, nil
}
