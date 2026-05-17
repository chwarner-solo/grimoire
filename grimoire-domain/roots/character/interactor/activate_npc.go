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

type ActivateNPCRequest struct {
	CallerID string
	NPCID    identity.NarrativeCharacterID
	GameID   identity.GameID
	Source   event.Source
}

type ActivateNPCResult struct {
	NPC    characterentity.ActiveNPC
	Events []event.Event
}

type ActivateNPCInteractor struct {
	gameRepo GameRepository
	npcRepo  NPCRepository
	bus      event.EventBus
}

func NewActivateNPCInteractor(gameRepo GameRepository, npcRepo NPCRepository, bus event.EventBus) *ActivateNPCInteractor {
	return &ActivateNPCInteractor{gameRepo: gameRepo, npcRepo: npcRepo, bus: bus}
}

func (i *ActivateNPCInteractor) Execute(ctx context.Context, req ActivateNPCRequest) (ActivateNPCResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ActivateNPCResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ActivateNPCResult{}, shared.ErrUnauthorized
	}

	snap, err := i.npcRepo.Load(ctx, req.NPCID)
	if err != nil {
		return ActivateNPCResult{}, fmt.Errorf("%w: %v", ErrNPCNotFound, err)
	}

	npc, err := characterentity.ReconstituteNPC(snap)
	if err != nil {
		return ActivateNPCResult{}, err
	}

	dn, ok := npc.(characterentity.DraftNPC)
	if !ok {
		return ActivateNPCResult{}, fmt.Errorf("%w: npc must be Draft to activate", ErrInvalidNPCState)
	}

	active, events, err := dn.Activate(req.Source)
	if err != nil {
		return ActivateNPCResult{}, err
	}

	if err := i.npcRepo.Save(ctx, active.Snapshot()); err != nil {
		return ActivateNPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ActivateNPCResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ActivateNPCResult{NPC: active, Events: events}, nil
}
