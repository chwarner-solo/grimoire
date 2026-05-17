package interactor

import "errors"

var (
	ErrRepositorySaveFailed = errors.New("interactor: failed to save game")
	ErrRepositoryLoadFailed = errors.New("interactor: failed to load game")
	ErrEventDispatchFailed  = errors.New("interactor: failed to dispatch event")
)
