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
	id                identity.CampaignNarrativeID
	campaignID        identity.CampaignID
	gameID            identity.GameID
	discoveredBeatIDs []identity.BeatID
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

// --- Snapshot / Reconstitute ---

type CampaignNarrativeSnapshot struct {
	ID                identity.CampaignNarrativeID
	CampaignID        identity.CampaignID
	GameID            identity.GameID
	DiscoveredBeatIDs []identity.BeatID
}

func (cn *CampaignNarrative) Snapshot() CampaignNarrativeSnapshot {
	ids := make([]identity.BeatID, len(cn.discoveredBeatIDs))
	copy(ids, cn.discoveredBeatIDs)
	return CampaignNarrativeSnapshot{
		ID:                cn.id,
		CampaignID:        cn.campaignID,
		GameID:            cn.gameID,
		DiscoveredBeatIDs: ids,
	}
}

func ReconstituteCampaignNarrative(snap CampaignNarrativeSnapshot) *CampaignNarrative {
	ids := make([]identity.BeatID, len(snap.DiscoveredBeatIDs))
	copy(ids, snap.DiscoveredBeatIDs)
	return &CampaignNarrative{
		id:                snap.ID,
		campaignID:        snap.CampaignID,
		gameID:            snap.GameID,
		discoveredBeatIDs: ids,
	}
}
