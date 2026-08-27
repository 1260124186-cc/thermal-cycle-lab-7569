package store

import (
	"context"
	"errors"
	"thermal-cycle-lab/internal/domain"
)

var ErrNotFound = errors.New("record not found")
var ErrConflict = errors.New("record already exists")

type SpecimenRepository interface {
	CreateSpecimen(context.Context, domain.Specimen) error
	GetSpecimen(context.Context, string) (domain.Specimen, error)
	UpdateSpecimen(context.Context, domain.Specimen) error
	ListSpecimens(context.Context) ([]domain.Specimen, error)
}
type ProfileRepository interface {
	CreateProfile(context.Context, domain.Profile) error
	GetProfile(context.Context, string) (domain.Profile, error)
	ListProfiles(context.Context) ([]domain.Profile, error)
}
type RunRepository interface {
	CreateRun(context.Context, domain.ExperimentRun) error
	GetRun(context.Context, string) (domain.ExperimentRun, error)
	UpdateRun(context.Context, domain.ExperimentRun) error
	DeleteRun(context.Context, string) error
	ListRuns(context.Context) ([]domain.ExperimentRun, error)
	AppendFrame(context.Context, domain.SensorFrame) error
	FramesForRun(context.Context, string) ([]domain.SensorFrame, error)
	GetCursor(context.Context, string, string) (domain.SensorCursor, error)
	PutCursor(context.Context, string, domain.SensorCursor) error
	SaveReport(context.Context, domain.RunReport) error
	GetReport(context.Context, string) (domain.RunReport, error)
}
type Repository interface {
	SpecimenRepository
	ProfileRepository
	RunRepository
}

func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }
