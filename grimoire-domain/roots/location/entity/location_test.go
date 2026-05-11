package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- helpers ---

func createLocation(t *testing.T) entity.NewLocation {
	t.Helper()
	l, _, err := entity.CreateLocation(identity.NewLocationID(), "Test Location", identity.NewGameID(), value.LocationTypeSettlement, identity.LocationID{}, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create location: %v", err)
	}
	return l
}

func draftLocation(t *testing.T) entity.DraftLocation {
	t.Helper()
	l := createLocation(t)
	d, _, err := l.BeginDraft(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to begin draft: %v", err)
	}
	return d
}

func activeLocation(t *testing.T) entity.ActiveLocation {
	t.Helper()
	d := draftLocation(t)
	d, _, _ = d.AddScene(identity.NewSceneID(), event.SourceGrimoire)
	a, _, err := d.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to activate: %v", err)
	}
	return a
}

func idleLocation(t *testing.T) entity.IdleLocation {
	t.Helper()
	a := activeLocation(t)
	i, _, err := a.GoIdle(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to go idle: %v", err)
	}
	return i
}

// --- Constructor tests ---

func TestCreateLocation_ValidInputs(t *testing.T) {
	id := identity.NewLocationID()
	gameID := identity.NewGameID()
	l, events, err := entity.CreateLocation(id, "Waterdeep", gameID, value.LocationTypeSettlement, identity.LocationID{}, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.LocationID().String() != id.String() {
		t.Fatalf("expected id %s, got %s", id, l.LocationID())
	}
	if l.LocationName() != "Waterdeep" {
		t.Fatalf("expected name 'Waterdeep', got %q", l.LocationName())
	}
	if l.GameID().String() != gameID.String() {
		t.Fatalf("expected gameID %s, got %s", gameID, l.GameID())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestCreateLocation_RequiresNonZeroID(t *testing.T) {
	_, _, err := entity.CreateLocation(identity.LocationID{}, "Name", identity.NewGameID(), value.LocationTypeRegion, identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrLocationIDRequired) {
		t.Fatalf("expected ErrLocationIDRequired, got %v", err)
	}
}

func TestCreateLocation_RequiresNonEmptyName(t *testing.T) {
	_, _, err := entity.CreateLocation(identity.NewLocationID(), "  ", identity.NewGameID(), value.LocationTypeRegion, identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrLocationNameRequired) {
		t.Fatalf("expected ErrLocationNameRequired, got %v", err)
	}
}

func TestCreateLocation_RequiresGameID(t *testing.T) {
	_, _, err := entity.CreateLocation(identity.NewLocationID(), "Name", identity.GameID{}, value.LocationTypeRegion, identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got %v", err)
	}
}

func TestCreateLocation_RequiresLocationType(t *testing.T) {
	_, _, err := entity.CreateLocation(identity.NewLocationID(), "Name", identity.NewGameID(), "", identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrLocationTypeRequired) {
		t.Fatalf("expected ErrLocationTypeRequired, got %v", err)
	}
}

func TestCreateLocation_DefaultsPlayerVisibleFalse(t *testing.T) {
	l := createLocation(t)
	snap := l.Snapshot()
	if snap.PlayerVisible {
		t.Fatal("expected playerVisible to default to false")
	}
}

func TestCreateLocation_ProducesEntityCreatedEvent(t *testing.T) {
	_, events, _ := entity.CreateLocation(identity.NewLocationID(), "Place", identity.NewGameID(), value.LocationTypeWorld, identity.LocationID{}, event.SourceGrimoire)
	created, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if created.EntityType != "location" {
		t.Fatalf("expected entity_type 'location', got %q", created.EntityType)
	}
}

func TestCreateLocation_ParentLocationIDOptional(t *testing.T) {
	parentID := identity.NewLocationID()
	l, _, err := entity.CreateLocation(identity.NewLocationID(), "Child", identity.NewGameID(), value.LocationTypeBuilding, parentID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := l.Snapshot()
	if snap.ParentLocationID.String() != parentID.String() {
		t.Fatalf("expected parentLocationID %s, got %s", parentID, snap.ParentLocationID)
	}
}

// --- Transition tests ---

func TestNewLocation_BeginDraft_TransitionsToDraft(t *testing.T) {
	l := createLocation(t)
	d, events, err := l.BeginDraft(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

// --- DraftLocation tests ---

func TestDraftLocation_AddScene_Succeeds(t *testing.T) {
	d := draftLocation(t)
	sceneID := identity.NewSceneID()
	result, events, err := d.AddScene(sceneID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftLocation, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "has_scene" {
		t.Fatalf("expected relationship 'has_scene', got %q", linked.Relationship)
	}
}

func TestDraftLocation_AddScene_RejectsDuplicate(t *testing.T) {
	d := draftLocation(t)
	sceneID := identity.NewSceneID()
	d, _, _ = d.AddScene(sceneID, event.SourceGrimoire)
	_, _, err := d.AddScene(sceneID, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrSceneAlreadyExists) {
		t.Fatalf("expected ErrSceneAlreadyExists, got %v", err)
	}
}

func TestDraftLocation_AddChild_Succeeds(t *testing.T) {
	d := draftLocation(t)
	childID := identity.NewLocationID()
	result, events, err := d.AddChild(childID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected DraftLocation, got nil")
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "contains" {
		t.Fatalf("expected relationship 'contains', got %q", linked.Relationship)
	}
}

func TestDraftLocation_AddChild_RejectsDuplicate(t *testing.T) {
	d := draftLocation(t)
	childID := identity.NewLocationID()
	d, _, _ = d.AddChild(childID, event.SourceGrimoire)
	_, _, err := d.AddChild(childID, event.SourceGrimoire)
	if !errors.Is(err, entity.ErrChildAlreadyExists) {
		t.Fatalf("expected ErrChildAlreadyExists, got %v", err)
	}
}

func TestDraftLocation_Activate_RequiresScene(t *testing.T) {
	d := draftLocation(t)
	_, _, err := d.Activate(event.SourceGrimoire)
	if !errors.Is(err, entity.ErrCannotActivateWithoutScenes) {
		t.Fatalf("expected ErrCannotActivateWithoutScenes, got %v", err)
	}
}

func TestDraftLocation_Activate_SucceedsWithScene(t *testing.T) {
	d := draftLocation(t)
	d, _, _ = d.AddScene(identity.NewSceneID(), event.SourceGrimoire)
	a, events, err := d.Activate(event.SourceGrimoire)
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

// --- ActiveLocation tests ---

func TestActiveLocation_ConnectTo_Succeeds(t *testing.T) {
	a := activeLocation(t)
	targetID := identity.NewLocationID()
	result, events, err := a.ConnectTo(targetID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected ActiveLocation, got nil")
	}
	linked, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if linked.Relationship != "connects_to" {
		t.Fatalf("expected relationship 'connects_to', got %q", linked.Relationship)
	}
}

func TestActiveLocation_ConnectTo_RejectsSelf(t *testing.T) {
	a := activeLocation(t)
	_, _, err := a.ConnectTo(a.LocationID(), event.SourceGrimoire)
	if !errors.Is(err, entity.ErrCannotConnectToSelf) {
		t.Fatalf("expected ErrCannotConnectToSelf, got %v", err)
	}
}

func TestActiveLocation_ConnectTo_RejectsDuplicate(t *testing.T) {
	a := activeLocation(t)
	targetID := identity.NewLocationID()
	a, _, _ = a.ConnectTo(targetID, event.SourceGrimoire)
	_, _, err := a.ConnectTo(targetID, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error for duplicate connection")
	}
	var typedErr entity.ErrAlreadyConnectedTo
	if !errors.As(err, &typedErr) {
		t.Fatalf("expected ErrAlreadyConnectedTo, got %v", err)
	}
}

func TestActiveLocation_GoIdle(t *testing.T) {
	a := activeLocation(t)
	i, events, err := a.GoIdle(event.SourceGrimoire)
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

func TestActiveLocation_Archive(t *testing.T) {
	a := activeLocation(t)
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

func TestActiveLocation_AddScene(t *testing.T) {
	a := activeLocation(t)
	sceneID := identity.NewSceneID()
	result, events, err := a.AddScene(sceneID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected ActiveLocation, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestActiveLocation_AddChild(t *testing.T) {
	a := activeLocation(t)
	childID := identity.NewLocationID()
	result, events, err := a.AddChild(childID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected ActiveLocation, got nil")
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

// --- IdleLocation tests ---

func TestIdleLocation_Reactivate(t *testing.T) {
	i := idleLocation(t)
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

func TestIdleLocation_Archive(t *testing.T) {
	i := idleLocation(t)
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

// --- ArchivedLocation tests ---

func TestArchivedLocation_PreservesIdentity(t *testing.T) {
	a := activeLocation(t)
	arch, _, _ := a.Archive(event.SourceGrimoire)
	if arch.LocationID().IsZero() {
		t.Fatal("expected non-zero LocationID on archived location")
	}
	if arch.LocationName() == "" {
		t.Fatal("expected non-empty name on archived location")
	}
	if arch.GameID().IsZero() {
		t.Fatal("expected non-zero GameID on archived location")
	}
}
