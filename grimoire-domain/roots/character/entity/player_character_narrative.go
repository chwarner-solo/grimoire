package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// PlayerCharacterNarrative is a purely additive event journal.
// No state machine. No Handle/Replay. Snapshot/Reconstitute only.
// Intentional departure from Game/Campaign/Session pattern —
// consistent with CampaignNarrative (ADR-016). See ADR-019.
type PlayerCharacterNarrative struct {
	id                    identity.PlayerCharacterNarrativeID
	characterID           identity.PlayerCharacterID
	campaignID            identity.CampaignID
	gameID                identity.GameID
	backgroundBeatIDs     []identity.BeatID
	personalBeatIDs       []identity.BeatID
	revealedBackgroundIDs []identity.BeatID
}

// CreatePlayerCharacterNarrative constructs a new narrative journal.
func CreatePlayerCharacterNarrative(
	id identity.PlayerCharacterNarrativeID,
	characterID identity.PlayerCharacterID,
	campaignID identity.CampaignID,
	gameID identity.GameID,
	source event.Source,
) (*PlayerCharacterNarrative, []event.Event, error) {
	if id.IsZero() {
		return nil, nil, ErrPlayerCharacterNarrativeIDRequired
	}
	if characterID.IsZero() {
		return nil, nil, ErrPCNCharacterIDRequired
	}
	if campaignID.IsZero() {
		return nil, nil, ErrPCNCampaignIDRequired
	}
	if gameID.IsZero() {
		return nil, nil, ErrPCNGameIDRequired
	}
	evt := event.EntityCreated{
		EntityID:   id.String(),
		EntityType: "player_character_narrative",
		Name:       "",
		Source:     source,
	}
	return &PlayerCharacterNarrative{
		id:          id,
		characterID: characterID,
		campaignID:  campaignID,
		gameID:      gameID,
	}, []event.Event{evt}, nil
}

// AddBackgroundBeat adds a pre-campaign backstory beat.
func (pcn *PlayerCharacterNarrative) AddBackgroundBeat(
	beatID identity.BeatID,
	source event.Source,
) (*PlayerCharacterNarrative, []event.Event, error) {
	for _, bid := range pcn.backgroundBeatIDs {
		if bid.String() == beatID.String() {
			return nil, nil, ErrBackgroundBeatAlreadyAdded
		}
	}
	pcn.backgroundBeatIDs = append(pcn.backgroundBeatIDs, beatID)
	evt := event.EntityLinked{
		EntityAID:    pcn.id.String(),
		EntityBID:    beatID.String(),
		Relationship: "has_background_beat",
		Source:       source,
	}
	return pcn, []event.Event{evt}, nil
}

// DiscoverPersonalBeat discovers a personal arc beat during play.
// prerequisiteSets is the Beat's prerequisite sets — the interactor loads the
// Beat and passes the sets in. Cross-aggregate imports are forbidden.
func (pcn *PlayerCharacterNarrative) DiscoverPersonalBeat(
	beatID identity.BeatID,
	prerequisiteSets [][]identity.BeatID,
	source event.Source,
) (*PlayerCharacterNarrative, []event.Event, error) {
	if !pcn.prerequisitesMet(prerequisiteSets) {
		return nil, nil, ErrPrerequisitesNotMet
	}
	pcn.personalBeatIDs = append(pcn.personalBeatIDs, beatID)
	evt := event.EntityLinked{
		EntityAID:    pcn.id.String(),
		EntityBID:    beatID.String(),
		Relationship: "discovered_personal_beat",
		Source:       source,
	}
	return pcn, []event.Event{evt}, nil
}

// RevealBackgroundBeat reveals a background beat to the owning player.
func (pcn *PlayerCharacterNarrative) RevealBackgroundBeat(
	beatID identity.BeatID,
	to []string,
	sessionID identity.SessionID,
	source event.Source,
) (*PlayerCharacterNarrative, []event.Event, error) {
	found := false
	for _, bid := range pcn.backgroundBeatIDs {
		if bid.String() == beatID.String() {
			found = true
			break
		}
	}
	if !found {
		return nil, nil, ErrBackgroundBeatNotFound
	}
	for _, rid := range pcn.revealedBackgroundIDs {
		if rid.String() == beatID.String() {
			return nil, nil, ErrBackgroundBeatAlreadyRevealed
		}
	}
	pcn.revealedBackgroundIDs = append(pcn.revealedBackgroundIDs, beatID)
	evt := event.EntityRevealed{
		EntityID:   beatID.String(),
		RevealedTo: to,
		SessionID:  sessionID.String(),
		Source:     source,
	}
	return pcn, []event.Event{evt}, nil
}

// allDiscoveredBeatIDs returns union of personalBeatIDs + revealedBackgroundIDs.
func (pcn *PlayerCharacterNarrative) allDiscoveredBeatIDs() []identity.BeatID {
	result := make([]identity.BeatID, 0, len(pcn.personalBeatIDs)+len(pcn.revealedBackgroundIDs))
	result = append(result, pcn.personalBeatIDs...)
	result = append(result, pcn.revealedBackgroundIDs...)
	return result
}

// prerequisitesMet checks whether at least one complete prerequisite set is satisfied.
func (pcn *PlayerCharacterNarrative) prerequisitesMet(prerequisiteSets [][]identity.BeatID) bool {
	if len(prerequisiteSets) == 0 {
		return true
	}
	discovered := pcn.allDiscoveredBeatIDs()
	discoveredSet := make(map[string]bool, len(discovered))
	for _, d := range discovered {
		discoveredSet[d.String()] = true
	}
	for _, set := range prerequisiteSets {
		allMet := true
		for _, prereq := range set {
			if !discoveredSet[prereq.String()] {
				allMet = false
				break
			}
		}
		if allMet {
			return true
		}
	}
	return false
}

// --- Getters ---

func (pcn *PlayerCharacterNarrative) PlayerCharacterNarrativeID() identity.PlayerCharacterNarrativeID {
	return pcn.id
}
func (pcn *PlayerCharacterNarrative) PCNCharacterID() identity.PlayerCharacterID { return pcn.characterID }
func (pcn *PlayerCharacterNarrative) PCNCampaignID() identity.CampaignID         { return pcn.campaignID }
func (pcn *PlayerCharacterNarrative) PCNGameID() identity.GameID                 { return pcn.gameID }
