package entity

import (
	"errors"
	"fmt"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- copyCore ---

func (n *npcCore) copyCore() npcCore {
	macGuffinIDs := make([]identity.MacGuffinID, len(n.macGuffinIDs))
	copy(macGuffinIDs, n.macGuffinIDs)
	return npcCore{
		id:                 n.id,
		gameID:             n.gameID,
		name:               n.name,
		description:        n.description,
		playerDescription:  n.playerDescription,
		playerVisible:      n.playerVisible,
		locationID:         n.locationID,
		macGuffinIDs:       macGuffinIDs,
		foundryCharacterID: n.foundryCharacterID,
	}
}

// --- newNPC Handle ---

func (n *newNPC) Handle(evt event.Event) (NPC, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" && e.NewValue == "draft" {
			core := n.copyCore()
			return &draftNPC{npcCore: core}, nil
		}
		return nil, ErrNPCUnexpectedEvent
	default:
		return nil, ErrNPCUnexpectedEvent
	}
}

// --- draftNPC Handle ---

func (n *draftNPC) Handle(evt event.Event) (NPC, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		core := n.copyCore()
		switch {
		case e.Field == "content":
			core.name = e.NewValue
			return &draftNPC{npcCore: core}, nil
		case e.Field == "status" && e.NewValue == "active":
			return &activeNPC{npcCore: core}, nil
		default:
			return nil, ErrNPCUnexpectedEvent
		}
	case event.EntityLinked:
		if e.Relationship == "located_at" {
			core := n.copyCore()
			locID, err := identity.ParseLocationID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("npc handle: %w", err)
			}
			core.locationID = locID
			return &draftNPC{npcCore: core}, nil
		}
		return nil, ErrNPCUnexpectedEvent
	default:
		return nil, ErrNPCUnexpectedEvent
	}
}

// --- activeNPC Handle ---

func (n *activeNPC) Handle(evt event.Event) (NPC, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		core := n.copyCore()
		switch {
		case e.Field == "content":
			core.name = e.NewValue
			return &activeNPC{npcCore: core}, nil
		case e.Field == "location":
			if e.NewValue != "" {
				locID, err := identity.ParseLocationID(e.NewValue)
				if err != nil {
					return nil, fmt.Errorf("npc handle: %w", err)
				}
				core.locationID = locID
			}
			return &activeNPC{npcCore: core}, nil
		case e.Field == "status":
			switch e.NewValue {
			case "idle":
				return &idleNPC{npcCore: core}, nil
			case "archived":
				return &archivedNPC{npcCore: core}, nil
			default:
				return nil, ErrNPCUnexpectedEvent
			}
		default:
			return nil, ErrNPCUnexpectedEvent
		}
	case event.EntityLinked:
		core := n.copyCore()
		switch e.Relationship {
		case "located_at":
			locID, err := identity.ParseLocationID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("npc handle: %w", err)
			}
			core.locationID = locID
			return &activeNPC{npcCore: core}, nil
		case "possesses":
			mgID, err := identity.ParseMacGuffinID(e.EntityBID)
			if err != nil {
				return nil, fmt.Errorf("npc handle: %w", err)
			}
			core.macGuffinIDs = append(core.macGuffinIDs, mgID)
			return &activeNPC{npcCore: core}, nil
		default:
			return nil, ErrNPCUnexpectedEvent
		}
	case event.EntityRevealed:
		core := n.copyCore()
		core.playerVisible = true
		return &activeNPC{npcCore: core}, nil
	default:
		return nil, ErrNPCUnexpectedEvent
	}
}

// --- idleNPC Handle ---

func (n *idleNPC) Handle(evt event.Event) (NPC, error) {
	switch e := evt.(type) {
	case event.EntityUpdated:
		if e.Field == "status" {
			core := n.copyCore()
			switch e.NewValue {
			case "active":
				return &activeNPC{npcCore: core}, nil
			case "archived":
				return &archivedNPC{npcCore: core}, nil
			default:
				return nil, ErrNPCUnexpectedEvent
			}
		}
		return nil, ErrNPCUnexpectedEvent
	default:
		return nil, ErrNPCUnexpectedEvent
	}
}

// --- archivedNPC Handle ---

func (n *archivedNPC) Handle(_ event.Event) (NPC, error) {
	return nil, ErrNPCUnexpectedEvent
}

// --- ReplayNPC ---

// ReplayNPC rebuilds an NPC aggregate from a sequence of events.
// The first event must be an EntityCreated for an npc.
// gameID is required because it is not stored in the event payload.
func ReplayNPC(gameID identity.GameID, events []event.Event) (NPC, error) {
	if len(events) == 0 {
		return nil, errors.New("npc: no events to replay")
	}

	first, ok := events[0].(event.EntityCreated)
	if !ok {
		return nil, errors.New("npc: first event must be EntityCreated")
	}

	npcID, err := identity.ParseNarrativeCharacterID(first.EntityID)
	if err != nil {
		return nil, fmt.Errorf("npc replay: %w", err)
	}

	n, _, err := CreateNPC(npcID, first.Name, gameID, event.SourceGrimoire)
	if err != nil {
		return nil, fmt.Errorf("npc replay: %w", err)
	}

	var current NPC = n
	for _, evt := range events[1:] {
		current, err = current.Handle(evt)
		if err != nil {
			return nil, fmt.Errorf("npc replay: %w", err)
		}
	}
	return current, nil
}
