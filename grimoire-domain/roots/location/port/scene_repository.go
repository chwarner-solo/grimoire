package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// SceneRepository defines persistence operations for Scene child entities.
type SceneRepository interface {
	Save(ctx context.Context, scene *entity.Scene) error
	Load(ctx context.Context, id identity.SceneID) (*entity.Scene, error)
	LoadByLocation(ctx context.Context, id identity.LocationID) ([]*entity.Scene, error)
	LoadPrerequisiteChain(ctx context.Context, id identity.SceneID) ([]*entity.Scene, error)
}
