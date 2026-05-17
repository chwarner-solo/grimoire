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

type CreateFactionRequest struct {
	CallerID  string
	ID        identity.FactionID
	GameID    identity.GameID
	Name      string
	Source    event.Source
}

type CreateFactionResult struct {
	Faction factionentity.NewFaction
	Events  []event.Event
}

type CreateFactionInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewCreateFactionInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *CreateFactionInteractor {
	return &CreateFactionInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *CreateFactionInteractor) Execute(ctx context.Context, req CreateFactionRequest) (CreateFactionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateFactionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateFactionResult{}, shared.ErrUnauthorized
	}

	faction, events, err := factionentity.CreateFaction(req.ID, req.Name, req.GameID, req.Source)
	if err != nil {
		return CreateFactionResult{}, err
	}

	if err := i.factionRepo.Save(ctx, faction); err != nil {
		return CreateFactionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.ID.GrimoireID),
			AggregateType: event.AggregateFaction,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateFactionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateFactionResult{Faction: faction, Events: events}, nil
}
