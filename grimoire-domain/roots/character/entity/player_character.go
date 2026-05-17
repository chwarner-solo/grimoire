package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// PlayerCharacterStatus tracks the lifecycle of a player character.
type PlayerCharacterStatus string

const (
	PlayerCharacterStatusActive  PlayerCharacterStatus = "active"
	PlayerCharacterStatusRetired PlayerCharacterStatus = "retired"
)

// PlayerCharacter uses Snapshot/Reconstitute, not the Handle/Replay pattern
// of Game/Campaign/Faction/Location/NPC. The Active → Retired lifecycle
// does not require event replay for reconstruction. Intentional — see ADR-019.
type PlayerCharacter struct {
	id                 identity.PlayerCharacterID
	gameID             identity.GameID
	name               string
	description        string
	playerDescription  string
	playerVisible      bool
	ownerPlayerID      string
	status             PlayerCharacterStatus
	campaignIDs        []identity.CampaignID
	macGuffinIDs       []identity.MacGuffinID
	foundryCharacterID string
}

// CreatePlayerCharacter constructs a new PlayerCharacter in the Active state.
func CreatePlayerCharacter(
	id identity.PlayerCharacterID,
	gameID identity.GameID,
	name string,
	ownerPlayerID string,
	source event.Source,
) (*PlayerCharacter, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrPlayerCharacterIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrPCGameIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrCharacterNameRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "player_character",
		Name:       name,
		Source:     source,
	}
	return &PlayerCharacter{
		id:            id,
		gameID:        gameID,
		name:          name,
		ownerPlayerID: ownerPlayerID,
		status:        PlayerCharacterStatusActive,
	}, []event.Event{evt}, nil
}

// --- Methods ---

func (pc *PlayerCharacter) Retire(source event.Source) (*PlayerCharacter, []event.Event, error) {
	if pc.status == PlayerCharacterStatusRetired {
		return nil, nil, ErrCharacterAlreadyRetired
	}
	pc.status = PlayerCharacterStatusRetired
	evt := event.EntityUpdated{
		EntityID: pc.id.String(),
		Field:    "status",
		OldValue: string(PlayerCharacterStatusActive),
		NewValue: string(PlayerCharacterStatusRetired),
		Source:   source,
	}
	return pc, []event.Event{evt}, nil
}

func (pc *PlayerCharacter) UpdateContent(name, description, playerDescription string, source event.Source) (*PlayerCharacter, []event.Event, error) {
	pc.name = name
	pc.description = description
	pc.playerDescription = playerDescription
	evt := event.EntityUpdated{
		EntityID: pc.id.String(),
		Field:    "content",
		OldValue: "",
		NewValue: name,
		Source:   source,
	}
	return pc, []event.Event{evt}, nil
}

func (pc *PlayerCharacter) AssignMacGuffin(id identity.MacGuffinID, source event.Source) (*PlayerCharacter, []event.Event, error) {
	if pc.status == PlayerCharacterStatusRetired {
		return nil, nil, ErrCharacterRetired
	}
	pc.macGuffinIDs = append(pc.macGuffinIDs, id)
	evt := event.EntityLinked{
		EntityAID:    pc.id.String(),
		EntityBID:    id.String(),
		Relationship: "possesses",
		Source:       source,
	}
	return pc, []event.Event{evt}, nil
}

func (pc *PlayerCharacter) ReleaseMacGuffin(id identity.MacGuffinID, source event.Source) (*PlayerCharacter, []event.Event, error) {
	idx := -1
	for i, mid := range pc.macGuffinIDs {
		if mid.String() == id.String() {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, nil, ErrMacGuffinNotPossessed
	}
	pc.macGuffinIDs = append(pc.macGuffinIDs[:idx], pc.macGuffinIDs[idx+1:]...)
	evt := event.EntityUpdated{
		EntityID: pc.id.String(),
		Field:    "macguffin_released",
		OldValue: id.String(),
		NewValue: "",
		Source:   source,
	}
	return pc, []event.Event{evt}, nil
}

func (pc *PlayerCharacter) JoinCampaign(id identity.CampaignID, source event.Source) (*PlayerCharacter, []event.Event, error) {
	for _, cid := range pc.campaignIDs {
		if cid.String() == id.String() {
			return nil, nil, ErrAlreadyInCampaign
		}
	}
	pc.campaignIDs = append(pc.campaignIDs, id)
	evt := event.EntityLinked{
		EntityAID:    pc.id.String(),
		EntityBID:    id.String(),
		Relationship: "participates_in",
		Source:       source,
	}
	return pc, []event.Event{evt}, nil
}

func (pc *PlayerCharacter) Reveal(to []string, sessionID identity.SessionID, source event.Source) (*PlayerCharacter, []event.Event, error) {
	pc.playerVisible = true
	evt := event.EntityRevealed{
		EntityID:   pc.id.String(),
		RevealedTo: to,
		SessionID:  sessionID.String(),
		Source:     source,
	}
	return pc, []event.Event{evt}, nil
}

// --- Getters ---

func (pc *PlayerCharacter) PlayerCharacterID() identity.PlayerCharacterID { return pc.id }
func (pc *PlayerCharacter) PCGameID() identity.GameID                     { return pc.gameID }
func (pc *PlayerCharacter) PCName() string                                { return pc.name }
func (pc *PlayerCharacter) Description() string                           { return pc.description }
func (pc *PlayerCharacter) PlayerDescription() string                     { return pc.playerDescription }
func (pc *PlayerCharacter) IsPlayerVisible() bool                         { return pc.playerVisible }
func (pc *PlayerCharacter) OwnerPlayerID() string                         { return pc.ownerPlayerID }
func (pc *PlayerCharacter) Status() PlayerCharacterStatus                 { return pc.status }
