package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// NPCRepository defines persistence operations for the NPC aggregate.
type NPCRepository interface {
	Save(ctx context.Context, snap entity.NPCSnapshot) error
	Load(ctx context.Context, id identity.NarrativeCharacterID) (entity.NPCSnapshot, error)
	FindByLocation(ctx context.Context, id identity.LocationID) ([]entity.NPCSnapshot, error)
	FindByGame(ctx context.Context, id identity.GameID) ([]entity.NPCSnapshot, error)
}
