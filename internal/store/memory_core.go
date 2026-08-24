package store

import (
	"sync"
	"thermal-cycle-lab/internal/domain"
)

type Memory struct {
	mu        sync.RWMutex
	specimens map[string]domain.Specimen
	profiles  map[string]domain.Profile
	runs      map[string]domain.ExperimentRun
	frames    map[string][]domain.SensorFrame
	cursors   map[string]domain.SensorCursor
	reports   map[string]domain.RunReport
}

func NewMemory() *Memory {
	return &Memory{specimens: map[string]domain.Specimen{}, profiles: map[string]domain.Profile{}, runs: map[string]domain.ExperimentRun{}, frames: map[string][]domain.SensorFrame{}, cursors: map[string]domain.SensorCursor{}, reports: map[string]domain.RunReport{}}
}
func cursorKey(runID, sensorID string) string { return runID + "\x00" + sensorID }
func cloneSpecimens(source map[string]domain.Specimen) []domain.Specimen {
	values := make([]domain.Specimen, 0, len(source))
	for _, item := range source {
		values = append(values, item)
	}
	return values
}
func cloneProfiles(source map[string]domain.Profile) []domain.Profile {
	values := make([]domain.Profile, 0, len(source))
	for _, item := range source {
		item.Stages = item.CloneStages()
		values = append(values, item)
	}
	return values
}
func cloneRuns(source map[string]domain.ExperimentRun) []domain.ExperimentRun {
	values := make([]domain.ExperimentRun, 0, len(source))
	for _, item := range source {
		values = append(values, item)
	}
	return values
}
func cloneFrames(source []domain.SensorFrame) []domain.SensorFrame {
	return append([]domain.SensorFrame(nil), source...)
}
