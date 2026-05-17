package interactor

import (
	"context"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository is the persistence port for the Game aggregate.
type GameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

// CampaignRepository is the persistence port for the Campaign aggregate.
type CampaignRepository interface {
	Save(ctx context.Context, campaign campaignentity.Campaign) error
	Load(ctx context.Context, id identity.CampaignID) (campaignentity.Campaign, error)
}
