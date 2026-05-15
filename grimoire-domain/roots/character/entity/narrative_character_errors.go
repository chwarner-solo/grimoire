package entity

import "errors"

var (
	ErrNPCIDRequired          = errors.New("npc: id is required")
	ErrNPCNameRequired        = errors.New("npc: name is required")
	ErrNPCGameIDRequired      = errors.New("npc: game id is required")
	ErrNPCDescriptionRequired = errors.New("npc: description is required for activation")
	ErrLocationIDRequired     = errors.New("npc: location id is required")
	ErrMacGuffinIDRequired    = errors.New("npc: macguffin id is required")
	ErrSessionIDRequired      = errors.New("npc: session id is required")
	ErrInvalidNPCState        = errors.New("npc: invalid state")
	ErrNPCUnexpectedEvent     = errors.New("npc: unexpected event for current state")
	ErrMacGuffinAlreadyAssigned = errors.New("npc: macguffin already assigned")
)
