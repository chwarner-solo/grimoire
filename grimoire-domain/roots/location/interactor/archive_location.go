package interactor

import (
	"context"
	"fmt"
	"time"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type ArchiveLocationRequest struct {
	CallerID   string
	LocationID identity.LocationID
	GameID     identity.GameID
	Source     event.Source
}

type ArchiveLocationResult struct {
	Location locationentity.ArchivedLocation
	Events   []event.Event
}

type ArchiveLocationInteractor struct {
	gameRepo     GameRepository
	locationRepo LocationRepository
	campaignRepo CampaignRepository
	bus          event.EventBus
}

func NewArchiveLocationInteractor(gameRepo GameRepository, locationRepo LocationRepository, campaignRepo CampaignRepository, bus event.EventBus) *ArchiveLocationInteractor {
	return &ArchiveLocationInteractor{gameRepo: gameRepo, locationRepo: locationRepo, campaignRepo: campaignRepo, bus: bus}
}

func (i *ArchiveLocationInteractor) Execute(ctx context.Context, req ArchiveLocationRequest) (ArchiveLocationResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return ArchiveLocationResult{}, shared.ErrUnauthorized
	}

	location, err := i.locationRepo.Load(ctx, req.LocationID)
	if err != nil {
		return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrLocationNotFound, err)
	}

	switch location.(type) {
	case locationentity.ActiveLocation, locationentity.IdleLocation:
		// valid — proceed to guard checks
	default:
		return ArchiveLocationResult{}, fmt.Errorf("%w: location must be Active or Idle to archive", ErrInvalidLocationState)
	}

	// Guard 1: no active child locations.
	children, err := i.locationRepo.FindChildren(ctx, req.LocationID)
	if err != nil {
		return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrRepositoryLoadFailed, err)
	}
	for _, child := range children {
		if _, ok := child.(locationentity.ActiveLocation); ok {
			return ArchiveLocationResult{}, ErrActiveChildrenExist
		}
	}

	// Guard 2: no active campaign has party here.
	campaigns, err := i.campaignRepo.FindByGame(ctx, req.GameID)
	if err != nil {
		return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrRepositoryLoadFailed, err)
	}
	for _, c := range campaigns {
		if _, ok := c.(campaignentity.ActiveCampaign); ok {
			if c.CurrentLocationID().String() == req.LocationID.String() {
				return ArchiveLocationResult{}, ErrPartyPresent
			}
		}
	}

	var archived locationentity.ArchivedLocation
	var events []event.Event

	switch l := location.(type) {
	case locationentity.ActiveLocation:
		archived, events, err = l.Archive(req.Source)
	case locationentity.IdleLocation:
		archived, events, err = l.Archive(req.Source)
	}
	if err != nil {
		return ArchiveLocationResult{}, err
	}

	if err := i.locationRepo.Save(ctx, archived); err != nil {
		return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
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
			return ArchiveLocationResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ArchiveLocationResult{Location: archived, Events: events}, nil
}
