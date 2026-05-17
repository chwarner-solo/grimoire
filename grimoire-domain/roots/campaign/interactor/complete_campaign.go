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

type CompleteCampaignRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	Source     event.Source
}

type CompleteCampaignResult struct {
	Campaign campaignentity.CompleteCampaign
	Events   []event.Event
}

type CompleteCampaignInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewCompleteCampaignInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *CompleteCampaignInteractor {
	return &CompleteCampaignInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *CompleteCampaignInteractor) Execute(ctx context.Context, req CompleteCampaignRequest) (CompleteCampaignResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CompleteCampaignResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CompleteCampaignResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return CompleteCampaignResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	ic, ok := campaign.(campaignentity.IdleCampaign)
	if !ok {
		return CompleteCampaignResult{}, fmt.Errorf("%w: campaign must be Idle to complete", ErrInvalidCampaignState)
	}

	complete, events, err := ic.Complete(req.Source)
	if err != nil {
		return CompleteCampaignResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, complete); err != nil {
		return CompleteCampaignResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return CompleteCampaignResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CompleteCampaignResult{Campaign: complete, Events: events}, nil
}