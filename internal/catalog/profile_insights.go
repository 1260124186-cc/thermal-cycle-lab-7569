package catalog

import (
	"context"
	"fmt"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

type ProfileInsight struct {
	Intensity           string                `json:"intensity"`
	EstimatedMinutes    int                   `json:"estimated_minutes"`
	Profile             domain.Profile        `json:"profile"`
	Metrics             domain.ProfileMetrics `json:"metrics"`
	CompatibleSpecimens int                   `json:"compatible_specimens"`
	AvailableSpecimens  int                   `json:"available_specimens"`
}

func (s *Service) ProfileInsights(ctx context.Context) ([]ProfileInsight, error) {
	profiles, err := s.repo.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles for insights: %w", err)
	}
	specimens, err := s.repo.ListSpecimens(ctx)
	if err != nil {
		return nil, fmt.Errorf("list specimens for insights: %w", err)
	}
	insights := make([]ProfileInsight, 0, len(profiles))
	for _, profile := range profiles {
		metrics, err := domain.MeasureProfile(profile)
		if err != nil {
			return nil, fmt.Errorf("measure profile %s: %w", profile.ID, err)
		}
		insight := ProfileInsight{Profile: profile, Metrics: metrics, Intensity: metrics.Intensity(), EstimatedMinutes: metrics.EstimatedMinutes()}
		low, high := profile.TemperatureBounds()
		for _, specimen := range specimens {
			if specimen.Supports(low, high) {
				insight.CompatibleSpecimens++
				if specimen.State == domain.SpecimenReady {
					insight.AvailableSpecimens++
				}
			}
		}
		insights = append(insights, insight)
	}
	sort.Slice(insights, func(i, j int) bool {
		if insights[i].AvailableSpecimens == insights[j].AvailableSpecimens {
			return insights[i].Profile.ID < insights[j].Profile.ID
		}
		return insights[i].AvailableSpecimens > insights[j].AvailableSpecimens
	})
	return insights, nil
}

type CatalogHealth struct {
	SpecimenCount int      `json:"specimen_count"`
	ProfileCount  int      `json:"profile_count"`
	UsablePairs   int      `json:"usable_pairs"`
	Warnings      []string `json:"warnings"`
}

func (s *Service) Health(ctx context.Context) (CatalogHealth, error) {
	inventory, err := s.Inventory(ctx)
	if err != nil {
		return CatalogHealth{}, err
	}
	result := CatalogHealth{SpecimenCount: len(inventory.Specimens), ProfileCount: len(inventory.Profiles)}
	for _, specimen := range inventory.Specimens {
		if specimen.State == domain.SpecimenRetired {
			continue
		}
		matches, err := s.CompatibleProfiles(ctx, specimen.ID)
		if err != nil {
			return CatalogHealth{}, err
		}
		result.UsablePairs += len(matches)
		if len(matches) == 0 {
			result.Warnings = append(result.Warnings, "specimen "+specimen.ID+" has no compatible profile")
		}
	}
	if len(inventory.Profiles) == 0 {
		result.Warnings = append(result.Warnings, "catalog has no temperature profile")
	}
	return result, nil
}
