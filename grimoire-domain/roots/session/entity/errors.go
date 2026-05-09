package entity

import "errors"

var (
	ErrSessionIDRequired   = errors.New("session: id is required")
	ErrCampaignIDRequired  = errors.New("session: campaign id is required")
	ErrNotesRequired       = errors.New("session: notes are required")
	ErrInvalidSessionState = errors.New("session: invalid state")
	ErrUnexpectedEvent     = errors.New("session: unexpected event for current state")
)
