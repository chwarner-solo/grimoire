package entity

import (
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// LocationSnapshot is the DTO used for persistence and reconstitution.
type LocationSnapshot struct {
	ID                   identity.LocationID
	GameID               identity.GameID
	LocationType         value.LocationType
	ParentLocationID     identity.LocationID
	ChildLocationIDs     []identity.LocationID
	SceneIDs             []identity.SceneID
	ConnectedLocationIDs []identity.LocationID
	PlayerVisible        bool
	Name                 string
	Description          string
	PlayerDescription    string
	RevealedNotes        string
	State                string
}

// ReconstituteLocation rebuilds a Location aggregate from a persisted snapshot.
func ReconstituteLocation(snap LocationSnapshot) (Location, error) {
	core := locationCore{
		id:                   snap.ID,
		gameID:               snap.GameID,
		locationType:         snap.LocationType,
		parentLocationID:     snap.ParentLocationID,
		childLocationIDs:     snap.ChildLocationIDs,
		sceneIDs:             snap.SceneIDs,
		connectedLocationIDs: snap.ConnectedLocationIDs,
		playerVisible:        snap.PlayerVisible,
		name:                 snap.Name,
		description:          snap.Description,
		playerDescription:    snap.PlayerDescription,
		revealedNotes:        snap.RevealedNotes,
	}
	switch snap.State {
	case "new":
		return &newLocation{locationCore: core}, nil
	case "draft":
		return &draftLocation{locationCore: core}, nil
	case "active":
		return &activeLocation{locationCore: core}, nil
	case "idle":
		return &idleLocation{locationCore: core}, nil
	case "archived":
		return &archivedLocation{locationCore: core}, nil
	default:
		return nil, fmt.Errorf("%w: unknown state %q", ErrInvalidLocationState, snap.State)
	}
}
