package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Location is the sealed interface for the Location aggregate root.
type Location interface {
	isLocation()
	LocationID() identity.LocationID
	LocationName() string
	GameID() identity.GameID
	Snapshot() LocationSnapshot
	Handle(evt event.Event) (Location, error)
}

// NewLocation represents a location that has just been created.
type NewLocation interface {
	Location
	BeginDraft(source event.Source) (DraftLocation, []event.Event, error)
}

// DraftLocation represents a location being built by the GM.
type DraftLocation interface {
	Location
	AddScene(sceneID identity.SceneID, source event.Source) (DraftLocation, []event.Event, error)
	AddChild(childID identity.LocationID, source event.Source) (DraftLocation, []event.Event, error)
	Activate(source event.Source) (ActiveLocation, []event.Event, error)
}

// ActiveLocation represents a location operating in the world.
type ActiveLocation interface {
	Location
	AddScene(sceneID identity.SceneID, source event.Source) (ActiveLocation, []event.Event, error)
	AddChild(childID identity.LocationID, source event.Source) (ActiveLocation, []event.Event, error)
	ConnectTo(locationID identity.LocationID, source event.Source) (ActiveLocation, []event.Event, error)
	GoIdle(source event.Source) (IdleLocation, []event.Event, error)
	Archive(source event.Source) (ArchivedLocation, []event.Event, error)
}

// IdleLocation represents a dormant location.
type IdleLocation interface {
	Location
	Reactivate(source event.Source) (ActiveLocation, []event.Event, error)
	Archive(source event.Source) (ArchivedLocation, []event.Event, error)
}

// ArchivedLocation is the terminal state. No transitions are possible.
type ArchivedLocation interface {
	Location
}
