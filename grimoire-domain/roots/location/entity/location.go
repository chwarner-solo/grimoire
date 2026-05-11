package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// locationCore holds the shared state for all Location states.
type locationCore struct {
	id                   identity.LocationID
	gameID               identity.GameID
	locationType         value.LocationType
	parentLocationID     identity.LocationID
	childLocationIDs     []identity.LocationID
	sceneIDs             []identity.SceneID
	connectedLocationIDs []identity.LocationID
	playerVisible        bool
	name                 string
	description          string
	playerDescription    string
	revealedNotes        string
}

func (c *locationCore) LocationID() identity.LocationID { return c.id }
func (c *locationCore) LocationName() string            { return c.name }
func (c *locationCore) GameID() identity.GameID         { return c.gameID }
func (c *locationCore) isLocation()                     {}

func (c *locationCore) hasScene(id identity.SceneID) bool {
	for _, sid := range c.sceneIDs {
		if sid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *locationCore) hasChild(id identity.LocationID) bool {
	for _, cid := range c.childLocationIDs {
		if cid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *locationCore) isConnectedTo(id identity.LocationID) bool {
	for _, cid := range c.connectedLocationIDs {
		if cid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *locationCore) snapshot(state string) LocationSnapshot {
	childIDs := make([]identity.LocationID, len(c.childLocationIDs))
	copy(childIDs, c.childLocationIDs)
	sceneIDs := make([]identity.SceneID, len(c.sceneIDs))
	copy(sceneIDs, c.sceneIDs)
	connIDs := make([]identity.LocationID, len(c.connectedLocationIDs))
	copy(connIDs, c.connectedLocationIDs)
	return LocationSnapshot{
		ID:                   c.id,
		GameID:               c.gameID,
		LocationType:         c.locationType,
		ParentLocationID:     c.parentLocationID,
		ChildLocationIDs:     childIDs,
		SceneIDs:             sceneIDs,
		ConnectedLocationIDs: connIDs,
		PlayerVisible:        c.playerVisible,
		Name:                 c.name,
		Description:          c.description,
		PlayerDescription:    c.playerDescription,
		RevealedNotes:        c.revealedNotes,
		State:                state,
	}
}

// CreateLocation constructs a new Location aggregate in the New state.
func CreateLocation(id identity.LocationID, name string, gameID identity.GameID, locationType value.LocationType, parentLocationID identity.LocationID, source event.Source) (NewLocation, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrLocationNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	if locationType == "" {
		return nil, nil, ErrLocationTypeRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "location",
		Name:       name,
		Source:     source,
	}
	return &newLocation{locationCore: locationCore{
		id:               id,
		gameID:           gameID,
		locationType:     locationType,
		parentLocationID: parentLocationID,
		name:             name,
	}}, []event.Event{evt}, nil
}

// --- State structs ---

type newLocation struct{ locationCore }
type draftLocation struct{ locationCore }
type activeLocation struct{ locationCore }
type idleLocation struct{ locationCore }
type archivedLocation struct{ locationCore }

// Snapshot implementations
func (l *newLocation) Snapshot() LocationSnapshot      { return l.snapshot("new") }
func (l *draftLocation) Snapshot() LocationSnapshot    { return l.snapshot("draft") }
func (l *activeLocation) Snapshot() LocationSnapshot   { return l.snapshot("active") }
func (l *idleLocation) Snapshot() LocationSnapshot     { return l.snapshot("idle") }
func (l *archivedLocation) Snapshot() LocationSnapshot { return l.snapshot("archived") }
