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

type PromoteBeatToMasterRequest struct {
	CallerID string
	BeatID   identity.BeatID
	GameID   identity.GameID
	Source   event.Source
}

type PromoteBeatToMasterResult struct {
	Beat   *narrativeentity.Beat
	Events []event.Event
}

type PromoteBeatToMasterInteractor struct {
	gameRepo GameRepository
	mnRepo   MasterNarrativeRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewPromoteBeatToMasterInteractor(gameRepo GameRepository, mnRepo MasterNarrativeRepository, beatRepo BeatRepository, bus event.EventBus) *PromoteBeatToMasterInteractor {
	return &PromoteBeatToMasterInteractor{gameRepo: gameRepo, mnRepo: mnRepo, beatRepo: beatRepo, bus: bus}
}

func (i *PromoteBeatToMasterInteractor) Execute(ctx context.Context, req PromoteBeatToMasterRequest) (PromoteBeatToMasterResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return PromoteBeatToMasterResult{}, shared.ErrUnauthorized
	}

	beat, err := i.beatRepo.Load(ctx, req.BeatID)
	if err != nil {
		return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrBeatNotFound, err)
	}

	events, err := beat.PromoteToMaster(req.Source)
	if err != nil {
		return PromoteBeatToMasterResult{}, err
	}

	mn, err := i.mnRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrMasterNarrativeNotFound, err)
	}

	mn.AddBeat(beat.BeatID())

	if err := i.beatRepo.Save(ctx, beat); err != nil {
		return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	if err := i.mnRepo.Save(ctx, mn); err != nil {
		return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return PromoteBeatToMasterResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return PromoteBeatToMasterResult{Beat: beat, Events: events}, nil
}
