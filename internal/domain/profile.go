package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Stage struct {
	Name          string  `json:"name"`
	TargetCelsius float64 `json:"target_celsius"`
	RampSeconds   int     `json:"ramp_seconds"`
	HoldSeconds   int     `json:"hold_seconds"`
	Tolerance     float64 `json:"tolerance"`
}

type Profile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Revision  int       `json:"revision"`
	Stages    []Stage   `json:"stages"`
	CreatedAt time.Time `json:"created_at"`
}

func NewProfile(id, name string, stages []Stage, now time.Time) (Profile, error) {
	copied := append([]Stage(nil), stages...)
	profile := Profile{ID: strings.TrimSpace(id), Name: strings.TrimSpace(name), Revision: 1, Stages: copied, CreatedAt: now.UTC()}
	if err := profile.Validate(); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func (p Profile) Validate() error {
	if p.ID == "" {
		return errors.New("profile id is required")
	}
	if len(p.Name) < 3 {
		return errors.New("profile name must contain at least three characters")
	}
	if p.Revision < 1 {
		return errors.New("profile revision must be positive")
	}
	if len(p.Stages) == 0 || len(p.Stages) > 24 {
		return errors.New("profile must contain between one and twenty-four stages")
	}
	seen := map[string]struct{}{}
	for index, stage := range p.Stages {
		key := strings.ToLower(strings.TrimSpace(stage.Name))
		if key == "" {
			return fmt.Errorf("stage %d has no name", index+1)
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("stage name %q is repeated", stage.Name)
		}
		seen[key] = struct{}{}
		if stage.TargetCelsius < -160 || stage.TargetCelsius > 400 {
			return fmt.Errorf("stage %q target exceeds chamber limits", stage.Name)
		}
		if stage.RampSeconds < 1 || stage.HoldSeconds < 0 {
			return fmt.Errorf("stage %q duration is invalid", stage.Name)
		}
		if stage.Tolerance <= 0 || stage.Tolerance > 25 {
			return fmt.Errorf("stage %q tolerance is invalid", stage.Name)
		}
	}
	return nil
}

func (p Profile) TemperatureBounds() (float64, float64) {
	low, high := p.Stages[0].TargetCelsius, p.Stages[0].TargetCelsius
	for _, stage := range p.Stages[1:] {
		if stage.TargetCelsius < low {
			low = stage.TargetCelsius
		}
		if stage.TargetCelsius > high {
			high = stage.TargetCelsius
		}
	}
	return low, high
}

func (p Profile) PlannedSeconds() int {
	total := 0
	for _, stage := range p.Stages {
		total += stage.RampSeconds + stage.HoldSeconds
	}
	return total
}

func (p Profile) StageAt(index int) (Stage, error) {
	if index < 0 || index >= len(p.Stages) {
		return Stage{}, errors.New("stage index is outside profile")
	}
	return p.Stages[index], nil
}

func (p Profile) CloneStages() []Stage { return append([]Stage(nil), p.Stages...) }
