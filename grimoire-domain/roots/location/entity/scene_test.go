package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestCreateScene_ValidInputs(t *testing.T) {
	id := identity.NewSceneID()
	locID := identity.NewLocationID()
	gameID := identity.NewGameID()
	s, events, err := CreateScene(id, locID, gameID, "Throne Room", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SceneID().String() != id.String() {
		t.Fatalf("expected id %s, got %s", id, s.SceneID())
	}
	if s.LocationID().String() != locID.String() {
		t.Fatalf("expected locationID %s, got %s", locID, s.LocationID())
	}
	if s.GameID().String() != gameID.String() {
		t.Fatalf("expected gameID %s, got %s", gameID, s.GameID())
	}
	if s.Name() != "Throne Room" {
		t.Fatalf("expected name 'Throne Room', got %q", s.Name())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityType != "scene" {
		t.Fatalf("expected entity_type 'scene', got %q", ec.EntityType)
	}
}

func TestCreateScene_RequiresSceneID(t *testing.T) {
	_, _, err := CreateScene(identity.SceneID{}, identity.NewLocationID(), identity.NewGameID(), "Name", event.SourceGrimoire)
	if !errors.Is(err, ErrSceneIDRequired) {
		t.Fatalf("expected ErrSceneIDRequired, got %v", err)
	}
}

func TestCreateScene_RequiresLocationID(t *testing.T) {
	_, _, err := CreateScene(identity.NewSceneID(), identity.LocationID{}, identity.NewGameID(), "Name", event.SourceGrimoire)
	if !errors.Is(err, ErrLocationIDRequired) {
		t.Fatalf("expected ErrLocationIDRequired, got %v", err)
	}
}

func TestCreateScene_RequiresGameID(t *testing.T) {
	_, _, err := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.GameID{}, "Name", event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got %v", err)
	}
}

func TestCreateScene_RequiresName(t *testing.T) {
	_, _, err := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "  ", event.SourceGrimoire)
	if !errors.Is(err, ErrSceneNameRequired) {
		t.Fatalf("expected ErrSceneNameRequired, got %v", err)
	}
}

func TestCreateScene_DefaultsPlayerVisibleFalse(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	if s.IsPlayerVisible() {
		t.Fatal("expected playerVisible to default to false")
	}
}

func TestScene_UpdateContent(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Original", event.SourceGrimoire)
	events, err := s.UpdateContent("New Name", "desc", "player desc", "explored desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	if s.Name() != "New Name" {
		t.Fatalf("expected name 'New Name', got %q", s.Name())
	}
	if s.Description() != "desc" {
		t.Fatalf("expected description 'desc', got %q", s.Description())
	}
}

func TestScene_UpdateContent_NoChangeProducesNoEvents(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Name", event.SourceGrimoire)
	events, _ := s.UpdateContent("Name", "", "", "")
	if len(events) != 0 {
		t.Fatalf("expected 0 events for no-op update, got %d", len(events))
	}
}

func TestScene_AddPrerequisite(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	prereqID := identity.NewSceneID()
	s.AddPrerequisite(prereqID)
	sets := s.PrerequisiteSets()
	if len(sets) != 1 || len(sets[0]) != 1 {
		t.Fatalf("expected 1 set with 1 prereq, got %v", sets)
	}
	if sets[0][0].String() != prereqID.String() {
		t.Fatalf("expected prereq %s, got %s", prereqID, sets[0][0])
	}
}

func TestScene_AddRecommendedBeat(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	beatID := identity.NewBeatID()
	s.AddRecommendedBeat(beatID)
	beats := s.RecommendedBeatIDs()
	if len(beats) != 1 {
		t.Fatalf("expected 1 recommended beat, got %d", len(beats))
	}
	if beats[0].String() != beatID.String() {
		t.Fatalf("expected beat %s, got %s", beatID, beats[0])
	}
}

func TestScene_WouldCreateCycle_DetectsCycle(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	targetID := identity.NewSceneID()
	chain := []identity.SceneID{s.SceneID()}
	err := s.WouldCreateCycle(targetID, chain)
	if err == nil {
		t.Fatal("expected cycle error")
	}
	var cycleErr ErrSceneCycleDetected
	if !errors.As(err, &cycleErr) {
		t.Fatalf("expected ErrSceneCycleDetected, got %v", err)
	}
}

func TestScene_WouldCreateCycle_NoCycle(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	targetID := identity.NewSceneID()
	chain := []identity.SceneID{identity.NewSceneID()} // different scene in chain
	err := s.WouldCreateCycle(targetID, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestScene_Snapshot_Roundtrip(t *testing.T) {
	s, _, _ := CreateScene(identity.NewSceneID(), identity.NewLocationID(), identity.NewGameID(), "Scene", event.SourceGrimoire)
	prereqID := identity.NewSceneID()
	s.AddPrerequisite(prereqID)
	beatID := identity.NewBeatID()
	s.AddRecommendedBeat(beatID)
	s.UpdateContent("Scene", "desc", "pdesc", "edesc")

	snap := s.Snapshot()
	restored := ReconstituteScene(snap)
	snap2 := restored.Snapshot()

	if snap.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap.LocationID.String() != snap2.LocationID.String() {
		t.Fatal("LocationID mismatch after round trip")
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
	if snap.ExploredDescription != snap2.ExploredDescription {
		t.Fatal("ExploredDescription mismatch after round trip")
	}
	if len(snap.ScenePrerequisiteSets) != len(snap2.ScenePrerequisiteSets) {
		t.Fatal("PrerequisiteSets length mismatch")
	}
	if len(snap.RecommendedBeatIDs) != len(snap2.RecommendedBeatIDs) {
		t.Fatal("RecommendedBeatIDs length mismatch")
	}
}
