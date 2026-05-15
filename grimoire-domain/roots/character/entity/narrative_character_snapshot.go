package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// NPCSnapshot is the DTO used for persistence and reconstitution.
type NPCSnapshot struct {
	ID                 identity.NarrativeCharacterID
	GameID             identity.GameID
	State              string
	Name               string
	Description        string
	PlayerDescription  string
	PlayerVisible      bool
	LocationID         identity.LocationID
	MacGuffinIDs       []identity.MacGuffinID
	FoundryCharacterID string
}

// ReconstituteNPC rebuilds an NPC aggregate from a persisted snapshot.
func ReconstituteNPC(snap NPCSnapshot) (NPC, error) {
	core := npcCore{
		id:                 snap.ID,
		gameID:             snap.GameID,
		name:               snap.Name,
		description:        snap.Description,
		playerDescription:  snap.PlayerDescription,
		playerVisible:      snap.PlayerVisible,
		locationID:         snap.LocationID,
		macGuffinIDs:       snap.MacGuffinIDs,
		foundryCharacterID: snap.FoundryCharacterID,
	}
	switch snap.State {
	case "new":
		return &newNPC{npcCore: core}, nil
	case "draft":
		return &draftNPC{npcCore: core}, nil
	case "active":
		return &activeNPC{npcCore: core}, nil
	case "idle":
		return &idleNPC{npcCore: core}, nil
	case "archived":
		return &archivedNPC{npcCore: core}, nil
	default:
		return nil, fmt.Errorf("%w: unknown state %q", ErrInvalidNPCState, snap.State)
	}
}
