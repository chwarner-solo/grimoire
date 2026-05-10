package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

var (
	ErrFactionIDRequired                  = errors.New("faction: id is required")
	ErrFactionNameRequired                = errors.New("faction: name is required")
	ErrGameIDRequired                     = errors.New("faction: game id is required")
	ErrMembershipIDRequired               = errors.New("faction: membership id is required")
	ErrRankRequired                       = errors.New("faction: rank is required")
	ErrCharacterIDRequired                = errors.New("faction: character id is required")
	ErrCannotActivateWithoutMembers       = errors.New("faction: cannot activate without members")
	ErrCannotActivateWithoutStandingLevels = errors.New("faction: cannot activate without standing levels")
	ErrMemberAlreadyExists                = errors.New("faction: member already exists")
	ErrAllianceAlreadyExists              = errors.New("faction: alliance already exists")
	ErrWarAlreadyExists                   = errors.New("faction: war already exists")
	ErrInvalidFactionState                = errors.New("faction: invalid state")
	ErrUnexpectedEvent                    = errors.New("faction: unexpected event for current state")
)

// ErrCannotBeAlliedAndAtWar is returned when attempting to ally with a faction
// that is currently at war.
type ErrCannotBeAlliedAndAtWar struct {
	FactionID identity.FactionID
}

func (e ErrCannotBeAlliedAndAtWar) Error() string {
	return fmt.Sprintf("faction: cannot ally with faction %s — currently at war", e.FactionID)
}

// ErrCannotBeAtWarAndAllied is returned when attempting to declare war on a faction
// that is currently allied.
type ErrCannotBeAtWarAndAllied struct {
	FactionID identity.FactionID
}

func (e ErrCannotBeAtWarAndAllied) Error() string {
	return fmt.Sprintf("faction: cannot declare war on faction %s — currently allied", e.FactionID)
}
