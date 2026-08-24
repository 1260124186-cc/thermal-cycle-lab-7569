package domain

import (
	"errors"
	"sort"
	"time"
)

type StageSummary struct {
	StageIndex int     `json:"stage_index"`
	Samples    int     `json:"samples"`
	Minimum    float64 `json:"minimum"`
	Maximum    float64 `json:"maximum"`
	Average    float64 `json:"average"`
	Alerts     int     `json:"alerts"`
}
type RunReport struct {
	RunID       string         `json:"run_id"`
	SpecimenID  string         `json:"specimen_id"`
	ProfileID   string         `json:"profile_id"`
	Result      string         `json:"result"`
	TotalFrames int            `json:"total_frames"`
	TotalAlerts int            `json:"total_alerts"`
	Stages      []StageSummary `json:"stages"`
	GeneratedAt time.Time      `json:"generated_at"`
}

func BuildRunReport(run ExperimentRun, frames []SensorFrame, profile Profile, now time.Time) (RunReport, error) {
	if run.State != RunCompleted {
		return RunReport{}, errors.New("report is available only for completed run")
	}
	if len(frames) == 0 {
		return RunReport{}, errors.New("report requires sensor frames")
	}
	grouped := make(map[int][]SensorFrame)
	for _, frame := range frames {
		grouped[frame.StageIndex] = append(grouped[frame.StageIndex], frame)
	}
	indexes := make([]int, 0, len(grouped))
	for index := range grouped {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	summaries := make([]StageSummary, 0, len(indexes))
	alerts := 0
	for _, index := range indexes {
		stage, err := profile.StageAt(index)
		if err != nil {
			return RunReport{}, err
		}
		values := grouped[index]
		summary := StageSummary{StageIndex: index, Samples: len(values), Minimum: values[0].Celsius, Maximum: values[0].Celsius}
		total := 0.0
		for _, frame := range values {
			total += frame.Celsius
			if frame.Celsius < summary.Minimum {
				summary.Minimum = frame.Celsius
			}
			if frame.Celsius > summary.Maximum {
				summary.Maximum = frame.Celsius
			}
			assessment, err := AssessFrame(frame, stage)
			if err != nil {
				return RunReport{}, err
			}
			if !assessment.WithinTolerance {
				summary.Alerts++
				alerts++
			}
		}
		summary.Average = total / float64(len(values))
		summaries = append(summaries, summary)
	}
	result := "stable"
	if alerts > 0 {
		result = "review_required"
	}
	return RunReport{RunID: run.ID, SpecimenID: run.SpecimenID, ProfileID: run.ProfileID, Result: result, TotalFrames: len(frames), TotalAlerts: alerts, Stages: summaries, GeneratedAt: now.UTC()}, nil
}

func (r RunReport) Validate() error {
	if r.RunID == "" || r.SpecimenID == "" || r.ProfileID == "" {
		return errors.New("report identity is incomplete")
	}
	if r.TotalFrames < 1 || len(r.Stages) < 1 {
		return errors.New("report contains no observations")
	}
	if r.Result != "stable" && r.Result != "review_required" {
		return errors.New("report result is invalid")
	}
	sum := 0
	alerts := 0
	for _, stage := range r.Stages {
		sum += stage.Samples
		alerts += stage.Alerts
	}
	if sum != r.TotalFrames || alerts != r.TotalAlerts {
		return errors.New("report totals do not match stage summaries")
	}
	return nil
}
