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

type DeclareAllyRequest struct {
	CallerID  string
	FactionID identity.FactionID
	AllyID    identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type DeclareAllyResult struct {
	Events []event.Event
}

type DeclareAllyInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewDeclareAllyInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *DeclareAllyInteractor {
	return &DeclareAllyInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *DeclareAllyInteractor) Execute(ctx context.Context, req DeclareAllyRequest) (DeclareAllyResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return DeclareAllyResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return DeclareAllyResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return DeclareAllyResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	var updated factionentity.Faction
	var events []event.Event

	switch f := faction.(type) {
	case factionentity.DraftFaction:
		updated, events, err = f.AddAlly(req.AllyID, req.Source)
	case factionentity.ActiveFaction:
		updated, events, err = f.AddAlly(req.AllyID, req.Source)
	default:
		return DeclareAllyResult{}, fmt.Errorf("%w: can only declare allies on Draft or Active factions", ErrInvalidFactionState)
	}
	if err != nil {
		return DeclareAllyResult{}, err
	}

	if err := i.factionRepo.Save(ctx, updated); err != nil {
		return DeclareAllyResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return DeclareAllyResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return DeclareAllyResult{Events: events}, nil
}
