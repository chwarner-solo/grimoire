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

type UpdatePlayerCharacterContentRequest struct {
	CallerID    string
	ID          identity.PlayerCharacterID
	GameID      identity.GameID
	Name        string
	Description string
	PlayerDesc  string
	Source      event.Source
}

type UpdatePlayerCharacterContentResult struct {
	PlayerCharacter *characterentity.PlayerCharacter
	Events          []event.Event
}

type UpdatePlayerCharacterContentInteractor struct {
	gameRepo GameRepository
	pcRepo   PlayerCharacterRepository
	bus      event.EventBus
}

func NewUpdatePlayerCharacterContentInteractor(gameRepo GameRepository, pcRepo PlayerCharacterRepository, bus event.EventBus) *UpdatePlayerCharacterContentInteractor {
	return &UpdatePlayerCharacterContentInteractor{gameRepo: gameRepo, pcRepo: pcRepo, bus: bus}
}

func (i *UpdatePlayerCharacterContentInteractor) Execute(ctx context.Context, req UpdatePlayerCharacterContentRequest) (UpdatePlayerCharacterContentResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return UpdatePlayerCharacterContentResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return UpdatePlayerCharacterContentResult{}, shared.ErrUnauthorized
	}

	snap, err := i.pcRepo.LoadCore(ctx, req.ID)
	if err != nil {
		return UpdatePlayerCharacterContentResult{}, fmt.Errorf("%w: %v", ErrPlayerCharacterNotFound, err)
	}

	pc := characterentity.ReconstitutePlayerCharacter(snap)

	if pc.Snapshot().Status == characterentity.PlayerCharacterStatusRetired {
		return UpdatePlayerCharacterContentResult{}, ErrCharacterRetired
	}

	updated, events, err := pc.UpdateContent(req.Name, req.Description, req.PlayerDesc, req.Source)
	if err != nil {
		return UpdatePlayerCharacterContentResult{}, err
	}

	if err := i.pcRepo.SaveCore(ctx, updated.Snapshot()); err != nil {
		return UpdatePlayerCharacterContentResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return UpdatePlayerCharacterContentResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return UpdatePlayerCharacterContentResult{PlayerCharacter: updated, Events: events}, nil
}
