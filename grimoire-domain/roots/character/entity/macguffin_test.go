package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- MacGuffin test helpers ---

func mustCreateMacGuffin(t *testing.T) *MacGuffin {
	t.Helper()
	mg, _, err := CreateMacGuffin(identity.NewMacGuffinID(), identity.NewGameID(), "The One Ring", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCreateMacGuffin: %v", err)
	}
	return mg
}

// --- Constructor tests ---

func TestCreateMacGuffin_ValidInputs(t *testing.T) {
	id := identity.NewMacGuffinID()
	gameID := identity.NewGameID()
	mg, events, err := CreateMacGuffin(id, gameID, "The One Ring", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.MGName() != "The One Ring" {
		t.Fatalf("expected name 'The One Ring', got %q", mg.MGName())
	}
	if !mg.NPCPossessorID().IsZero() || !mg.PCPossessorID().IsZero() || !mg.MGLocationID().IsZero() {
		t.Fatal("expected all possession fields zero")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec := events[0].(event.EntityCreated)
	if ec.EntityType != "macguffin" {
		t.Fatalf("expected entity type 'macguffin', got %q", ec.EntityType)
	}
}

func TestCreateMacGuffin_RequiresID(t *testing.T) {
	_, _, err := CreateMacGuffin(identity.MacGuffinID{}, identity.NewGameID(), "Name", event.SourceGrimoire)
	if !errors.Is(err, ErrMGMacGuffinIDRequired) {
		t.Fatalf("expected ErrMGMacGuffinIDRequired, got: %v", err)
	}
}

func TestCreateMacGuffin_RequiresGameID(t *testing.T) {
	_, _, err := CreateMacGuffin(identity.NewMacGuffinID(), identity.GameID{}, "Name", event.SourceGrimoire)
	if !errors.Is(err, ErrMGGameIDRequired) {
		t.Fatalf("expected ErrMGGameIDRequired, got: %v", err)
	}
}

func TestCreateMacGuffin_RequiresName(t *testing.T) {
	_, _, err := CreateMacGuffin(identity.NewMacGuffinID(), identity.NewGameID(), "", event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinNameRequired) {
		t.Fatalf("expected ErrMacGuffinNameRequired, got: %v", err)
	}
}

// --- Possession model tests ---

func TestMacGuffin_PlaceAt_SetsLocation(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	locID := identity.NewLocationID()
	mg, events, err := mg.PlaceAt(locID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.MGLocationID().String() != locID.String() {
		t.Fatalf("expected location %s, got %s", locID, mg.MGLocationID())
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "located_at" {
		t.Fatalf("expected 'located_at', got %q", el.Relationship)
	}
}

func TestMacGuffin_PlaceAt_RequiresLocationID(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	_, _, err := mg.PlaceAt(identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrMGLocationIDRequired) {
		t.Fatalf("expected ErrMGLocationIDRequired, got: %v", err)
	}
}

func TestMacGuffin_TransferToNPC_ClearsOtherPossession(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	locID := identity.NewLocationID()
	mg, _, _ = mg.PlaceAt(locID, event.SourceGrimoire)

	npcID := identity.NewNarrativeCharacterID()
	mg, _, err := mg.TransferToNPC(npcID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.NPCPossessorID().String() != npcID.String() {
		t.Fatalf("expected NPC possessor %s, got %s", npcID, mg.NPCPossessorID())
	}
	if !mg.MGLocationID().IsZero() {
		t.Fatal("expected location cleared")
	}
	if !mg.PCPossessorID().IsZero() {
		t.Fatal("expected PC possessor cleared")
	}
}

func TestMacGuffin_TransferToNPC_RequiresID(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	_, _, err := mg.TransferToNPC(identity.NarrativeCharacterID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrMGCharacterIDRequired) {
		t.Fatalf("expected ErrMGCharacterIDRequired, got: %v", err)
	}
}

func TestMacGuffin_TransferToPC_ClearsOtherPossession(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	npcID := identity.NewNarrativeCharacterID()
	mg, _, _ = mg.TransferToNPC(npcID, event.SourceGrimoire)

	pcID := identity.NewPlayerCharacterID()
	mg, _, err := mg.TransferToPC(pcID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.PCPossessorID().String() != pcID.String() {
		t.Fatalf("expected PC possessor %s, got %s", pcID, mg.PCPossessorID())
	}
	if !mg.NPCPossessorID().IsZero() {
		t.Fatal("expected NPC possessor cleared")
	}
}

func TestMacGuffin_TransferToPC_RequiresID(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	_, _, err := mg.TransferToPC(identity.PlayerCharacterID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrMGCharacterIDRequired) {
		t.Fatalf("expected ErrMGCharacterIDRequired, got: %v", err)
	}
}

func TestMacGuffin_Drop_ClearsPossessors(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	npcID := identity.NewNarrativeCharacterID()
	mg, _, _ = mg.TransferToNPC(npcID, event.SourceGrimoire)
	locID := identity.NewLocationID()
	mg, events, err := mg.Drop(locID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mg.NPCPossessorID().IsZero() {
		t.Fatal("expected NPC possessor cleared")
	}
	if mg.MGLocationID().String() != locID.String() {
		t.Fatalf("expected location %s, got %s", locID, mg.MGLocationID())
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	eu := events[0].(event.EntityUpdated)
	if eu.Field != "possessor" {
		t.Fatalf("expected field 'possessor', got %q", eu.Field)
	}
}

func TestMacGuffin_Reveal_SetsPlayerVisible(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	sessID := identity.NewSessionID()
	mg, _, err := mg.Reveal([]string{"player1"}, sessID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mg.IsPlayerVisible() {
		t.Fatal("expected playerVisible true")
	}
}

func TestMacGuffin_UpdateContent_SetsFields(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	mg, _, err := mg.UpdateContent("New Name", "Secret desc", "Player desc", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mg.MGName() != "New Name" {
		t.Fatalf("expected 'New Name', got %q", mg.MGName())
	}
}

// --- Destroy tests ---

func TestMacGuffin_Destroy_SetsDestroyedFlag(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	npcID := identity.NewNarrativeCharacterID()
	mg, _, _ = mg.TransferToNPC(npcID, event.SourceGrimoire)
	mg, events, err := mg.Destroy(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mg.IsDestroyed() {
		t.Fatal("expected destroyed true")
	}
	if !mg.NPCPossessorID().IsZero() {
		t.Fatal("expected NPC possessor cleared on destroy")
	}
	eu := events[0].(event.EntityUpdated)
	if eu.NewValue != "destroyed" {
		t.Fatalf("expected 'destroyed', got %q", eu.NewValue)
	}
}

func TestMacGuffin_Destroy_AlreadyDestroyed_ReturnsError(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	mg, _, _ = mg.Destroy(event.SourceGrimoire)
	_, _, err := mg.Destroy(event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinAlreadyDestroyed) {
		t.Fatalf("expected ErrMacGuffinAlreadyDestroyed, got: %v", err)
	}
}

func TestMacGuffin_MethodsOnDestroyed_ReturnError(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	mg, _, _ = mg.Destroy(event.SourceGrimoire)

	_, _, err := mg.PlaceAt(identity.NewLocationID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("PlaceAt: expected ErrMacGuffinDestroyed, got: %v", err)
	}
	_, _, err = mg.TransferToNPC(identity.NewNarrativeCharacterID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("TransferToNPC: expected ErrMacGuffinDestroyed, got: %v", err)
	}
	_, _, err = mg.TransferToPC(identity.NewPlayerCharacterID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("TransferToPC: expected ErrMacGuffinDestroyed, got: %v", err)
	}
	_, _, err = mg.Drop(identity.NewLocationID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("Drop: expected ErrMacGuffinDestroyed, got: %v", err)
	}
	_, _, err = mg.Reveal([]string{"p1"}, identity.NewSessionID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("Reveal: expected ErrMacGuffinDestroyed, got: %v", err)
	}
	_, _, err = mg.UpdateContent("n", "d", "pd", event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("UpdateContent: expected ErrMacGuffinDestroyed, got: %v", err)
	}
}

// --- Possession invariant ---

func TestMacGuffin_PossessionInvariant_AtMostOne(t *testing.T) {
	// After PlaceAt, only locationID is set
	mg := mustCreateMacGuffin(t)
	locID := identity.NewLocationID()
	mg, _, _ = mg.PlaceAt(locID, event.SourceGrimoire)
	if err := mg.validatePossessionInvariant(); err != nil {
		t.Fatalf("expected valid possession, got: %v", err)
	}

	// After TransferToNPC, only npcPossessorID is set
	npcID := identity.NewNarrativeCharacterID()
	mg, _, _ = mg.TransferToNPC(npcID, event.SourceGrimoire)
	if err := mg.validatePossessionInvariant(); err != nil {
		t.Fatalf("expected valid possession, got: %v", err)
	}

	// After TransferToPC, only pcPossessorID is set
	pcID := identity.NewPlayerCharacterID()
	mg, _, _ = mg.TransferToPC(pcID, event.SourceGrimoire)
	if err := mg.validatePossessionInvariant(); err != nil {
		t.Fatalf("expected valid possession, got: %v", err)
	}
}

// --- Snapshot round-trip ---

func TestMacGuffin_Snapshot_RoundTrip(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	npcID := identity.NewNarrativeCharacterID()
	mg, _, _ = mg.TransferToNPC(npcID, event.SourceGrimoire)
	mg, _, _ = mg.UpdateContent("Updated Ring", "Secret", "A mysterious ring", event.SourceGrimoire)

	snap1 := mg.Snapshot()
	reconstituted := ReconstituteMacGuffin(snap1)
	snap2 := reconstituted.Snapshot()

	if snap1.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch")
	}
	if snap1.Name != snap2.Name {
		t.Fatal("Name mismatch")
	}
	if snap1.NPCPossessorID.String() != snap2.NPCPossessorID.String() {
		t.Fatal("NPCPossessorID mismatch")
	}
	if snap1.Destroyed != snap2.Destroyed {
		t.Fatal("Destroyed mismatch")
	}
	if snap1.Description != snap2.Description {
		t.Fatal("Description mismatch")
	}
}

func TestMacGuffin_Snapshot_RoundTrip_Destroyed(t *testing.T) {
	mg := mustCreateMacGuffin(t)
	mg, _, _ = mg.Destroy(event.SourceGrimoire)

	snap := mg.Snapshot()
	reconstituted := ReconstituteMacGuffin(snap)
	if !reconstituted.IsDestroyed() {
		t.Fatal("expected destroyed after reconstitution")
	}
}
