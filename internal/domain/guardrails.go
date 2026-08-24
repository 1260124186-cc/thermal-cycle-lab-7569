package domain

import (
	"errors"
	"fmt"
	"time"
)

type LaboratoryLimits struct {
	MinimumCelsius     float64       `json:"minimum_celsius"`
	MaximumCelsius     float64       `json:"maximum_celsius"`
	MaximumStages      int           `json:"maximum_stages"`
	MaximumRunDuration time.Duration `json:"maximum_run_duration"`
	MaximumFrameAge    time.Duration `json:"maximum_frame_age"`
}

func DefaultLaboratoryLimits() LaboratoryLimits {
	return LaboratoryLimits{MinimumCelsius: -160, MaximumCelsius: 400, MaximumStages: 24, MaximumRunDuration: 72 * time.Hour, MaximumFrameAge: 10 * time.Minute}
}
func (l LaboratoryLimits) Validate() error {
	if l.MinimumCelsius >= l.MaximumCelsius {
		return errors.New("laboratory temperature bounds are invalid")
	}
	if l.MaximumStages < 1 {
		return errors.New("laboratory stage limit must be positive")
	}
	if l.MaximumRunDuration <= 0 || l.MaximumFrameAge <= 0 {
		return errors.New("laboratory time limits must be positive")
	}
	return nil
}
func (l LaboratoryLimits) CheckProfile(profile Profile) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if len(profile.Stages) > l.MaximumStages {
		return fmt.Errorf("profile contains %d stages above laboratory limit", len(profile.Stages))
	}
	low, high := profile.TemperatureBounds()
	if low < l.MinimumCelsius || high > l.MaximumCelsius {
		return fmt.Errorf("profile range %.1f..%.1f exceeds laboratory bounds", low, high)
	}
	if time.Duration(profile.PlannedSeconds())*time.Second > l.MaximumRunDuration {
		return errors.New("profile planned duration exceeds laboratory limit")
	}
	return nil
}
func (l LaboratoryLimits) CheckFrame(frame SensorFrame, now time.Time) error {
	if err := l.Validate(); err != nil {
		return err
	}
	if err := frame.Validate(); err != nil {
		return err
	}
	age := now.UTC().Sub(frame.CapturedAt.UTC())
	if age > l.MaximumFrameAge {
		return fmt.Errorf("sensor frame is too old by %s", age-l.MaximumFrameAge)
	}
	if age < -time.Minute {
		return errors.New("sensor frame capture time is too far in the future")
	}
	return nil
}
