package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// FactionRepository defines persistence operations for the Faction aggregate.
type FactionRepository interface {
	Save(ctx context.Context, faction entity.Faction) error
	Load(ctx context.Context, id identity.FactionID) (entity.Faction, error)
	SaveMembership(ctx context.Context, m *entity.FactionMembership) error
	LoadMembership(ctx context.Context, id identity.FactionMembershipID) (*entity.FactionMembership, error)
}
