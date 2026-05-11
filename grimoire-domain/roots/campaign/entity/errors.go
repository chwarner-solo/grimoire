package entity

import "errors"

var (
	ErrCampaignIDRequired    = errors.New("campaign: id is required")
	ErrCampaignNameRequired  = errors.New("campaign: name is required")
	ErrGameIDRequired        = errors.New("campaign: game id is required")
	ErrCharacterIDRequired   = errors.New("campaign: character id is required")
	ErrCharacterAlreadyAdded = errors.New("campaign: character already added")
	ErrNoCharacters          = errors.New("campaign: at least one character required")
	ErrSessionIDRequired     = errors.New("campaign: session id is required")
	ErrLocationIDRequired    = errors.New("campaign: location id is required")
	ErrInvalidCampaignState  = errors.New("campaign: invalid state")
	ErrUnexpectedEvent       = errors.New("campaign: unexpected event for current state")
)
