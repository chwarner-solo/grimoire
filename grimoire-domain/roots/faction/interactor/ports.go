package interactor

import (
	"context"

	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository is the persistence port for the Game aggregate.
type GameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

// FactionRepository is the persistence port for the Faction aggregate.
type FactionRepository interface {
	Save(ctx context.Context, faction factionentity.Faction) error
	Load(ctx context.Context, id identity.FactionID) (factionentity.Faction, error)
	SaveMembership(ctx context.Context, m *factionentity.FactionMembership) error
	LoadMembership(ctx context.Context, id identity.FactionMembershipID) (*factionentity.FactionMembership, error)
}
