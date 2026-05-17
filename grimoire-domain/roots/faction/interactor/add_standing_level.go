package interactor

import (
	"context"
	"fmt"
	"time"

	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	factionvalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddStandingLevelRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Level     factionvalue.StandingLevel
	Source    event.Source
}

type AddStandingLevelResult struct {
	Events []event.Event
}

type AddStandingLevelInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewAddStandingLevelInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *AddStandingLevelInteractor {
	return &AddStandingLevelInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *AddStandingLevelInteractor) Execute(ctx context.Context, req AddStandingLevelRequest) (AddStandingLevelResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddStandingLevelResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddStandingLevelResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return AddStandingLevelResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	var updated factionentity.Faction
	var events []event.Event

	switch f := faction.(type) {
	case factionentity.DraftFaction:
		updated, events, err = f.AddStandingLevel(req.Level, req.Source)
	case factionentity.ActiveFaction:
		updated, events, err = f.AddStandingLevel(req.Level, req.Source)
	default:
		return AddStandingLevelResult{}, fmt.Errorf("%w: can only add standing levels to Draft or Active factions", ErrInvalidFactionState)
	}
	if err != nil {
		return AddStandingLevelResult{}, err
	}

	if err := i.factionRepo.Save(ctx, updated); err != nil {
		return AddStandingLevelResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return AddStandingLevelResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AddStandingLevelResult{Events: events}, nil
}
