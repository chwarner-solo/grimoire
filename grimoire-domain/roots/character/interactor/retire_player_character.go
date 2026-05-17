package interactor

import (
	"context"
	"fmt"
	"time"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type RetirePlayerCharacterRequest struct {
	CallerID string
	ID       identity.PlayerCharacterID
	GameID   identity.GameID
	Source   event.Source
}

type RetirePlayerCharacterResult struct {
	PlayerCharacter *characterentity.PlayerCharacter
	Events          []event.Event
}

type RetirePlayerCharacterInteractor struct {
	gameRepo GameRepository
	pcRepo   PlayerCharacterRepository
	bus      event.EventBus
}

func NewRetirePlayerCharacterInteractor(gameRepo GameRepository, pcRepo PlayerCharacterRepository, bus event.EventBus) *RetirePlayerCharacterInteractor {
	return &RetirePlayerCharacterInteractor{gameRepo: gameRepo, pcRepo: pcRepo, bus: bus}
}

func (i *RetirePlayerCharacterInteractor) Execute(ctx context.Context, req RetirePlayerCharacterRequest) (RetirePlayerCharacterResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return RetirePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return RetirePlayerCharacterResult{}, shared.ErrUnauthorized
	}

	snap, err := i.pcRepo.LoadCore(ctx, req.ID)
	if err != nil {
		return RetirePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrPlayerCharacterNotFound, err)
	}

	pc := characterentity.ReconstitutePlayerCharacter(snap)

	retired, events, err := pc.Retire(req.Source)
	if err != nil {
		return RetirePlayerCharacterResult{}, err
	}

	if err := i.pcRepo.SaveCore(ctx, retired.Snapshot()); err != nil {
		return RetirePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.ID.GrimoireID),
			AggregateType: event.AggregateCharacter,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return RetirePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return RetirePlayerCharacterResult{PlayerCharacter: retired, Events: events}, nil
}
