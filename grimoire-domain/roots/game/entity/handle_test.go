package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- handle test helpers ---

func newGameForHandle(t *testing.T) entity.Game {
	t.Helper()
	id := identity.NewGameID()
	g, _, err := entity.CreateGame(id, "Test Game", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	return g
}

func draftGameForHandle(t *testing.T) entity.Game {
	t.Helper()
	g := newGameForHandle(t)
	result, err := g.Handle(event.EntityCreated{
		EntityID:   identity.NewNarrativeID().String(),
		EntityType: "narrative",
		Name:       "Lore Entry",
		Source:     event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("failed to reach draft: %v", err)
	}
	return result
}

func activeGameForHandle(t *testing.T) entity.Game {
	t.Helper()
	g := draftGameForHandle(t)
	campID := identity.NewCampaignID()
	g2, err := g.Handle(event.EntityLinked{
		EntityAID: g.GameID().String(), EntityBID: campID.String(),
		Relationship: "campaign", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	g3, err := g2.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: campID.String(),
		Date: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return g3
}

func idleGameForHandle(t *testing.T) entity.Game {
	t.Helper()
	g := activeGameForHandle(t)
	g2, err := g.Handle(event.SessionEnded{SessionID: identity.NewSessionID().String()})
	if err != nil {
		t.Fatal(err)
	}
	return g2
}

// --- tests ---

func TestHandle_NewGame_EntityCreated_TransitionsToDraft(t *testing.T) {
	g := newGameForHandle(t)
	evt := event.EntityCreated{
		EntityID:   identity.NewNarrativeID().String(),
		EntityType: "narrative",
		Name:       "The Dark Forest",
		Source:     event.SourceGrimoire,
	}

	result, err := g.Handle(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftGame); !ok {
		t.Fatalf("expected DraftGame, got %T", result)
	}
}

func TestHandle_NewGame_UnexpectedEvent_ReturnsError(t *testing.T) {
	g := newGameForHandle(t)
	evt := event.SessionStarted{
		SessionID:  identity.NewSessionID().String(),
		CampaignID: identity.NewCampaignID().String(),
		Date:       time.Now(),
	}

	_, err := g.Handle(evt)
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_DraftGame_EntityLinked_AddsCampaign(t *testing.T) {
	g := draftGameForHandle(t)
	campID := identity.NewCampaignID()
	evt := event.EntityLinked{
		EntityAID:    g.GameID().String(),
		EntityBID:    campID.String(),
		Relationship: "campaign",
		Source:       event.SourceGrimoire,
	}

	result, err := g.Handle(evt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.DraftGame); !ok {
		t.Fatalf("expected DraftGame, got %T", result)
	}
}

func TestHandle_DraftGame_SessionStarted_TransitionsToActive(t *testing.T) {
	g := draftGameForHandle(t)
	campID := identity.NewCampaignID()

	linked, err := g.Handle(event.EntityLinked{
		EntityAID:    g.GameID().String(),
		EntityBID:    campID.String(),
		Relationship: "campaign",
		Source:       event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error linking campaign: %v", err)
	}

	result, err := linked.Handle(event.SessionStarted{
		SessionID:  identity.NewSessionID().String(),
		CampaignID: campID.String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveGame); !ok {
		t.Fatalf("expected ActiveGame, got %T", result)
	}
}

func TestHandle_ActiveGame_SessionEnded_TransitionsToIdle_WhenLastCampaign(t *testing.T) {
	g := activeGameForHandle(t)

	result, err := g.Handle(event.SessionEnded{
		SessionID: identity.NewSessionID().String(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleGame); !ok {
		t.Fatalf("expected IdleGame, got %T", result)
	}
}

func TestHandle_ActiveGame_SessionEnded_StaysActive_WhenOtherCampaigns(t *testing.T) {
	g := draftGameForHandle(t)
	camp1 := identity.NewCampaignID()
	camp2 := identity.NewCampaignID()

	g2, err := g.Handle(event.EntityLinked{
		EntityAID: g.GameID().String(), EntityBID: camp1.String(),
		Relationship: "campaign", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	g3, err := g2.Handle(event.EntityLinked{
		EntityAID: g2.GameID().String(), EntityBID: camp2.String(),
		Relationship: "campaign", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	g4, err := g3.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: camp1.String(),
		Date: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	g5, err := g4.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: camp2.String(),
		Date: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	g6, err := g5.Handle(event.SessionEnded{SessionID: identity.NewSessionID().String()})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g6.(entity.ActiveGame); !ok {
		t.Fatalf("expected ActiveGame, got %T", g6)
	}
}

func TestHandle_IdleGame_SessionStarted_TransitionsToActive(t *testing.T) {
	g := idleGameForHandle(t)

	result, err := g.Handle(event.SessionStarted{
		SessionID:  identity.NewSessionID().String(),
		CampaignID: identity.NewCampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveGame); !ok {
		t.Fatalf("expected ActiveGame, got %T", result)
	}
}

func TestHandle_IdleGame_EntityUpdated_Archived_TransitionsToArchived(t *testing.T) {
	g := idleGameForHandle(t)

	result, err := g.Handle(event.EntityUpdated{
		EntityID: g.GameID().String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "archived",
		Source:   event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedGame); !ok {
		t.Fatalf("expected ArchivedGame, got %T", result)
	}
}

func TestHandle_ArchivedGame_AnyEvent_ReturnsError(t *testing.T) {
	g := idleGameForHandle(t)
	archived, err := g.Handle(event.EntityUpdated{
		EntityID: g.GameID().String(), Field: "status",
		OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = archived.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: identity.NewCampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_Immutability_OriginalUnchanged(t *testing.T) {
	g := newGameForHandle(t)
	snapBefore := g.Snapshot()

	_, err := g.Handle(event.EntityCreated{
		EntityID:   identity.NewNarrativeID().String(),
		EntityType: "narrative",
		Name:       "Some Lore",
		Source:     event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	snapAfter := g.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

func TestReplayGame_FullLifecycle(t *testing.T) {
	gameID := identity.NewGameID()
	campID := identity.NewCampaignID()
	sessID := identity.NewSessionID()

	events := []event.Event{
		event.EntityCreated{EntityID: gameID.String(), EntityType: "game", Name: "My Game", Source: event.SourceGrimoire},
		event.EntityCreated{EntityID: identity.NewNarrativeID().String(), EntityType: "narrative", Name: "Lore", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: gameID.String(), EntityBID: campID.String(), Relationship: "campaign", Source: event.SourceGrimoire},
		event.SessionStarted{SessionID: sessID.String(), CampaignID: campID.String(), Date: time.Now()},
		event.SessionEnded{SessionID: sessID.String()},
		event.EntityUpdated{EntityID: gameID.String(), Field: "status", OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire},
	}

	result, err := entity.ReplayGame(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ArchivedGame); !ok {
		t.Fatalf("expected ArchivedGame, got %T", result)
	}
	if result.GameName() != "My Game" {
		t.Fatalf("expected name 'My Game', got '%s'", result.GameName())
	}
}

func TestReplayGame_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := entity.ReplayGame(nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
