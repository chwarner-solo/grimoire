package interactor

import "errors"

var (
	ErrRepositorySaveFailed    = errors.New("interactor: failed to save")
	ErrRepositoryLoadFailed    = errors.New("interactor: failed to load")
	ErrEventDispatchFailed     = errors.New("interactor: failed to dispatch event")
	ErrGameNotFound            = errors.New("interactor: game not found")
	ErrNPCNotFound             = errors.New("interactor: npc not found")
	ErrPlayerCharacterNotFound = errors.New("interactor: player character not found")
	ErrMacGuffinNotFound       = errors.New("interactor: macguffin not found")
	ErrMacGuffinDestroyed      = errors.New("interactor: macguffin is destroyed")
	ErrInvalidNPCState         = errors.New("interactor: npc is not in required state")
	ErrCharacterRetired        = errors.New("interactor: player character is retired")
)
