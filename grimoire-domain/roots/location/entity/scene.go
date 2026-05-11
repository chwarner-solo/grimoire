package entity

import (
	"strings"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Scene is a child entity of Location. It has no state machine — status
// is per-campaign in Neo4j. Scene is purely structural on the command side.
type Scene struct {
	id                   identity.SceneID
	locationID           identity.LocationID
	gameID               identity.GameID
	scenePrerequisiteSets [][]identity.SceneID
	recommendedBeatIDs   []identity.BeatID
	playerVisible        bool
	name                 string
	description          string
	playerDescription    string
	exploredDescription  string
}

// CreateScene constructs a new Scene child entity.
func CreateScene(id identity.SceneID, locationID identity.LocationID, gameID identity.GameID, name string, source event.Source) (*Scene, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrSceneIDRequired
	}
	if locationID.IsZero() {
		return nil, nil, ErrLocationIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	if strings.TrimSpace(name) == "" {
		return nil, nil, ErrSceneNameRequired
	}
	s := &Scene{
		id:            id,
		locationID:    locationID,
		gameID:        gameID,
		name:          name,
		playerVisible: false,
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "scene",
		Name:       name,
		Source:     source,
	}
	return s, []event.Event{evt}, nil
}

// UpdateContent updates the scene's content fields and emits events for each changed field.
func (s *Scene) UpdateContent(name, description, playerDescription, exploredDescription string) ([]event.Event, error) {
	var events []event.Event
	if name != s.name {
		events = append(events, event.EntityUpdated{
			EntityID: s.id.String(),
			Field:    "name",
			OldValue: s.name,
			NewValue: name,
		})
		s.name = name
	}
	if description != s.description {
		events = append(events, event.EntityUpdated{
			EntityID: s.id.String(),
			Field:    "description",
			OldValue: s.description,
			NewValue: description,
		})
		s.description = description
	}
	if playerDescription != s.playerDescription {
		events = append(events, event.EntityUpdated{
			EntityID: s.id.String(),
			Field:    "player_description",
			OldValue: s.playerDescription,
			NewValue: playerDescription,
		})
		s.playerDescription = playerDescription
	}
	if exploredDescription != s.exploredDescription {
		events = append(events, event.EntityUpdated{
			EntityID: s.id.String(),
			Field:    "explored_description",
			OldValue: s.exploredDescription,
			NewValue: exploredDescription,
		})
		s.exploredDescription = exploredDescription
	}
	return events, nil
}

// AddPrerequisite appends a scene ID to the first prerequisite set.
func (s *Scene) AddPrerequisite(sceneID identity.SceneID) {
	if len(s.scenePrerequisiteSets) == 0 {
		s.scenePrerequisiteSets = [][]identity.SceneID{{sceneID}}
	} else {
		s.scenePrerequisiteSets[0] = append(s.scenePrerequisiteSets[0], sceneID)
	}
}

// AddRecommendedBeat appends a beat ID to recommendedBeatIDs.
func (s *Scene) AddRecommendedBeat(beatID identity.BeatID) {
	s.recommendedBeatIDs = append(s.recommendedBeatIDs, beatID)
}

// WouldCreateCycle checks if adding targetID as a prerequisite would create a cycle.
// chain is the set of scene IDs already visited in this traversal.
func (s *Scene) WouldCreateCycle(targetID identity.SceneID, chain []identity.SceneID) error {
	for _, c := range chain {
		if c.String() == s.id.String() {
			return ErrSceneCycleDetected{SceneID: targetID}
		}
	}
	return nil
}

// --- Getters ---

func (s *Scene) SceneID() identity.SceneID       { return s.id }
func (s *Scene) LocationID() identity.LocationID  { return s.locationID }
func (s *Scene) GameID() identity.GameID          { return s.gameID }
func (s *Scene) Name() string                     { return s.name }
func (s *Scene) Description() string              { return s.description }
func (s *Scene) PlayerDescription() string        { return s.playerDescription }
func (s *Scene) ExploredDescription() string      { return s.exploredDescription }
func (s *Scene) IsPlayerVisible() bool            { return s.playerVisible }
func (s *Scene) RecommendedBeatIDs() []identity.BeatID {
	ids := make([]identity.BeatID, len(s.recommendedBeatIDs))
	copy(ids, s.recommendedBeatIDs)
	return ids
}

func (s *Scene) PrerequisiteSets() [][]identity.SceneID {
	sets := make([][]identity.SceneID, len(s.scenePrerequisiteSets))
	for i, set := range s.scenePrerequisiteSets {
		cp := make([]identity.SceneID, len(set))
		copy(cp, set)
		sets[i] = cp
	}
	return sets
}

// --- Snapshot / Reconstitute ---

// SceneSnapshot is the DTO used for persistence.
type SceneSnapshot struct {
	ID                    identity.SceneID
	LocationID            identity.LocationID
	GameID                identity.GameID
	ScenePrerequisiteSets [][]identity.SceneID
	RecommendedBeatIDs    []identity.BeatID
	PlayerVisible         bool
	Name                  string
	Description           string
	PlayerDescription     string
	ExploredDescription   string
}

func (s *Scene) Snapshot() SceneSnapshot {
	return SceneSnapshot{
		ID:                    s.id,
		LocationID:            s.locationID,
		GameID:                s.gameID,
		ScenePrerequisiteSets: s.PrerequisiteSets(),
		RecommendedBeatIDs:    s.RecommendedBeatIDs(),
		PlayerVisible:         s.playerVisible,
		Name:                  s.name,
		Description:           s.description,
		PlayerDescription:     s.playerDescription,
		ExploredDescription:   s.exploredDescription,
	}
}

func ReconstituteScene(snap SceneSnapshot) *Scene {
	sets := make([][]identity.SceneID, len(snap.ScenePrerequisiteSets))
	for i, set := range snap.ScenePrerequisiteSets {
		cp := make([]identity.SceneID, len(set))
		copy(cp, set)
		sets[i] = cp
	}
	beatIDs := make([]identity.BeatID, len(snap.RecommendedBeatIDs))
	copy(beatIDs, snap.RecommendedBeatIDs)
	return &Scene{
		id:                    snap.ID,
		locationID:            snap.LocationID,
		gameID:                snap.GameID,
		scenePrerequisiteSets: sets,
		recommendedBeatIDs:    beatIDs,
		playerVisible:         snap.PlayerVisible,
		name:                  snap.Name,
		description:           snap.Description,
		playerDescription:     snap.PlayerDescription,
		exploredDescription:   snap.ExploredDescription,
	}
}
