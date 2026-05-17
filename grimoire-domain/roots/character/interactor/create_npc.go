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

type CreateNPCRequest struct {
	CallerID string
	ID       identity.NarrativeCharacterID
	GameID   identity.GameID
	Name     string
	Source   event.Source
}

type CreateNPCResult struct {
	NPC    characterentity.NewNPC
	Events []event.Event
}

type CreateNPCInteractor struct {
	gameRepo GameRepository
	npcRepo  NPCRepository
	bus      event.EventBus
}

func NewCreateNPCInteractor(gameRepo GameRepository, npcRepo NPCRepository, bus event.EventBus) *CreateNPCInteractor {
	return &CreateNPCInteractor{gameRepo: gameRepo, npcRepo: npcRepo, bus: bus}
}

func (i *CreateNPCInteractor) Execute(ctx context.Context, req CreateNPCRequest) (CreateNPCResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateNPCResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateNPCResult{}, shared.ErrUnauthorized
	}

	npc, events, err := characterentity.CreateNPC(req.ID, req.Name, req.GameID, req.Source)
	if err != nil {
		return CreateNPCResult{}, err
	}

	if err := i.npcRepo.Save(ctx, npc.Snapshot()); err != nil {
		return CreateNPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return CreateNPCResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateNPCResult{NPC: npc, Events: events}, nil
}
