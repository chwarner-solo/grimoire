package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- handle test helpers ---

func newCampaignForHandle(t *testing.T) entity.Campaign {
	t.Helper()
	c, _, err := entity.CreateCampaign(identity.NewCampaignID(), "Test Campaign", identity.NewGameID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	return c
}

func formingCampaignForHandle(t *testing.T) entity.Campaign {
	t.Helper()
	c := newCampaignForHandle(t)
	// Add a character first
	c2, err := c.Handle(event.EntityLinked{
		EntityAID: c.CampaignID().String(), EntityBID: identity.NewCharacterID().String(),
		Relationship: "character", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Transition to forming
	c3, err := c2.Handle(event.EntityUpdated{
		EntityID: c2.CampaignID().String(), Field: "status",
		OldValue: "new", NewValue: "forming", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return c3
}

func activeCampaignForHandle(t *testing.T) entity.Campaign {
	t.Helper()
	c := formingCampaignForHandle(t)
	c2, err := c.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: c.CampaignID().String(),
		Date: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return c2
}

func idleCampaignForHandle(t *testing.T) entity.Campaign {
	t.Helper()
	c := activeCampaignForHandle(t)
	c2, err := c.Handle(event.SessionEnded{SessionID: identity.NewSessionID().String()})
	if err != nil {
		t.Fatal(err)
	}
	return c2
}

// --- tests ---

func TestHandle_NewCampaign_EntityLinked_AddsCharacter(t *testing.T) {
	c := newCampaignForHandle(t)
	result, err := c.Handle(event.EntityLinked{
		EntityAID:    c.CampaignID().String(),
		EntityBID:    identity.NewCharacterID().String(),
		Relationship: "character",
		Source:       event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.NewCampaign); !ok {
		t.Fatalf("expected NewCampaign, got %T", result)
	}
}

func TestHandle_NewCampaign_EntityUpdated_TransitionsToForming(t *testing.T) {
	c := newCampaignForHandle(t)
	result, err := c.Handle(event.EntityUpdated{
		EntityID: c.CampaignID().String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "forming",
		Source:   event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.FormingCampaign); !ok {
		t.Fatalf("expected FormingCampaign, got %T", result)
	}
}

func TestHandle_NewCampaign_UnexpectedEvent_ReturnsError(t *testing.T) {
	c := newCampaignForHandle(t)
	_, err := c.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: c.CampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_FormingCampaign_EntityLinked_AddsCharacter(t *testing.T) {
	c := formingCampaignForHandle(t)
	result, err := c.Handle(event.EntityLinked{
		EntityAID:    c.CampaignID().String(),
		EntityBID:    identity.NewCharacterID().String(),
		Relationship: "character",
		Source:       event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.FormingCampaign); !ok {
		t.Fatalf("expected FormingCampaign, got %T", result)
	}
}

func TestHandle_FormingCampaign_SessionStarted_TransitionsToActive(t *testing.T) {
	c := formingCampaignForHandle(t)
	result, err := c.Handle(event.SessionStarted{
		SessionID:  identity.NewSessionID().String(),
		CampaignID: c.CampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveCampaign); !ok {
		t.Fatalf("expected ActiveCampaign, got %T", result)
	}
}

func TestHandle_ActiveCampaign_SessionEnded_TransitionsToIdle(t *testing.T) {
	c := activeCampaignForHandle(t)
	result, err := c.Handle(event.SessionEnded{SessionID: identity.NewSessionID().String()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleCampaign); !ok {
		t.Fatalf("expected IdleCampaign, got %T", result)
	}
}

func TestHandle_ActiveCampaign_UnexpectedEvent_ReturnsError(t *testing.T) {
	c := activeCampaignForHandle(t)
	_, err := c.Handle(event.EntityCreated{
		EntityID: identity.NewMasterNarrativeID().String(), EntityType: "master_narrative",
		Name: "Lore", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_IdleCampaign_SessionStarted_TransitionsToActive(t *testing.T) {
	c := idleCampaignForHandle(t)
	result, err := c.Handle(event.SessionStarted{
		SessionID:  identity.NewSessionID().String(),
		CampaignID: c.CampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.ActiveCampaign); !ok {
		t.Fatalf("expected ActiveCampaign, got %T", result)
	}
}

func TestHandle_IdleCampaign_EntityUpdated_Complete_TransitionsToComplete(t *testing.T) {
	c := idleCampaignForHandle(t)
	result, err := c.Handle(event.EntityUpdated{
		EntityID: c.CampaignID().String(), Field: "status",
		OldValue: "idle", NewValue: "complete", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.CompleteCampaign); !ok {
		t.Fatalf("expected CompleteCampaign, got %T", result)
	}
}

func TestHandle_CompleteCampaign_AnyEvent_ReturnsError(t *testing.T) {
	c := idleCampaignForHandle(t)
	complete, err := c.Handle(event.EntityUpdated{
		EntityID: c.CampaignID().String(), Field: "status",
		OldValue: "idle", NewValue: "complete", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = complete.Handle(event.SessionStarted{
		SessionID: identity.NewSessionID().String(), CampaignID: c.CampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_Campaign_Immutability_OriginalUnchanged(t *testing.T) {
	c := newCampaignForHandle(t)
	snapBefore := c.Snapshot()

	_, err := c.Handle(event.EntityUpdated{
		EntityID: c.CampaignID().String(), Field: "status",
		OldValue: "new", NewValue: "forming", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	snapAfter := c.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

func TestReplayCampaign_FullLifecycle(t *testing.T) {
	campID := identity.NewCampaignID()
	gameID := identity.NewGameID()
	charID := identity.NewCharacterID()
	sessID := identity.NewSessionID()

	events := []event.Event{
		event.EntityCreated{EntityID: campID.String(), EntityType: "campaign", Name: "Heroes of Light", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: campID.String(), EntityBID: charID.String(), Relationship: "character", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: campID.String(), Field: "status", OldValue: "new", NewValue: "forming", Source: event.SourceGrimoire},
		event.SessionStarted{SessionID: sessID.String(), CampaignID: campID.String(), Date: time.Now()},
		event.SessionEnded{SessionID: sessID.String()},
		event.EntityUpdated{EntityID: campID.String(), Field: "status", OldValue: "idle", NewValue: "complete", Source: event.SourceGrimoire},
	}

	result, err := entity.ReplayCampaign(gameID, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.CompleteCampaign); !ok {
		t.Fatalf("expected CompleteCampaign, got %T", result)
	}
	if result.CampaignName() != "Heroes of Light" {
		t.Fatalf("expected name 'Heroes of Light', got '%s'", result.CampaignName())
	}
}

func TestReplayCampaign_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := entity.ReplayCampaign(identity.NewGameID(), nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
