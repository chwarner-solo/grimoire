package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddSecretToMasterNarrativeRequest struct {
	CallerID string
	SecretID identity.SecretID
	GameID   identity.GameID
	Source   event.Source
}

type AddSecretToMasterNarrativeResult struct {
	Events []event.Event
}

type AddSecretToMasterNarrativeInteractor struct {
	gameRepo GameRepository
	mnRepo   MasterNarrativeRepository
	bus      event.EventBus
}

func NewAddSecretToMasterNarrativeInteractor(gameRepo GameRepository, mnRepo MasterNarrativeRepository, bus event.EventBus) *AddSecretToMasterNarrativeInteractor {
	return &AddSecretToMasterNarrativeInteractor{gameRepo: gameRepo, mnRepo: mnRepo, bus: bus}
}

func (i *AddSecretToMasterNarrativeInteractor) Execute(ctx context.Context, req AddSecretToMasterNarrativeRequest) (AddSecretToMasterNarrativeResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddSecretToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddSecretToMasterNarrativeResult{}, shared.ErrUnauthorized
	}

	mn, err := i.mnRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return AddSecretToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrMasterNarrativeNotFound, err)
	}

	mn.AddSecret(req.SecretID)

	if err := i.mnRepo.Save(ctx, mn); err != nil {
		return AddSecretToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    mn.MasterNarrativeID().String(),
		EntityBID:    req.SecretID.String(),
		Relationship: "contains_secret",
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
		return AddSecretToMasterNarrativeResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
	}

	return AddSecretToMasterNarrativeResult{Events: events}, nil
}
