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

type CreateMacGuffinRequest struct {
	CallerID string
	ID       identity.MacGuffinID
	GameID   identity.GameID
	Name     string
	Source   event.Source
}

type CreateMacGuffinResult struct {
	MacGuffin *characterentity.MacGuffin
	Events    []event.Event
}

type CreateMacGuffinInteractor struct {
	gameRepo      GameRepository
	macguffinRepo MacGuffinRepository
	bus           event.EventBus
}

func NewCreateMacGuffinInteractor(gameRepo GameRepository, macguffinRepo MacGuffinRepository, bus event.EventBus) *CreateMacGuffinInteractor {
	return &CreateMacGuffinInteractor{gameRepo: gameRepo, macguffinRepo: macguffinRepo, bus: bus}
}

func (i *CreateMacGuffinInteractor) Execute(ctx context.Context, req CreateMacGuffinRequest) (CreateMacGuffinResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateMacGuffinResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateMacGuffinResult{}, shared.ErrUnauthorized
	}

	mg, events, err := characterentity.CreateMacGuffin(req.ID, req.GameID, req.Name, req.Source)
	if err != nil {
		return CreateMacGuffinResult{}, err
	}

	if err := i.macguffinRepo.Save(ctx, mg.Snapshot()); err != nil {
		return CreateMacGuffinResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return CreateMacGuffinResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateMacGuffinResult{MacGuffin: mg, Events: events}, nil
}
