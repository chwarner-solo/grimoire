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

type AssignMacGuffinToPCRequest struct {
	CallerID    string
	MacGuffinID identity.MacGuffinID
	PCID        identity.PlayerCharacterID
	GameID      identity.GameID
	Source      event.Source
}

type AssignMacGuffinToPCResult struct {
	Events []event.Event
}

type AssignMacGuffinToPCInteractor struct {
	gameRepo      GameRepository
	pcRepo        PlayerCharacterRepository
	macguffinRepo MacGuffinRepository
	bus           event.EventBus
}

func NewAssignMacGuffinToPCInteractor(gameRepo GameRepository, pcRepo PlayerCharacterRepository, macguffinRepo MacGuffinRepository, bus event.EventBus) *AssignMacGuffinToPCInteractor {
	return &AssignMacGuffinToPCInteractor{gameRepo: gameRepo, pcRepo: pcRepo, macguffinRepo: macguffinRepo, bus: bus}
}

func (i *AssignMacGuffinToPCInteractor) Execute(ctx context.Context, req AssignMacGuffinToPCRequest) (AssignMacGuffinToPCResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AssignMacGuffinToPCResult{}, shared.ErrUnauthorized
	}

	mgSnap, err := i.macguffinRepo.Load(ctx, req.MacGuffinID)
	if err != nil {
		return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrMacGuffinNotFound, err)
	}

	mg := characterentity.ReconstituteMacGuffin(mgSnap)
	if mg.Snapshot().Destroyed {
		return AssignMacGuffinToPCResult{}, ErrMacGuffinDestroyed
	}

	pcSnap, err := i.pcRepo.LoadCore(ctx, req.PCID)
	if err != nil {
		return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrPlayerCharacterNotFound, err)
	}

	pc := characterentity.ReconstitutePlayerCharacter(pcSnap)

	updatedPC, pcEvents, err := pc.AssignMacGuffin(req.MacGuffinID, req.Source)
	if err != nil {
		return AssignMacGuffinToPCResult{}, err
	}

	updatedMG, _, err := mg.TransferToPC(req.PCID, req.Source)
	if err != nil {
		return AssignMacGuffinToPCResult{}, err
	}

	if err := i.pcRepo.SaveCore(ctx, updatedPC.Snapshot()); err != nil {
		return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	if err := i.macguffinRepo.Save(ctx, updatedMG.Snapshot()); err != nil {
		return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range pcEvents {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.PCID.GrimoireID),
			AggregateType: event.AggregateCharacter,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return AssignMacGuffinToPCResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AssignMacGuffinToPCResult{Events: pcEvents}, nil
}
