package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// PlayerCharacterRepository defines persistence operations for PlayerCharacter and its narrative.
type PlayerCharacterRepository interface {
	SaveCore(ctx context.Context, snap entity.PlayerCharacterSnapshot) error
	LoadCore(ctx context.Context, id identity.PlayerCharacterID) (entity.PlayerCharacterSnapshot, error)
	SaveNarrative(ctx context.Context, snap entity.PlayerCharacterNarrativeSnapshot) error
	LoadNarrative(ctx context.Context, id identity.PlayerCharacterNarrativeID) (entity.PlayerCharacterNarrativeSnapshot, error)
	FindNarrativesByCharacter(ctx context.Context, id identity.PlayerCharacterID) ([]entity.PlayerCharacterNarrativeSnapshot, error)
	FindByGame(ctx context.Context, id identity.GameID) ([]entity.PlayerCharacterSnapshot, error)
	FindByCampaign(ctx context.Context, id identity.CampaignID) ([]entity.PlayerCharacterSnapshot, error)
}
