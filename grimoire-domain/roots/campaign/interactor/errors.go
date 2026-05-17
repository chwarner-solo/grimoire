package interactor

import "errors"

var (
	ErrRepositorySaveFailed = errors.New("interactor: failed to save campaign")
	ErrRepositoryLoadFailed = errors.New("interactor: failed to load campaign")
	ErrEventDispatchFailed  = errors.New("interactor: failed to dispatch event")
	ErrGameNotFound         = errors.New("interactor: game not found")
	ErrCampaignNotFound     = errors.New("interactor: campaign not found")
	ErrInvalidCampaignState = errors.New("interactor: campaign is not in required state")
)