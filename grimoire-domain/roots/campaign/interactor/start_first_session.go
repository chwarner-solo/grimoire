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

type StartFirstSessionRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	SessionID  identity.SessionID
	Date       time.Time
	Source     event.Source
}

type StartFirstSessionResult struct {
	Campaign campaignentity.ActiveCampaign
	Events   []event.Event
}

type StartFirstSessionInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewStartFirstSessionInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *StartFirstSessionInteractor {
	return &StartFirstSessionInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *StartFirstSessionInteractor) Execute(ctx context.Context, req StartFirstSessionRequest) (StartFirstSessionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return StartFirstSessionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return StartFirstSessionResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return StartFirstSessionResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	fc, ok := campaign.(campaignentity.FormingCampaign)
	if !ok {
		return StartFirstSessionResult{}, fmt.Errorf("%w: campaign must be Forming to start first session", ErrInvalidCampaignState)
	}

	active, events, err := fc.StartFirstSession(req.SessionID, req.Date)
	if err != nil {
		return StartFirstSessionResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, active); err != nil {
		return StartFirstSessionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return StartFirstSessionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return StartFirstSessionResult{Campaign: active, Events: events}, nil
}