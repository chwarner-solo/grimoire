package interactor

import (
	"context"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

type GameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type LocationRepository interface {
	Save(ctx context.Context, location locationentity.Location) error
	Load(ctx context.Context, id identity.LocationID) (locationentity.Location, error)
	FindChildren(ctx context.Context, id identity.LocationID) ([]locationentity.Location, error)
	FindByGame(ctx context.Context, id identity.GameID) ([]locationentity.Location, error)
}

type CampaignRepository interface {
	FindByGame(ctx context.Context, id identity.GameID) ([]campaignentity.Campaign, error)
}
