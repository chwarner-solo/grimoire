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

type AddBeatPrerequisiteRequest struct {
	CallerID       string
	BeatID         identity.BeatID
	PrerequisiteID identity.BeatID
	GameID         identity.GameID
	Source         event.Source
}

type AddBeatPrerequisiteResult struct {
	Beat   *narrativeentity.Beat
	Events []event.Event
}

type AddBeatPrerequisiteInteractor struct {
	gameRepo GameRepository
	beatRepo BeatRepository
	bus      event.EventBus
}

func NewAddBeatPrerequisiteInteractor(gameRepo GameRepository, beatRepo BeatRepository, bus event.EventBus) *AddBeatPrerequisiteInteractor {
	return &AddBeatPrerequisiteInteractor{gameRepo: gameRepo, beatRepo: beatRepo, bus: bus}
}

func (i *AddBeatPrerequisiteInteractor) Execute(ctx context.Context, req AddBeatPrerequisiteRequest) (AddBeatPrerequisiteResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddBeatPrerequisiteResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddBeatPrerequisiteResult{}, shared.ErrUnauthorized
	}

	beat, err := i.beatRepo.Load(ctx, req.BeatID)
	if err != nil {
		return AddBeatPrerequisiteResult{}, fmt.Errorf("%w: %v", ErrBeatNotFound, err)
	}

	if req.BeatID.String() == req.PrerequisiteID.String() {
		return AddBeatPrerequisiteResult{}, narrativeentity.ErrCycleDetected{BeatID: req.BeatID}
	}

	chain, err := i.beatRepo.LoadPrerequisiteChain(ctx, req.PrerequisiteID)
	if err != nil {
		return AddBeatPrerequisiteResult{}, fmt.Errorf("%w: %v", ErrPrerequisiteChainLoadFailed, err)
	}

	if err := beat.WouldCreateCycle(req.BeatID, chain); err != nil {
		return AddBeatPrerequisiteResult{}, err
	}

	beat.AddPrerequisite(req.PrerequisiteID)

	if err := i.beatRepo.Save(ctx, beat); err != nil {
		return AddBeatPrerequisiteResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    req.PrerequisiteID.String(),
		EntityBID:    req.BeatID.String(),
		Relationship: "prerequisite_of",
		Source:       req.Source,
	}
	events := []event.Event{linkedEvt}

	envelope := event.EventEnvelope{
		Type:          linkedEvt.EventType(),
		AggregateID:   identity.GrimoireID(req.BeatID.GrimoireID),
		AggregateType: event.AggregateBeat,
		Source:        req.Source,
		OccurredAt:    time.Now(),
		Payload:       linkedEvt,
	}
	if err := i.bus.Dispatch(ctx, envelope); err != nil {
		return AddBeatPrerequisiteResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
	}

	return AddBeatPrerequisiteResult{Beat: beat, Events: events}, nil
}
