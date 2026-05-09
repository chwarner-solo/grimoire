package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- newCampaign Handle ---

func (c *newCampaign) Handle(evt event.Event) (Campaign, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		if e.Relationship != "character" {
			return nil, ErrUnexpectedEvent
		}
		charID, err := identity.ParseCharacterID(e.EntityBID)
		if err != nil {
			return nil, fmt.Errorf("campaign handle: %w", err)
		}
		core := c.copyCore()
		core.characterIDs = append(core.characterIDs, charID)
		return &newCampaign{campaignCore: core}, nil
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "forming" {
			core := c.copyCore()
			return &formingCampaign{campaignCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- formingCampaign Handle ---

func (c *formingCampaign) Handle(evt event.Event) (Campaign, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		if e.Relationship != "character" {
			return nil, ErrUnexpectedEvent
		}
		charID, err := identity.ParseCharacterID(e.EntityBID)
		if err != nil {
			return nil, fmt.Errorf("campaign handle: %w", err)
		}
		core := c.copyCore()
		core.characterIDs = append(core.characterIDs, charID)
		return &formingCampaign{campaignCore: core}, nil
	case event.SessionStarted:
		core := c.copyCore()
		return &activeCampaign{campaignCore: core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- activeCampaign Handle ---

func (c *activeCampaign) Handle(evt event.Event) (Campaign, error) {
	switch evt.(type) {
	case event.SessionEnded:
		core := c.copyCore()
		return &idleCampaign{campaignCore: core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- idleCampaign Handle ---

func (c *idleCampaign) Handle(evt event.Event) (Campaign, error) {
	switch e := evt.(type) {
	case event.SessionStarted:
		core := c.copyCore()
		return &activeCampaign{campaignCore: core}, nil
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "complete" {
			core := c.copyCore()
			return &completeCampaign{campaignCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- completeCampaign Handle ---

func (c *completeCampaign) Handle(_ event.Event) (Campaign, error) {
	return nil, ErrUnexpectedEvent
}

// --- copyCore ---

func (c *campaignCore) copyCore() campaignCore {
	ids := make([]identity.CharacterID, len(c.characterIDs))
	copy(ids, c.characterIDs)
	return campaignCore{
		id:           c.id,
		name:         c.name,
		gameID:       c.gameID,
		characterIDs: ids,
	}
}

// --- ReplayCampaign ---

// ReplayCampaign rebuilds a Campaign aggregate from a sequence of events.
// The first event must be an EntityCreated for a campaign.
// The gameID is required because it's not stored in the event payload.
func ReplayCampaign(gameID identity.GameID, events []event.Event) (Campaign, error) {
	if len(events) == 0 {
		return nil, errors.New("campaign: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("campaign: first event must be EntityCreated")
	}

	campID, err := identity.ParseCampaignID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("campaign replay: %w", err)
	}

	c, err := CreateCampaign(campID, first.Name, gameID)
	if err != nil {
		return nil, fmt.Errorf("campaign replay: %w", err)
	}

	var current Campaign = c
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("campaign replay: %w", err)
		}
	}
	return current, nil
}
