package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// CampaignRepository defines the persistence port for the Campaign aggregate.
type CampaignRepository interface {
	Save(ctx context.Context, campaign entity.Campaign) error
	Load(ctx context.Context, id identity.CampaignID) (entity.Campaign, error)
}
