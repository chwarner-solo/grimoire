package entity

import "errors"

var (
	ErrPlayerCharacterIDRequired = errors.New("player character: id is required")
	ErrCharacterNameRequired     = errors.New("player character: name is required")
	ErrOwnerPlayerIDRequired     = errors.New("player character: owner player id is required")
	ErrPCGameIDRequired          = errors.New("player character: game id is required")
	ErrCharacterAlreadyRetired   = errors.New("player character: already retired")
	ErrCharacterRetired          = errors.New("player character: character is retired")
	ErrMacGuffinNotPossessed     = errors.New("player character: macguffin not possessed")
	ErrAlreadyInCampaign         = errors.New("player character: already in campaign")
)
