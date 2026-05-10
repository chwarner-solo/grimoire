package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// factionCore holds the shared state for all Faction states.
type factionCore struct {
	id             identity.FactionID
	gameID         identity.GameID
	name           string
	description    string
	goals          string
	playerVisible  bool
	memberIDs      []identity.FactionMembershipID
	alliedWithIDs  []identity.FactionID
	atWarWithIDs   []identity.FactionID
	standingLevels []value.StandingLevel
}

func (c *factionCore) FactionID() identity.FactionID { return c.id }
func (c *factionCore) FactionName() string           { return c.name }
func (c *factionCore) GameID() identity.GameID       { return c.gameID }
func (c *factionCore) Description() string           { return c.description }
func (c *factionCore) Goals() string                 { return c.goals }
func (c *factionCore) IsPlayerVisible() bool         { return c.playerVisible }
func (c *factionCore) isFaction()                    {}

func (c *factionCore) hasMember(id identity.FactionMembershipID) bool {
	for _, mid := range c.memberIDs {
		if mid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *factionCore) isAlliedWith(id identity.FactionID) bool {
	for _, aid := range c.alliedWithIDs {
		if aid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *factionCore) isAtWarWith(id identity.FactionID) bool {
	for _, wid := range c.atWarWithIDs {
		if wid.String() == id.String() {
			return true
		}
	}
	return false
}

func (c *factionCore) snapshot(state string) FactionSnapshot {
	memberIDs := make([]identity.FactionMembershipID, len(c.memberIDs))
	copy(memberIDs, c.memberIDs)
	alliedWithIDs := make([]identity.FactionID, len(c.alliedWithIDs))
	copy(alliedWithIDs, c.alliedWithIDs)
	atWarWithIDs := make([]identity.FactionID, len(c.atWarWithIDs))
	copy(atWarWithIDs, c.atWarWithIDs)
	standingLevels := make([]value.StandingLevel, len(c.standingLevels))
	copy(standingLevels, c.standingLevels)
	return FactionSnapshot{
		ID:             c.id,
		GameID:         c.gameID,
		Name:           c.name,
		Description:    c.description,
		Goals:          c.goals,
		PlayerVisible:  c.playerVisible,
		State:          state,
		MemberIDs:      memberIDs,
		AlliedWithIDs:  alliedWithIDs,
		AtWarWithIDs:   atWarWithIDs,
		StandingLevels: standingLevels,
	}
}

// CreateFaction constructs a new Faction aggregate in the New state.
func CreateFaction(id identity.FactionID, name string, gameID identity.GameID, source event.Source) (NewFaction, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrFactionIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrFactionNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "faction",
		Name:       name,
		Source:     source,
	}
	return &newFaction{factionCore: factionCore{
		id:     id,
		gameID: gameID,
		name:   name,
	}}, []event.Event{evt}, nil
}

// --- State structs ---

type newFaction struct{ factionCore }
type draftFaction struct{ factionCore }
type activeFaction struct{ factionCore }
type idleFaction struct{ factionCore }
type archivedFaction struct{ factionCore }

// Snapshot implementations
func (f *newFaction) Snapshot() FactionSnapshot      { return f.snapshot("new") }
func (f *draftFaction) Snapshot() FactionSnapshot    { return f.snapshot("draft") }
func (f *activeFaction) Snapshot() FactionSnapshot   { return f.snapshot("active") }
func (f *idleFaction) Snapshot() FactionSnapshot     { return f.snapshot("idle") }
func (f *archivedFaction) Snapshot() FactionSnapshot { return f.snapshot("archived") }
