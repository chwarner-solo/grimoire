package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Handle helpers ---

func newFactionForHandle(t *testing.T) entity.Faction {
	t.Helper()
	f, _, err := entity.CreateFaction(identity.NewFactionID(), "Test Faction", identity.NewGameID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create faction: %v", err)
	}
	return f
}

func draftFactionForHandle(t *testing.T) entity.Faction {
	t.Helper()
	f := newFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "draft",
		Source:   event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("failed to reach draft: %v", err)
	}
	return result
}

func activeFactionForHandle(t *testing.T) entity.Faction {
	t.Helper()
	f := draftFactionForHandle(t)
	memberID := identity.NewFactionMembershipID()
	f2, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: memberID.String(),
		Relationship: "member_of", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	f3, err := f2.Handle(event.EntityUpdated{
		EntityID: f2.FactionID().String(), Field: "standing_levels",
		OldValue: "", NewValue: "Neutral", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	f4, err := f3.Handle(event.EntityUpdated{
		EntityID: f3.FactionID().String(), Field: "status",
		OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return f4
}

func idleFactionForHandle(t *testing.T) entity.Faction {
	t.Helper()
	f := activeFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// --- Handle tests: NewFaction ---

func TestHandle_NewFaction_StatusDraft_TransitionsToDraft(t *testing.T) {
	f := newFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result)
	}
}

func TestHandle_NewFaction_UnexpectedEvent_ReturnsError(t *testing.T) {
	f := newFactionForHandle(t)
	_, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: identity.NewFactionMembershipID().String(),
		Relationship: "member_of", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

// --- Handle tests: DraftFaction ---

func TestHandle_DraftFaction_AddMember(t *testing.T) {
	f := draftFactionForHandle(t)
	memberID := identity.NewFactionMembershipID()
	result, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: memberID.String(),
		Relationship: "member_of", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result)
	}
}

func TestHandle_DraftFaction_AddAlly(t *testing.T) {
	f := draftFactionForHandle(t)
	allyID := identity.NewFactionID()
	result, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: allyID.String(),
		Relationship: "allied_with", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result)
	}
}

func TestHandle_DraftFaction_DeclareWar(t *testing.T) {
	f := draftFactionForHandle(t)
	enemyID := identity.NewFactionID()
	result, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: enemyID.String(),
		Relationship: "at_war_with", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result)
	}
}

func TestHandle_DraftFaction_AddStandingLevel(t *testing.T) {
	f := draftFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "standing_levels",
		OldValue: "", NewValue: "Kindly", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result)
	}
}

func TestHandle_DraftFaction_Activate(t *testing.T) {
	f := draftFactionForHandle(t)
	// Add member
	f2, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: identity.NewFactionMembershipID().String(),
		Relationship: "member_of", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Add standing level
	f3, err := f2.Handle(event.EntityUpdated{
		EntityID: f2.FactionID().String(), Field: "standing_levels",
		OldValue: "", NewValue: "Neutral", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Activate
	result, err := f3.Handle(event.EntityUpdated{
		EntityID: f3.FactionID().String(), Field: "status",
		OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", result)
	}
}

// --- Handle tests: ActiveFaction ---

func TestHandle_ActiveFaction_MarkDormant(t *testing.T) {
	f := activeFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleFaction); !ok {
		t.Fatalf("expected IdleFaction, got %T", result)
	}
}

func TestHandle_ActiveFaction_Archive(t *testing.T) {
	f := activeFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "active", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", result)
	}
}

func TestHandle_ActiveFaction_AddMember(t *testing.T) {
	f := activeFactionForHandle(t)
	result, err := f.Handle(event.EntityLinked{
		EntityAID: f.FactionID().String(), EntityBID: identity.NewFactionMembershipID().String(),
		Relationship: "member_of", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", result)
	}
}

// --- Handle tests: IdleFaction ---

func TestHandle_IdleFaction_Reactivate(t *testing.T) {
	f := idleFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "idle", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", result)
	}
}

func TestHandle_IdleFaction_Archive(t *testing.T) {
	f := idleFactionForHandle(t)
	result, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", result)
	}
}

// --- Handle tests: ArchivedFaction ---

func TestHandle_ArchivedFaction_AnyEvent_ReturnsError(t *testing.T) {
	f := activeFactionForHandle(t)
	archived, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "active", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = archived.Handle(event.EntityUpdated{
		EntityID: archived.FactionID().String(), Field: "status",
		OldValue: "archived", NewValue: "active", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

// --- Immutability ---

func TestHandle_Immutability_OriginalUnchanged(t *testing.T) {
	f := newFactionForHandle(t)
	snapBefore := f.Snapshot()

	_, err := f.Handle(event.EntityUpdated{
		EntityID: f.FactionID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	snapAfter := f.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

// --- ReplayFaction ---

func TestReplayFaction_FullLifecycle(t *testing.T) {
	factionID := identity.NewFactionID()
	gameID := identity.NewGameID()
	memberID := identity.NewFactionMembershipID()

	events := []event.Event{
		event.EntityCreated{EntityID: factionID.String(), EntityType: "faction", Name: "The Guild", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: factionID.String(), Field: "status", OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: factionID.String(), EntityBID: memberID.String(), Relationship: "member_of", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: factionID.String(), Field: "standing_levels", OldValue: "", NewValue: "Neutral", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: factionID.String(), Field: "status", OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: factionID.String(), Field: "status", OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: factionID.String(), Field: "status", OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire},
	}

	result, err := entity.ReplayFaction(gameID, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", result)
	}
	if result.FactionName() != "The Guild" {
		t.Fatalf("expected name 'The Guild', got %q", result.FactionName())
	}
}

func TestReplayFaction_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := entity.ReplayFaction(identity.NewGameID(), nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
