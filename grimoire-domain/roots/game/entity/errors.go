package entity

import "errors"

var (
	ErrGameIDRequired        = errors.New("game: id is required")
	ErrGameNameRequired      = errors.New("game: name is required")
	ErrNarrativeIDRequired   = errors.New("game: narrative id is required")
	ErrNarrativeNameRequired = errors.New("game: narrative name is required")
	ErrCampaignIDRequired    = errors.New("game: campaign id is required")
	ErrCampaignAlreadyLinked = errors.New("game: campaign already linked")
	ErrCampaignNotLinked     = errors.New("game: campaign not linked")
	ErrInvalidGameState      = errors.New("game: invalid state")
)
