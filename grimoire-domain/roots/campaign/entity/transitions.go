package entity

import (
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- NewCampaign transitions ---

func (c *newCampaign) AddCharacter(id identity.CharacterID, source event.Source) (NewCampaign, []event.Event, error) {
	if err := validateCharacterID(id); err != nil {
		return nil, nil, err
	}
	if c.hasCharacter(id) {
		return nil, nil, ErrCharacterAlreadyAdded
	}
	c.characterIDs = append(c.characterIDs, id)
	evt := event.EntityLinked{
		EntityAID:    c.id.String(),
		EntityBID:    id.String(),
		Relationship: "character",
		Source:       source,
	}
	return c, []event.Event{evt}, nil
}

func (c *newCampaign) BeginFormation(source event.Source) (FormingCampaign, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: c.id.String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "forming",
		Source:   source,
	}
	return &formingCampaign{campaignCore: c.campaignCore}, []event.Event{evt}, nil
}

// --- FormingCampaign transitions ---

func (c *formingCampaign) AddCharacter(id identity.CharacterID, source event.Source) (FormingCampaign, []event.Event, error) {
	if err := validateCharacterID(id); err != nil {
		return nil, nil, err
	}
	if c.hasCharacter(id) {
		return nil, nil, ErrCharacterAlreadyAdded
	}
	c.characterIDs = append(c.characterIDs, id)
	evt := event.EntityLinked{
		EntityAID:    c.id.String(),
		EntityBID:    id.String(),
		Relationship: "character",
		Source:       source,
	}
	return c, []event.Event{evt}, nil
}

func (c *formingCampaign) StartFirstSession(id identity.SessionID, date time.Time) (ActiveCampaign, []event.Event, error) {
	if err := validateSessionID(id); err != nil {
		return nil, nil, err
	}
	if len(c.characterIDs) == 0 {
		return nil, nil, ErrNoCharacters
	}
	evt := event.SessionStarted{
		SessionID:  id.String(),
		CampaignID: c.id.String(),
		Date:       date,
	}
	return &activeCampaign{campaignCore: c.campaignCore}, []event.Event{evt}, nil
}

// --- ActiveCampaign transitions ---

func (c *activeCampaign) NotifySessionSummarized() (IdleCampaign, error) {
	return &idleCampaign{campaignCore: c.campaignCore}, nil
}

func (c *activeCampaign) MoveToLocation(locationID identity.LocationID, source event.Source) (ActiveCampaign, []event.Event, error) {
	if locationID.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	oldValue := ""
	if !c.currentLocationID.IsZero() {
		oldValue = c.currentLocationID.String()
	}
	c.currentLocationID = locationID
	evt := event.EntityUpdated{
		EntityID: c.id.String(),
		Field:    "current_location_id",
		OldValue: oldValue,
		NewValue: locationID.String(),
		Source:   source,
	}
	return c, []event.Event{evt}, nil
}

// --- IdleCampaign transitions ---

func (c *idleCampaign) StartNewSession(id identity.SessionID, date time.Time) (ActiveCampaign, []event.Event, error) {
	if err := validateSessionID(id); err != nil {
		return nil, nil, err
	}
	evt := event.SessionStarted{
		SessionID:  id.String(),
		CampaignID: c.id.String(),
		Date:       date,
	}
	return &activeCampaign{campaignCore: c.campaignCore}, []event.Event{evt}, nil
}

func (c *idleCampaign) Complete(source event.Source) (CompleteCampaign, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: c.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "complete",
		Source:   source,
	}
	return &completeCampaign{campaignCore: c.campaignCore}, []event.Event{evt}, nil
}

// --- Helpers ---

func validateCharacterID(id identity.CharacterID) error {
	if id.IsZero() {
		return ErrCharacterIDRequired
	}
	return nil
}

func validateSessionID(id identity.SessionID) error {
	if id.IsZero() {
		return ErrSessionIDRequired
	}
	return nil
}
