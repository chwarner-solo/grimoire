package interactor

import (
	"context"
	"fmt"
	"time"

	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type BeginFactionDraftRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type BeginFactionDraftResult struct {
	Faction factionentity.DraftFaction
	Events  []event.Event
}

type BeginFactionDraftInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewBeginFactionDraftInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *BeginFactionDraftInteractor {
	return &BeginFactionDraftInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *BeginFactionDraftInteractor) Execute(ctx context.Context, req BeginFactionDraftRequest) (BeginFactionDraftResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return BeginFactionDraftResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return BeginFactionDraftResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return BeginFactionDraftResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	nf, ok := faction.(factionentity.NewFaction)
	if !ok {
		return BeginFactionDraftResult{}, fmt.Errorf("%w: faction must be New to begin draft", ErrInvalidFactionState)
	}

	draft, events, err := nf.BeginDraft(req.Source)
	if err != nil {
		return BeginFactionDraftResult{}, err
	}

	if err := i.factionRepo.Save(ctx, draft); err != nil {
		return BeginFactionDraftResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.FactionID.GrimoireID),
			AggregateType: event.AggregateFaction,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return BeginFactionDraftResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return BeginFactionDraftResult{Faction: draft, Events: events}, nil
}
