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

type UpdateNPCContentRequest struct {
	CallerID    string
	NPCID       identity.NarrativeCharacterID
	GameID      identity.GameID
	Name        string
	Description string
	PlayerDesc  string
	Source      event.Source
}

type UpdateNPCContentResult struct {
	NPC    characterentity.NPC
	Events []event.Event
}

type UpdateNPCContentInteractor struct {
	gameRepo GameRepository
	npcRepo  NPCRepository
	bus      event.EventBus
}

func NewUpdateNPCContentInteractor(gameRepo GameRepository, npcRepo NPCRepository, bus event.EventBus) *UpdateNPCContentInteractor {
	return &UpdateNPCContentInteractor{gameRepo: gameRepo, npcRepo: npcRepo, bus: bus}
}

func (i *UpdateNPCContentInteractor) Execute(ctx context.Context, req UpdateNPCContentRequest) (UpdateNPCContentResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return UpdateNPCContentResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return UpdateNPCContentResult{}, shared.ErrUnauthorized
	}

	snap, err := i.npcRepo.Load(ctx, req.NPCID)
	if err != nil {
		return UpdateNPCContentResult{}, fmt.Errorf("%w: %v", ErrNPCNotFound, err)
	}

	npc, err := characterentity.ReconstituteNPC(snap)
	if err != nil {
		return UpdateNPCContentResult{}, err
	}

	var updated characterentity.NPC
	var events []event.Event

	switch n := npc.(type) {
	case characterentity.DraftNPC:
		updated, events, err = n.UpdateContent(req.Name, req.Description, req.PlayerDesc, req.Source)
	case characterentity.ActiveNPC:
		updated, events, err = n.UpdateContent(req.Name, req.Description, req.PlayerDesc, req.Source)
	default:
		return UpdateNPCContentResult{}, fmt.Errorf("%w: npc must be Draft or Active to update content", ErrInvalidNPCState)
	}
	if err != nil {
		return UpdateNPCContentResult{}, err
	}

	if err := i.npcRepo.Save(ctx, updated.Snapshot()); err != nil {
		return UpdateNPCContentResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.NPCID.GrimoireID),
			AggregateType: event.AggregateCharacter,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return UpdateNPCContentResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return UpdateNPCContentResult{NPC: updated, Events: events}, nil
}
