package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

var (
	ErrBeatIDRequired                  = errors.New("narrative: beat id is required")
	ErrBeatNameRequired                = errors.New("narrative: beat name is required")
	ErrBeatDescriptionRequired         = errors.New("narrative: beat description is required")
	ErrBeatPlayerDescriptionRequired   = errors.New("narrative: beat player description is required")
	ErrGameIDRequired                  = errors.New("narrative: game id is required")
	ErrCampaignIDRequired              = errors.New("narrative: campaign id is required")
	ErrMasterNarrativeIDRequired       = errors.New("narrative: master narrative id is required")
	ErrCampaignNarrativeIDRequired     = errors.New("narrative: campaign narrative id is required")
	ErrInvalidScope                    = errors.New("narrative: invalid scope")
)

// ErrCycleDetected indicates that adding a prerequisite would create a cycle in the beat DAG.
type ErrCycleDetected struct {
	BeatID identity.BeatID
}

func (e ErrCycleDetected) Error() string {
	return fmt.Sprintf("narrative: cycle detected at beat %s", e.BeatID.String())
}

// ErrPrerequisiteNotMet indicates that a beat cannot be discovered because its prerequisites have not been met.
type ErrPrerequisiteNotMet struct {
	BeatID identity.BeatID
}

func (e ErrPrerequisiteNotMet) Error() string {
	return fmt.Sprintf("narrative: prerequisite not met for beat %s", e.BeatID.String())
}
