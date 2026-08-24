package store

import (
	"context"
	"errors"
	"sort"
	"thermal-cycle-lab/internal/domain"
)

type Snapshot struct {
	Specimens []domain.Specimen               `json:"specimens"`
	Profiles  []domain.Profile                `json:"profiles"`
	Runs      []domain.ExperimentRun          `json:"runs"`
	Frames    map[string][]domain.SensorFrame `json:"frames"`
	Reports   []domain.RunReport              `json:"reports"`
}

func (m *Memory) Snapshot(ctx context.Context) (Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return Snapshot{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := Snapshot{Specimens: cloneSpecimens(m.specimens), Profiles: cloneProfiles(m.profiles), Runs: cloneRuns(m.runs), Frames: make(map[string][]domain.SensorFrame, len(m.frames)), Reports: make([]domain.RunReport, 0, len(m.reports))}
	for runID, frames := range m.frames {
		result.Frames[runID] = cloneFrames(frames)
	}
	for _, report := range m.reports {
		report.Stages = append([]domain.StageSummary(nil), report.Stages...)
		result.Reports = append(result.Reports, report)
	}
	sort.Slice(result.Specimens, func(i, j int) bool { return result.Specimens[i].ID < result.Specimens[j].ID })
	sort.Slice(result.Profiles, func(i, j int) bool { return result.Profiles[i].ID < result.Profiles[j].ID })
	sort.Slice(result.Runs, func(i, j int) bool { return result.Runs[i].ID < result.Runs[j].ID })
	sort.Slice(result.Reports, func(i, j int) bool { return result.Reports[i].RunID < result.Reports[j].RunID })
	return result, nil
}
func (s Snapshot) Validate() error {
	specimens := map[string]domain.Specimen{}
	for _, item := range s.Specimens {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, ok := specimens[item.ID]; ok {
			return errors.New("snapshot repeats specimen")
		}
		specimens[item.ID] = item
	}
	profiles := map[string]domain.Profile{}
	for _, item := range s.Profiles {
		if err := item.Validate(); err != nil {
			return err
		}
		if _, ok := profiles[item.ID]; ok {
			return errors.New("snapshot repeats profile")
		}
		profiles[item.ID] = item
	}
	for _, run := range s.Runs {
		if err := run.Validate(); err != nil {
			return err
		}
		if _, ok := specimens[run.SpecimenID]; !ok {
			return errors.New("snapshot run references missing specimen")
		}
		if _, ok := profiles[run.ProfileID]; !ok {
			return errors.New("snapshot run references missing profile")
		}
	}
	return nil
}
