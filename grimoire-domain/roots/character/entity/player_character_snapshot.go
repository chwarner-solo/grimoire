package entity

import "github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"

// PlayerCharacterSnapshot is the DTO used for persistence and reconstitution.
type PlayerCharacterSnapshot struct {
	ID                 identity.PlayerCharacterID
	GameID             identity.GameID
	Name               string
	Description        string
	PlayerDescription  string
	PlayerVisible      bool
	OwnerPlayerID      string
	Status             PlayerCharacterStatus
	CampaignIDs        []identity.CampaignID
	MacGuffinIDs       []identity.MacGuffinID
	FoundryCharacterID string
}

// Snapshot returns a serializable representation of this PlayerCharacter.
func (pc *PlayerCharacter) Snapshot() PlayerCharacterSnapshot {
	campaignIDs := make([]identity.CampaignID, len(pc.campaignIDs))
	copy(campaignIDs, pc.campaignIDs)
	macGuffinIDs := make([]identity.MacGuffinID, len(pc.macGuffinIDs))
	copy(macGuffinIDs, pc.macGuffinIDs)
	return PlayerCharacterSnapshot{
		ID:                 pc.id,
		GameID:             pc.gameID,
		Name:               pc.name,
		Description:        pc.description,
		PlayerDescription:  pc.playerDescription,
		PlayerVisible:      pc.playerVisible,
		OwnerPlayerID:      pc.ownerPlayerID,
		Status:             pc.status,
		CampaignIDs:        campaignIDs,
		MacGuffinIDs:       macGuffinIDs,
		FoundryCharacterID: pc.foundryCharacterID,
	}
}

// ReconstitutePlayerCharacter rebuilds a PlayerCharacter from a persisted snapshot.
func ReconstitutePlayerCharacter(snap PlayerCharacterSnapshot) *PlayerCharacter {
	campaignIDs := make([]identity.CampaignID, len(snap.CampaignIDs))
	copy(campaignIDs, snap.CampaignIDs)
	macGuffinIDs := make([]identity.MacGuffinID, len(snap.MacGuffinIDs))
	copy(macGuffinIDs, snap.MacGuffinIDs)
	return &PlayerCharacter{
		id:                 snap.ID,
		gameID:             snap.GameID,
		name:               snap.Name,
		description:        snap.Description,
		playerDescription:  snap.PlayerDescription,
		playerVisible:      snap.PlayerVisible,
		ownerPlayerID:      snap.OwnerPlayerID,
		status:             snap.Status,
		campaignIDs:        campaignIDs,
		macGuffinIDs:       macGuffinIDs,
		foundryCharacterID: snap.FoundryCharacterID,
	}
}
