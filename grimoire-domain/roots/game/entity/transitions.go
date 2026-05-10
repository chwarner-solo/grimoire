package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- NewGame transitions ---

func (g *newGame) OnNarrativeCreated(id identity.MasterNarrativeID) (DraftGame, error) {
	if id.IsZero() {
		return nil, ErrMasterNarrativeIDRequired
	}
	core := g.gameCore
	core.masterNarrativeID = id
	return &draftGame{gameCore: core}, nil
}

// --- DraftGame transitions ---

func (g *draftGame) LinkCampaign(id identity.CampaignID, source event.Source) (DraftGame, []event.Event, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, nil, err
	}
	if g.hasCampaign(id) {
		return nil, nil, ErrCampaignAlreadyLinked
	}
	g.campaignIDs = append(g.campaignIDs, id)
	evt := event.EntityLinked{
		EntityAID:    g.id.String(),
		EntityBID:    id.String(),
		Relationship: "campaign",
		Source:       source,
	}
	return g, []event.Event{evt}, nil
}

func (g *draftGame) ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount = 1
	return &activeGame{gameCore: g.gameCore}, nil
}

// --- ActiveGame transitions ---

func (g *activeGame) LinkCampaign(id identity.CampaignID, source event.Source) (ActiveGame, []event.Event, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, nil, err
	}
	if g.hasCampaign(id) {
		return nil, nil, ErrCampaignAlreadyLinked
	}
	g.campaignIDs = append(g.campaignIDs, id)
	evt := event.EntityLinked{
		EntityAID:    g.id.String(),
		EntityBID:    id.String(),
		Relationship: "campaign",
		Source:       source,
	}
	return g, []event.Event{evt}, nil
}

func (g *activeGame) NotifyCampaignIdle(id identity.CampaignID) (Game, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount--
	if g.activeCampaignCount <= 0 {
		g.activeCampaignCount = 0
		return &idleGame{gameCore: g.gameCore}, nil
	}
	return g, nil
}

// --- IdleGame transitions ---

func (g *idleGame) ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount = 1
	return &activeGame{gameCore: g.gameCore}, nil
}

func (g *idleGame) Archive(source event.Source) (ArchivedGame, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: g.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedGame{gameCore: g.gameCore}, []event.Event{evt}, nil
}

// --- Helpers ---

func validateCampaignID(id identity.CampaignID) error {
	if id.IsZero() {
		return ErrCampaignIDRequired
	}
	return nil
}
