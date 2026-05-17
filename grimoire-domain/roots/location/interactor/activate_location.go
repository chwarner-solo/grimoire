package interactor

import (
	"context"
	"fmt"
	"time"

	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type ActivateLocationRequest struct {
	CallerID   string
	LocationID identity.LocationID
	GameID     identity.GameID
	Source     event.Source
}

type ActivateLocationResult struct {
	Location locationentity.ActiveLocation
	Events   []event.Event
}

type ActivateLocationInteractor struct {
	gameRepo     GameRepository
	locationRepo LocationRepository
	bus          event.EventBus
}

func NewActivateLocationInteractor(gameRepo GameRepository, locationRepo LocationRepository, bus event.EventBus) *ActivateLocationInteractor {
	return &ActivateLocationInteractor{gameRepo: gameRepo, locationRepo: locationRepo, bus: bus}
}

func (i *ActivateLocationInteractor) Execute(ctx context.Context, req ActivateLocationRequest) (ActivateLocationResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ActivateLocationResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ActivateLocationResult{}, shared.ErrUnauthorized
	}

	location, err := i.locationRepo.Load(ctx, req.LocationID)
	if err != nil {
		return ActivateLocationResult{}, fmt.Errorf("%w: %v", ErrLocationNotFound, err)
	}

	dl, ok := location.(locationentity.DraftLocation)
	if !ok {
		return ActivateLocationResult{}, fmt.Errorf("%w: location must be Draft to activate", ErrInvalidLocationState)
	}

	active, events, err := dl.Activate(req.Source)
	if err != nil {
		return ActivateLocationResult{}, err
	}

	if err := i.locationRepo.Save(ctx, active); err != nil {
		return ActivateLocationResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.LocationID.GrimoireID),
			AggregateType: event.AggregateLocation,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return ActivateLocationResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ActivateLocationResult{Location: active, Events: events}, nil
}
