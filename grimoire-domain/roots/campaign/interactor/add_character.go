package interactor

import (
	"context"
	"fmt"
	"time"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddCharacterToCampaignRequest struct {
	CallerID    string
	CampaignID  identity.CampaignID
	GameID      identity.GameID
	CharacterID identity.PlayerCharacterID
	Source      event.Source
}

type AddCharacterToCampaignResult struct {
	Campaign campaignentity.Campaign
	Events   []event.Event
}

type AddCharacterToCampaignInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewAddCharacterToCampaignInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *AddCharacterToCampaignInteractor {
	return &AddCharacterToCampaignInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *AddCharacterToCampaignInteractor) Execute(ctx context.Context, req AddCharacterToCampaignRequest) (AddCharacterToCampaignResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddCharacterToCampaignResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddCharacterToCampaignResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return AddCharacterToCampaignResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	var updated campaignentity.Campaign
	var events []event.Event

	switch c := campaign.(type) {
	case campaignentity.NewCampaign:
		updated, events, err = c.AddCharacter(req.CharacterID, req.Source)
	case campaignentity.FormingCampaign:
		updated, events, err = c.AddCharacter(req.CharacterID, req.Source)
	default:
		return AddCharacterToCampaignResult{}, fmt.Errorf("%w: can only add characters to New or Forming campaigns", ErrInvalidCampaignState)
	}
	if err != nil {
		return AddCharacterToCampaignResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, updated); err != nil {
		return AddCharacterToCampaignResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.CampaignID.GrimoireID),
			AggregateType: event.AggregateCampaign,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return AddCharacterToCampaignResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AddCharacterToCampaignResult{Campaign: updated, Events: events}, nil
}