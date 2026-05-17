package interactor

import "errors"

var (
	ErrRevealRepositorySaveFailed  = errors.New("reveal: failed to save entity")
	ErrRevealRepositoryLoadFailed  = errors.New("reveal: failed to load entity")
	ErrRevealEventDispatchFailed   = errors.New("reveal: failed to dispatch event")
	ErrRevealGameNotFound          = errors.New("reveal: game not found")
	ErrEntityNotFound              = errors.New("reveal: entity not found")
	ErrInvalidEntityState          = errors.New("reveal: entity is not in a revealable state")
	ErrUnsupportedEntityType       = errors.New("reveal: unsupported entity type")
)
