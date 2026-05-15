package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// CampaignSnapshot is the DTO used for persistence and reconstitution.
type CampaignSnapshot struct {
	ID                identity.CampaignID
	Name              string
	GameID            identity.GameID
	State             string
	CharacterIDs      []identity.PlayerCharacterID
	CurrentLocationID identity.LocationID
}

// ReconstituteCampaign rebuilds a Campaign aggregate from a persisted snapshot.
// It trusts persisted data and bypasses constructor validation.
func ReconstituteCampaign(snap CampaignSnapshot) (Campaign, error) {
	core := campaignCore{
		id:                snap.ID,
		name:              snap.Name,
		gameID:            snap.GameID,
		characterIDs:      snap.CharacterIDs,
		currentLocationID: snap.CurrentLocationID,
	}
	switch snap.State {
	case "new":
		return &newCampaign{campaignCore: core}, nil
	case "forming":
		return &formingCampaign{campaignCore: core}, nil
	case "active":
		return &activeCampaign{campaignCore: core}, nil
	case "idle":
		return &idleCampaign{campaignCore: core}, nil
	case "complete":
		return &completeCampaign{campaignCore: core}, nil
	default:
		return nil, fmt.Errorf("%w: unknown state %q", ErrInvalidCampaignState, snap.State)
	}
}
