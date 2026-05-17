package interactor

import (
	"context"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

type RevealGameRepository interface {
	Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type RevealNPCRepository interface {
	Save(ctx context.Context, snap characterentity.NPCSnapshot) error
	Load(ctx context.Context, id identity.NarrativeCharacterID) (characterentity.NPCSnapshot, error)
}

type RevealPlayerCharacterRepository interface {
	SaveCore(ctx context.Context, snap characterentity.PlayerCharacterSnapshot) error
	LoadCore(ctx context.Context, id identity.PlayerCharacterID) (characterentity.PlayerCharacterSnapshot, error)
}

type RevealMacGuffinRepository interface {
	Save(ctx context.Context, snap characterentity.MacGuffinSnapshot) error
	Load(ctx context.Context, id identity.MacGuffinID) (characterentity.MacGuffinSnapshot, error)
}

type RevealFactionRepository interface {
	Save(ctx context.Context, faction factionentity.Faction) error
	Load(ctx context.Context, id identity.FactionID) (factionentity.Faction, error)
}
