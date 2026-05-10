package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestCreateCampaignNarrative_ValidInputs(t *testing.T) {
	id := identity.NewCampaignNarrativeID()
	campID := identity.NewCampaignID()
	gameID := identity.NewGameID()
	cn, events, err := CreateCampaignNarrative(id, campID, gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cn.CampaignNarrativeID().String() != id.String() {
		t.Fatalf("expected ID %s, got %s", id.String(), cn.CampaignNarrativeID().String())
	}
	if cn.CampaignID().String() != campID.String() {
		t.Fatalf("expected CampaignID %s, got %s", campID.String(), cn.CampaignID().String())
	}
	if cn.GameID().String() != gameID.String() {
		t.Fatalf("expected GameID %s, got %s", gameID.String(), cn.GameID().String())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityType != "campaign_narrative" {
		t.Fatalf("expected entity type 'campaign_narrative', got %q", ec.EntityType)
	}
}

func TestCreateCampaignNarrative_RequiresID(t *testing.T) {
	_, _, err := CreateCampaignNarrative(identity.CampaignNarrativeID{}, identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignNarrativeIDRequired) {
		t.Fatalf("expected ErrCampaignNarrativeIDRequired, got: %v", err)
	}
}

func TestCreateCampaignNarrative_RequiresCampaignID(t *testing.T) {
	_, _, err := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.CampaignID{}, identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignIDRequired) {
		t.Fatalf("expected ErrCampaignIDRequired, got: %v", err)
	}
}

func TestCreateCampaignNarrative_RequiresGameID(t *testing.T) {
	_, _, err := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.GameID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got: %v", err)
	}
}

func TestCampaignNarrative_DiscoverBeat_NoPrerequisites(t *testing.T) {
	cn, _, _ := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Simple Beat", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)

	events, err := cn.DiscoverBeat(beat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	el, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if el.Relationship != "discovered" {
		t.Fatalf("expected relationship 'discovered', got %q", el.Relationship)
	}
}

func TestCampaignNarrative_DiscoverBeat_PrerequisiteNotMet(t *testing.T) {
	cn, _, _ := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	gameID := identity.NewGameID()
	prereqBeat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Prereq", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	mainBeat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Main", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	mainBeat.AddPrerequisite(prereqBeat.BeatID())

	_, err := cn.DiscoverBeat(mainBeat)
	if err == nil {
		t.Fatal("expected error for unmet prerequisite")
	}
	var prereqErr ErrPrerequisiteNotMet
	if !errors.As(err, &prereqErr) {
		t.Fatalf("expected ErrPrerequisiteNotMet, got: %v", err)
	}
}

func TestCampaignNarrative_DiscoverBeat_PrerequisiteMet(t *testing.T) {
	cn, _, _ := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	gameID := identity.NewGameID()
	prereqBeat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Prereq", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	mainBeat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Main", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	mainBeat.AddPrerequisite(prereqBeat.BeatID())

	// Discover the prerequisite first
	_, err := cn.DiscoverBeat(prereqBeat)
	if err != nil {
		t.Fatalf("unexpected error discovering prereq: %v", err)
	}

	// Now discover the main beat
	_, err = cn.DiscoverBeat(mainBeat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCampaignNarrative_DiscoverBeat_MultiplePathsOR(t *testing.T) {
	cn, _, _ := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	gameID := identity.NewGameID()
	pathA, _, _ := CreateMasterBeat(identity.NewBeatID(), "Path A", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	pathB, _, _ := CreateMasterBeat(identity.NewBeatID(), "Path B", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	mainBeat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Main", value.BeatTypeRequired, gameID, event.SourceGrimoire)

	// Two separate prerequisite sets (OR): either {pathA} or {pathB}
	mainBeat.AddPrerequisite(pathA.BeatID())                              // set 0: [pathA]
	mainBeat.prerequisiteSets = append(mainBeat.prerequisiteSets, []identity.BeatID{pathB.BeatID()}) // set 1: [pathB]

	// Discover only pathB — should satisfy set 1
	_, err := cn.DiscoverBeat(pathB)
	if err != nil {
		t.Fatalf("unexpected error discovering pathB: %v", err)
	}

	_, err = cn.DiscoverBeat(mainBeat)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCampaignNarrative_Snapshot_RoundTrips(t *testing.T) {
	cn, _, _ := CreateCampaignNarrative(identity.NewCampaignNarrativeID(), identity.NewCampaignID(), identity.NewGameID(), event.SourceGrimoire)
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Beat", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	cn.DiscoverBeat(beat)

	snap := cn.Snapshot()
	reconstituted := ReconstituteCampaignNarrative(snap)
	snap2 := reconstituted.Snapshot()

	if snap.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap.CampaignID.String() != snap2.CampaignID.String() {
		t.Fatal("CampaignID mismatch after round trip")
	}
	if snap.GameID.String() != snap2.GameID.String() {
		t.Fatal("GameID mismatch after round trip")
	}
	if len(snap.DiscoveredBeatIDs) != len(snap2.DiscoveredBeatIDs) {
		t.Fatal("DiscoveredBeatIDs length mismatch after round trip")
	}
}
