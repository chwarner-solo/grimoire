package interactor

import "errors"

var (
	ErrRepositorySaveFailed        = errors.New("interactor: failed to save")
	ErrRepositoryLoadFailed        = errors.New("interactor: failed to load")
	ErrEventDispatchFailed         = errors.New("interactor: failed to dispatch event")
	ErrGameNotFound                = errors.New("interactor: game not found")
	ErrBeatNotFound                = errors.New("interactor: beat not found")
	ErrMasterNarrativeNotFound     = errors.New("interactor: master narrative not found")
	ErrCampaignNarrativeNotFound   = errors.New("interactor: campaign narrative not found")
	ErrPrerequisiteChainLoadFailed = errors.New("interactor: failed to load prerequisite chain")
	ErrInvalidBeatType             = errors.New("interactor: invalid beat type for this operation")
)
