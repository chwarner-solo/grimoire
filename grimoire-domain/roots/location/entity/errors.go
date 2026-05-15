package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

var (
	ErrLocationIDRequired        = errors.New("location: id is required")
	ErrLocationNameRequired      = errors.New("location: name is required")
	ErrGameIDRequired            = errors.New("location: game id is required")
	ErrLocationTypeRequired      = errors.New("location: location type is required")
	ErrSceneIDRequired           = errors.New("location: scene id is required")
	ErrSceneNameRequired         = errors.New("location: scene name is required")
	ErrSceneAlreadyExists        = errors.New("location: scene already exists")
	ErrChildAlreadyExists        = errors.New("location: child already exists")
	ErrCannotConnectToSelf       = errors.New("location: cannot connect to self")
	ErrAlreadyConnected          = errors.New("location: already connected")
	ErrConnectionIDRequired      = errors.New("location: connection id is required")
	ErrInvalidLocationState      = errors.New("location: invalid state")
	ErrUnexpectedEvent           = errors.New("location: unexpected event for current state")
)

// ErrLocationCycleDetected is returned when a location hierarchy would form a cycle.
type ErrLocationCycleDetected struct {
	LocationID identity.LocationID
}

func (e ErrLocationCycleDetected) Error() string {
	return fmt.Sprintf("location: cycle detected involving location %s", e.LocationID)
}

// ErrSceneCycleDetected is returned when scene prerequisites would form a cycle.
type ErrSceneCycleDetected struct {
	SceneID identity.SceneID
}

func (e ErrSceneCycleDetected) Error() string {
	return fmt.Sprintf("location: cycle detected involving scene %s", e.SceneID)
}

// ErrAlreadyConnectedTo is returned when a connection already exists to the target.
type ErrAlreadyConnectedTo struct {
	LocationID identity.LocationID
}

func (e ErrAlreadyConnectedTo) Error() string {
	return fmt.Sprintf("location: already connected to %s", e.LocationID)
}
