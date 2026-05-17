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

type BeginCampaignFormationRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	Source     event.Source
}

type BeginCampaignFormationResult struct {
	Campaign campaignentity.FormingCampaign
	Events   []event.Event
}

type BeginCampaignFormationInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewBeginCampaignFormationInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *BeginCampaignFormationInteractor {
	return &BeginCampaignFormationInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *BeginCampaignFormationInteractor) Execute(ctx context.Context, req BeginCampaignFormationRequest) (BeginCampaignFormationResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return BeginCampaignFormationResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return BeginCampaignFormationResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return BeginCampaignFormationResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	nc, ok := campaign.(campaignentity.NewCampaign)
	if !ok {
		return BeginCampaignFormationResult{}, fmt.Errorf("%w: campaign must be New to begin formation", ErrInvalidCampaignState)
	}

	forming, events, err := nc.BeginFormation(req.Source)
	if err != nil {
		return BeginCampaignFormationResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, forming); err != nil {
		return BeginCampaignFormationResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return BeginCampaignFormationResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return BeginCampaignFormationResult{Campaign: forming, Events: events}, nil
}