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

type ReactivateFactionRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type ReactivateFactionResult struct {
	Faction factionentity.ActiveFaction
	Events  []event.Event
}

type ReactivateFactionInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewReactivateFactionInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *ReactivateFactionInteractor {
	return &ReactivateFactionInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *ReactivateFactionInteractor) Execute(ctx context.Context, req ReactivateFactionRequest) (ReactivateFactionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ReactivateFactionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ReactivateFactionResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return ReactivateFactionResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	idle, ok := faction.(factionentity.IdleFaction)
	if !ok {
		return ReactivateFactionResult{}, fmt.Errorf("%w: faction must be Idle to reactivate", ErrInvalidFactionState)
	}

	active, events, err := idle.Reactivate(req.Source)
	if err != nil {
		return ReactivateFactionResult{}, err
	}

	if err := i.factionRepo.Save(ctx, active); err != nil {
		return ReactivateFactionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ReactivateFactionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ReactivateFactionResult{Faction: active, Events: events}, nil
}
