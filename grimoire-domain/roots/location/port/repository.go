package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// LocationRepository defines persistence operations for the Location aggregate.
type LocationRepository interface {
	Save(ctx context.Context, location entity.Location) error
	Load(ctx context.Context, id identity.LocationID) (entity.Location, error)
	LoadChildChain(ctx context.Context, id identity.LocationID) ([]entity.Location, error)
}
