package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestCreateMasterBeat_ValidInputs(t *testing.T) {
	id := identity.NewBeatID()
	gameID := identity.NewGameID()
	beat, events, err := CreateMasterBeat(id, "The Dark Forest", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if beat.BeatID().String() != id.String() {
		t.Fatalf("expected beat ID %s, got %s", id.String(), beat.BeatID().String())
	}
	if beat.Name() != "The Dark Forest" {
		t.Fatalf("expected name 'The Dark Forest', got %q", beat.Name())
	}
	if beat.Scope() != value.BeatScopeMaster {
		t.Fatalf("expected scope master, got %q", beat.Scope())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityType != "beat" {
		t.Fatalf("expected entity type 'beat', got %q", ec.EntityType)
	}
}

func TestCreateMasterBeat_RequiresNonZeroID(t *testing.T) {
	_, _, err := CreateMasterBeat(identity.BeatID{}, "Name", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrBeatIDRequired) {
		t.Fatalf("expected ErrBeatIDRequired, got: %v", err)
	}
}

func TestCreateMasterBeat_RequiresNonEmptyName(t *testing.T) {
	_, _, err := CreateMasterBeat(identity.NewBeatID(), "", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrBeatNameRequired) {
		t.Fatalf("expected ErrBeatNameRequired, got: %v", err)
	}
}

func TestCreateMasterBeat_RequiresNonZeroGameID(t *testing.T) {
	_, _, err := CreateMasterBeat(identity.NewBeatID(), "Name", value.BeatTypeRequired, identity.GameID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got: %v", err)
	}
}

func TestCreateCampaignBeat_RequiresCampaignID(t *testing.T) {
	_, _, err := CreateCampaignBeat(identity.NewBeatID(), "Name", identity.NewGameID(), identity.CampaignID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignIDRequired) {
		t.Fatalf("expected ErrCampaignIDRequired, got: %v", err)
	}
}

func TestCreateCampaignBeat_SetsScopeCampaign(t *testing.T) {
	beat, _, err := CreateCampaignBeat(identity.NewBeatID(), "Side Quest", identity.NewGameID(), identity.NewCampaignID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if beat.Scope() != value.BeatScopeCampaign {
		t.Fatalf("expected scope campaign, got %q", beat.Scope())
	}
	if beat.BeatType() != value.BeatTypeCampaignSpecific {
		t.Fatalf("expected beat type campaign-specific, got %q", beat.BeatType())
	}
}

func TestBeat_UpdateContent_SetsAllFields(t *testing.T) {
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Original", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	events, err := beat.UpdateContent("New Name", "A description", "Player sees this")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if beat.Name() != "New Name" {
		t.Fatalf("expected name 'New Name', got %q", beat.Name())
	}
	if beat.Description() != "A description" {
		t.Fatalf("expected description 'A description', got %q", beat.Description())
	}
	if beat.PlayerDescription() != "Player sees this" {
		t.Fatalf("expected player description 'Player sees this', got %q", beat.PlayerDescription())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if _, ok := events[0].(event.EntityUpdated); !ok {
		t.Fatalf("expected EntityUpdated, got %T", events[0])
	}
}

func TestBeat_UpdateContent_RequiresName(t *testing.T) {
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Original", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	_, err := beat.UpdateContent("", "desc", "player desc")
	if !errors.Is(err, ErrBeatNameRequired) {
		t.Fatalf("expected ErrBeatNameRequired, got: %v", err)
	}
}

func TestBeat_UpdateContent_RequiresDescription(t *testing.T) {
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Original", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	_, err := beat.UpdateContent("Name", "", "player desc")
	if !errors.Is(err, ErrBeatDescriptionRequired) {
		t.Fatalf("expected ErrBeatDescriptionRequired, got: %v", err)
	}
}

func TestBeat_UpdateContent_RequiresPlayerDescription(t *testing.T) {
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Original", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	_, err := beat.UpdateContent("Name", "desc", "")
	if !errors.Is(err, ErrBeatPlayerDescriptionRequired) {
		t.Fatalf("expected ErrBeatPlayerDescriptionRequired, got: %v", err)
	}
}

func TestBeat_WouldCreateCycle_DetectsCycle(t *testing.T) {
	gameID := identity.NewGameID()
	beatA, _, _ := CreateMasterBeat(identity.NewBeatID(), "A", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	beatB, _, _ := CreateMasterBeat(identity.NewBeatID(), "B", value.BeatTypeRequired, gameID, event.SourceGrimoire)

	// A requires B, B requires A → cycle
	beatA.AddPrerequisite(beatB.BeatID())
	// chain: B → A (B's prerequisite chain includes A)
	err := beatB.WouldCreateCycle(beatA.BeatID(), []*Beat{beatA})
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
	var cycleErr ErrCycleDetected
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected ErrCycleDetected, got: %v", err)
	}
}

func TestBeat_WouldCreateCycle_AllowsValidPrerequisite(t *testing.T) {
	gameID := identity.NewGameID()
	beatA, _, _ := CreateMasterBeat(identity.NewBeatID(), "A", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	beatB, _, _ := CreateMasterBeat(identity.NewBeatID(), "B", value.BeatTypeRequired, gameID, event.SourceGrimoire)

	// B wants to require A — no cycle because A has no prerequisites
	err := beatB.WouldCreateCycle(beatA.BeatID(), []*Beat{beatA})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBeat_AddPrerequisite_AppendsToFirstSet(t *testing.T) {
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Beat", value.BeatTypeRequired, identity.NewGameID(), event.SourceGrimoire)
	prereqID := identity.NewBeatID()
	beat.AddPrerequisite(prereqID)

	snap := beat.Snapshot()
	if len(snap.PrerequisiteSets) != 1 {
		t.Fatalf("expected 1 prerequisite set, got %d", len(snap.PrerequisiteSets))
	}
	if len(snap.PrerequisiteSets[0]) != 1 {
		t.Fatalf("expected 1 prerequisite in first set, got %d", len(snap.PrerequisiteSets[0]))
	}
	if snap.PrerequisiteSets[0][0].String() != prereqID.String() {
		t.Fatalf("expected prerequisite %s, got %s", prereqID.String(), snap.PrerequisiteSets[0][0].String())
	}
}

func TestBeat_Snapshot_RoundTrips(t *testing.T) {
	gameID := identity.NewGameID()
	beat, _, _ := CreateMasterBeat(identity.NewBeatID(), "Original", value.BeatTypeRequired, gameID, event.SourceGrimoire)
	beat.UpdateContent("Updated", "A description", "Player desc")
	prereqID := identity.NewBeatID()
	beat.AddPrerequisite(prereqID)

	snap := beat.Snapshot()
	reconstituted := ReconstituteBeat(snap)

	snap2 := reconstituted.Snapshot()
	if snap.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap.Name != snap2.Name {
		t.Fatal("Name mismatch after round trip")
	}
	if snap.Description != snap2.Description {
		t.Fatal("Description mismatch after round trip")
	}
	if snap.PlayerDescription != snap2.PlayerDescription {
		t.Fatal("PlayerDescription mismatch after round trip")
	}
	if snap.Scope != snap2.Scope {
		t.Fatal("Scope mismatch after round trip")
	}
	if snap.Type != snap2.Type {
		t.Fatal("Type mismatch after round trip")
	}
	if snap.GameID.String() != snap2.GameID.String() {
		t.Fatal("GameID mismatch after round trip")
	}
	if len(snap.PrerequisiteSets) != len(snap2.PrerequisiteSets) {
		t.Fatal("PrerequisiteSets length mismatch after round trip")
	}
}
