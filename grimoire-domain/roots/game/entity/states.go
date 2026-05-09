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
	GameName() string
	Snapshot() GameSnapshot
	Handle(evt event.Event) (Game, error)
}

// NewGame represents a game that has just been created.
// The only valid transition is adding a narrative element, which moves to Draft.
type NewGame interface {
	Game
	AddNarrativeElement(id identity.NarrativeID, name string) (DraftGame, error)
}

// DraftGame represents a game that has narrative content but no active campaigns.
type DraftGame interface {
	Game
	AddNarrativeElement(id identity.NarrativeID, name string) (DraftGame, error)
	LinkCampaign(id identity.CampaignID) (DraftGame, error)
	ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error)
}

// ActiveGame represents a game with at least one active campaign.
type ActiveGame interface {
	Game
	LinkCampaign(id identity.CampaignID) (ActiveGame, error)
	NotifyCampaignIdle(id identity.CampaignID) (Game, error)
}

// IdleGame represents a game whose campaigns have all gone idle.
type IdleGame interface {
	Game
	ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error)
	Archive() (ArchivedGame, error)
}

// ArchivedGame is the terminal state. No transitions are possible.
type ArchivedGame interface {
	Game
}
