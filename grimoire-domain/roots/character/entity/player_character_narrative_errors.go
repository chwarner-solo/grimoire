package entity

import "errors"

var (
	ErrPlayerCharacterNarrativeIDRequired = errors.New("player character narrative: id is required")
	ErrPCNCharacterIDRequired             = errors.New("player character narrative: character id is required")
	ErrPCNCampaignIDRequired              = errors.New("player character narrative: campaign id is required")
	ErrPCNGameIDRequired                  = errors.New("player character narrative: game id is required")
	ErrBackgroundBeatAlreadyAdded         = errors.New("player character narrative: background beat already added")
	ErrPrerequisitesNotMet                = errors.New("player character narrative: prerequisites not met")
	ErrBackgroundBeatNotFound             = errors.New("player character narrative: background beat not found")
	ErrBackgroundBeatAlreadyRevealed      = errors.New("player character narrative: background beat already revealed")
)
