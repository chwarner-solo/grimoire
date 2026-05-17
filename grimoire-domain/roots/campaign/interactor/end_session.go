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

type EndSessionRequest struct {
	CallerID   string
	CampaignID identity.CampaignID
	GameID     identity.GameID
	SessionID  identity.SessionID
	Notes      string
	Source     event.Source
}

type EndSessionResult struct {
	Campaign campaignentity.IdleCampaign
	Events   []event.Event
}

type EndSessionInteractor struct {
	gameRepo     GameRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewEndSessionInteractor(gameRepo GameRepository, campaignRepo CampaignRepository, bus event.EventBus) *EndSessionInteractor {
	return &EndSessionInteractor{gameRepo: gameRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *EndSessionInteractor) Execute(ctx context.Context, req EndSessionRequest) (EndSessionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return EndSessionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return EndSessionResult{}, shared.ErrUnauthorized
	}

	campaign, err := i.campaignRepo.Load(ctx, req.CampaignID)
	if err != nil {
		return EndSessionResult{}, fmt.Errorf("%w: %v", ErrCampaignNotFound, err)
	}

	ac, ok := campaign.(campaignentity.ActiveCampaign)
	if !ok {
		return EndSessionResult{}, fmt.Errorf("%w: campaign must be Active to end a session", ErrInvalidCampaignState)
	}

	idle, err := ac.NotifySessionSummarized()
	if err != nil {
		return EndSessionResult{}, err
	}

	if err := i.campaignRepo.Save(ctx, idle); err != nil {
		return EndSessionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	sessionEnded := event.SessionEnded{
		SessionID: req.SessionID.String(),
		Notes:     req.Notes,
	}
	events := []event.Event{sessionEnded}

	envelope := event.EventEnvelope{
		Type:          sessionEnded.EventType(),
		AggregateID:   identity.GrimoireID(req.CampaignID.GrimoireID),
		AggregateType: event.AggregateCampaign,
		Source:        req.Source,
		OccurredAt:    time.Now(),
		Payload:       sessionEnded,
	}
	if err := i.bus.Dispatch(ctx, envelope); err != nil {
		return EndSessionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
	}

	return EndSessionResult{Campaign: idle, Events: events}, nil
}