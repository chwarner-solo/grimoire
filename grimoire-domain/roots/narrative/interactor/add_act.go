package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddActToMasterNarrativeRequest struct {
	CallerID string
	ActID    identity.ActID
	GameID   identity.GameID
	Source   event.Source
}

type AddActToMasterNarrativeResult struct {
	Events []event.Event
}

type AddActToMasterNarrativeInteractor struct {
	gameRepo GameRepository
	mnRepo   MasterNarrativeRepository
	bus      event.EventBus
}

func NewAddActToMasterNarrativeInteractor(gameRepo GameRepository, mnRepo MasterNarrativeRepository, bus event.EventBus) *AddActToMasterNarrativeInteractor {
	return &AddActToMasterNarrativeInteractor{gameRepo: gameRepo, mnRepo: mnRepo, bus: bus}
}

func (i *AddActToMasterNarrativeInteractor) Execute(ctx context.Context, req AddActToMasterNarrativeRequest) (AddActToMasterNarrativeResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddActToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddActToMasterNarrativeResult{}, shared.ErrUnauthorized
	}

	mn, err := i.mnRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return AddActToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrMasterNarrativeNotFound, err)
	}

	mn.AddAct(req.ActID)

	if err := i.mnRepo.Save(ctx, mn); err != nil {
		return AddActToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    mn.MasterNarrativeID().String(),
		EntityBID:    req.ActID.String(),
		Relationship: "contains_act",
		Source:       req.Source,
	}
	events := []event.Event{linkedEvt}

	envelope := event.EventEnvelope{
		Type:          linkedEvt.EventType(),
		AggregateID:   identity.GrimoireID(mn.MasterNarrativeID().GrimoireID),
		AggregateType: event.AggregateMasterNarrative,
		Source:        req.Source,
		OccurredAt:    time.Now(),
		Payload:       linkedEvt,
	}
	if err := i.bus.Dispatch(ctx, envelope); err != nil {
		return AddActToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
	}

	return AddActToMasterNarrativeResult{Events: events}, nil
}
