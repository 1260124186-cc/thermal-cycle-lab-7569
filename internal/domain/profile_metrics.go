package domain

import (
	"errors"
	"math"
)

type ProfileMetrics struct {
	StageCount       int     `json:"stage_count"`
	PlannedSeconds   int     `json:"planned_seconds"`
	LowestTarget     float64 `json:"lowest_target"`
	HighestTarget    float64 `json:"highest_target"`
	LargestRamp      float64 `json:"largest_ramp"`
	ThermalDistance  float64 `json:"thermal_distance"`
	AverageTolerance float64 `json:"average_tolerance"`
}

func MeasureProfile(profile Profile) (ProfileMetrics, error) {
	if err := profile.Validate(); err != nil {
		return ProfileMetrics{}, err
	}
	low, high := profile.TemperatureBounds()
	metrics := ProfileMetrics{StageCount: len(profile.Stages), PlannedSeconds: profile.PlannedSeconds(), LowestTarget: low, HighestTarget: high}
	toleranceTotal := 0.0
	previous := profile.Stages[0].TargetCelsius
	for index, stage := range profile.Stages {
		toleranceTotal += stage.Tolerance
		if index > 0 {
			distance := math.Abs(stage.TargetCelsius - previous)
			metrics.ThermalDistance += distance
			rate := distance / float64(stage.RampSeconds)
			if rate > metrics.LargestRamp {
				metrics.LargestRamp = rate
			}
		}
		previous = stage.TargetCelsius
	}
	metrics.AverageTolerance = toleranceTotal / float64(len(profile.Stages))
	return metrics, nil
}

func (m ProfileMetrics) Validate() error {
	if m.StageCount < 1 {
		return errors.New("profile metrics require stages")
	}
	if m.PlannedSeconds < 1 {
		return errors.New("profile metrics require positive duration")
	}
	if m.LowestTarget > m.HighestTarget {
		return errors.New("profile metric bounds are inverted")
	}
	if m.LargestRamp < 0 || m.ThermalDistance < 0 {
		return errors.New("profile metric movement cannot be negative")
	}
	if m.AverageTolerance <= 0 {
		return errors.New("profile metric tolerance must be positive")
	}
	return nil
}

func (m ProfileMetrics) Intensity() string {
	if m.ThermalDistance >= 250 || m.LargestRamp >= 3 {
		return "severe"
	}
	if m.ThermalDistance >= 100 || m.LargestRamp >= 1 {
		return "moderate"
	}
	return "gentle"
}

func (m ProfileMetrics) EstimatedMinutes() int {
	if m.PlannedSeconds == 0 {
		return 0
	}
	return (m.PlannedSeconds + 59) / 60
}
