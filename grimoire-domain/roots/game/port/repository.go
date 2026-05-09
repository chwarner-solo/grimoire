package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository defines the persistence port for the Game aggregate.
type GameRepository interface {
	Save(ctx context.Context, game entity.Game) error
	Load(ctx context.Context, id identity.GameID) (entity.Game, error)
}
