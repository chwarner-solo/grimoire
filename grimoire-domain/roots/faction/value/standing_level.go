package value

import "errors"

var (
	ErrStandingLevelNameRequired    = errors.New("standing level: name is required")
	ErrStandingLevelThresholdNegative = errors.New("standing level: threshold must be >= 0")
	ErrStandingLevelOrdinalNegative   = errors.New("standing level: ordinal must be >= 0")
)

// StandingLevel defines a named tier in a faction's standing scale.
// It is a value object — immutable after construction.
type StandingLevel struct {
	name      string
	threshold int
	ordinal   int
}

// NewStandingLevel constructs a StandingLevel with validation.
func NewStandingLevel(name string, threshold int, ordinal int) (StandingLevel, error) {
	if name == "" {
		return StandingLevel{}, ErrStandingLevelNameRequired
	}
	if threshold < 0 {
		return StandingLevel{}, ErrStandingLevelThresholdNegative
	}
	if ordinal < 0 {
		return StandingLevel{}, ErrStandingLevelOrdinalNegative
	}
	return StandingLevel{
		name:      name,
		threshold: threshold,
		ordinal:   ordinal,
	}, nil
}

func (s StandingLevel) Name() string  { return s.name }
func (s StandingLevel) Threshold() int { return s.threshold }
func (s StandingLevel) Ordinal() int   { return s.ordinal }
