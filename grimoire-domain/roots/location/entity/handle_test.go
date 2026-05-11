package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Handle helpers ---

func newLocationForHandle(t *testing.T) entity.Location {
	t.Helper()
	l, _, err := entity.CreateLocation(identity.NewLocationID(), "Test Location", identity.NewGameID(), value.LocationTypeSettlement, identity.LocationID{}, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	return l
}

func draftLocationForHandle(t *testing.T) entity.Location {
	t.Helper()
	l := newLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(),
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

func activeLocationForHandle(t *testing.T) entity.Location {
	t.Helper()
	l := draftLocationForHandle(t)
	sceneID := identity.NewSceneID()
	l2, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: sceneID.String(),
		Relationship: "has_scene", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	l3, err := l2.Handle(event.EntityUpdated{
		EntityID: l2.LocationID().String(), Field: "status",
		OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return l3
}

func idleLocationForHandle(t *testing.T) entity.Location {
	t.Helper()
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

// --- Handle tests: NewLocation ---

func TestHandle_NewLocation_StatusDraft_TransitionsToDraft(t *testing.T) {
	l := newLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftLocation); !ok {
		t.Fatalf("expected DraftLocation, got %T", result)
	}
}

func TestHandle_NewLocation_UnexpectedEvent_ReturnsError(t *testing.T) {
	l := newLocationForHandle(t)
	_, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: identity.NewSceneID().String(),
		Relationship: "has_scene", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

// --- Handle tests: DraftLocation ---

func TestHandle_DraftLocation_AddScene(t *testing.T) {
	l := draftLocationForHandle(t)
	sceneID := identity.NewSceneID()
	result, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: sceneID.String(),
		Relationship: "has_scene", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftLocation); !ok {
		t.Fatalf("expected DraftLocation, got %T", result)
	}
}

func TestHandle_DraftLocation_AddChild(t *testing.T) {
	l := draftLocationForHandle(t)
	childID := identity.NewLocationID()
	result, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: childID.String(),
		Relationship: "contains", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftLocation); !ok {
		t.Fatalf("expected DraftLocation, got %T", result)
	}
}

func TestHandle_DraftLocation_Activate(t *testing.T) {
	l := draftLocationForHandle(t)
	// Add scene first
	l2, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: identity.NewSceneID().String(),
		Relationship: "has_scene", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Activate
	result, err := l2.Handle(event.EntityUpdated{
		EntityID: l2.LocationID().String(), Field: "status",
		OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", result)
	}
}

// --- Handle tests: ActiveLocation ---

func TestHandle_ActiveLocation_AddScene(t *testing.T) {
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: identity.NewSceneID().String(),
		Relationship: "has_scene", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", result)
	}
}

func TestHandle_ActiveLocation_AddChild(t *testing.T) {
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: identity.NewLocationID().String(),
		Relationship: "contains", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", result)
	}
}

func TestHandle_ActiveLocation_ConnectTo(t *testing.T) {
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityLinked{
		EntityAID: l.LocationID().String(), EntityBID: identity.NewLocationID().String(),
		Relationship: "connects_to", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", result)
	}
}

func TestHandle_ActiveLocation_GoIdle(t *testing.T) {
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleLocation); !ok {
		t.Fatalf("expected IdleLocation, got %T", result)
	}
}

func TestHandle_ActiveLocation_Archive(t *testing.T) {
	l := activeLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "active", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedLocation); !ok {
		t.Fatalf("expected ArchivedLocation, got %T", result)
	}
}

// --- Handle tests: IdleLocation ---

func TestHandle_IdleLocation_Reactivate(t *testing.T) {
	l := idleLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "idle", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", result)
	}
}

func TestHandle_IdleLocation_Archive(t *testing.T) {
	l := idleLocationForHandle(t)
	result, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedLocation); !ok {
		t.Fatalf("expected ArchivedLocation, got %T", result)
	}
}

// --- Handle tests: ArchivedLocation ---

func TestHandle_ArchivedLocation_AnyEvent_ReturnsError(t *testing.T) {
	l := activeLocationForHandle(t)
	archived, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "active", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = archived.Handle(event.EntityUpdated{
		EntityID: archived.LocationID().String(), Field: "status",
		OldValue: "archived", NewValue: "active", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

// --- Immutability ---

func TestHandle_Immutability_OriginalUnchanged(t *testing.T) {
	l := newLocationForHandle(t)
	snapBefore := l.Snapshot()

	_, err := l.Handle(event.EntityUpdated{
		EntityID: l.LocationID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	snapAfter := l.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

// --- ReplayLocation ---

func TestReplayLocation_FullLifecycle(t *testing.T) {
	locationID := identity.NewLocationID()
	gameID := identity.NewGameID()
	sceneID := identity.NewSceneID()
	childID := identity.NewLocationID()
	connID := identity.NewLocationID()

	events := []event.Event{
		event.EntityCreated{EntityID: locationID.String(), EntityType: "location", Name: "Waterdeep", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: locationID.String(), Field: "status", OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: locationID.String(), EntityBID: sceneID.String(), Relationship: "has_scene", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: locationID.String(), EntityBID: childID.String(), Relationship: "contains", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: locationID.String(), Field: "status", OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: locationID.String(), EntityBID: connID.String(), Relationship: "connects_to", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: locationID.String(), Field: "status", OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: locationID.String(), Field: "status", OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire},
	}

	result, err := entity.ReplayLocation(gameID, value.LocationTypeSettlement, identity.LocationID{}, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedLocation); !ok {
		t.Fatalf("expected ArchivedLocation, got %T", result)
	}
	if result.LocationName() != "Waterdeep" {
		t.Fatalf("expected name 'Waterdeep', got %q", result.LocationName())
	}
	snap := result.Snapshot()
	if len(snap.SceneIDs) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(snap.SceneIDs))
	}
	if len(snap.ChildLocationIDs) != 1 {
		t.Fatalf("expected 1 child, got %d", len(snap.ChildLocationIDs))
	}
	if len(snap.ConnectedLocationIDs) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(snap.ConnectedLocationIDs))
	}
}

func TestReplayLocation_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := entity.ReplayLocation(identity.NewGameID(), value.LocationTypeRegion, identity.LocationID{}, nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
