package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddLoreToMasterNarrativeRequest struct {
	CallerID string
	LoreID   identity.LoreID
	GameID   identity.GameID
	Source   event.Source
}

type AddLoreToMasterNarrativeResult struct {
	Events []event.Event
}

type AddLoreToMasterNarrativeInteractor struct {
	gameRepo GameRepository
	mnRepo   MasterNarrativeRepository
	bus      event.EventBus
}

func NewAddLoreToMasterNarrativeInteractor(gameRepo GameRepository, mnRepo MasterNarrativeRepository, bus event.EventBus) *AddLoreToMasterNarrativeInteractor {
	return &AddLoreToMasterNarrativeInteractor{gameRepo: gameRepo, mnRepo: mnRepo, bus: bus}
}

func (i *AddLoreToMasterNarrativeInteractor) Execute(ctx context.Context, req AddLoreToMasterNarrativeRequest) (AddLoreToMasterNarrativeResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddLoreToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddLoreToMasterNarrativeResult{}, shared.ErrUnauthorized
	}

	mn, err := i.mnRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return AddLoreToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrMasterNarrativeNotFound, err)
	}

	mn.AddLore(req.LoreID)

	if err := i.mnRepo.Save(ctx, mn); err != nil {
		return AddLoreToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    mn.MasterNarrativeID().String(),
		EntityBID:    req.LoreID.String(),
		Relationship: "contains_lore",
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
		return AddLoreToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
	}

	return AddLoreToMasterNarrativeResult{Events: events}, nil
}
