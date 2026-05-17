package interactor

import "errors"

var (
	ErrRepositorySaveFailed    = errors.New("interactor: failed to save location")
	ErrRepositoryLoadFailed    = errors.New("interactor: failed to load location")
	ErrEventDispatchFailed     = errors.New("interactor: failed to dispatch event")
	ErrGameNotFound            = errors.New("interactor: game not found")
	ErrLocationNotFound        = errors.New("interactor: location not found")
	ErrInvalidLocationState    = errors.New("interactor: location is not in required state")
	ErrActiveChildrenExist     = errors.New("interactor: cannot archive — active child locations exist")
	ErrPartyPresent            = errors.New("interactor: cannot archive — party is currently here")
	ErrConnectionAlreadyExists = errors.New("interactor: connection already exists")
	ErrSelfConnection          = errors.New("interactor: location cannot connect to itself")
)
