package interactor

import "errors"

var (
	ErrRepositorySaveFailed = errors.New("interactor: failed to save game")
	ErrEventDispatchFailed  = errors.New("interactor: failed to dispatch event")
)
