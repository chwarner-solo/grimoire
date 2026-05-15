package entity

import "github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"

// MacGuffinSnapshot is the DTO used for persistence and reconstitution.
type MacGuffinSnapshot struct {
	ID                identity.MacGuffinID
	GameID            identity.GameID
	Name              string
	Description       string
	PlayerDescription string
	PlayerVisible     bool
	Destroyed         bool
	NPCPossessorID    identity.NarrativeCharacterID
	PCPossessorID     identity.PlayerCharacterID
	LocationID        identity.LocationID
}

// Snapshot returns a serializable representation of this MacGuffin.
func (m *MacGuffin) Snapshot() MacGuffinSnapshot {
	return MacGuffinSnapshot{
		ID:                m.id,
		GameID:            m.gameID,
		Name:              m.name,
		Description:       m.description,
		PlayerDescription: m.playerDescription,
		PlayerVisible:     m.playerVisible,
		Destroyed:         m.destroyed,
		NPCPossessorID:    m.npcPossessorID,
		PCPossessorID:     m.pcPossessorID,
		LocationID:        m.locationID,
	}
}

// ReconstituteMacGuffin rebuilds a MacGuffin from a persisted snapshot.
func ReconstituteMacGuffin(snap MacGuffinSnapshot) *MacGuffin {
	return &MacGuffin{
		id:                snap.ID,
		gameID:            snap.GameID,
		name:              snap.Name,
		description:       snap.Description,
		playerDescription: snap.PlayerDescription,
		playerVisible:     snap.PlayerVisible,
		destroyed:         snap.Destroyed,
		npcPossessorID:    snap.NPCPossessorID,
		pcPossessorID:     snap.PCPossessorID,
		locationID:        snap.LocationID,
	}
}
