package interactor

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository is the persistence port for the Game aggregate.
// Defined locally to avoid importing the port package.
type GameRepository interface {
	Save(ctx context.Context, game entity.Game) error
	Load(ctx context.Context, id identity.GameID) (entity.Game, error)
}