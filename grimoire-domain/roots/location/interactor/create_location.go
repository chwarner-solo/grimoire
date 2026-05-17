package interactor

import (
	"context"
	"fmt"
	"time"

	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	locationvalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type CreateLocationRequest struct {
	CallerID     string
	ID           identity.LocationID
	GameID       identity.GameID
	Name         string
	LocationType locationvalue.LocationType
	ParentID     identity.LocationID
	Source       event.Source
}

type CreateLocationResult struct {
	Location locationentity.NewLocation
	Events   []event.Event
}

type CreateLocationInteractor struct {
	gameRepo     GameRepository
	locationRepo LocationRepository
	bus          event.EventBus
}

func NewCreateLocationInteractor(gameRepo GameRepository, locationRepo LocationRepository, bus event.EventBus) *CreateLocationInteractor {
	return &CreateLocationInteractor{gameRepo: gameRepo, locationRepo: locationRepo, bus: bus}
}

func (i *CreateLocationInteractor) Execute(ctx context.Context, req CreateLocationRequest) (CreateLocationResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return CreateLocationResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return CreateLocationResult{}, shared.ErrUnauthorized
	}

	location, events, err := locationentity.CreateLocation(req.ID, req.Name, req.GameID, req.LocationType, req.ParentID, req.Source)
	if err != nil {
		return CreateLocationResult{}, err
	}

	if err := i.locationRepo.Save(ctx, location); err != nil {
		return CreateLocationResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	var allEvents []event.Event
	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.ID.GrimoireID),
			AggregateType: event.AggregateLocation,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateLocationResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
		allEvents = append(allEvents, evt)
	}

	if !req.ParentID.IsZero() {
		linked := event.EntityLinked{
			EntityAID:    req.ID.String(),
			EntityBID:    req.ParentID.String(),
			Relationship: "child_of",
			Source:       req.Source,
		}
		envelope := event.EventEnvelope{
			Type:          linked.EventType(),
			AggregateID:   identity.GrimoireID(req.ID.GrimoireID),
			AggregateType: event.AggregateLocation,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       linked,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateLocationResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
		allEvents = append(allEvents, linked)
	}

	return CreateLocationResult{Location: location, Events: allEvents}, nil
}
