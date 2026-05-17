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

type DestroyMacGuffinRequest struct {
	CallerID    string
	MacGuffinID identity.MacGuffinID
	GameID      identity.GameID
	Source      event.Source
}

type DestroyMacGuffinResult struct {
	Events []event.Event
}

type DestroyMacGuffinInteractor struct {
	gameRepo      GameRepository
	macguffinRepo MacGuffinRepository
	bus           event.EventBus
}

func NewDestroyMacGuffinInteractor(gameRepo GameRepository, macguffinRepo MacGuffinRepository, bus event.EventBus) *DestroyMacGuffinInteractor {
	return &DestroyMacGuffinInteractor{gameRepo: gameRepo, macguffinRepo: macguffinRepo, bus: bus}
}

func (i *DestroyMacGuffinInteractor) Execute(ctx context.Context, req DestroyMacGuffinRequest) (DestroyMacGuffinResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return DestroyMacGuffinResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return DestroyMacGuffinResult{}, shared.ErrUnauthorized
	}

	mgSnap, err := i.macguffinRepo.Load(ctx, req.MacGuffinID)
	if err != nil {
		return DestroyMacGuffinResult{}, fmt.Errorf("%w: %v", ErrMacGuffinNotFound, err)
	}

	mg := characterentity.ReconstituteMacGuffin(mgSnap)

	destroyed, events, err := mg.Destroy(req.Source)
	if err != nil {
		return DestroyMacGuffinResult{}, err
	}

	if err := i.macguffinRepo.Save(ctx, destroyed.Snapshot()); err != nil {
		return DestroyMacGuffinResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.MacGuffinID.GrimoireID),
			AggregateType: event.AggregateCharacter,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return DestroyMacGuffinResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return DestroyMacGuffinResult{Events: events}, nil
}
