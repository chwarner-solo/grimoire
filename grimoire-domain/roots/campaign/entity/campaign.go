package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// campaignCore holds the shared state for all Campaign states.
type campaignCore struct {
	id           identity.CampaignID
	name         string
	gameID       identity.GameID
	characterIDs []identity.CharacterID
}

func (c *campaignCore) CampaignID() identity.CampaignID { return c.id }
func (c *campaignCore) CampaignName() string            { return c.name }
func (c *campaignCore) isCampaign()                     {}

func (c *campaignCore) hasCharacter(id identity.CharacterID) bool {
	for _, cid := range c.characterIDs {
		if cid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *campaignCore) snapshot(state string) CampaignSnapshot {
	ids := make([]identity.CharacterID, len(c.characterIDs))
	copy(ids, c.characterIDs)
	return CampaignSnapshot{
		ID:           c.id,
		Name:         c.name,
		GameID:       c.gameID,
		State:        state,
		CharacterIDs: ids,
	}
}

// CreateCampaign constructs a new Campaign aggregate in the New state.
func CreateCampaign(id identity.CampaignID, name string, gameID identity.GameID, source event.Source) (NewCampaign, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrCampaignIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrCampaignNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "campaign",
		Name:       name,
		Source:     source,
	}
	return &newCampaign{campaignCore: campaignCore{
		id:     id,
		name:   name,
		gameID: gameID,
	}}, []event.Event{evt}, nil
}

// --- State structs ---

type newCampaign struct{ campaignCore }
type formingCampaign struct{ campaignCore }
type activeCampaign struct{ campaignCore }
type idleCampaign struct{ campaignCore }
type completeCampaign struct{ campaignCore }

// Snapshot implementations
func (c *newCampaign) Snapshot() CampaignSnapshot      { return c.snapshot("new") }
func (c *formingCampaign) Snapshot() CampaignSnapshot   { return c.snapshot("forming") }
func (c *activeCampaign) Snapshot() CampaignSnapshot    { return c.snapshot("active") }
func (c *idleCampaign) Snapshot() CampaignSnapshot      { return c.snapshot("idle") }
func (c *completeCampaign) Snapshot() CampaignSnapshot  { return c.snapshot("complete") }
