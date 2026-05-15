package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- PlayerCharacterNarrative test helpers ---

func mustCreatePCN(t *testing.T) *PlayerCharacterNarrative {
	t.Helper()
	pcn, _, err := CreatePlayerCharacterNarrative(
		identity.NewPlayerCharacterNarrativeID(),
		identity.NewPlayerCharacterID(),
		identity.NewCampaignID(),
		identity.NewGameID(),
		event.SourceGrimoire,
	)
	if err != nil {
		t.Fatalf("mustCreatePCN: %v", err)
	}
	return pcn
}

// --- Constructor tests ---

func TestCreatePlayerCharacterNarrative_ValidInputs(t *testing.T) {
	pcn, events, err := CreatePlayerCharacterNarrative(
		identity.NewPlayerCharacterNarrativeID(),
		identity.NewPlayerCharacterID(),
		identity.NewCampaignID(),
		identity.NewGameID(),
		event.SourceGrimoire,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pcn.PlayerCharacterNarrativeID().IsZero() {
		t.Fatal("expected non-zero ID")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec := events[0].(event.EntityCreated)
	if ec.EntityType != "player_character_narrative" {
		t.Fatalf("expected entity type 'player_character_narrative', got %q", ec.EntityType)
	}
}

func TestCreatePlayerCharacterNarrative_RequiresID(t *testing.T) {
	_, _, err := CreatePlayerCharacterNarrative(
		identity.PlayerCharacterNarrativeID{},
		identity.NewPlayerCharacterID(),
		identity.NewCampaignID(),
		identity.NewGameID(),
		event.SourceGrimoire,
	)
	if !errors.Is(err, ErrPlayerCharacterNarrativeIDRequired) {
		t.Fatalf("expected ErrPlayerCharacterNarrativeIDRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacterNarrative_RequiresCharacterID(t *testing.T) {
	_, _, err := CreatePlayerCharacterNarrative(
		identity.NewPlayerCharacterNarrativeID(),
		identity.PlayerCharacterID{},
		identity.NewCampaignID(),
		identity.NewGameID(),
		event.SourceGrimoire,
	)
	if !errors.Is(err, ErrPCNCharacterIDRequired) {
		t.Fatalf("expected ErrPCNCharacterIDRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacterNarrative_RequiresCampaignID(t *testing.T) {
	_, _, err := CreatePlayerCharacterNarrative(
		identity.NewPlayerCharacterNarrativeID(),
		identity.NewPlayerCharacterID(),
		identity.CampaignID{},
		identity.NewGameID(),
		event.SourceGrimoire,
	)
	if !errors.Is(err, ErrPCNCampaignIDRequired) {
		t.Fatalf("expected ErrPCNCampaignIDRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacterNarrative_RequiresGameID(t *testing.T) {
	_, _, err := CreatePlayerCharacterNarrative(
		identity.NewPlayerCharacterNarrativeID(),
		identity.NewPlayerCharacterID(),
		identity.NewCampaignID(),
		identity.GameID{},
		event.SourceGrimoire,
	)
	if !errors.Is(err, ErrPCNGameIDRequired) {
		t.Fatalf("expected ErrPCNGameIDRequired, got: %v", err)
	}
}

// --- AddBackgroundBeat ---

func TestPCN_AddBackgroundBeat_Succeeds(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	pcn, events, err := pcn.AddBackgroundBeat(beatID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := pcn.Snapshot()
	if len(snap.BackgroundBeatIDs) != 1 {
		t.Fatalf("expected 1 background beat, got %d", len(snap.BackgroundBeatIDs))
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "has_background_beat" {
		t.Fatalf("expected 'has_background_beat', got %q", el.Relationship)
	}
}

func TestPCN_AddBackgroundBeat_Duplicate_ReturnsError(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	pcn, _, _ = pcn.AddBackgroundBeat(beatID, event.SourceGrimoire)
	_, _, err := pcn.AddBackgroundBeat(beatID, event.SourceGrimoire)
	if !errors.Is(err, ErrBackgroundBeatAlreadyAdded) {
		t.Fatalf("expected ErrBackgroundBeatAlreadyAdded, got: %v", err)
	}
}

// --- DiscoverPersonalBeat ---

func TestPCN_DiscoverPersonalBeat_NoPrereqs_Succeeds(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	pcn, events, err := pcn.DiscoverPersonalBeat(beatID, nil, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := pcn.Snapshot()
	if len(snap.PersonalBeatIDs) != 1 {
		t.Fatalf("expected 1 personal beat, got %d", len(snap.PersonalBeatIDs))
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "discovered_personal_beat" {
		t.Fatalf("expected 'discovered_personal_beat', got %q", el.Relationship)
	}
}

func TestPCN_DiscoverPersonalBeat_PrereqsMet_Succeeds(t *testing.T) {
	pcn := mustCreatePCN(t)
	prereqBeatID := identity.NewBeatID()
	// Discover the prerequisite first (no prereqs of its own)
	pcn, _, _ = pcn.DiscoverPersonalBeat(prereqBeatID, nil, event.SourceGrimoire)
	// Now discover a beat that requires the prereq
	newBeatID := identity.NewBeatID()
	prereqSets := [][]identity.BeatID{{prereqBeatID}}
	pcn, _, err := pcn.DiscoverPersonalBeat(newBeatID, prereqSets, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pcn.Snapshot().PersonalBeatIDs) != 2 {
		t.Fatal("expected 2 personal beats")
	}
}

func TestPCN_DiscoverPersonalBeat_PrereqsNotMet_ReturnsError(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	prereqSets := [][]identity.BeatID{{identity.NewBeatID()}}
	_, _, err := pcn.DiscoverPersonalBeat(beatID, prereqSets, event.SourceGrimoire)
	if !errors.Is(err, ErrPrerequisitesNotMet) {
		t.Fatalf("expected ErrPrerequisitesNotMet, got: %v", err)
	}
}

// --- RevealBackgroundBeat ---

func TestPCN_RevealBackgroundBeat_Succeeds(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	pcn, _, _ = pcn.AddBackgroundBeat(beatID, event.SourceGrimoire)
	sessID := identity.NewSessionID()
	pcn, events, err := pcn.RevealBackgroundBeat(beatID, []string{"player1"}, sessID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := pcn.Snapshot()
	if len(snap.RevealedBackgroundIDs) != 1 {
		t.Fatalf("expected 1 revealed bg, got %d", len(snap.RevealedBackgroundIDs))
	}
	er := events[0].(event.EntityRevealed)
	if er.EntityID != beatID.String() {
		t.Fatalf("expected entity ID %s, got %s", beatID, er.EntityID)
	}
}

func TestPCN_RevealBackgroundBeat_NotFound_ReturnsError(t *testing.T) {
	pcn := mustCreatePCN(t)
	sessID := identity.NewSessionID()
	_, _, err := pcn.RevealBackgroundBeat(identity.NewBeatID(), []string{"p1"}, sessID, event.SourceGrimoire)
	if !errors.Is(err, ErrBackgroundBeatNotFound) {
		t.Fatalf("expected ErrBackgroundBeatNotFound, got: %v", err)
	}
}

func TestPCN_RevealBackgroundBeat_AlreadyRevealed_ReturnsError(t *testing.T) {
	pcn := mustCreatePCN(t)
	beatID := identity.NewBeatID()
	pcn, _, _ = pcn.AddBackgroundBeat(beatID, event.SourceGrimoire)
	sessID := identity.NewSessionID()
	pcn, _, _ = pcn.RevealBackgroundBeat(beatID, []string{"p1"}, sessID, event.SourceGrimoire)
	_, _, err := pcn.RevealBackgroundBeat(beatID, []string{"p1"}, sessID, event.SourceGrimoire)
	if !errors.Is(err, ErrBackgroundBeatAlreadyRevealed) {
		t.Fatalf("expected ErrBackgroundBeatAlreadyRevealed, got: %v", err)
	}
}

// --- Snapshot round-trip ---

func TestPCN_Snapshot_RoundTrip(t *testing.T) {
	pcn := mustCreatePCN(t)
	bgBeatID := identity.NewBeatID()
	pcn, _, _ = pcn.AddBackgroundBeat(bgBeatID, event.SourceGrimoire)
	personalBeatID := identity.NewBeatID()
	pcn, _, _ = pcn.DiscoverPersonalBeat(personalBeatID, nil, event.SourceGrimoire)
	sessID := identity.NewSessionID()
	pcn, _, _ = pcn.RevealBackgroundBeat(bgBeatID, []string{"p1"}, sessID, event.SourceGrimoire)

	snap1 := pcn.Snapshot()
	reconstituted := ReconstitutePlayerCharacterNarrative(snap1)
	snap2 := reconstituted.Snapshot()

	if snap1.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch")
	}
	if snap1.CharacterID.String() != snap2.CharacterID.String() {
		t.Fatal("CharacterID mismatch")
	}
	if snap1.CampaignID.String() != snap2.CampaignID.String() {
		t.Fatal("CampaignID mismatch")
	}
	if len(snap1.BackgroundBeatIDs) != len(snap2.BackgroundBeatIDs) {
		t.Fatal("BackgroundBeatIDs mismatch")
	}
	if len(snap1.PersonalBeatIDs) != len(snap2.PersonalBeatIDs) {
		t.Fatal("PersonalBeatIDs mismatch")
	}
	if len(snap1.RevealedBackgroundIDs) != len(snap2.RevealedBackgroundIDs) {
		t.Fatal("RevealedBackgroundIDs mismatch")
	}
}
