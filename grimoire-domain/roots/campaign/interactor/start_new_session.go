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

type StartNewSessionRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	SessionID  identity.SessionID
	Date       time.Time
	Source     event.Source
}

type StartNewSessionResult struct {
	Campaign campaignentity.ActiveCampaign
	Events   []event.Event
}

type StartNewSessionInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewStartNewSessionInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *StartNewSessionInteractor {
	return &StartNewSessionInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *StartNewSessionInteractor) Execute(ctx context.Context, req StartNewSessionRequest) (StartNewSessionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return StartNewSessionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return StartNewSessionResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return StartNewSessionResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	ic, ok := campaign.(campaignentity.IdleCampaign)
	if !ok {
		return StartNewSessionResult{}, fmt.Errorf("%w: campaign must be Idle to start a new session", ErrInvalidCampaignState)
	}

	active, events, err := ic.StartNewSession(req.SessionID, req.Date)
	if err != nil {
		return StartNewSessionResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, active); err != nil {
		return StartNewSessionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return StartNewSessionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return StartNewSessionResult{Campaign: active, Events: events}, nil
}