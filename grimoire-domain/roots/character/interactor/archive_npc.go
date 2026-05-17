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

type ArchiveNPCRequest struct {
	CallerID string
	NPCID    identity.NarrativeCharacterID
	GameID   identity.GameID
	Source   event.Source
}

type ArchiveNPCResult struct {
	NPC    characterentity.ArchivedNPC
	Events []event.Event
}

type ArchiveNPCInteractor struct {
	gameRepo GameRepository
	npcRepo  NPCRepository
	bus      event.EventBus
}

func NewArchiveNPCInteractor(gameRepo GameRepository, npcRepo NPCRepository, bus event.EventBus) *ArchiveNPCInteractor {
	return &ArchiveNPCInteractor{gameRepo: gameRepo, npcRepo: npcRepo, bus: bus}
}

func (i *ArchiveNPCInteractor) Execute(ctx context.Context, req ArchiveNPCRequest) (ArchiveNPCResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ArchiveNPCResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ArchiveNPCResult{}, shared.ErrUnauthorized
	}

	snap, err := i.npcRepo.Load(ctx, req.NPCID)
	if err != nil {
		return ArchiveNPCResult{}, fmt.Errorf("%w: %v", ErrNPCNotFound, err)
	}

	npc, err := characterentity.ReconstituteNPC(snap)
	if err != nil {
		return ArchiveNPCResult{}, err
	}

	var archived characterentity.ArchivedNPC
	var events []event.Event

	switch n := npc.(type) {
	case characterentity.ActiveNPC:
		archived, events, err = n.Archive(req.Source)
	case characterentity.IdleNPC:
		archived, events, err = n.Archive(req.Source)
	default:
		return ArchiveNPCResult{}, fmt.Errorf("%w: npc must be Active or Idle to archive", ErrInvalidNPCState)
	}
	if err != nil {
		return ArchiveNPCResult{}, err
	}

	if err := i.npcRepo.Save(ctx, archived.Snapshot()); err != nil {
		return ArchiveNPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ArchiveNPCResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ArchiveNPCResult{NPC: archived, Events: events}, nil
}
