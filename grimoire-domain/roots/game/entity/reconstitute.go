package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameSnapshot is the DTO used for persistence and reconstitution.
type GameSnapshot struct {
	ID                  identity.GameID
	GMID                string
	Name                string
	MasterNarrativeID   identity.MasterNarrativeID
	State               string
	CampaignIDs         []identity.CampaignID
	ActiveCampaignCount int
}

// ReconstituteGame rebuilds a Game aggregate from a persisted snapshot.
// It trusts persisted data and bypasses constructor validation.
func ReconstituteGame(snap GameSnapshot) (Game, error) {
	core := gameCore{
		id:                  snap.ID,
		gmID:                snap.GMID,
		name:                snap.Name,
		masterNarrativeID:   snap.MasterNarrativeID,
		campaignIDs:         snap.CampaignIDs,
		activeCampaignCount: snap.ActiveCampaignCount,
	}
	switch snap.State {
	case "new":
		return &newGame{gameCore: core}, nil
	case "draft":
		return &draftGame{gameCore: core}, nil
	case "active":
		return &activeGame{gameCore: core}, nil
	case "idle":
		return &idleGame{gameCore: core}, nil
	case "archived":
		return &archivedGame{gameCore: core}, nil
	default:
		return nil, fmt.Errorf("%w: unknown state %q", ErrInvalidGameState, snap.State)
	}
}
