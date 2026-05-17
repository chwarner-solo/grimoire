package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// CampaignNarrative is an aggregate root representing one Campaign's path
// through the story DAG. It tracks which beats have been discovered.
//
// Intentional departure from the state-machine pattern used by Game and Campaign:
// CampaignNarrative is purely additive — beats are discovered but never
// un-discovered. There is no lifecycle state machine.
type CampaignNarrative struct {
	id                 identity.CampaignNarrativeID
	campaignID         identity.CampaignID
	gameID             identity.GameID
	discoveredBeatIDs  []identity.BeatID
	campaignBeatIDs    []identity.BeatID
	exploredSceneIDs   []identity.SceneID
	visitedLocationIDs []identity.LocationID
}

// CreateCampaignNarrative constructs a new CampaignNarrative aggregate.
func CreateCampaignNarrative(id identity.CampaignNarrativeID, campaignID identity.CampaignID, gameID identity.GameID, source event.Source) (*CampaignNarrative, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrCampaignNarrativeIDRequired
	}
	if campaignID.IsZero() {
		return nil, nil, ErrCampaignIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrGameIDRequired
	}
	cn := &CampaignNarrative{
		id:         id,
		campaignID: campaignID,
		gameID:     gameID,
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "campaign_narrative",
		Name:       "",
		Source:     source,
	}
	return cn, []event.Event{evt}, nil
}

// DiscoverBeat records that this campaign has discovered the given beat.
// Prerequisites are checked: at least one complete prerequisite set must be satisfied.
func (cn *CampaignNarrative) DiscoverBeat(beat *Beat) ([]event.Event, error) {
	if !cn.prerequisitesMet(beat) {
		return nil, ErrPrerequisiteNotMet{BeatID: beat.BeatID()}
	}
	cn.discoveredBeatIDs = append(cn.discoveredBeatIDs, beat.BeatID())
	evt := event.EntityLinked{
		EntityAID:    cn.id.String(),
		EntityBID:    beat.BeatID().String(),
		Relationship: "discovered",
	}
	return []event.Event{evt}, nil
}

// prerequisitesMet checks whether at least one complete prerequisite set is satisfied.
// If the beat has no prerequisites, it is always satisfiable.
func (cn *CampaignNarrative) prerequisitesMet(beat *Beat) bool {
	sets := beat.PrerequisiteSets()
	if len(sets) == 0 {
		return true
	}
	for _, set := range sets {
		if cn.allDiscovered(set) {
			return true
		}
	}
	return false
}

func (cn *CampaignNarrative) allDiscovered(beatIDs []identity.BeatID) bool {
	for _, id := range beatIDs {
		if !cn.hasDiscovered(id) {
			return false
		}
	}
	return true
}

func (cn *CampaignNarrative) hasDiscovered(id identity.BeatID) bool {
	for _, discovered := range cn.discoveredBeatIDs {
		if discovered.String() == id.String() {
			return true
		}
	}
	return false
}

// --- Getters ---

func (cn *CampaignNarrative) CampaignNarrativeID() identity.CampaignNarrativeID { return cn.id }
func (cn *CampaignNarrative) CampaignID() identity.CampaignID                   { return cn.campaignID }
func (cn *CampaignNarrative) GameID() identity.GameID                            { return cn.gameID }

// ExploredSceneIDs returns a copy of the explored scene IDs.
func (cn *CampaignNarrative) ExploredSceneIDs() []identity.SceneID {
	ids := make([]identity.SceneID, len(cn.exploredSceneIDs))
	copy(ids, cn.exploredSceneIDs)
	return ids
}

// VisitedLocationIDs returns a copy of the visited location IDs.
func (cn *CampaignNarrative) VisitedLocationIDs() []identity.LocationID {
	ids := make([]identity.LocationID, len(cn.visitedLocationIDs))
	copy(ids, cn.visitedLocationIDs)
	return ids
}

// HasExploredScene returns true if the given scene has been explored.
func (cn *CampaignNarrative) HasExploredScene(id identity.SceneID) bool {
	for _, sid := range cn.exploredSceneIDs {
		if sid.String() == id.String() {
			return true
		}
	}
	return false
}

// HasVisitedLocation returns true if the given location has been visited.
func (cn *CampaignNarrative) HasVisitedLocation(id identity.LocationID) bool {
	for _, lid := range cn.visitedLocationIDs {
		if lid.String() == id.String() {
			return true
		}
	}
	return false
}

// CanAccessScene checks if the campaign has explored all scenes in at least
// one complete prerequisite set.
func (cn *CampaignNarrative) CanAccessScene(prerequisiteSets [][]identity.SceneID) bool {
	if len(prerequisiteSets) == 0 {
		return true
	}
	for _, set := range prerequisiteSets {
		if cn.allScenesExplored(set) {
			return true
		}
	}
	return false
}

func (cn *CampaignNarrative) allScenesExplored(sceneIDs []identity.SceneID) bool {
	for _, id := range sceneIDs {
		if !cn.HasExploredScene(id) {
			return false
		}
	}
	return true
}

// MarkSceneExplored records that this campaign has explored the given scene.
func (cn *CampaignNarrative) MarkSceneExplored(sceneID identity.SceneID) ([]event.Event, error) {
	if cn.HasExploredScene(sceneID) {
		return nil, nil
	}
	cn.exploredSceneIDs = append(cn.exploredSceneIDs, sceneID)
	evt := event.EntityLinked{
		EntityAID:    cn.id.String(),
		EntityBID:    sceneID.String(),
		Relationship: "explored",
	}
	return []event.Event{evt}, nil
}

// MarkLocationVisited records that this campaign has visited the given location.
func (cn *CampaignNarrative) MarkLocationVisited(locationID identity.LocationID) ([]event.Event, error) {
	if cn.HasVisitedLocation(locationID) {
		return nil, nil
	}
	cn.visitedLocationIDs = append(cn.visitedLocationIDs, locationID)
	evt := event.EntityLinked{
		EntityAID:    cn.id.String(),
		EntityBID:    locationID.String(),
		Relationship: "visited",
	}
	return []event.Event{evt}, nil
}

// AddCampaignBeat records a campaign-specific beat created at the table.
func (cn *CampaignNarrative) AddCampaignBeat(id identity.BeatID) {
	cn.campaignBeatIDs = append(cn.campaignBeatIDs, id)
}

// --- Snapshot / Reconstitute ---

type CampaignNarrativeSnapshot struct {
	ID                 identity.CampaignNarrativeID
	CampaignID         identity.CampaignID
	GameID             identity.GameID
	DiscoveredBeatIDs  []identity.BeatID
	CampaignBeatIDs    []identity.BeatID
	ExploredSceneIDs   []identity.SceneID
	VisitedLocationIDs []identity.LocationID
}

func (cn *CampaignNarrative) Snapshot() CampaignNarrativeSnapshot {
	beatIDs := make([]identity.BeatID, len(cn.discoveredBeatIDs))
	copy(beatIDs, cn.discoveredBeatIDs)
	campaignBeatIDs := make([]identity.BeatID, len(cn.campaignBeatIDs))
	copy(campaignBeatIDs, cn.campaignBeatIDs)
	sceneIDs := make([]identity.SceneID, len(cn.exploredSceneIDs))
	copy(sceneIDs, cn.exploredSceneIDs)
	locIDs := make([]identity.LocationID, len(cn.visitedLocationIDs))
	copy(locIDs, cn.visitedLocationIDs)
	return CampaignNarrativeSnapshot{
		ID:                 cn.id,
		CampaignID:         cn.campaignID,
		GameID:             cn.gameID,
		DiscoveredBeatIDs:  beatIDs,
		CampaignBeatIDs:    campaignBeatIDs,
		ExploredSceneIDs:   sceneIDs,
		VisitedLocationIDs: locIDs,
	}
}

func ReconstituteCampaignNarrative(snap CampaignNarrativeSnapshot) *CampaignNarrative {
	beatIDs := make([]identity.BeatID, len(snap.DiscoveredBeatIDs))
	copy(beatIDs, snap.DiscoveredBeatIDs)
	campaignBeatIDs := make([]identity.BeatID, len(snap.CampaignBeatIDs))
	copy(campaignBeatIDs, snap.CampaignBeatIDs)
	sceneIDs := make([]identity.SceneID, len(snap.ExploredSceneIDs))
	copy(sceneIDs, snap.ExploredSceneIDs)
	locIDs := make([]identity.LocationID, len(snap.VisitedLocationIDs))
	copy(locIDs, snap.VisitedLocationIDs)
	return &CampaignNarrative{
		id:                 snap.ID,
		campaignID:         snap.CampaignID,
		gameID:             snap.GameID,
		discoveredBeatIDs:  beatIDs,
		campaignBeatIDs:    campaignBeatIDs,
		exploredSceneIDs:   sceneIDs,
		visitedLocationIDs: locIDs,
	}
}
