package interactor

import (
	"context"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

type GameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type NPCRepository interface {
	Save(ctx context.Context, snap characterentity.NPCSnapshot) error
	Load(ctx context.Context, id identity.NarrativeCharacterID) (characterentity.NPCSnapshot, error)
}

type PlayerCharacterRepository interface {
	SaveCore(ctx context.Context, snap characterentity.PlayerCharacterSnapshot) error
	LoadCore(ctx context.Context, id identity.PlayerCharacterID) (characterentity.PlayerCharacterSnapshot, error)
}

type MacGuffinRepository interface {
	Save(ctx context.Context, snap characterentity.MacGuffinSnapshot) error
	Load(ctx context.Context, id identity.MacGuffinID) (characterentity.MacGuffinSnapshot, error)
}
