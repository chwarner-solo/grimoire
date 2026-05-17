package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Game is the sealed interface for the Game aggregate root.
// The unexported isGame() marker prevents external implementations.
type Game interface {
	isGame()
	GameID() identity.GameID
	GMID() string
	GameName() string
	MasterNarrativeID() identity.MasterNarrativeID
	Snapshot() GameSnapshot
	Handle(evt event.Event) (Game, error)
}

// NewGame represents a game that has just been created.
// The only valid transition is receiving a MasterNarrative creation, which moves to Draft.
type NewGame interface {
	Game
	OnNarrativeCreated(id identity.MasterNarrativeID) (DraftGame, error)
}

// DraftGame represents a game that has narrative content but no active campaigns.
type DraftGame interface {
	Game
	LinkCampaign(id identity.CampaignID, source event.Source) (DraftGame, []event.Event, error)
	ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error)
}

// ActiveGame represents a game with at least one active campaign.
type ActiveGame interface {
	Game
	LinkCampaign(id identity.CampaignID, source event.Source) (ActiveGame, []event.Event, error)
	NotifyCampaignIdle(id identity.CampaignID) (Game, error)
}

// IdleGame represents a game whose campaigns have all gone idle.
type IdleGame interface {
	Game
	ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error)
	Archive(source event.Source) (ArchivedGame, []event.Event, error)
}

// ArchivedGame is the terminal state. No transitions are possible.
type ArchivedGame interface {
	Game
}
