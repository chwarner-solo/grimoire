package interactor

import "errors"

var (
	ErrRepositorySaveFailed = errors.New("interactor: failed to save faction")
	ErrRepositoryLoadFailed = errors.New("interactor: failed to load faction")
	ErrEventDispatchFailed  = errors.New("interactor: failed to dispatch event")
	ErrGameNotFound         = errors.New("interactor: game not found")
	ErrFactionNotFound      = errors.New("interactor: faction not found")
	ErrInvalidFactionState  = errors.New("interactor: faction is not in required state")
)
