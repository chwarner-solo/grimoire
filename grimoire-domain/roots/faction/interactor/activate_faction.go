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

type ActivateFactionRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type ActivateFactionResult struct {
	Faction factionentity.ActiveFaction
	Events  []event.Event
}

type ActivateFactionInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewActivateFactionInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *ActivateFactionInteractor {
	return &ActivateFactionInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *ActivateFactionInteractor) Execute(ctx context.Context, req ActivateFactionRequest) (ActivateFactionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ActivateFactionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ActivateFactionResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return ActivateFactionResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	df, ok := faction.(factionentity.DraftFaction)
	if !ok {
		return ActivateFactionResult{}, fmt.Errorf("%w: faction must be Draft to activate", ErrInvalidFactionState)
	}

	active, events, err := df.Activate(req.Source)
	if err != nil {
		return ActivateFactionResult{}, err
	}

	if err := i.factionRepo.Save(ctx, active); err != nil {
		return ActivateFactionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ActivateFactionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ActivateFactionResult{Faction: active, Events: events}, nil
}
