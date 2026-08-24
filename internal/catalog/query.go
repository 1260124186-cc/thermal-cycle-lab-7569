package catalog

import (
	"context"
	"fmt"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

type Inventory struct {
	Specimens        []domain.Specimen `json:"specimens"`
	Profiles         []domain.Profile  `json:"profiles"`
	ReadySpecimens   int               `json:"ready_specimens"`
	RetiredSpecimens int               `json:"retired_specimens"`
}

func (s *Service) Inventory(ctx context.Context) (Inventory, error) {
	specimens, err := s.repo.ListSpecimens(ctx)
	if err != nil {
		return Inventory{}, fmt.Errorf("list specimens: %w", err)
	}
	profiles, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return Inventory{}, fmt.Errorf("list profiles: %w", err)
	}
	result := Inventory{Specimens: specimens, Profiles: profiles}
	for _, item := range specimens {
		if item.State == domain.SpecimenReady {
			result.ReadySpecimens++
		}
		if item.State == domain.SpecimenRetired {
			result.RetiredSpecimens++
		}
	}
	sort.Slice(result.Profiles, func(i, j int) bool { return result.Profiles[i].Name < result.Profiles[j].Name })
	return result, nil
}
func (s *Service) CompatibleProfiles(ctx context.Context, specimenID string) ([]domain.Profile, error) {
	specimen, err := s.GetSpecimen(ctx, specimenID)
	if err != nil {
		return nil, err
	}
	profiles, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles for compatibility: %w", err)
	}
	matches := make([]domain.Profile, 0)
	for _, profile := range profiles {
		low, high := profile.TemperatureBounds()
		if specimen.Supports(low, high) {
			matches = append(matches, profile)
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].PlannedSeconds() < matches[j].PlannedSeconds() })
	return matches, nil
}
