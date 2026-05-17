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

type ArchiveFactionRequest struct {
	CallerID  string
	FactionID identity.FactionID
	GameID    identity.GameID
	Source    event.Source
}

type ArchiveFactionResult struct {
	Faction factionentity.ArchivedFaction
	Events  []event.Event
}

type ArchiveFactionInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewArchiveFactionInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *ArchiveFactionInteractor {
	return &ArchiveFactionInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *ArchiveFactionInteractor) Execute(ctx context.Context, req ArchiveFactionRequest) (ArchiveFactionResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ArchiveFactionResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ArchiveFactionResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return ArchiveFactionResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	var archived factionentity.ArchivedFaction
	var events []event.Event

	switch f := faction.(type) {
	case factionentity.ActiveFaction:
		archived, events, err = f.Archive(req.Source)
	case factionentity.IdleFaction:
		archived, events, err = f.Archive(req.Source)
	default:
		return ArchiveFactionResult{}, fmt.Errorf("%w: faction must be Active or Idle to archive", ErrInvalidFactionState)
	}
	if err != nil {
		return ArchiveFactionResult{}, err
	}

	if err := i.factionRepo.Save(ctx, archived); err != nil {
		return ArchiveFactionResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ArchiveFactionResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ArchiveFactionResult{Faction: archived, Events: events}, nil
}
