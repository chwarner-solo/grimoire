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

type CreatePlayerCharacterRequest struct {
	CallerID      string
	ID            identity.PlayerCharacterID
	GameID        identity.GameID
	Name          string
	OwnerPlayerID string
	Source        event.Source
}

type CreatePlayerCharacterResult struct {
	PlayerCharacter *characterentity.PlayerCharacter
	Events          []event.Event
}

type CreatePlayerCharacterInteractor struct {
	gameRepo GameRepository
	pcRepo   PlayerCharacterRepository
	bus      event.EventBus
}

func NewCreatePlayerCharacterInteractor(gameRepo GameRepository, pcRepo PlayerCharacterRepository, bus event.EventBus) *CreatePlayerCharacterInteractor {
	return &CreatePlayerCharacterInteractor{gameRepo: gameRepo, pcRepo: pcRepo, bus: bus}
}

func (i *CreatePlayerCharacterInteractor) Execute(ctx context.Context, req CreatePlayerCharacterRequest) (CreatePlayerCharacterResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreatePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreatePlayerCharacterResult{}, shared.ErrUnauthorized
	}

	pc, events, err := characterentity.CreatePlayerCharacter(req.ID, req.GameID, req.Name, req.OwnerPlayerID, req.Source)
	if err != nil {
		return CreatePlayerCharacterResult{}, err
	}

	if err := i.pcRepo.SaveCore(ctx, pc.Snapshot()); err != nil {
		return CreatePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return CreatePlayerCharacterResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreatePlayerCharacterResult{PlayerCharacter: pc, Events: events}, nil
}
