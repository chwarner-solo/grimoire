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

type MovePartyRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	LocationID identity.LocationID
	Source     event.Source
}

type MovePartyResult struct {
	Campaign campaignentity.ActiveCampaign
	Events   []event.Event
}

type MovePartyInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewMovePartyInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *MovePartyInteractor {
	return &MovePartyInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *MovePartyInteractor) Execute(ctx context.Context, req MovePartyRequest) (MovePartyResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return MovePartyResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return MovePartyResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return MovePartyResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	ac, ok := campaign.(campaignentity.ActiveCampaign)
	if !ok {
		return MovePartyResult{}, fmt.Errorf("%w: party can only move during an active session", ErrInvalidCampaignState)
	}

	updated, entityEvents, err := ac.MoveToLocation(req.LocationID, req.Source)
	if err != nil {
		return MovePartyResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, updated); err != nil {
		return MovePartyResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    req.CampaignID.String(),
		EntityBID:    req.LocationID.String(),
		Relationship: "located_at",
		Source:       req.Source,
	}
	allEvents := append(entityEvents, linkedEvt)

	for _, evt := range allEvents {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.CampaignID.GrimoireID),
			AggregateType: event.AggregateCampaign,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return MovePartyResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return MovePartyResult{Campaign: updated, Events: allEvents}, nil
}