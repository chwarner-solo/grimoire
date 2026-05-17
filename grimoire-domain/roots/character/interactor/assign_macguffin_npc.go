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

type AssignMacGuffinToNPCRequest struct {
	CallerID    string
	MacGuffinID identity.MacGuffinID
	NPCID       identity.NarrativeCharacterID
	GameID      identity.GameID
	Source      event.Source
}

type AssignMacGuffinToNPCResult struct {
	Events []event.Event
}

type AssignMacGuffinToNPCInteractor struct {
	gameRepo      GameRepository
	npcRepo       NPCRepository
	macguffinRepo MacGuffinRepository
	bus           event.EventBus
}

func NewAssignMacGuffinToNPCInteractor(gameRepo GameRepository, npcRepo NPCRepository, macguffinRepo MacGuffinRepository, bus event.EventBus) *AssignMacGuffinToNPCInteractor {
	return &AssignMacGuffinToNPCInteractor{gameRepo: gameRepo, npcRepo: npcRepo, macguffinRepo: macguffinRepo, bus: bus}
}

func (i *AssignMacGuffinToNPCInteractor) Execute(ctx context.Context, req AssignMacGuffinToNPCRequest) (AssignMacGuffinToNPCResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AssignMacGuffinToNPCResult{}, shared.ErrUnauthorized
	}

	mgSnap, err := i.macguffinRepo.Load(ctx, req.MacGuffinID)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrMacGuffinNotFound, err)
	}

	mg := characterentity.ReconstituteMacGuffin(mgSnap)
	if mg.Snapshot().Destroyed {
		return AssignMacGuffinToNPCResult{}, ErrMacGuffinDestroyed
	}

	npcSnap, err := i.npcRepo.Load(ctx, req.NPCID)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrNPCNotFound, err)
	}

	npc, err := characterentity.ReconstituteNPC(npcSnap)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, err
	}

	an, ok := npc.(characterentity.ActiveNPC)
	if !ok {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: npc must be Active to receive macguffin", ErrInvalidNPCState)
	}

	updatedNPC, npcEvents, err := an.AssignMacGuffin(req.MacGuffinID, req.Source)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, err
	}

	updatedMG, _, err := mg.TransferToNPC(req.NPCID, req.Source)
	if err != nil {
		return AssignMacGuffinToNPCResult{}, err
	}

	if err := i.npcRepo.Save(ctx, updatedNPC.Snapshot()); err != nil {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	if err := i.macguffinRepo.Save(ctx, updatedMG.Snapshot()); err != nil {
		return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range npcEvents {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.NPCID.GrimoireID),
			AggregateType: event.AggregateCharacter,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return AssignMacGuffinToNPCResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AssignMacGuffinToNPCResult{Events: npcEvents}, nil
}
