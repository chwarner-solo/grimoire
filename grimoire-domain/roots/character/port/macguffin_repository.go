package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// MacGuffinRepository defines persistence operations for the MacGuffin aggregate.
type MacGuffinRepository interface {
	Save(ctx context.Context, snap entity.MacGuffinSnapshot) error
	Load(ctx context.Context, id identity.MacGuffinID) (entity.MacGuffinSnapshot, error)
	FindByNPCPossessor(ctx context.Context, id identity.NarrativeCharacterID) ([]entity.MacGuffinSnapshot, error)
	FindByPCPossessor(ctx context.Context, id identity.PlayerCharacterID) ([]entity.MacGuffinSnapshot, error)
	FindByLocation(ctx context.Context, id identity.LocationID) ([]entity.MacGuffinSnapshot, error)
	FindByGame(ctx context.Context, id identity.GameID) ([]entity.MacGuffinSnapshot, error)
}
