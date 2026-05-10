package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// MasterNarrative is an aggregate root representing the GM's world truth,
// shared across all Campaigns of a Game.
//
// Intentional departure from the state-machine pattern used by Game and Campaign:
// MasterNarrative is purely additive — beats, acts, secrets, and lore are added
// but never removed. There is no lifecycle state machine.
type MasterNarrative struct {
	id        identity.MasterNarrativeID
	gameID    identity.GameID
	beatIDs   []identity.BeatID
	actIDs    []identity.ActID
	secretIDs []identity.SecretID
	loreIDs   []identity.LoreID
}

// CreateMasterNarrative constructs a new MasterNarrative aggregate.
func CreateMasterNarrative(id identity.MasterNarrativeID, gameID identity.GameID, source event.Source) (*MasterNarrative, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrMasterNarrativeIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	mn := &MasterNarrative{
		id:     id,
		gameID: gameID,
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "master_narrative",
		Name:       "",
		Source:     source,
	}
	return mn, []event.Event{evt}, nil
}

func (mn *MasterNarrative) AddBeat(id identity.BeatID)     { mn.beatIDs = append(mn.beatIDs, id) }
func (mn *MasterNarrative) AddAct(id identity.ActID)       { mn.actIDs = append(mn.actIDs, id) }
func (mn *MasterNarrative) AddSecret(id identity.SecretID) { mn.secretIDs = append(mn.secretIDs, id) }
func (mn *MasterNarrative) AddLore(id identity.LoreID)     { mn.loreIDs = append(mn.loreIDs, id) }

// --- Getters ---

func (mn *MasterNarrative) MasterNarrativeID() identity.MasterNarrativeID { return mn.id }
func (mn *MasterNarrative) GameID() identity.GameID                       { return mn.gameID }

// --- Snapshot / Reconstitute ---

type MasterNarrativeSnapshot struct {
	ID        identity.MasterNarrativeID
	GameID    identity.GameID
	BeatIDs   []identity.BeatID
	ActIDs    []identity.ActID
	SecretIDs []identity.SecretID
	LoreIDs   []identity.LoreID
}

func (mn *MasterNarrative) Snapshot() MasterNarrativeSnapshot {
	return MasterNarrativeSnapshot{
		ID:        mn.id,
		GameID:    mn.gameID,
		BeatIDs:   copyBeatIDs(mn.beatIDs),
		ActIDs:    copyActIDs(mn.actIDs),
		SecretIDs: copySecretIDs(mn.secretIDs),
		LoreIDs:   copyLoreIDs(mn.loreIDs),
	}
}

func ReconstituteMasterNarrative(snap MasterNarrativeSnapshot) *MasterNarrative {
	return &MasterNarrative{
		id:        snap.ID,
		gameID:    snap.GameID,
		beatIDs:   copyBeatIDs(snap.BeatIDs),
		actIDs:    copyActIDs(snap.ActIDs),
		secretIDs: copySecretIDs(snap.SecretIDs),
		loreIDs:   copyLoreIDs(snap.LoreIDs),
	}
}

// --- copy helpers ---

func copyBeatIDs(ids []identity.BeatID) []identity.BeatID {
	if ids == nil {
		return nil
	}
	cp := make([]identity.BeatID, len(ids))
	copy(cp, ids)
	return cp
}

func copyActIDs(ids []identity.ActID) []identity.ActID {
	if ids == nil {
		return nil
	}
	cp := make([]identity.ActID, len(ids))
	copy(cp, ids)
	return cp
}

func copySecretIDs(ids []identity.SecretID) []identity.SecretID {
	if ids == nil {
		return nil
	}
	cp := make([]identity.SecretID, len(ids))
	copy(cp, ids)
	return cp
}

func copyLoreIDs(ids []identity.LoreID) []identity.LoreID {
	if ids == nil {
		return nil
	}
	cp := make([]identity.LoreID, len(ids))
	copy(cp, ids)
	return cp
}
