package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestReconstituteFaction_NewState(t *testing.T) {
	snap := entity.FactionSnapshot{
		ID:    identity.NewFactionID(),
		GameID: identity.NewGameID(),
		Name:  "Faction",
		State: "new",
	}
	f, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(entity.NewFaction); !ok {
		t.Fatalf("expected NewFaction, got %T", f)
	}
}

func TestReconstituteFaction_DraftState(t *testing.T) {
	snap := entity.FactionSnapshot{
		ID:    identity.NewFactionID(),
		GameID: identity.NewGameID(),
		Name:  "Faction",
		State: "draft",
	}
	f, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", f)
	}
}

func TestReconstituteFaction_ActiveState(t *testing.T) {
	sl, _ := value.NewStandingLevel("Neutral", 0, 0)
	snap := entity.FactionSnapshot{
		ID:             identity.NewFactionID(),
		GameID:         identity.NewGameID(),
		Name:           "Faction",
		State:          "active",
		MemberIDs:      []identity.FactionMembershipID{identity.NewFactionMembershipID()},
		StandingLevels: []value.StandingLevel{sl},
	}
	f, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(entity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", f)
	}
}

func TestReconstituteFaction_IdleState(t *testing.T) {
	snap := entity.FactionSnapshot{
		ID:    identity.NewFactionID(),
		GameID: identity.NewGameID(),
		Name:  "Faction",
		State: "idle",
	}
	f, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(entity.IdleFaction); !ok {
		t.Fatalf("expected IdleFaction, got %T", f)
	}
}

func TestReconstituteFaction_ArchivedState(t *testing.T) {
	snap := entity.FactionSnapshot{
		ID:    identity.NewFactionID(),
		GameID: identity.NewGameID(),
		Name:  "Faction",
		State: "archived",
	}
	f, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := f.(entity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", f)
	}
}

func TestReconstituteFaction_UnknownState_ReturnsError(t *testing.T) {
	snap := entity.FactionSnapshot{
		ID:    identity.NewFactionID(),
		GameID: identity.NewGameID(),
		Name:  "Faction",
		State: "bogus",
	}
	_, err := entity.ReconstituteFaction(snap)
	if !errors.Is(err, entity.ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got %v", err)
	}
}

func TestReconstituteFaction_Roundtrip(t *testing.T) {
	// Create a faction, move to draft, add member, snapshot, reconstitute
	f := createFaction(t)
	d, _, _ := f.BeginDraft(event.SourceGrimoire)
	memberID := identity.NewFactionMembershipID()
	d, _, _ = d.AddMember(memberID, event.SourceGrimoire)

	snap := d.Snapshot()
	restored, err := entity.ReconstituteFaction(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	restoredSnap := restored.Snapshot()
	if restoredSnap.Name != snap.Name {
		t.Fatalf("expected name %q, got %q", snap.Name, restoredSnap.Name)
	}
	if restoredSnap.State != "draft" {
		t.Fatalf("expected state 'draft', got %q", restoredSnap.State)
	}
	if len(restoredSnap.MemberIDs) != 1 {
		t.Fatalf("expected 1 member, got %d", len(restoredSnap.MemberIDs))
	}
}
