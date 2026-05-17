package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// gameCore holds the shared state for all Game states.
type gameCore struct {
	id                  identity.GameID
	gmID                string
	name                string
	masterNarrativeID   identity.MasterNarrativeID
	campaignIDs         []identity.CampaignID
	activeCampaignCount int
}

func (c *gameCore) GameID() identity.GameID                       { return c.id }
func (c *gameCore) GMID() string                                  { return c.gmID }
func (c *gameCore) GameName() string                              { return c.name }
func (c *gameCore) MasterNarrativeID() identity.MasterNarrativeID { return c.masterNarrativeID }
func (c *gameCore) isGame()                                       {}

func (c *gameCore) hasCampaign(id identity.CampaignID) bool {
	for _, cid := range c.campaignIDs {
		if cid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *gameCore) snapshot(state string) GameSnapshot {
	ids := make([]identity.CampaignID, len(c.campaignIDs))
	copy(ids, c.campaignIDs)
	return GameSnapshot{
		ID:                  c.id,
		GMID:                c.gmID,
		Name:                c.name,
		MasterNarrativeID:   c.masterNarrativeID,
		State:               state,
		CampaignIDs:         ids,
		ActiveCampaignCount: c.activeCampaignCount,
	}
}

// CreateGame constructs a new Game aggregate in the New state.
// gmID is the Firebase UID of the GM creating the game.
func CreateGame(id identity.GameID, gmID string, name string, source event.Source) (NewGame, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	if strings.TrimSpace(gmID) == "" {
		return nil, nil, ErrGMIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrGameNameRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "game",
		Name:       name,
		Source:     source,
	}
	return &newGame{gameCore: gameCore{
		id:   id,
		gmID: gmID,
		name: name,
	}}, []event.Event{evt}, nil
}

// --- State structs ---

type newGame struct{ gameCore }
type draftGame struct{ gameCore }
type activeGame struct{ gameCore }
type idleGame struct{ gameCore }
type archivedGame struct{ gameCore }

// Snapshot implementations
func (g *newGame) Snapshot() GameSnapshot      { return g.snapshot("new") }
func (g *draftGame) Snapshot() GameSnapshot    { return g.snapshot("draft") }
func (g *activeGame) Snapshot() GameSnapshot   { return g.snapshot("active") }
func (g *idleGame) Snapshot() GameSnapshot     { return g.snapshot("idle") }
func (g *archivedGame) Snapshot() GameSnapshot { return g.snapshot("archived") }
