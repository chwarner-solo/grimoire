package entity_test

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestReconstituteLocation_NewState(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeSettlement,
		Name:         "Location",
		State:        "new",
	}
	l, err := entity.ReconstituteLocation(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := l.(entity.NewLocation); !ok {
		t.Fatalf("expected NewLocation, got %T", l)
	}
}

func TestReconstituteLocation_DraftState(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeRegion,
		Name:         "Location",
		State:        "draft",
	}
	l, err := entity.ReconstituteLocation(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := l.(entity.DraftLocation); !ok {
		t.Fatalf("expected DraftLocation, got %T", l)
	}
}

func TestReconstituteLocation_ActiveState(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeSettlement,
		Name:         "Location",
		State:        "active",
		SceneIDs:     []identity.SceneID{identity.NewSceneID()},
	}
	l, err := entity.ReconstituteLocation(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := l.(entity.ActiveLocation); !ok {
		t.Fatalf("expected ActiveLocation, got %T", l)
	}
}

func TestReconstituteLocation_IdleState(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeDungeon,
		Name:         "Location",
		State:        "idle",
	}
	l, err := entity.ReconstituteLocation(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := l.(entity.IdleLocation); !ok {
		t.Fatalf("expected IdleLocation, got %T", l)
	}
}

func TestReconstituteLocation_ArchivedState(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeWilderness,
		Name:         "Location",
		State:        "archived",
	}
	l, err := entity.ReconstituteLocation(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := l.(entity.ArchivedLocation); !ok {
		t.Fatalf("expected ArchivedLocation, got %T", l)
	}
}

func TestReconstituteLocation_UnknownState_ReturnsError(t *testing.T) {
	snap := entity.LocationSnapshot{
		ID:           identity.NewLocationID(),
		GameID:       identity.NewGameID(),
		LocationType: value.LocationTypeWorld,
		Name:         "Location",
		State:        "bogus",
	}
	_, err := entity.ReconstituteLocation(snap)
	if !errors.Is(err, entity.ErrInvalidLocationState) {
		t.Fatalf("expected ErrInvalidLocationState, got %v", err)
	}
}

func TestReconstituteLocation_Roundtrip(t *testing.T) {
	l := createLocation(t)
	d, _, _ := l.BeginDraft(event.SourceGrimoire)
	sceneID := identity.NewSceneID()
	d, _, _ = d.AddScene(sceneID, event.SourceGrimoire)

	snap := d.Snapshot()
	restored, err := entity.ReconstituteLocation(snap)
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
	if len(restoredSnap.SceneIDs) != 1 {
		t.Fatalf("expected 1 scene, got %d", len(restoredSnap.SceneIDs))
	}
}
