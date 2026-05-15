package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// MacGuffin uses Snapshot/Reconstitute, not Handle/Replay.
// No state machine — possession and location are the only mutable state.
// Intentional — see ADR-019.
type MacGuffin struct {
	id                identity.MacGuffinID
	gameID            identity.GameID
	name              string
	description       string
	playerDescription string
	playerVisible     bool
	destroyed         bool
	npcPossessorID    identity.NarrativeCharacterID
	pcPossessorID     identity.PlayerCharacterID
	locationID        identity.LocationID
}

func (m *MacGuffin) validatePossessionInvariant() error {
	owned := 0
	if !m.npcPossessorID.IsZero() {
		owned++
	}
	if !m.pcPossessorID.IsZero() {
		owned++
	}
	if !m.locationID.IsZero() {
		owned++
	}
	if owned > 1 {
		return ErrMacGuffinPossessionConflict
	}
	return nil
}

// CreateMacGuffin constructs a new MacGuffin. Starts unowned.
func CreateMacGuffin(
	id identity.MacGuffinID,
	gameID identity.GameID,
	name string,
	source event.Source,
) (*MacGuffin, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrMGMacGuffinIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrMGGameIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrMacGuffinNameRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "macguffin",
		Name:       name,
		Source:     source,
	}
	return &MacGuffin{
		id:     id,
		gameID: gameID,
		name:   name,
	}, []event.Event{evt}, nil
}

// --- Methods ---

func (m *MacGuffin) PlaceAt(locationID identity.LocationID, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	if locationID.IsZero() {
		return nil, nil, ErrMGLocationIDRequired
	}
	m.npcPossessorID = identity.NarrativeCharacterID{}
	m.pcPossessorID = identity.PlayerCharacterID{}
	m.locationID = locationID
	evt := event.EntityLinked{
		EntityAID:    m.id.String(),
		EntityBID:    locationID.String(),
		Relationship: "located_at",
		Source:       source,
	}
	return m, []event.Event{evt}, nil
}

func (m *MacGuffin) TransferToNPC(id identity.NarrativeCharacterID, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	if id.IsZero() {
		return nil, nil, ErrMGCharacterIDRequired
	}
	m.locationID = identity.LocationID{}
	m.pcPossessorID = identity.PlayerCharacterID{}
	m.npcPossessorID = id
	evt := event.EntityLinked{
		EntityAID:    id.String(),
		EntityBID:    m.id.String(),
		Relationship: "possesses",
		Source:       source,
	}
	return m, []event.Event{evt}, nil
}

func (m *MacGuffin) TransferToPC(id identity.PlayerCharacterID, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	if id.IsZero() {
		return nil, nil, ErrMGCharacterIDRequired
	}
	m.locationID = identity.LocationID{}
	m.npcPossessorID = identity.NarrativeCharacterID{}
	m.pcPossessorID = id
	evt := event.EntityLinked{
		EntityAID:    id.String(),
		EntityBID:    m.id.String(),
		Relationship: "possesses",
		Source:       source,
	}
	return m, []event.Event{evt}, nil
}

func (m *MacGuffin) Drop(locationID identity.LocationID, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	oldPossessor := ""
	if !m.npcPossessorID.IsZero() {
		oldPossessor = m.npcPossessorID.String()
	} else if !m.pcPossessorID.IsZero() {
		oldPossessor = m.pcPossessorID.String()
	}
	m.npcPossessorID = identity.NarrativeCharacterID{}
	m.pcPossessorID = identity.PlayerCharacterID{}
	m.locationID = locationID
	evts := []event.Event{
		event.EntityUpdated{
			EntityID: m.id.String(),
			Field:    "possessor",
			OldValue: oldPossessor,
			NewValue: "",
			Source:   source,
		},
		event.EntityLinked{
			EntityAID:    m.id.String(),
			EntityBID:    locationID.String(),
			Relationship: "located_at",
			Source:       source,
		},
	}
	return m, evts, nil
}

func (m *MacGuffin) Reveal(to []string, sessionID identity.SessionID, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	m.playerVisible = true
	evt := event.EntityRevealed{
		EntityID:   m.id.String(),
		RevealedTo: to,
		SessionID:  sessionID.String(),
		Source:     source,
	}
	return m, []event.Event{evt}, nil
}

func (m *MacGuffin) UpdateContent(name, description, playerDescription string, source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinDestroyed
	}
	m.name = name
	m.description = description
	m.playerDescription = playerDescription
	evt := event.EntityUpdated{
		EntityID: m.id.String(),
		Field:    "content",
		OldValue: "",
		NewValue: name,
		Source:   source,
	}
	return m, []event.Event{evt}, nil
}

func (m *MacGuffin) Destroy(source event.Source) (*MacGuffin, []event.Event, error) {
	if m.destroyed {
		return nil, nil, ErrMacGuffinAlreadyDestroyed
	}
	m.destroyed = true
	m.npcPossessorID = identity.NarrativeCharacterID{}
	m.pcPossessorID = identity.PlayerCharacterID{}
	m.locationID = identity.LocationID{}
	evt := event.EntityUpdated{
		EntityID: m.id.String(),
		Field:    "status",
		OldValue: "",
		NewValue: "destroyed",
		Source:   source,
	}
	return m, []event.Event{evt}, nil
}

// --- Getters ---

func (m *MacGuffin) MacGuffinID() identity.MacGuffinID                   { return m.id }
func (m *MacGuffin) MGGameID() identity.GameID                           { return m.gameID }
func (m *MacGuffin) MGName() string                                      { return m.name }
func (m *MacGuffin) MGDescription() string                               { return m.description }
func (m *MacGuffin) MGPlayerDescription() string                         { return m.playerDescription }
func (m *MacGuffin) IsPlayerVisible() bool                               { return m.playerVisible }
func (m *MacGuffin) IsDestroyed() bool                                   { return m.destroyed }
func (m *MacGuffin) NPCPossessorID() identity.NarrativeCharacterID       { return m.npcPossessorID }
func (m *MacGuffin) PCPossessorID() identity.PlayerCharacterID           { return m.pcPossessorID }
func (m *MacGuffin) MGLocationID() identity.LocationID                   { return m.locationID }
