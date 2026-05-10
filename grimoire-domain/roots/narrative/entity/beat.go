package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Beat is a shared entity owned by the Narrative subsystem.
// Content (name, description, playerDescription) lives in Neo4j only for reads.
// The command side (Firestore) holds structural fields: IDs, type, scope, prerequisite sets.
// Events carry all fields; GCS preserves full content for Neo4j rebuilds.
type Beat struct {
	id                identity.BeatID
	name              string
	description       string
	playerDescription string
	scope             value.BeatScope
	beatType          value.BeatType
	status            value.BeatStatus
	gameID            identity.GameID
	campaignID        identity.CampaignID
	prerequisiteSets  [][]identity.BeatID // OR-of-ANDs: any complete set satisfies
}

// CreateMasterBeat constructs a new Beat with master scope.
func CreateMasterBeat(id identity.BeatID, name string, beatType value.BeatType, gameID identity.GameID, source event.Source) (*Beat, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrBeatIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrBeatNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	beat := &Beat{
		id:       id,
		name:     name,
		scope:    value.BeatScopeMaster,
		beatType: beatType,
		status:   value.BeatStatusDraft,
		gameID:   gameID,
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "beat",
		Name:       name,
		Source:     source,
	}
	return beat, []event.Event{evt}, nil
}

// CreateCampaignBeat constructs a new Beat with campaign scope.
func CreateCampaignBeat(id identity.BeatID, name string, gameID identity.GameID, campaignID identity.CampaignID, source event.Source) (*Beat, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrBeatIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrBeatNameRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	if campaignID.IsZero() {
		return nil, nil, ErrCampaignIDRequired
	}
	beat := &Beat{
		id:         id,
		name:       name,
		scope:      value.BeatScopeCampaign,
		beatType:   value.BeatTypeCampaignSpecific,
		status:     value.BeatStatusDraft,
		gameID:     gameID,
		campaignID: campaignID,
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "beat",
		Name:       name,
		Source:     source,
	}
	return beat, []event.Event{evt}, nil
}

// UpdateContent sets the beat's content fields. All three are required.
func (b *Beat) UpdateContent(name, description, playerDescription string) ([]event.Event, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrBeatNameRequired
	}
	if strings.TrimSpace(description) == "" {
		return nil, ErrBeatDescriptionRequired
	}
	if strings.TrimSpace(playerDescription) == "" {
		return nil, ErrBeatPlayerDescriptionRequired
	}
	b.name = name
	b.description = description
	b.playerDescription = playerDescription
	evt := event.EntityUpdated{
		EntityID: b.id.String(),
		Field:    "content",
		OldValue: "",
		NewValue: name,
	}
	return []event.Event{evt}, nil
}

// WouldCreateCycle checks whether adding targetID as a prerequisite of this beat
// would create a cycle. chain is the prerequisite chain of targetID (all beats
// reachable from targetID by following prerequisites).
func (b *Beat) WouldCreateCycle(targetID identity.BeatID, chain []*Beat) error {
	for _, ancestor := range chain {
		for _, set := range ancestor.prerequisiteSets {
			for _, prereqID := range set {
				if prereqID.String() == b.id.String() {
					return ErrCycleDetected{BeatID: b.id}
				}
			}
		}
	}
	return nil
}

// AddPrerequisite appends a prerequisite beat ID to the first prerequisite set.
// If no sets exist, a new set is created.
func (b *Beat) AddPrerequisite(prerequisiteID identity.BeatID) {
	if len(b.prerequisiteSets) == 0 {
		b.prerequisiteSets = [][]identity.BeatID{{prerequisiteID}}
		return
	}
	b.prerequisiteSets[0] = append(b.prerequisiteSets[0], prerequisiteID)
}

// --- Getters ---

func (b *Beat) BeatID() identity.BeatID          { return b.id }
func (b *Beat) Name() string                      { return b.name }
func (b *Beat) Description() string               { return b.description }
func (b *Beat) PlayerDescription() string          { return b.playerDescription }
func (b *Beat) Scope() value.BeatScope            { return b.scope }
func (b *Beat) BeatType() value.BeatType          { return b.beatType }
func (b *Beat) Status() value.BeatStatus          { return b.status }
func (b *Beat) GameID() identity.GameID           { return b.gameID }
func (b *Beat) CampaignID() identity.CampaignID   { return b.campaignID }
func (b *Beat) PrerequisiteSets() [][]identity.BeatID {
	result := make([][]identity.BeatID, len(b.prerequisiteSets))
	for i, set := range b.prerequisiteSets {
		cp := make([]identity.BeatID, len(set))
		copy(cp, set)
		result[i] = cp
	}
	return result
}

// --- Snapshot / Reconstitute ---

// BeatSnapshot is the DTO used for persistence and reconstitution.
type BeatSnapshot struct {
	ID                identity.BeatID
	Name              string
	Description       string
	PlayerDescription string
	Scope             value.BeatScope
	Type              value.BeatType
	Status            value.BeatStatus
	GameID            identity.GameID
	CampaignID        identity.CampaignID
	PrerequisiteSets  [][]identity.BeatID
}

func (b *Beat) Snapshot() BeatSnapshot {
	return BeatSnapshot{
		ID:                b.id,
		Name:              b.name,
		Description:       b.description,
		PlayerDescription: b.playerDescription,
		Scope:             b.scope,
		Type:              b.beatType,
		Status:            b.status,
		GameID:            b.gameID,
		CampaignID:        b.campaignID,
		PrerequisiteSets:  b.PrerequisiteSets(),
	}
}

// ReconstituteBeat rebuilds a Beat from a persisted snapshot.
func ReconstituteBeat(snap BeatSnapshot) *Beat {
	sets := make([][]identity.BeatID, len(snap.PrerequisiteSets))
	for i, set := range snap.PrerequisiteSets {
		cp := make([]identity.BeatID, len(set))
		copy(cp, set)
		sets[i] = cp
	}
	return &Beat{
		id:                snap.ID,
		name:              snap.Name,
		description:       snap.Description,
		playerDescription: snap.PlayerDescription,
		scope:             snap.Scope,
		beatType:          snap.Type,
		status:            snap.Status,
		gameID:            snap.GameID,
		campaignID:        snap.CampaignID,
		prerequisiteSets:  sets,
	}
}
