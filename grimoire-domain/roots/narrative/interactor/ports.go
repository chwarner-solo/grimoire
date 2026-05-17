package interactor

import (
	"context"

	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository is the persistence port for the Game aggregate.
type GameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

// MasterNarrativeRepository is the persistence port for MasterNarrative.
type MasterNarrativeRepository interface {
	Save(ctx context.Context, mn *narrativeentity.MasterNarrative) error
	Load(ctx context.Context, id identity.MasterNarrativeID) (*narrativeentity.MasterNarrative, error)
	FindByGame(ctx context.Context, id identity.GameID) (*narrativeentity.MasterNarrative, error)
}

// CampaignNarrativeRepository is the persistence port for CampaignNarrative.
type CampaignNarrativeRepository interface {
	Save(ctx context.Context, cn *narrativeentity.CampaignNarrative) error
	Load(ctx context.Context, id identity.CampaignNarrativeID) (*narrativeentity.CampaignNarrative, error)
	FindByCampaign(ctx context.Context, id identity.CampaignID) (*narrativeentity.CampaignNarrative, error)
}

// BeatRepository is the persistence port for Beat.
type BeatRepository interface {
	Save(ctx context.Context, beat *narrativeentity.Beat) error
	Load(ctx context.Context, id identity.BeatID) (*narrativeentity.Beat, error)
	LoadPrerequisiteChain(ctx context.Context, id identity.BeatID) ([]*narrativeentity.Beat, error)
	FindByGame(ctx context.Context, id identity.GameID) ([]*narrativeentity.Beat, error)
}
