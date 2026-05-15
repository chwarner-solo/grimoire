package entity

import "github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"

// PlayerCharacterNarrativeSnapshot is the DTO used for persistence and reconstitution.
type PlayerCharacterNarrativeSnapshot struct {
	ID                    identity.PlayerCharacterNarrativeID
	CharacterID           identity.PlayerCharacterID
	CampaignID            identity.CampaignID
	GameID                identity.GameID
	BackgroundBeatIDs     []identity.BeatID
	PersonalBeatIDs       []identity.BeatID
	RevealedBackgroundIDs []identity.BeatID
}

// Snapshot returns a serializable representation of this PlayerCharacterNarrative.
func (pcn *PlayerCharacterNarrative) Snapshot() PlayerCharacterNarrativeSnapshot {
	bgIDs := make([]identity.BeatID, len(pcn.backgroundBeatIDs))
	copy(bgIDs, pcn.backgroundBeatIDs)
	personalIDs := make([]identity.BeatID, len(pcn.personalBeatIDs))
	copy(personalIDs, pcn.personalBeatIDs)
	revealedIDs := make([]identity.BeatID, len(pcn.revealedBackgroundIDs))
	copy(revealedIDs, pcn.revealedBackgroundIDs)
	return PlayerCharacterNarrativeSnapshot{
		ID:                    pcn.id,
		CharacterID:           pcn.characterID,
		CampaignID:            pcn.campaignID,
		GameID:                pcn.gameID,
		BackgroundBeatIDs:     bgIDs,
		PersonalBeatIDs:       personalIDs,
		RevealedBackgroundIDs: revealedIDs,
	}
}

// ReconstitutePlayerCharacterNarrative rebuilds from a persisted snapshot.
func ReconstitutePlayerCharacterNarrative(snap PlayerCharacterNarrativeSnapshot) *PlayerCharacterNarrative {
	bgIDs := make([]identity.BeatID, len(snap.BackgroundBeatIDs))
	copy(bgIDs, snap.BackgroundBeatIDs)
	personalIDs := make([]identity.BeatID, len(snap.PersonalBeatIDs))
	copy(personalIDs, snap.PersonalBeatIDs)
	revealedIDs := make([]identity.BeatID, len(snap.RevealedBackgroundIDs))
	copy(revealedIDs, snap.RevealedBackgroundIDs)
	return &PlayerCharacterNarrative{
		id:                    snap.ID,
		characterID:           snap.CharacterID,
		campaignID:            snap.CampaignID,
		gameID:                snap.GameID,
		backgroundBeatIDs:     bgIDs,
		personalBeatIDs:       personalIDs,
		revealedBackgroundIDs: revealedIDs,
	}
}
