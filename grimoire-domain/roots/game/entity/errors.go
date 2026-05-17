package entity

import "errors"

var (
	ErrGameIDRequired            = errors.New("game: id is required")
	ErrGMIDRequired              = errors.New("game: gm id is required")
	ErrGameNameRequired          = errors.New("game: name is required")
	ErrMasterNarrativeIDRequired = errors.New("game: master narrative id is required")
	ErrCampaignIDRequired        = errors.New("game: campaign id is required")
	ErrCampaignAlreadyLinked = errors.New("game: campaign already linked")
	ErrCampaignNotLinked     = errors.New("game: campaign not linked")
	ErrInvalidGameState      = errors.New("game: invalid state")
	ErrUnexpectedEvent       = errors.New("game: unexpected event for current state")
)
