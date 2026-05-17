package interactor

import (
	"context"
	"fmt"
	"time"

	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type UpdateBeatContentRequest struct {
	CallerID    string
	BeatID      identity.BeatID
	GameID      identity.GameID
	Name        string
	Description string
	PlayerDesc  string
	Source      event.Source
}

type UpdateBeatContentResult struct {
	Beat   *narrativeentity.Beat
	Events []event.Event
}

type UpdateBeatContentInteractor struct {
	gameRepo GameRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewUpdateBeatContentInteractor(gameRepo GameRepository, beatRepo BeatRepository, bus event.EventBus) *UpdateBeatContentInteractor {
	return &UpdateBeatContentInteractor{gameRepo: gameRepo, beatRepo: beatRepo, bus: bus}
}

func (i *UpdateBeatContentInteractor) Execute(ctx context.Context, req UpdateBeatContentRequest) (UpdateBeatContentResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return UpdateBeatContentResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return UpdateBeatContentResult{}, shared.ErrUnauthorized
	}

	beat, err := i.beatRepo.Load(ctx, req.BeatID)
	if err != nil {
		return UpdateBeatContentResult{}, fmt.Errorf("%w: %v", ErrBeatNotFound, err)
	}

	events, err := beat.UpdateContent(req.Name, req.Description, req.PlayerDesc)
	if err != nil {
		return UpdateBeatContentResult{}, err
	}

	if err := i.beatRepo.Save(ctx, beat); err != nil {
		return UpdateBeatContentResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.BeatID.GrimoireID),
			AggregateType: event.AggregateBeat,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return UpdateBeatContentResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return UpdateBeatContentResult{Beat: beat, Events: events}, nil
}
