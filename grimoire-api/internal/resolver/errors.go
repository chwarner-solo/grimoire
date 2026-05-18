package resolver

import (
	"errors"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	campaigninteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/interactor"
	characterinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/interactor"
	factioninteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/interactor"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	locationinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/interactor"
	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	narrativeinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/interactor"
	sharedinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var errUnauthenticated = &gqlerror.Error{Message: "UNAUTHENTICATED: no valid token provided"}

type errInvalidInput struct {
	field  string
	reason string
}

func (e *errInvalidInput) Error() string {
	return "INVALID_INPUT: " + e.field + ": " + e.reason
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, sharedinteractor.ErrUnauthorized):
		return &gqlerror.Error{Message: "UNAUTHORIZED: caller is not authorized"}

	case errors.Is(err, gameentity.ErrInvalidGameState):
		return &gqlerror.Error{Message: "INVALID_STATE: game is not in the required state"}

	case errors.Is(err, campaignentity.ErrInvalidCampaignState):
		return &gqlerror.Error{Message: "INVALID_STATE: campaign is not in the required state"}

	case errors.Is(err, campaignentity.ErrNoCharacters):
		return &gqlerror.Error{Message: "PRECONDITION_FAILED: campaign has no characters"}

	case isCycleDetected(err):
		return &gqlerror.Error{Message: "CYCLE_DETECTED: prerequisite would create a cycle"}

	case isPrerequisiteNotMet(err):
		return &gqlerror.Error{Message: "PRECONDITION_FAILED: beat prerequisites are not met"}

	case isNotFound(err):
		return &gqlerror.Error{Message: "NOT_FOUND: entity does not exist"}

	default:
		return &gqlerror.Error{Message: "INTERNAL: an internal error occurred"}
	}
}

func isCycleDetected(err error) bool {
	var target narrativeentity.ErrCycleDetected
	return errors.As(err, &target)
}

func isPrerequisiteNotMet(err error) bool {
	var target narrativeentity.ErrPrerequisiteNotMet
	return errors.As(err, &target)
}

func isNotFound(err error) bool {
	notFound := []error{
		// campaign
		campaigninteractor.ErrGameNotFound,
		campaigninteractor.ErrCampaignNotFound,
		// character
		characterinteractor.ErrGameNotFound,
		characterinteractor.ErrNPCNotFound,
		characterinteractor.ErrMacGuffinNotFound,
		characterinteractor.ErrPlayerCharacterNotFound,
		// faction
		factioninteractor.ErrGameNotFound,
		factioninteractor.ErrFactionNotFound,
		// location
		locationinteractor.ErrGameNotFound,
		locationinteractor.ErrLocationNotFound,
		// narrative
		narrativeinteractor.ErrGameNotFound,
		narrativeinteractor.ErrBeatNotFound,
		narrativeinteractor.ErrCampaignNarrativeNotFound,
		narrativeinteractor.ErrMasterNarrativeNotFound,
		// reveal
		sharedinteractor.ErrRevealGameNotFound,
		sharedinteractor.ErrEntityNotFound,
	}
	for _, sentinel := range notFound {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}
