package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- PlayerCharacter test helpers ---

func mustCreatePlayerCharacter(t *testing.T) *PlayerCharacter {
	t.Helper()
	pc, _, err := CreatePlayerCharacter(identity.NewPlayerCharacterID(), identity.NewGameID(), "Elara", "player1", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCreatePlayerCharacter: %v", err)
	}
	return pc
}

// --- Constructor tests ---

func TestCreatePlayerCharacter_ValidInputs(t *testing.T) {
	id := identity.NewPlayerCharacterID()
	gameID := identity.NewGameID()
	pc, events, err := CreatePlayerCharacter(id, gameID, "Elara", "player1", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pc.PCName() != "Elara" {
		t.Fatalf("expected name 'Elara', got %q", pc.PCName())
	}
	if pc.Status() != PlayerCharacterStatusActive {
		t.Fatalf("expected active status, got %q", pc.Status())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec := events[0].(event.EntityCreated)
	if ec.EntityType != "player_character" {
		t.Fatalf("expected entity type 'player_character', got %q", ec.EntityType)
	}
}

func TestCreatePlayerCharacter_RequiresID(t *testing.T) {
	_, _, err := CreatePlayerCharacter(identity.PlayerCharacterID{}, identity.NewGameID(), "Name", "player1", event.SourceGrimoire)
	if !errors.Is(err, ErrPlayerCharacterIDRequired) {
		t.Fatalf("expected ErrPlayerCharacterIDRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacter_RequiresGameID(t *testing.T) {
	_, _, err := CreatePlayerCharacter(identity.NewPlayerCharacterID(), identity.GameID{}, "Name", "player1", event.SourceGrimoire)
	if !errors.Is(err, ErrPCGameIDRequired) {
		t.Fatalf("expected ErrPCGameIDRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacter_RequiresName(t *testing.T) {
	_, _, err := CreatePlayerCharacter(identity.NewPlayerCharacterID(), identity.NewGameID(), "", "player1", event.SourceGrimoire)
	if !errors.Is(err, ErrCharacterNameRequired) {
		t.Fatalf("expected ErrCharacterNameRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacter_RequiresOwnerPlayerID(t *testing.T) {
	_, _, err := CreatePlayerCharacter(identity.NewPlayerCharacterID(), identity.NewGameID(), "Name", "", event.SourceGrimoire)
	if !errors.Is(err, ErrOwnerPlayerIDRequired) {
		t.Fatalf("expected ErrOwnerPlayerIDRequired, got: %v", err)
	}
}

// --- Method tests ---

func TestPlayerCharacter_Retire_Succeeds(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	pc, events, err := pc.Retire(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pc.Status() != PlayerCharacterStatusRetired {
		t.Fatalf("expected retired status, got %q", pc.Status())
	}
	eu := events[0].(event.EntityUpdated)
	if eu.NewValue != "retired" {
		t.Fatalf("expected new value 'retired', got %q", eu.NewValue)
	}
}

func TestPlayerCharacter_Retire_AlreadyRetired_ReturnsError(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	pc, _, _ = pc.Retire(event.SourceGrimoire)
	_, _, err := pc.Retire(event.SourceGrimoire)
	if !errors.Is(err, ErrCharacterAlreadyRetired) {
		t.Fatalf("expected ErrCharacterAlreadyRetired, got: %v", err)
	}
}

func TestPlayerCharacter_UpdateContent_SetsFields(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	pc, _, err := pc.UpdateContent("New Name", "Backstory", "Public face", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pc.PCName() != "New Name" {
		t.Fatalf("expected 'New Name', got %q", pc.PCName())
	}
	if pc.Description() != "Backstory" {
		t.Fatalf("expected 'Backstory', got %q", pc.Description())
	}
}

func TestPlayerCharacter_AssignMacGuffin_Succeeds(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	mgID := identity.NewMacGuffinID()
	pc, events, err := pc.AssignMacGuffin(mgID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := pc.Snapshot()
	if len(snap.MacGuffinIDs) != 1 {
		t.Fatalf("expected 1 macguffin, got %d", len(snap.MacGuffinIDs))
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "possesses" {
		t.Fatalf("expected 'possesses', got %q", el.Relationship)
	}
}

func TestPlayerCharacter_AssignMacGuffin_WhenRetired_ReturnsError(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	pc, _, _ = pc.Retire(event.SourceGrimoire)
	_, _, err := pc.AssignMacGuffin(identity.NewMacGuffinID(), event.SourceGrimoire)
	if !errors.Is(err, ErrCharacterRetired) {
		t.Fatalf("expected ErrCharacterRetired, got: %v", err)
	}
}

func TestPlayerCharacter_ReleaseMacGuffin_Succeeds(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	mgID := identity.NewMacGuffinID()
	pc, _, _ = pc.AssignMacGuffin(mgID, event.SourceGrimoire)
	pc, _, err := pc.ReleaseMacGuffin(mgID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pc.Snapshot().MacGuffinIDs) != 0 {
		t.Fatal("expected 0 macguffins after release")
	}
}

func TestPlayerCharacter_ReleaseMacGuffin_NotPossessed_ReturnsError(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	_, _, err := pc.ReleaseMacGuffin(identity.NewMacGuffinID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinNotPossessed) {
		t.Fatalf("expected ErrMacGuffinNotPossessed, got: %v", err)
	}
}

func TestPlayerCharacter_JoinCampaign_Succeeds(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	campID := identity.NewCampaignID()
	pc, events, err := pc.JoinCampaign(campID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pc.Snapshot().CampaignIDs) != 1 {
		t.Fatal("expected 1 campaign")
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "participates_in" {
		t.Fatalf("expected 'participates_in', got %q", el.Relationship)
	}
}

func TestPlayerCharacter_JoinCampaign_Duplicate_ReturnsError(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	campID := identity.NewCampaignID()
	pc, _, _ = pc.JoinCampaign(campID, event.SourceGrimoire)
	_, _, err := pc.JoinCampaign(campID, event.SourceGrimoire)
	if !errors.Is(err, ErrAlreadyInCampaign) {
		t.Fatalf("expected ErrAlreadyInCampaign, got: %v", err)
	}
}

func TestPlayerCharacter_Reveal_SetsPlayerVisible(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	sessID := identity.NewSessionID()
	pc, events, err := pc.Reveal([]string{"player2"}, sessID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pc.IsPlayerVisible() {
		t.Fatal("expected playerVisible true")
	}
	er := events[0].(event.EntityRevealed)
	if er.SessionID != sessID.String() {
		t.Fatalf("expected sessionID %s, got %s", sessID, er.SessionID)
	}
}

// --- Snapshot round-trip ---

func TestPlayerCharacter_Snapshot_RoundTrip(t *testing.T) {
	pc := mustCreatePlayerCharacter(t)
	campID := identity.NewCampaignID()
	pc, _, _ = pc.JoinCampaign(campID, event.SourceGrimoire)
	mgID := identity.NewMacGuffinID()
	pc, _, _ = pc.AssignMacGuffin(mgID, event.SourceGrimoire)

	snap1 := pc.Snapshot()
	reconstituted := ReconstitutePlayerCharacter(snap1)
	snap2 := reconstituted.Snapshot()

	if snap1.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap1.Name != snap2.Name {
		t.Fatal("Name mismatch after round trip")
	}
	if snap1.Status != snap2.Status {
		t.Fatal("Status mismatch after round trip")
	}
	if snap1.OwnerPlayerID != snap2.OwnerPlayerID {
		t.Fatal("OwnerPlayerID mismatch after round trip")
	}
	if len(snap1.CampaignIDs) != len(snap2.CampaignIDs) {
		t.Fatal("CampaignIDs length mismatch after round trip")
	}
	if len(snap1.MacGuffinIDs) != len(snap2.MacGuffinIDs) {
		t.Fatal("MacGuffinIDs length mismatch after round trip")
	}
}
