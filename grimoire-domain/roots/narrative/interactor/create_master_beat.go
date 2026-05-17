package interactor

import (
	"context"
	"fmt"
	"time"

	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	narrativevalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type CreateMasterBeatRequest struct {
	CallerID  string
	BeatID    identity.BeatID
	GameID    identity.GameID
	Name      string
	BeatType  narrativevalue.BeatType
	Source    event.Source
}

type CreateMasterBeatResult struct {
	Beat   *narrativeentity.Beat
	Events []event.Event
}

type CreateMasterBeatInteractor struct {
	gameRepo GameRepository
	mnRepo   MasterNarrativeRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewCreateMasterBeatInteractor(gameRepo GameRepository, mnRepo MasterNarrativeRepository, beatRepo BeatRepository, bus event.EventBus) *CreateMasterBeatInteractor {
	return &CreateMasterBeatInteractor{gameRepo: gameRepo, mnRepo: mnRepo, beatRepo: beatRepo, bus: bus}
}

func (i *CreateMasterBeatInteractor) Execute(ctx context.Context, req CreateMasterBeatRequest) (CreateMasterBeatResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateMasterBeatResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateMasterBeatResult{}, shared.ErrUnauthorized
	}

	mn, err := i.mnRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return CreateMasterBeatResult{}, fmt.Errorf("%w: %v", ErrMasterNarrativeNotFound, err)
	}

	beat, events, err := narrativeentity.CreateMasterBeat(req.BeatID, req.Name, req.BeatType, req.GameID, req.Source)
	if err != nil {
		return CreateMasterBeatResult{}, err
	}

	if err := i.beatRepo.Save(ctx, beat); err != nil {
		return CreateMasterBeatResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	mn.AddBeat(beat.BeatID())
	if err := i.mnRepo.Save(ctx, mn); err != nil {
		return CreateMasterBeatResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return CreateMasterBeatResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateMasterBeatResult{Beat: beat, Events: events}, nil
}
