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

type AddSceneRequest struct {
	CallerID   string
	LocationID identity.LocationID
	SceneID    identity.SceneID
	GameID     identity.GameID
	Name       string
	Source     event.Source
}

type AddSceneResult struct {
	Events []event.Event
}

type AddSceneInteractor struct {
	gameRepo     GameRepository
	locationRepo LocationRepository
	bus          event.EventBus
}

func NewAddSceneInteractor(gameRepo GameRepository, locationRepo LocationRepository, bus event.EventBus) *AddSceneInteractor {
	return &AddSceneInteractor{gameRepo: gameRepo, locationRepo: locationRepo, bus: bus}
}

func (i *AddSceneInteractor) Execute(ctx context.Context, req AddSceneRequest) (AddSceneResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddSceneResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddSceneResult{}, shared.ErrUnauthorized
	}

	location, err := i.locationRepo.Load(ctx, req.LocationID)
	if err != nil {
		return AddSceneResult{}, fmt.Errorf("%w: %v", ErrLocationNotFound, err)
	}

	// Validate scene fields before mutating location.
	_, sceneEvents, err := locationentity.CreateScene(req.SceneID, req.LocationID, req.GameID, req.Name, req.Source)
	if err != nil {
		return AddSceneResult{}, err
	}

	var updated locationentity.Location
	var linkEvents []event.Event

	switch l := location.(type) {
	case locationentity.DraftLocation:
		updated, linkEvents, err = l.AddScene(req.SceneID, req.Source)
	case locationentity.ActiveLocation:
		updated, linkEvents, err = l.AddScene(req.SceneID, req.Source)
	default:
		return AddSceneResult{}, fmt.Errorf("%w: location must be Draft or Active to add a scene", ErrInvalidLocationState)
	}
	if err != nil {
		return AddSceneResult{}, err
	}

	if err := i.locationRepo.Save(ctx, updated); err != nil {
		return AddSceneResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	allEvents := append(sceneEvents, linkEvents...)
	for _, evt := range allEvents {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.LocationID.GrimoireID),
			AggregateType: event.AggregateLocation,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return AddSceneResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AddSceneResult{Events: allEvents}, nil
}
