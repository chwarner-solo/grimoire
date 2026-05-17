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

type DeclareWarRequest struct {
	CallerID  string
	FactionID identity.FactionID
	EnemyID   identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type DeclareWarResult struct {
	Events []event.Event
}

type DeclareWarInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewDeclareWarInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *DeclareWarInteractor {
	return &DeclareWarInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *DeclareWarInteractor) Execute(ctx context.Context, req DeclareWarRequest) (DeclareWarResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return DeclareWarResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return DeclareWarResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return DeclareWarResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	var updated factionentity.Faction
	var events []event.Event

	switch f := faction.(type) {
	case factionentity.DraftFaction:
		updated, events, err = f.DeclareWar(req.EnemyID, req.Source)
	case factionentity.ActiveFaction:
		updated, events, err = f.DeclareWar(req.EnemyID, req.Source)
	default:
		return DeclareWarResult{}, fmt.Errorf("%w: can only declare war from Draft or Active factions", ErrInvalidFactionState)
	}
	if err != nil {
		return DeclareWarResult{}, err
	}

	if err := i.factionRepo.Save(ctx, updated); err != nil {
		return DeclareWarResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return DeclareWarResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return DeclareWarResult{Events: events}, nil
}
