package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// FactionSnapshot is the DTO used for persistence and reconstitution.
type FactionSnapshot struct {
	ID             identity.FactionID
	GameID         identity.GameID
	Name           string
	Description    string
	Goals          string
	PlayerVisible  bool
	State          string
	MemberIDs      []identity.FactionMembershipID
	AlliedWithIDs  []identity.FactionID
	AtWarWithIDs   []identity.FactionID
	StandingLevels []value.StandingLevel
}

// ReconstituteFaction rebuilds a Faction aggregate from a persisted snapshot.
func ReconstituteFaction(snap FactionSnapshot) (Faction, error) {
	core := factionCore{
		id:             snap.ID,
		gameID:         snap.GameID,
		name:           snap.Name,
		description:    snap.Description,
		goals:          snap.Goals,
		playerVisible:  snap.PlayerVisible,
		memberIDs:      snap.MemberIDs,
		alliedWithIDs:  snap.AlliedWithIDs,
		atWarWithIDs:   snap.AtWarWithIDs,
		standingLevels: snap.StandingLevels,
	}
	switch snap.State {
	case "new":
		return &newFaction{factionCore: core}, nil
	case "draft":
		return &draftFaction{factionCore: core}, nil
	case "active":
		return &activeFaction{factionCore: core}, nil
	case "idle":
		return &idleFaction{factionCore: core}, nil
	case "archived":
		return &archivedFaction{factionCore: core}, nil
	default:
		return nil, fmt.Errorf("%w: unknown state %q", ErrInvalidFactionState, snap.State)
	}
}
