package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- newGame Handle ---

func (g *newGame) Handle(evt event.Event) (Game, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "master_narrative_id" {
			mnID, err := identity.ParseMasterNarrativeID(e.NewValue)
			if err != nil {
				return nil, fmt.Errorf("game handle: %w", err)
			}
			core := g.gameCore
			core.masterNarrativeID = mnID
			return &draftGame{gameCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- draftGame Handle ---

func (g *draftGame) Handle(evt event.Event) (Game, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		if e.Relationship != "campaign" {
			return nil, ErrUnexpectedEvent
		}
		campID, err := identity.ParseCampaignID(e.EntityBID)
		if err != nil {
			return nil, fmt.Errorf("game handle: %w", err)
		}
		core := g.copyCore()
		core.campaignIDs = append(core.campaignIDs, campID)
		return &draftGame{gameCore: core}, nil
	case event.SessionStarted:
		core := g.copyCore()
		core.activeCampaignCount++
		return &activeGame{gameCore: core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- activeGame Handle ---

func (g *activeGame) Handle(evt event.Event) (Game, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		if e.Relationship != "campaign" {
			return nil, ErrUnexpectedEvent
		}
		campID, err := identity.ParseCampaignID(e.EntityBID)
		if err != nil {
			return nil, fmt.Errorf("game handle: %w", err)
		}
		core := g.copyCore()
		core.campaignIDs = append(core.campaignIDs, campID)
		return &activeGame{gameCore: core}, nil
	case event.SessionStarted:
		core := g.copyCore()
		core.activeCampaignCount++
		return &activeGame{gameCore: core}, nil
	case event.SessionEnded:
		core := g.copyCore()
		core.activeCampaignCount--
		if core.activeCampaignCount <= 0 {
			core.activeCampaignCount = 0
			return &idleGame{gameCore: core}, nil
		}
		return &activeGame{gameCore: core}, nil
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- idleGame Handle ---

func (g *idleGame) Handle(evt event.Event) (Game, error) {
	switch e := evt.(type) {
	case event.SessionStarted:
		core := g.copyCore()
		core.activeCampaignCount = 1
		return &activeGame{gameCore: core}, nil
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "archived" {
			core := g.copyCore()
			return &archivedGame{gameCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- archivedGame Handle ---

func (g *archivedGame) Handle(_ event.Event) (Game, error) {
	return nil, ErrUnexpectedEvent
}

// --- copyCore ---

func (c *gameCore) copyCore() gameCore {
	ids := make([]identity.CampaignID, len(c.campaignIDs))
	copy(ids, c.campaignIDs)
	return gameCore{
		id:                  c.id,
		gmID:                c.gmID,
		name:                c.name,
		masterNarrativeID:   c.masterNarrativeID,
		campaignIDs:         ids,
		activeCampaignCount: c.activeCampaignCount,
	}
}

// --- ReplayGame ---

// ReplayGame rebuilds a Game aggregate from a sequence of events.
// The first event must be an EntityCreated for a game.
// gmID must be provided by the caller — it is not carried in events (ADR-020).
func ReplayGame(gmID string, events []event.Event) (Game, error) {
	if len(events) == 0 {
		return nil, errors.New("game: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("game: first event must be EntityCreated")
	}

	gameID, err := identity.ParseGameID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("game replay: %w", err)
	}

	g, _, err := CreateGame(gameID, gmID, first.Name, event.SourceGrimoire)
	if err != nil {
		return nil, fmt.Errorf("game replay: %w", err)
	}

	var current Game = g
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("game replay: %w", err)
		}
	}
	return current, nil
}
