package domain

import (
	"errors"
	"strings"
	"time"
)

type SpecimenState string

const (
	SpecimenReady   SpecimenState = "ready"
	SpecimenInUse   SpecimenState = "in_use"
	SpecimenRetired SpecimenState = "retired"
)

type Specimen struct {
	ID          string        `json:"id"`
	Label       string        `json:"label"`
	Material    string        `json:"material"`
	MinCelsius  float64       `json:"min_celsius"`
	MaxCelsius  float64       `json:"max_celsius"`
	State       SpecimenState `json:"state"`
	ActiveRunID string        `json:"active_run_id,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func NewSpecimen(id, label, material string, minCelsius, maxCelsius float64, now time.Time) (Specimen, error) {
	item := Specimen{ID: strings.TrimSpace(id), Label: strings.TrimSpace(label), Material: strings.TrimSpace(material), MinCelsius: minCelsius, MaxCelsius: maxCelsius, State: SpecimenReady, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if err := item.Validate(); err != nil {
		return Specimen{}, err
	}
	return item, nil
}

func (s Specimen) Validate() error {
	if s.ID == "" {
		return errors.New("specimen id is required")
	}
	if len(s.Label) < 3 {
		return errors.New("specimen label must contain at least three characters")
	}
	if s.Material == "" {
		return errors.New("specimen material is required")
	}
	if s.MinCelsius >= s.MaxCelsius {
		return errors.New("specimen temperature window is invalid")
	}
	if s.MinCelsius < -180 || s.MaxCelsius > 420 {
		return errors.New("specimen temperature window exceeds laboratory limits")
	}
	switch s.State {
	case SpecimenReady, SpecimenInUse, SpecimenRetired:
	default:
		return errors.New("specimen state is invalid")
	}
	if s.State == SpecimenInUse && s.ActiveRunID == "" {
		return errors.New("active specimen must reference a run")
	}
	if s.State != SpecimenInUse && s.ActiveRunID != "" {
		return errors.New("inactive specimen cannot reference a run")
	}
	return nil
}

func (s Specimen) Acquire(runID string, now time.Time) (Specimen, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return Specimen{}, errors.New("run id is required to acquire specimen")
	}
	if s.State != SpecimenReady {
		return Specimen{}, errors.New("specimen is not available")
	}
	s.State, s.ActiveRunID, s.UpdatedAt = SpecimenInUse, runID, now.UTC()
	return s, s.Validate()
}

func (s Specimen) Release(now time.Time) (Specimen, error) {
	if s.State != SpecimenInUse {
		return Specimen{}, errors.New("only an active specimen can be released")
	}
	s.State, s.ActiveRunID, s.UpdatedAt = SpecimenReady, "", now.UTC()
	return s, s.Validate()
}

func (s Specimen) Retire(now time.Time) (Specimen, error) {
	if s.State == SpecimenInUse {
		return Specimen{}, errors.New("active specimen cannot be retired")
	}
	s.State, s.ActiveRunID, s.UpdatedAt = SpecimenRetired, "", now.UTC()
	return s, s.Validate()
}

func (s Specimen) Supports(low, high float64) bool {
	return low >= s.MinCelsius && high <= s.MaxCelsius
}
