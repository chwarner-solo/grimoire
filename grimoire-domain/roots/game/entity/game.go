package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// gameCore holds the shared state for all Game states.
type gameCore struct {
	id                  identity.GameID
	name                string
	campaignIDs         []identity.CampaignID
	activeCampaignCount int
}

func (c *gameCore) GameID() identity.GameID   { return c.id }
func (c *gameCore) GameName() string          { return c.name }
func (c *gameCore) isGame()                   {}

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
		Name:                c.name,
		State:               state,
		CampaignIDs:         ids,
		ActiveCampaignCount: c.activeCampaignCount,
	}
}

// CreateGame constructs a new Game aggregate in the New state.
func CreateGame(id identity.GameID, name string) (NewGame, error) {
	if id.IsZero() {
		return nil, ErrGameIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrGameNameRequired
	}
	return &newGame{gameCore: gameCore{
		id:   id,
		name: name,
	}}, nil
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
