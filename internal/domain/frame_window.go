package domain

import (
	"errors"
	"math"
	"sort"
	"time"
)

type FrameWindow struct {
	SensorID        string    `json:"sensor_id"`
	FirstSequence   int64     `json:"first_sequence"`
	LastSequence    int64     `json:"last_sequence"`
	FirstCapturedAt time.Time `json:"first_captured_at"`
	LastCapturedAt  time.Time `json:"last_captured_at"`
	Minimum         float64   `json:"minimum"`
	Maximum         float64   `json:"maximum"`
	Mean            float64   `json:"mean"`
	Spread          float64   `json:"spread"`
	Samples         int       `json:"samples"`
}

func SummarizeFrameWindow(frames []SensorFrame) (FrameWindow, error) {
	if len(frames) == 0 {
		return FrameWindow{}, errors.New("frame window cannot be empty")
	}
	copied := append([]SensorFrame(nil), frames...)
	sort.Slice(copied, func(i, j int) bool {
		if copied[i].Sequence == copied[j].Sequence {
			return copied[i].CapturedAt.Before(copied[j].CapturedAt)
		}
		return copied[i].Sequence < copied[j].Sequence
	})
	sensorID := copied[0].SensorID
	summary := FrameWindow{SensorID: sensorID, FirstSequence: copied[0].Sequence, LastSequence: copied[len(copied)-1].Sequence, FirstCapturedAt: copied[0].CapturedAt, LastCapturedAt: copied[len(copied)-1].CapturedAt, Minimum: copied[0].Celsius, Maximum: copied[0].Celsius, Samples: len(copied)}
	total := 0.0
	for index, frame := range copied {
		if err := frame.Validate(); err != nil {
			return FrameWindow{}, err
		}
		if frame.SensorID != sensorID {
			return FrameWindow{}, errors.New("frame window mixes multiple sensors")
		}
		if index > 0 && frame.Sequence == copied[index-1].Sequence {
			return FrameWindow{}, errors.New("frame window contains repeated sequence")
		}
		total += frame.Celsius
		if frame.Celsius < summary.Minimum {
			summary.Minimum = frame.Celsius
		}
		if frame.Celsius > summary.Maximum {
			summary.Maximum = frame.Celsius
		}
	}
	summary.Mean = total / float64(len(copied))
	summary.Spread = summary.Maximum - summary.Minimum
	return summary, nil
}

func (w FrameWindow) Duration() time.Duration { return w.LastCapturedAt.Sub(w.FirstCapturedAt) }
func (w FrameWindow) SequenceCoverage() float64 {
	expected := w.LastSequence - w.FirstSequence + 1
	if expected <= 0 {
		return 0
	}
	return float64(w.Samples) / float64(expected)
}
func (w FrameWindow) Stable(maxSpread float64) bool {
	return w.Samples >= 2 && !math.IsNaN(w.Spread) && w.Spread <= maxSpread && w.SequenceCoverage() >= 0.9
}
