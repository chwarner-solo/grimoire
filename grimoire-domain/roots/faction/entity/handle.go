package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- copyCore ---

func (c *factionCore) copyCore() factionCore {
	memberIDs := make([]identity.FactionMembershipID, len(c.memberIDs))
	copy(memberIDs, c.memberIDs)
	alliedWithIDs := make([]identity.FactionID, len(c.alliedWithIDs))
	copy(alliedWithIDs, c.alliedWithIDs)
	atWarWithIDs := make([]identity.FactionID, len(c.atWarWithIDs))
	copy(atWarWithIDs, c.atWarWithIDs)
	standingLevels := make([]value.StandingLevel, len(c.standingLevels))
	copy(standingLevels, c.standingLevels)
	return factionCore{
		id:             c.id,
		gameID:         c.gameID,
		name:           c.name,
		description:    c.description,
		goals:          c.goals,
		playerVisible:  c.playerVisible,
		memberIDs:      memberIDs,
		alliedWithIDs:  alliedWithIDs,
		atWarWithIDs:   atWarWithIDs,
		standingLevels: standingLevels,
	}
}

// --- newFaction Handle ---

func (f *newFaction) Handle(evt event.Event) (Faction, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "draft" {
			core := f.copyCore()
			return &draftFaction{factionCore: core}, nil
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- draftFaction Handle ---

func (f *draftFaction) Handle(evt event.Event) (Faction, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		core := f.copyCore()
		switch e.Relationship {
		case "member_of":
			memberID, err := identity.ParseFactionMembershipID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.memberIDs = append(core.memberIDs, memberID)
			return &draftFaction{factionCore: core}, nil
		case "allied_with":
			allyID, err := identity.ParseFactionID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.alliedWithIDs = append(core.alliedWithIDs, allyID)
			return &draftFaction{factionCore: core}, nil
		case "at_war_with":
			enemyID, err := identity.ParseFactionID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.atWarWithIDs = append(core.atWarWithIDs, enemyID)
			return &draftFaction{factionCore: core}, nil
		default:
			return nil, ErrUnexpectedEvent
		}
	case event.EntityUpdated:
		core := f.copyCore()
		switch {
		case e.Field == "standing_levels":
			// During replay, standing level details are not preserved in the event.
			// We create a minimal StandingLevel with the name from the event.
			sl, err := value.NewStandingLevel(e.NewValue, 0, len(core.standingLevels))
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.standingLevels = append(core.standingLevels, sl)
			return &draftFaction{factionCore: core}, nil
		case e.Field == "status" && e.NewValue == "active":
			return &activeFaction{factionCore: core}, nil
		default:
			return nil, ErrUnexpectedEvent
		}
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- activeFaction Handle ---

func (f *activeFaction) Handle(evt event.Event) (Faction, error) {
	switch e := evt.(type) {
	case event.EntityLinked:
		core := f.copyCore()
		switch e.Relationship {
		case "member_of":
			memberID, err := identity.ParseFactionMembershipID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.memberIDs = append(core.memberIDs, memberID)
			return &activeFaction{factionCore: core}, nil
		case "allied_with":
			allyID, err := identity.ParseFactionID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.alliedWithIDs = append(core.alliedWithIDs, allyID)
			return &activeFaction{factionCore: core}, nil
		case "at_war_with":
			enemyID, err := identity.ParseFactionID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("faction handle: %w", err)
			}
			core.atWarWithIDs = append(core.atWarWithIDs, enemyID)
			return &activeFaction{factionCore: core}, nil
		default:
			return nil, ErrUnexpectedEvent
		}
	case event.EntityUpdated:
		if e.Field == "status" {
			core := f.copyCore()
			switch e.NewValue {
			case "idle":
				return &idleFaction{factionCore: core}, nil
			case "archived":
				return &archivedFaction{factionCore: core}, nil
			default:
				return nil, ErrUnexpectedEvent
			}
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- idleFaction Handle ---

func (f *idleFaction) Handle(evt event.Event) (Faction, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" {
			core := f.copyCore()
			switch e.NewValue {
			case "active":
				return &activeFaction{factionCore: core}, nil
			case "archived":
				return &archivedFaction{factionCore: core}, nil
			default:
				return nil, ErrUnexpectedEvent
			}
		}
		return nil, ErrUnexpectedEvent
	default:
		return nil, ErrUnexpectedEvent
	}
}

// --- archivedFaction Handle ---

func (f *archivedFaction) Handle(_ event.Event) (Faction, error) {
	return nil, ErrUnexpectedEvent
}

// --- ReplayFaction ---

// ReplayFaction rebuilds a Faction aggregate from a sequence of events.
// The first event must be an EntityCreated for a faction.
func ReplayFaction(gameID identity.GameID, events []event.Event) (Faction, error) {
	if len(events) == 0 {
		return nil, errors.New("faction: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("faction: first event must be EntityCreated")
	}

	factionID, err := identity.ParseFactionID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("faction replay: %w", err)
	}

	f, _, err := CreateFaction(factionID, first.Name, gameID, event.SourceGrimoire)
	if err != nil {
		return nil, fmt.Errorf("faction replay: %w", err)
	}

	var current Faction = f
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("faction replay: %w", err)
		}
	}
	return current, nil
}
