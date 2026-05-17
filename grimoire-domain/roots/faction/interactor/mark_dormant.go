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

type MarkFactionDormantRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type MarkFactionDormantResult struct {
	Faction factionentity.IdleFaction
	Events  []event.Event
}

type MarkFactionDormantInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewMarkFactionDormantInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *MarkFactionDormantInteractor {
	return &MarkFactionDormantInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *MarkFactionDormantInteractor) Execute(ctx context.Context, req MarkFactionDormantRequest) (MarkFactionDormantResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return MarkFactionDormantResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return MarkFactionDormantResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return MarkFactionDormantResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	af, ok := faction.(factionentity.ActiveFaction)
	if !ok {
		return MarkFactionDormantResult{}, fmt.Errorf("%w: faction must be Active to mark dormant", ErrInvalidFactionState)
	}

	idle, events, err := af.MarkDormant(req.Source)
	if err != nil {
		return MarkFactionDormantResult{}, err
	}

	if err := i.factionRepo.Save(ctx, idle); err != nil {
		return MarkFactionDormantResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return MarkFactionDormantResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return MarkFactionDormantResult{Faction: idle, Events: events}, nil
}
