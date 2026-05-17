package interactor

import (
	"context"
	"errors"
	"fmt"
	"time"

	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type ConnectLocationsRequest struct {
	CallerID string
	FromID   identity.LocationID
	ToID     identity.LocationID
	GameID   identity.GameID
	Source   event.Source
}

type ConnectLocationsResult struct {
	Events []event.Event
}

type ConnectLocationsInteractor struct {
	gameRepo     GameRepository
	locationRepo LocationRepository
	bus          event.EventBus
}

func NewConnectLocationsInteractor(gameRepo GameRepository, locationRepo LocationRepository, bus event.EventBus) *ConnectLocationsInteractor {
	return &ConnectLocationsInteractor{gameRepo: gameRepo, locationRepo: locationRepo, bus: bus}
}

func (i *ConnectLocationsInteractor) Execute(ctx context.Context, req ConnectLocationsRequest) (ConnectLocationsResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ConnectLocationsResult{}, shared.ErrUnauthorized
	}

	if req.FromID.String() == req.ToID.String() {
		return ConnectLocationsResult{}, ErrSelfConnection
	}

	fromLocation, err := i.locationRepo.Load(ctx, req.FromID)
	if err != nil {
		return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrLocationNotFound, err)
	}

	// Verify the target location exists.
	if _, err := i.locationRepo.Load(ctx, req.ToID); err != nil {
		return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrLocationNotFound, err)
	}

	al, ok := fromLocation.(locationentity.ActiveLocation)
	if !ok {
		return ConnectLocationsResult{}, fmt.Errorf("%w: from-location must be Active to connect", ErrInvalidLocationState)
	}

	updated, events, err := al.ConnectTo(req.ToID, req.Source)
	if err != nil {
		var alreadyConn locationentity.ErrAlreadyConnectedTo
		if errors.As(err, &alreadyConn) {
			return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrConnectionAlreadyExists, err)
		}
		return ConnectLocationsResult{}, err
	}

	if err := i.locationRepo.Save(ctx, updated); err != nil {
		return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.FromID.GrimoireID),
			AggregateType: event.AggregateLocation,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return ConnectLocationsResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ConnectLocationsResult{Events: events}, nil
}
