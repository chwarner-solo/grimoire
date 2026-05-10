package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestNewFactionMembership_ValidInputs_Succeeds(t *testing.T) {
	m, err := entity.NewFactionMembership(
		identity.NewFactionMembershipID(),
		identity.NewFactionID(),
		identity.NewCharacterID(),
		"Grunt",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Rank() != "Grunt" {
		t.Fatalf("expected rank 'Grunt', got %q", m.Rank())
	}
}

func TestNewFactionMembership_RequiresNonZeroMembershipID(t *testing.T) {
	var zeroID identity.FactionMembershipID
	_, err := entity.NewFactionMembership(zeroID, identity.NewFactionID(), identity.NewCharacterID(), "Grunt")
	if !errors.Is(err, entity.ErrMembershipIDRequired) {
		t.Fatalf("expected ErrMembershipIDRequired, got %v", err)
	}
}

func TestNewFactionMembership_RequiresNonZeroFactionID(t *testing.T) {
	var zeroID identity.FactionID
	_, err := entity.NewFactionMembership(identity.NewFactionMembershipID(), zeroID, identity.NewCharacterID(), "Grunt")
	if !errors.Is(err, entity.ErrFactionIDRequired) {
		t.Fatalf("expected ErrFactionIDRequired, got %v", err)
	}
}

func TestNewFactionMembership_RequiresNonZeroCharacterID(t *testing.T) {
	var zeroID identity.CharacterID
	_, err := entity.NewFactionMembership(identity.NewFactionMembershipID(), identity.NewFactionID(), zeroID, "Grunt")
	if !errors.Is(err, entity.ErrCharacterIDRequired) {
		t.Fatalf("expected ErrCharacterIDRequired, got %v", err)
	}
}

func TestNewFactionMembership_RequiresNonEmptyRank(t *testing.T) {
	_, err := entity.NewFactionMembership(identity.NewFactionMembershipID(), identity.NewFactionID(), identity.NewCharacterID(), "")
	if !errors.Is(err, entity.ErrRankRequired) {
		t.Fatalf("expected ErrRankRequired, got %v", err)
	}
}

func TestNewFactionMembership_DefaultsPlayerVisibleFalse(t *testing.T) {
	m, _ := entity.NewFactionMembership(identity.NewFactionMembershipID(), identity.NewFactionID(), identity.NewCharacterID(), "Elder")
	if m.IsPlayerVisible() {
		t.Fatal("expected playerVisible to default to false")
	}
}

func TestFactionMembership_Reveal_SetsPlayerVisible(t *testing.T) {
	m, _ := entity.NewFactionMembership(identity.NewFactionMembershipID(), identity.NewFactionID(), identity.NewCharacterID(), "Elder")
	m.Reveal()
	if !m.IsPlayerVisible() {
		t.Fatal("expected playerVisible to be true after Reveal")
	}
}

func TestFactionMembership_Snapshot_Roundtrip(t *testing.T) {
	m, _ := entity.NewFactionMembership(identity.NewFactionMembershipID(), identity.NewFactionID(), identity.NewCharacterID(), "Inquisitor")
	m.Reveal()
	snap := m.Snapshot()

	restored := entity.ReconstituteFactionMembership(snap)
	if restored.MembershipID().String() != m.MembershipID().String() {
		t.Fatalf("expected id %s, got %s", m.MembershipID(), restored.MembershipID())
	}
	if restored.Rank() != "Inquisitor" {
		t.Fatalf("expected rank 'Inquisitor', got %q", restored.Rank())
	}
	if !restored.IsPlayerVisible() {
		t.Fatal("expected playerVisible to be true after reconstitution")
	}
}
