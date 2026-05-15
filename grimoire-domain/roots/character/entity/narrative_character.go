package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// NPC is the sealed interface for the NPC aggregate root.
// The unexported isNPC() marker prevents external implementations.
// Handle enables event replay — required for Bigtable scaling path (ADR-014).
type NPC interface {
	isNPC()
	NarrativeCharacterID() identity.NarrativeCharacterID
	NPCName() string
	GameID() identity.GameID
	Snapshot() NPCSnapshot
	Handle(evt event.Event) (NPC, error)
}

// NewNPC represents an NPC that has just been created.
type NewNPC interface {
	NPC
	BeginDraft(source event.Source) (DraftNPC, []event.Event, error)
}

// DraftNPC represents an NPC being built by the GM.
type DraftNPC interface {
	NPC
	UpdateContent(name, description, playerDescription string, source event.Source) (DraftNPC, []event.Event, error)
	AssignToLocation(id identity.LocationID, source event.Source) (DraftNPC, []event.Event, error)
	Activate(source event.Source) (ActiveNPC, []event.Event, error)
}

// ActiveNPC represents an NPC operating in the world.
type ActiveNPC interface {
	NPC
	UpdateContent(name, description, playerDescription string, source event.Source) (ActiveNPC, []event.Event, error)
	MoveTo(id identity.LocationID, source event.Source) (ActiveNPC, []event.Event, error)
	GoIdle(source event.Source) (IdleNPC, []event.Event, error)
	Archive(source event.Source) (ArchivedNPC, []event.Event, error)
	Reveal(to []string, sessionID identity.SessionID, source event.Source) (ActiveNPC, []event.Event, error)
	AssignMacGuffin(id identity.MacGuffinID, source event.Source) (ActiveNPC, []event.Event, error)
}

// IdleNPC represents a dormant NPC.
type IdleNPC interface {
	NPC
	Activate(source event.Source) (ActiveNPC, []event.Event, error)
	Archive(source event.Source) (ArchivedNPC, []event.Event, error)
}

// ArchivedNPC is the terminal state. No transitions are possible.
type ArchivedNPC interface {
	NPC
}

// --- npcCore ---

type npcCore struct {
	id                identity.NarrativeCharacterID
	gameID            identity.GameID
	name              string
	description       string
	playerDescription string
	playerVisible     bool
	locationID        identity.LocationID
	macGuffinIDs      []identity.MacGuffinID
	foundryCharacterID string
}

func (n *npcCore) NarrativeCharacterID() identity.NarrativeCharacterID { return n.id }
func (n *npcCore) NPCName() string                                     { return n.name }
func (n *npcCore) GameID() identity.GameID                             { return n.gameID }
func (n *npcCore) isNPC()                                              {}

func (n *npcCore) hasMacGuffin(id identity.MacGuffinID) bool {
	for _, mid := range n.macGuffinIDs {
		if mid.String() == id.String() {
			return true
		}
	}
	return false
}

func (n *npcCore) snapshot(state string) NPCSnapshot {
	macGuffinIDs := make([]identity.MacGuffinID, len(n.macGuffinIDs))
	copy(macGuffinIDs, n.macGuffinIDs)
	return NPCSnapshot{
		ID:                 n.id,
		GameID:             n.gameID,
		State:              state,
		Name:               n.name,
		Description:        n.description,
		PlayerDescription:  n.playerDescription,
		PlayerVisible:      n.playerVisible,
		LocationID:         n.locationID,
		MacGuffinIDs:       macGuffinIDs,
		FoundryCharacterID: n.foundryCharacterID,
	}
}

// CreateNPC constructs a new NPC aggregate in the New state.
func CreateNPC(
	id identity.NarrativeCharacterID,
	name string,
	gameID identity.GameID,
	source event.Source,
) (NewNPC, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrNPCIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrNPCNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrNPCGameIDRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "npc",
		Name:       name,
		Source:     source,
	}
	return &newNPC{npcCore: npcCore{
		id:     id,
		name:   name,
		gameID: gameID,
	}}, []event.Event{evt}, nil
}

// --- State structs ---

type newNPC struct{ npcCore }
type draftNPC struct{ npcCore }
type activeNPC struct{ npcCore }
type idleNPC struct{ npcCore }
type archivedNPC struct{ npcCore }

// Snapshot implementations
func (n *newNPC) Snapshot() NPCSnapshot      { return n.snapshot("new") }
func (n *draftNPC) Snapshot() NPCSnapshot    { return n.snapshot("draft") }
func (n *activeNPC) Snapshot() NPCSnapshot   { return n.snapshot("active") }
func (n *idleNPC) Snapshot() NPCSnapshot     { return n.snapshot("idle") }
func (n *archivedNPC) Snapshot() NPCSnapshot { return n.snapshot("archived") }

// --- NewNPC transitions ---

func (n *newNPC) BeginDraft(source event.Source) (DraftNPC, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "draft",
		Source:   source,
	}
	return &draftNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}

// --- DraftNPC transitions ---

func (n *draftNPC) UpdateContent(name, description, playerDescription string, source event.Source) (DraftNPC, []event.Event, error) {
	n.name = name
	n.description = description
	n.playerDescription = playerDescription
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "content",
		OldValue: "",
		NewValue: name,
		Source:   source,
	}
	return n, []event.Event{evt}, nil
}

func (n *draftNPC) AssignToLocation(id identity.LocationID, source event.Source) (DraftNPC, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	n.locationID = id
	evt := event.EntityLinked{
		EntityAID:    n.id.String(),
		EntityBID:    id.String(),
		Relationship: "located_at",
		Source:       source,
	}
	return n, []event.Event{evt}, nil
}

func (n *draftNPC) Activate(source event.Source) (ActiveNPC, []event.Event, error) {
	if strings.TrimSpace(n.name) == "" {
		return nil, nil, ErrNPCNameRequired
	}
	if strings.TrimSpace(n.description) == "" {
		return nil, nil, ErrNPCDescriptionRequired
	}
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "draft",
		NewValue: "active",
		Source:   source,
	}
	return &activeNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}

// --- ActiveNPC transitions ---

func (n *activeNPC) UpdateContent(name, description, playerDescription string, source event.Source) (ActiveNPC, []event.Event, error) {
	n.name = name
	n.description = description
	n.playerDescription = playerDescription
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "content",
		OldValue: "",
		NewValue: name,
		Source:   source,
	}
	return n, []event.Event{evt}, nil
}

func (n *activeNPC) MoveTo(id identity.LocationID, source event.Source) (ActiveNPC, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	oldValue := ""
	if !n.locationID.IsZero() {
		oldValue = n.locationID.String()
	}
	n.locationID = id
	evts := []event.Event{
		event.EntityUpdated{
			EntityID: n.id.String(),
			Field:    "location",
			OldValue: oldValue,
			NewValue: id.String(),
			Source:   source,
		},
		event.EntityLinked{
			EntityAID:    n.id.String(),
			EntityBID:    id.String(),
			Relationship: "located_at",
			Source:       source,
		},
	}
	return n, evts, nil
}

func (n *activeNPC) GoIdle(source event.Source) (IdleNPC, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "idle",
		Source:   source,
	}
	return &idleNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}

func (n *activeNPC) Archive(source event.Source) (ArchivedNPC, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}

func (n *activeNPC) Reveal(to []string, sessionID identity.SessionID, source event.Source) (ActiveNPC, []event.Event, error) {
	if sessionID.IsZero() {
		return nil, nil, ErrSessionIDRequired
	}
	n.playerVisible = true
	evt := event.EntityRevealed{
		EntityID:   n.id.String(),
		RevealedTo: to,
		SessionID:  sessionID.String(),
		Source:     source,
	}
	return n, []event.Event{evt}, nil
}

func (n *activeNPC) AssignMacGuffin(id identity.MacGuffinID, source event.Source) (ActiveNPC, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrMacGuffinIDRequired
	}
	if n.hasMacGuffin(id) {
		return nil, nil, ErrMacGuffinAlreadyAssigned
	}
	n.macGuffinIDs = append(n.macGuffinIDs, id)
	evt := event.EntityLinked{
		EntityAID:    n.id.String(),
		EntityBID:    id.String(),
		Relationship: "possesses",
		Source:       source,
	}
	return n, []event.Event{evt}, nil
}

// --- IdleNPC transitions ---

func (n *idleNPC) Activate(source event.Source) (ActiveNPC, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "active",
		Source:   source,
	}
	return &activeNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}

func (n *idleNPC) Archive(source event.Source) (ArchivedNPC, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: n.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedNPC{npcCore: n.npcCore}, []event.Event{evt}, nil
}
