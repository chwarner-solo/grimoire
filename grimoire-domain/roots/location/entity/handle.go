package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- copyCore ---

func (c *locationCore) copyCore() locationCore {
	childIDs := make([]identity.LocationID, len(c.childLocationIDs))
	copy(childIDs, c.childLocationIDs)
	sceneIDs := make([]identity.SceneID, len(c.sceneIDs))
	copy(sceneIDs, c.sceneIDs)
	connIDs := make([]identity.LocationID, len(c.connectedLocationIDs))
	copy(connIDs, c.connectedLocationIDs)
	return locationCore{
		id:                   c.id,
		gameID:               c.gameID,
		locationType:         c.locationType,
		parentLocationID:     c.parentLocationID,
		childLocationIDs:     childIDs,
		sceneIDs:             sceneIDs,
		connectedLocationIDs: connIDs,
		playerVisible:        c.playerVisible,
		name:                 c.name,
		description:          c.description,
		playerDescription:    c.playerDescription,
		revealedNotes:        c.revealedNotes,
	}
}

// --- newLocation Handle ---

func (l *newLocation) Handle(evt event.Event) (Location, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "draft" {
			core := l.copyCore()
			return &draftLocation{locationCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- draftLocation Handle ---

func (l *draftLocation) Handle(evt event.Event) (Location, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		core := l.copyCore()
		switch e.Relationship {
		case "has_scene":
			sceneID, err := identity.ParseSceneID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("location handle: %w", err)
			}
			core.sceneIDs = append(core.sceneIDs, sceneID)
			return &draftLocation{locationCore: core}, nil
		case "contains":
			childID, err := identity.ParseLocationID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("location handle: %w", err)
			}
			core.childLocationIDs = append(core.childLocationIDs, childID)
			return &draftLocation{locationCore: core}, nil
		default:
			return nil, ErrUnexpectedEvent
		}
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "active" {
			core := l.copyCore()
			return &activeLocation{locationCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- activeLocation Handle ---

func (l *activeLocation) Handle(evt event.Event) (Location, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		core := l.copyCore()
		switch e.Relationship {
		case "has_scene":
			sceneID, err := identity.ParseSceneID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("location handle: %w", err)
			}
			core.sceneIDs = append(core.sceneIDs, sceneID)
			return &activeLocation{locationCore: core}, nil
		case "contains":
			childID, err := identity.ParseLocationID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("location handle: %w", err)
			}
			core.childLocationIDs = append(core.childLocationIDs, childID)
			return &activeLocation{locationCore: core}, nil
		case "connects_to":
			connID, err := identity.ParseLocationID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("location handle: %w", err)
			}
			core.connectedLocationIDs = append(core.connectedLocationIDs, connID)
			return &activeLocation{locationCore: core}, nil
		default:
			return nil, ErrUnexpectedEvent
		}
	case event.EntityUpdated:
		if e.Field == "status" {
			core := l.copyCore()
			switch e.NewValue {
			case "idle":
				return &idleLocation{locationCore: core}, nil
			case "archived":
				return &archivedLocation{locationCore: core}, nil
			default:
				return nil, ErrUnexpectedEvent
			}
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- idleLocation Handle ---

func (l *idleLocation) Handle(evt event.Event) (Location, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" {
			core := l.copyCore()
			switch e.NewValue {
			case "active":
				return &activeLocation{locationCore: core}, nil
			case "archived":
				return &archivedLocation{locationCore: core}, nil
			default:
				return nil, ErrUnexpectedEvent
			}
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- archivedLocation Handle ---

func (l *archivedLocation) Handle(_ event.Event) (Location, error) {
	return nil, ErrUnexpectedEvent
}

// --- ReplayLocation ---

// ReplayLocation rebuilds a Location aggregate from a sequence of events.
// The first event must be an EntityCreated for a location.
// gameID, locationType, and parentLocationID are required because they are not
// stored in the event payload.
func ReplayLocation(gameID identity.GameID, locationType value.LocationType, parentLocationID identity.LocationID, events []event.Event) (Location, error) {
	if len(events) == 0 {
		return nil, errors.New("location: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("location: first event must be EntityCreated")
	}

	locationID, err := identity.ParseLocationID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("location replay: %w", err)
	}

	l, _, err := CreateLocation(locationID, first.Name, gameID, locationType, parentLocationID, event.SourceGrimoire)
	if err != nil {
		return nil, fmt.Errorf("location replay: %w", err)
	}

	var current Location = l
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("location replay: %w", err)
		}
	}
	return current, nil
}
