package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- helpers ---

func createFaction(t *testing.T) entity.NewFaction {
	t.Helper()
	f, _, err := entity.CreateFaction(identity.NewFactionID(), "Test Faction", identity.NewGameID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	return f
}

func draftFaction(t *testing.T) entity.DraftFaction {
	t.Helper()
	f := createFaction(t)
	d, _, err := f.BeginDraft(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to begin draft: %v", err)
	}
	return d
}

func activeFaction(t *testing.T) entity.ActiveFaction {
	t.Helper()
	d := draftFaction(t)
	d, _, _ = d.AddMember(identity.NewFactionMembershipID(), event.SourceGrimoire)
	sl, _ := value.NewStandingLevel("Neutral", 0, 0)
	d, _, _ = d.AddStandingLevel(sl, event.SourceGrimoire)
	a, _, err := d.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to activate: %v", err)
	}
	return a
}

func idleFaction(t *testing.T) entity.IdleFaction {
	t.Helper()
	a := activeFaction(t)
	i, _, err := a.MarkDormant(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to mark dormant: %v", err)
	}
	return i
}

// --- Constructor tests ---

func TestCreateFaction_ValidInputs(t *testing.T) {
	id := identity.NewFactionID()
	gameID := identity.NewGameID()
	f, events, err := entity.CreateFaction(id, "The Inquisition", gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.FactionID().String() != id.String() {
		t.Fatalf("expected id %s, got %s", id, f.FactionID())
	}
	if f.FactionName() != "The Inquisition" {
		t.Fatalf("expected name 'The Inquisition', got %q", f.FactionName())
	}
	if f.GameID().String() != gameID.String() {
		t.Fatalf("expected gameID %s, got %s", gameID, f.GameID())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestCreateFaction_RequiresNonZeroID(t *testing.T) {
	var zeroID identity.FactionID
	_, _, err := entity.CreateFaction(zeroID, "Name", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, entity.ErrFactionIDRequired) {
		t.Fatalf("expected ErrFactionIDRequired, got %v", err)
	}
}

func TestCreateFaction_RequiresNonEmptyName(t *testing.T) {
	_, _, err := entity.CreateFaction(identity.NewFactionID(), "  ", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, entity.ErrFactionNameRequired) {
		t.Fatalf("expected ErrFactionNameRequired, got %v", err)
	}
}

func TestCreateFaction_RequiresGameID(t *testing.T) {
	var zeroID identity.GameID
	_, _, err := entity.CreateFaction(identity.NewFactionID(), "Name", zeroID, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got %v", err)
	}
}

func TestCreateFaction_DefaultsPlayerVisibleFalse(t *testing.T) {
	f := createFaction(t)
	snap := f.Snapshot()
	if snap.PlayerVisible {
		t.Fatal("expected playerVisible to default to false")
	}
}

func TestCreateFaction_ProducesEntityCreatedEvent(t *testing.T) {
	_, events, _ := entity.CreateFaction(identity.NewFactionID(), "Faction", identity.NewGameID(), event.SourceGrimoire)
	created, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if created.EntityType != "faction" {
		t.Fatalf("expected entity_type 'faction', got %q", created.EntityType)
	}
}

// --- Transition tests ---

func TestNewFaction_BeginDraft_TransitionsToDraft(t *testing.T) {
	f := createFaction(t)
	d, events, err := f.BeginDraft(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected DraftFaction, got nil")
	}
	if d.Snapshot().State != "draft" {
		t.Fatalf("expected state 'draft', got %q", d.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	updated, ok := events[0].(event.EntityUpdated)
	if !ok {
		t.Fatalf("expected EntityUpdated, got %T", events[0])
	}
	if updated.Field != "status" || updated.NewValue != "draft" {
		t.Fatalf("unexpected event: field=%q new=%q", updated.Field, updated.NewValue)
	}
}

// --- DraftFaction tests ---

func TestDraftFaction_AddMember_Succeeds(t *testing.T) {
	d := draftFaction(t)
	memberID := identity.NewFactionMembershipID()
	result, events, err := d.AddMember(memberID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftFaction, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "member_of" {
		t.Fatalf("expected relationship 'member_of', got %q", linked.Relationship)
	}
}

func TestDraftFaction_AddMember_RejectsDuplicate(t *testing.T) {
	d := draftFaction(t)
	memberID := identity.NewFactionMembershipID()
	d, _, _ = d.AddMember(memberID, event.SourceGrimoire)
	_, _, err := d.AddMember(memberID, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrMemberAlreadyExists) {
		t.Fatalf("expected ErrMemberAlreadyExists, got %v", err)
	}
}

func TestDraftFaction_AddStandingLevel_Succeeds(t *testing.T) {
	d := draftFaction(t)
	sl, _ := value.NewStandingLevel("Kindly", 500, 2)
	result, events, err := d.AddStandingLevel(sl, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftFaction, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestDraftFaction_AddAlly_Succeeds(t *testing.T) {
	d := draftFaction(t)
	allyID := identity.NewFactionID()
	result, events, err := d.AddAlly(allyID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftFaction, got nil")
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "allied_with" {
		t.Fatalf("expected relationship 'allied_with', got %q", linked.Relationship)
	}
}

func TestDraftFaction_DeclareWar_Succeeds(t *testing.T) {
	d := draftFaction(t)
	enemyID := identity.NewFactionID()
	result, events, err := d.DeclareWar(enemyID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftFaction, got nil")
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "at_war_with" {
		t.Fatalf("expected relationship 'at_war_with', got %q", linked.Relationship)
	}
}

func TestDraftFaction_AddAlly_RejectsIfAtWar(t *testing.T) {
	d := draftFaction(t)
	enemyID := identity.NewFactionID()
	d, _, _ = d.DeclareWar(enemyID, event.SourceGrimoire)
	_, _, err := d.AddAlly(enemyID, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error when allying with enemy")
	}
	var typedErr entity.ErrCannotBeAlliedAndAtWar
	if !errors.As(err, &typedErr) {
		t.Fatalf("expected ErrCannotBeAlliedAndAtWar, got %v", err)
	}
}

func TestDraftFaction_DeclareWar_RejectsIfAllied(t *testing.T) {
	d := draftFaction(t)
	allyID := identity.NewFactionID()
	d, _, _ = d.AddAlly(allyID, event.SourceGrimoire)
	_, _, err := d.DeclareWar(allyID, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error when declaring war on ally")
	}
	var typedErr entity.ErrCannotBeAtWarAndAllied
	if !errors.As(err, &typedErr) {
		t.Fatalf("expected ErrCannotBeAtWarAndAllied, got %v", err)
	}
}

func TestDraftFaction_Activate_RequiresMember(t *testing.T) {
	d := draftFaction(t)
	sl, _ := value.NewStandingLevel("Neutral", 0, 0)
	d, _, _ = d.AddStandingLevel(sl, event.SourceGrimoire)
	_, _, err := d.Activate(event.SourceGrimoire)
	if !errors.Is(err, entity.ErrCannotActivateWithoutMembers) {
		t.Fatalf("expected ErrCannotActivateWithoutMembers, got %v", err)
	}
}

func TestDraftFaction_Activate_RequiresStandingLevel(t *testing.T) {
	d := draftFaction(t)
	d, _, _ = d.AddMember(identity.NewFactionMembershipID(), event.SourceGrimoire)
	_, _, err := d.Activate(event.SourceGrimoire)
	if !errors.Is(err, entity.ErrCannotActivateWithoutStandingLevels) {
		t.Fatalf("expected ErrCannotActivateWithoutStandingLevels, got %v", err)
	}
}

func TestDraftFaction_Activate_SucceedsWithBoth(t *testing.T) {
	d := draftFaction(t)
	d, _, _ = d.AddMember(identity.NewFactionMembershipID(), event.SourceGrimoire)
	sl, _ := value.NewStandingLevel("Neutral", 0, 0)
	d, _, _ = d.AddStandingLevel(sl, event.SourceGrimoire)
	a, events, err := d.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected ActiveFaction, got nil")
	}
	if a.Snapshot().State != "active" {
		t.Fatalf("expected state 'active', got %q", a.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

// --- ActiveFaction tests ---

func TestActiveFaction_MarkDormant(t *testing.T) {
	a := activeFaction(t)
	i, events, err := a.MarkDormant(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i.Snapshot().State != "idle" {
		t.Fatalf("expected state 'idle', got %q", i.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestActiveFaction_Archive(t *testing.T) {
	a := activeFaction(t)
	arch, events, err := a.Archive(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arch.Snapshot().State != "archived" {
		t.Fatalf("expected state 'archived', got %q", arch.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestActiveFaction_AddMember(t *testing.T) {
	a := activeFaction(t)
	memberID := identity.NewFactionMembershipID()
	result, events, err := a.AddMember(memberID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected ActiveFaction, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

// --- IdleFaction tests ---

func TestIdleFaction_Reactivate(t *testing.T) {
	i := idleFaction(t)
	a, events, err := i.Reactivate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Snapshot().State != "active" {
		t.Fatalf("expected state 'active', got %q", a.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestIdleFaction_Archive(t *testing.T) {
	i := idleFaction(t)
	arch, events, err := i.Archive(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arch.Snapshot().State != "archived" {
		t.Fatalf("expected state 'archived', got %q", arch.Snapshot().State)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

// --- ArchivedFaction tests ---

func TestArchivedFaction_PreservesIdentity(t *testing.T) {
	a := activeFaction(t)
	arch, _, _ := a.Archive(event.SourceGrimoire)
	if arch.FactionID().IsZero() {
		t.Fatal("expected non-zero FactionID on archived faction")
	}
	if arch.FactionName() == "" {
		t.Fatal("expected non-empty name on archived faction")
	}
	if arch.GameID().IsZero() {
		t.Fatal("expected non-zero GameID on archived faction")
	}
}

// --- Duplicate alliance/war tests on active ---

func TestActiveFaction_AddAlly_RejectsIfAtWar(t *testing.T) {
	a := activeFaction(t)
	enemyID := identity.NewFactionID()
	a, _, _ = a.DeclareWar(enemyID, event.SourceGrimoire)
	_, _, err := a.AddAlly(enemyID, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error when allying with enemy")
	}
}

func TestActiveFaction_DeclareWar_RejectsIfAllied(t *testing.T) {
	a := activeFaction(t)
	allyID := identity.NewFactionID()
	a, _, _ = a.AddAlly(allyID, event.SourceGrimoire)
	_, _, err := a.DeclareWar(allyID, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error when declaring war on ally")
	}
}
