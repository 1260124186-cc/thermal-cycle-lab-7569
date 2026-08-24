package domain

import (
	"errors"
	"math"
	"strings"
	"time"
)

type SensorFrame struct {
	RunID      string    `json:"run_id"`
	SensorID   string    `json:"sensor_id"`
	Sequence   int64     `json:"sequence"`
	StageIndex int       `json:"stage_index"`
	Celsius    float64   `json:"celsius"`
	Humidity   float64   `json:"humidity"`
	CapturedAt time.Time `json:"captured_at"`
}

type FrameAssessment struct {
	Accepted        bool    `json:"accepted"`
	WithinTolerance bool    `json:"within_tolerance"`
	Delta           float64 `json:"delta"`
	Message         string  `json:"message"`
}

func (f SensorFrame) Validate() error {
	if strings.TrimSpace(f.RunID) == "" || strings.TrimSpace(f.SensorID) == "" {
		return errors.New("frame run and sensor identity are required")
	}
	if f.Sequence < 1 {
		return errors.New("frame sequence must be positive")
	}
	if f.StageIndex < 0 {
		return errors.New("frame stage index cannot be negative")
	}
	if math.IsNaN(f.Celsius) || math.IsInf(f.Celsius, 0) {
		return errors.New("frame temperature is not finite")
	}
	if f.Celsius < -200 || f.Celsius > 500 {
		return errors.New("frame temperature is outside sensor range")
	}
	if f.Humidity < 0 || f.Humidity > 100 {
		return errors.New("frame humidity must be between zero and one hundred")
	}
	if f.CapturedAt.IsZero() {
		return errors.New("frame capture time is required")
	}
	return nil
}

func AssessFrame(frame SensorFrame, stage Stage) (FrameAssessment, error) {
	if err := frame.Validate(); err != nil {
		return FrameAssessment{}, err
	}
	delta := math.Abs(frame.Celsius - stage.TargetCelsius)
	within := delta <= stage.Tolerance
	message := "frame accepted within stage tolerance"
	if !within {
		message = "frame accepted with temperature alert"
	}
	return FrameAssessment{Accepted: true, WithinTolerance: within, Delta: delta, Message: message}, nil
}

type SensorCursor struct {
	SensorID       string    `json:"sensor_id"`
	LastSequence   int64     `json:"last_sequence"`
	LastCapturedAt time.Time `json:"last_captured_at"`
}

func AdvanceCursor(cursor SensorCursor, frame SensorFrame) (SensorCursor, error) {
	if cursor.SensorID != "" && cursor.SensorID != frame.SensorID {
		return SensorCursor{}, errors.New("sensor cursor belongs to another sensor")
	}
	if frame.Sequence <= cursor.LastSequence {
		return SensorCursor{}, errors.New("sensor frame sequence did not advance")
	}
	if !cursor.LastCapturedAt.IsZero() && frame.CapturedAt.Before(cursor.LastCapturedAt) {
		return SensorCursor{}, errors.New("sensor frame capture time moved backwards")
	}
	return SensorCursor{SensorID: frame.SensorID, LastSequence: frame.Sequence, LastCapturedAt: frame.CapturedAt.UTC()}, nil
}
