package catalog

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"thermal-cycle-lab/internal/clock"
	"thermal-cycle-lab/internal/domain"
	"thermal-cycle-lab/internal/store"
)

type Service struct {
	repo        store.Repository
	clock       clock.Source
	specimenSeq atomic.Uint64
	profileSeq  atomic.Uint64
}

func New(repo store.Repository, source clock.Source) *Service {
	return &Service{repo: repo, clock: source}
}

type CreateSpecimenInput struct {
	Label      string  `json:"label"`
	Material   string  `json:"material"`
	MinCelsius float64 `json:"min_celsius"`
	MaxCelsius float64 `json:"max_celsius"`
}
type CreateProfileInput struct {
	Name   string         `json:"name"`
	Stages []domain.Stage `json:"stages"`
}

func (s *Service) CreateSpecimen(ctx context.Context, input CreateSpecimenInput) (domain.Specimen, error) {
	if err := ctx.Err(); err != nil {
		return domain.Specimen{}, err
	}
	id := fmt.Sprintf("sp-%03d", s.specimenSeq.Add(1))
	item, err := domain.NewSpecimen(id, input.Label, input.Material, input.MinCelsius, input.MaxCelsius, s.clock.Now())
	if err != nil {
		return domain.Specimen{}, fmt.Errorf("create specimen: %w", err)
	}
	if err := s.repo.CreateSpecimen(ctx, item); err != nil {
		return domain.Specimen{}, fmt.Errorf("store specimen: %w", err)
	}
	return item, nil
}
func (s *Service) CreateProfile(ctx context.Context, input CreateProfileInput) (domain.Profile, error) {
	if err := ctx.Err(); err != nil {
		return domain.Profile{}, err
	}
	normalized := make([]domain.Stage, 0, len(input.Stages))
	for _, stage := range input.Stages {
		stage.Name = strings.TrimSpace(stage.Name)
		normalized = append(normalized, stage)
	}
	id := fmt.Sprintf("pr-%03d", s.profileSeq.Add(1))
	profile, err := domain.NewProfile(id, input.Name, normalized, s.clock.Now())
	if err != nil {
		return domain.Profile{}, fmt.Errorf("create profile: %w", err)
	}
	limits := domain.DefaultLaboratoryLimits()
	if err := limits.CheckProfile(profile); err != nil {
		return domain.Profile{}, fmt.Errorf("check laboratory limits: %w", err)
	}
	if err := s.repo.CreateProfile(ctx, profile); err != nil {
		return domain.Profile{}, fmt.Errorf("store profile: %w", err)
	}
	return profile, nil
}
func (s *Service) GetSpecimen(ctx context.Context, id string) (domain.Specimen, error) {
	item, err := s.repo.GetSpecimen(ctx, strings.TrimSpace(id))
	if err != nil {
		return domain.Specimen{}, fmt.Errorf("load specimen: %w", err)
	}
	return item, nil
}
func (s *Service) GetProfile(ctx context.Context, id string) (domain.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, strings.TrimSpace(id))
	if err != nil {
		return domain.Profile{}, fmt.Errorf("load profile: %w", err)
	}
	return profile, nil
}

func (s *Service) RetireSpecimen(ctx context.Context, id string) (domain.Specimen, error) {
	item, err := s.repo.GetSpecimen(ctx, strings.TrimSpace(id))
	if err != nil {
		return domain.Specimen{}, fmt.Errorf("load specimen to retire: %w", err)
	}
	retired, err := item.Retire(s.clock.Now())
	if err != nil {
		return domain.Specimen{}, fmt.Errorf("retire specimen: %w", err)
	}
	if err := s.repo.UpdateSpecimen(ctx, retired); err != nil {
		return domain.Specimen{}, fmt.Errorf("persist retired specimen: %w", err)
	}
	return retired, nil
}
