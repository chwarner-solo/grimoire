package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/session/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- handle test helpers ---

func newSessionForHandle(t *testing.T) entity.Session {
	t.Helper()
	s, err := entity.CreateSession(identity.NewSessionID(), identity.NewCampaignID())
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	return s
}

func inProgressSessionForHandle(t *testing.T) entity.Session {
	t.Helper()
	s := newSessionForHandle(t)
	s2, err := s.Handle(event.SessionStarted{
		SessionID:  s.SessionID().String(),
		CampaignID: s.CampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return s2
}

func completedSessionForHandle(t *testing.T) entity.Session {
	t.Helper()
	s := inProgressSessionForHandle(t)
	s2, err := s.Handle(event.SessionEnded{SessionID: s.SessionID().String()})
	if err != nil {
		t.Fatal(err)
	}
	return s2
}

func summarizedSessionForHandle(t *testing.T) entity.Session {
	t.Helper()
	s := completedSessionForHandle(t)
	s2, err := s.Handle(event.EntityUpdated{
		EntityID: s.SessionID().String(), Field: "notes",
		OldValue: "", NewValue: "Great session notes", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	return s2
}

// --- tests ---

func TestHandle_NewSession_SessionStarted_TransitionsToInProgress(t *testing.T) {
	s := newSessionForHandle(t)
	result, err := s.Handle(event.SessionStarted{
		SessionID:  s.SessionID().String(),
		CampaignID: s.CampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.InProgressSession); !ok {
		t.Fatalf("expected InProgressSession, got %T", result)
	}
}

func TestHandle_NewSession_UnexpectedEvent_ReturnsError(t *testing.T) {
	s := newSessionForHandle(t)
	_, err := s.Handle(event.SessionEnded{SessionID: s.SessionID().String()})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_InProgressSession_SessionEnded_TransitionsToCompleted(t *testing.T) {
	s := inProgressSessionForHandle(t)
	result, err := s.Handle(event.SessionEnded{SessionID: s.SessionID().String()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.CompletedSession); !ok {
		t.Fatalf("expected CompletedSession, got %T", result)
	}
}

func TestHandle_InProgressSession_UnexpectedEvent_ReturnsError(t *testing.T) {
	s := inProgressSessionForHandle(t)
	_, err := s.Handle(event.SessionStarted{
		SessionID: s.SessionID().String(), CampaignID: s.CampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_CompletedSession_EntityUpdated_Notes_TransitionsToSummarized(t *testing.T) {
	s := completedSessionForHandle(t)
	result, err := s.Handle(event.EntityUpdated{
		EntityID: s.SessionID().String(), Field: "notes",
		OldValue: "", NewValue: "Session recap here", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.SummarizedSession); !ok {
		t.Fatalf("expected SummarizedSession, got %T", result)
	}
}

func TestHandle_CompletedSession_UnexpectedEvent_ReturnsError(t *testing.T) {
	s := completedSessionForHandle(t)
	_, err := s.Handle(event.SessionStarted{
		SessionID: s.SessionID().String(), CampaignID: s.CampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_SummarizedSession_EntityUpdated_Idle_TransitionsToIdle(t *testing.T) {
	s := summarizedSessionForHandle(t)
	result, err := s.Handle(event.EntityUpdated{
		EntityID: s.SessionID().String(), Field: "status",
		OldValue: "summarized", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleSession); !ok {
		t.Fatalf("expected IdleSession, got %T", result)
	}
}

func TestHandle_IdleSession_AnyEvent_ReturnsError(t *testing.T) {
	s := summarizedSessionForHandle(t)
	idle, err := s.Handle(event.EntityUpdated{
		EntityID: s.SessionID().String(), Field: "status",
		OldValue: "summarized", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = idle.Handle(event.SessionStarted{
		SessionID: s.SessionID().String(), CampaignID: s.CampaignID().String(),
		Date: time.Now(),
	})
	if !errors.Is(err, entity.ErrUnexpectedEvent) {
		t.Fatalf("expected ErrUnexpectedEvent, got %v", err)
	}
}

func TestHandle_Session_Immutability_OriginalUnchanged(t *testing.T) {
	s := newSessionForHandle(t)
	snapBefore := s.Snapshot()

	_, err := s.Handle(event.SessionStarted{
		SessionID:  s.SessionID().String(),
		CampaignID: s.CampaignID().String(),
		Date:       time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	snapAfter := s.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

func TestReplaySession_FullLifecycle(t *testing.T) {
	sessID := identity.NewSessionID()
	campID := identity.NewCampaignID()

	events := []event.Event{
		event.EntityCreated{EntityID: sessID.String(), EntityType: "session", Name: "Session 1", Source: event.SourceGrimoire},
		event.SessionStarted{SessionID: sessID.String(), CampaignID: campID.String(), Date: time.Now()},
		event.SessionEnded{SessionID: sessID.String()},
		event.EntityUpdated{EntityID: sessID.String(), Field: "notes", OldValue: "", NewValue: "Good session", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: sessID.String(), Field: "status", OldValue: "summarized", NewValue: "idle", Source: event.SourceGrimoire},
	}

	result, err := entity.ReplaySession(campID, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(entity.IdleSession); !ok {
		t.Fatalf("expected IdleSession, got %T", result)
	}
}

func TestReplaySession_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := entity.ReplaySession(identity.NewCampaignID(), nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
