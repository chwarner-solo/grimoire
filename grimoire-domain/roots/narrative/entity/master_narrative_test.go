package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestCreateMasterNarrative_ValidInputs(t *testing.T) {
	id := identity.NewMasterNarrativeID()
	gameID := identity.NewGameID()
	mn, events, err := CreateMasterNarrative(id, gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mn.MasterNarrativeID().String() != id.String() {
		t.Fatalf("expected ID %s, got %s", id.String(), mn.MasterNarrativeID().String())
	}
	if mn.GameID().String() != gameID.String() {
		t.Fatalf("expected GameID %s, got %s", gameID.String(), mn.GameID().String())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityType != "master_narrative" {
		t.Fatalf("expected entity type 'master_narrative', got %q", ec.EntityType)
	}
}

func TestCreateMasterNarrative_RequiresID(t *testing.T) {
	_, _, err := CreateMasterNarrative(identity.MasterNarrativeID{}, identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrMasterNarrativeIDRequired) {
		t.Fatalf("expected ErrMasterNarrativeIDRequired, got: %v", err)
	}
}

func TestCreateMasterNarrative_RequiresGameID(t *testing.T) {
	_, _, err := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.GameID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got: %v", err)
	}
}

func TestMasterNarrative_AddBeat_RecordsBeatID(t *testing.T) {
	mn, _, _ := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.NewGameID(), event.SourceGrimoire)
	beatID := identity.NewBeatID()
	mn.AddBeat(beatID)

	snap := mn.Snapshot()
	if len(snap.BeatIDs) != 1 {
		t.Fatalf("expected 1 beat, got %d", len(snap.BeatIDs))
	}
	if snap.BeatIDs[0].String() != beatID.String() {
		t.Fatalf("expected beat ID %s, got %s", beatID.String(), snap.BeatIDs[0].String())
	}
}

func TestMasterNarrative_AddAct_RecordsActID(t *testing.T) {
	mn, _, _ := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.NewGameID(), event.SourceGrimoire)
	actID := identity.NewActID()
	mn.AddAct(actID)

	snap := mn.Snapshot()
	if len(snap.ActIDs) != 1 {
		t.Fatalf("expected 1 act, got %d", len(snap.ActIDs))
	}
	if snap.ActIDs[0].String() != actID.String() {
		t.Fatalf("expected act ID %s, got %s", actID.String(), snap.ActIDs[0].String())
	}
}

func TestMasterNarrative_AddSecret_RecordsSecretID(t *testing.T) {
	mn, _, _ := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.NewGameID(), event.SourceGrimoire)
	secretID := identity.NewSecretID()
	mn.AddSecret(secretID)

	snap := mn.Snapshot()
	if len(snap.SecretIDs) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(snap.SecretIDs))
	}
	if snap.SecretIDs[0].String() != secretID.String() {
		t.Fatalf("expected secret ID %s, got %s", secretID.String(), snap.SecretIDs[0].String())
	}
}

func TestMasterNarrative_AddLore_RecordsLoreID(t *testing.T) {
	mn, _, _ := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.NewGameID(), event.SourceGrimoire)
	loreID := identity.NewLoreID()
	mn.AddLore(loreID)

	snap := mn.Snapshot()
	if len(snap.LoreIDs) != 1 {
		t.Fatalf("expected 1 lore, got %d", len(snap.LoreIDs))
	}
	if snap.LoreIDs[0].String() != loreID.String() {
		t.Fatalf("expected lore ID %s, got %s", loreID.String(), snap.LoreIDs[0].String())
	}
}

func TestMasterNarrative_Snapshot_RoundTrips(t *testing.T) {
	mn, _, _ := CreateMasterNarrative(identity.NewMasterNarrativeID(), identity.NewGameID(), event.SourceGrimoire)
	mn.AddBeat(identity.NewBeatID())
	mn.AddAct(identity.NewActID())
	mn.AddSecret(identity.NewSecretID())
	mn.AddLore(identity.NewLoreID())

	snap := mn.Snapshot()
	reconstituted := ReconstituteMasterNarrative(snap)
	snap2 := reconstituted.Snapshot()

	if snap.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap.GameID.String() != snap2.GameID.String() {
		t.Fatal("GameID mismatch after round trip")
	}
	if len(snap.BeatIDs) != len(snap2.BeatIDs) {
		t.Fatal("BeatIDs length mismatch after round trip")
	}
	if len(snap.ActIDs) != len(snap2.ActIDs) {
		t.Fatal("ActIDs length mismatch after round trip")
	}
	if len(snap.SecretIDs) != len(snap2.SecretIDs) {
		t.Fatal("SecretIDs length mismatch after round trip")
	}
	if len(snap.LoreIDs) != len(snap2.LoreIDs) {
		t.Fatal("LoreIDs length mismatch after round trip")
	}
}
