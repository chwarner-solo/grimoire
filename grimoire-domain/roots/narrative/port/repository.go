package port

import (
	"context"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// BeatRepository defines persistence operations for Beat entities.
type BeatRepository interface {
	FindByID(ctx context.Context, id identity.BeatID) (*entity.Beat, error)
	LoadPrerequisiteChain(ctx context.Context, id identity.BeatID) ([]*entity.Beat, error)
	Save(ctx context.Context, beat *entity.Beat) error
}

// MasterNarrativeRepository defines persistence operations for the MasterNarrative aggregate.
type MasterNarrativeRepository interface {
	Save(ctx context.Context, mn *entity.MasterNarrative) error
	Load(ctx context.Context, id identity.MasterNarrativeID) (*entity.MasterNarrative, error)
}

// CampaignNarrativeRepository defines persistence operations for the CampaignNarrative aggregate.
type CampaignNarrativeRepository interface {
	Save(ctx context.Context, cn *entity.CampaignNarrative) error
	Load(ctx context.Context, id identity.CampaignNarrativeID) (*entity.CampaignNarrative, error)
}
