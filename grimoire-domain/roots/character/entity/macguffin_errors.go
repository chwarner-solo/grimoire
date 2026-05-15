package entity

import "errors"

var (
	ErrMGMacGuffinIDRequired        = errors.New("macguffin: id is required")
	ErrMacGuffinNameRequired        = errors.New("macguffin: name is required")
	ErrMGGameIDRequired             = errors.New("macguffin: game id is required")
	ErrMacGuffinPossessionConflict  = errors.New("macguffin: possession invariant violated — at most one possessor or location")
	ErrMacGuffinDestroyed           = errors.New("macguffin: macguffin is destroyed")
	ErrMacGuffinAlreadyDestroyed    = errors.New("macguffin: already destroyed")
	ErrMGLocationIDRequired         = errors.New("macguffin: location id is required")
	ErrMGCharacterIDRequired        = errors.New("macguffin: character id is required")
)
