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

type UpdateMacGuffinContentRequest struct {
	CallerID    string
	MacGuffinID identity.MacGuffinID
	GameID      identity.GameID
	Name        string
	Description string
	PlayerDesc  string
	Source      event.Source
}

type UpdateMacGuffinContentResult struct {
	Events []event.Event
}

type UpdateMacGuffinContentInteractor struct {
	gameRepo      GameRepository
	macguffinRepo MacGuffinRepository
	bus           event.EventBus
}

func NewUpdateMacGuffinContentInteractor(
	gameRepo GameRepository,
	macguffinRepo MacGuffinRepository,
	bus event.EventBus,
) *UpdateMacGuffinContentInteractor {
	return &UpdateMacGuffinContentInteractor{
		gameRepo:      gameRepo,
		macguffinRepo: macguffinRepo,
		bus:           bus,
	}
}

func (i *UpdateMacGuffinContentInteractor) Execute(ctx context.Context, req UpdateMacGuffinContentRequest) (UpdateMacGuffinContentResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return UpdateMacGuffinContentResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return UpdateMacGuffinContentResult{}, shared.ErrUnauthorized
	}

	snap, err := i.macguffinRepo.Load(ctx, req.MacGuffinID)
	if err != nil {
		return UpdateMacGuffinContentResult{}, fmt.Errorf("%w: %v", ErrMacGuffinNotFound, err)
	}

	mg := characterentity.ReconstituteMacGuffin(snap)

	updated, events, err := mg.UpdateContent(req.Name, req.Description, req.PlayerDesc, req.Source)
	if err != nil {
		return UpdateMacGuffinContentResult{}, err
	}

	if err := i.macguffinRepo.Save(ctx, updated.Snapshot()); err != nil {
		return UpdateMacGuffinContentResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return UpdateMacGuffinContentResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return UpdateMacGuffinContentResult{Events: events}, nil
}
