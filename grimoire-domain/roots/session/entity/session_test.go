package entity_test

import (
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/session/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- helpers ---

func mustCreateNewSession(t *testing.T) entity.NewSession {
	t.Helper()
	s, err := entity.CreateSession(identity.NewSessionID(), identity.NewCampaignID())
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}
	return s
}

func mustReachInProgress(t *testing.T) entity.InProgressSession {
	t.Helper()
	s := mustCreateNewSession(t)
	ip, err := s.Start()
	if err != nil {
		t.Fatalf("unexpected error starting session: %v", err)
	}
	return ip
}

func mustReachCompleted(t *testing.T) entity.CompletedSession {
	t.Helper()
	ip := mustReachInProgress(t)
	c, err := ip.End()
	if err != nil {
		t.Fatalf("unexpected error ending session: %v", err)
	}
	return c
}

func mustReachSummarized(t *testing.T) entity.SummarizedSession {
	t.Helper()
	c := mustReachCompleted(t)
	s, err := c.Summarize("The party defeated the dragon.")
	if err != nil {
		t.Fatalf("unexpected error summarizing session: %v", err)
	}
	return s
}

// --- creation tests ---

func TestCreateSession_RequiresNonZeroID(t *testing.T) {
	_, err := entity.CreateSession(identity.SessionID{}, identity.NewCampaignID())
	if err == nil {
		t.Fatal("expected error for zero session ID")
	}
	if err != entity.ErrSessionIDRequired {
		t.Fatalf("expected ErrSessionIDRequired, got: %v", err)
	}
}

func TestCreateSession_RequiresNonZeroCampaignID(t *testing.T) {
	_, err := entity.CreateSession(identity.NewSessionID(), identity.CampaignID{})
	if err == nil {
		t.Fatal("expected error for zero campaign ID")
	}
	if err != entity.ErrCampaignIDRequired {
		t.Fatalf("expected ErrCampaignIDRequired, got: %v", err)
	}
}

func TestCreateSession_ValidInputs_ReturnsNewSession(t *testing.T) {
	sid := identity.NewSessionID()
	cid := identity.NewCampaignID()
	s, err := entity.CreateSession(sid, cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SessionID() != sid {
		t.Errorf("expected session ID %v, got %v", sid, s.SessionID())
	}
	if s.CampaignID() != cid {
		t.Errorf("expected campaign ID %v, got %v", cid, s.CampaignID())
	}
	snap := s.Snapshot()
	if snap.State != "New" {
		t.Errorf("expected state New, got %s", snap.State)
	}
}

// --- transition tests ---

func TestNewSession_Start_TransitionsToInProgress(t *testing.T) {
	s := mustCreateNewSession(t)
	ip, err := s.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := ip.Snapshot()
	if snap.State != "InProgress" {
		t.Errorf("expected state InProgress, got %s", snap.State)
	}
	if ip.SessionID() != s.SessionID() {
		t.Error("session ID should be preserved across transition")
	}
}

func TestInProgressSession_End_TransitionsToCompleted(t *testing.T) {
	ip := mustReachInProgress(t)
	c, err := ip.End()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := c.Snapshot()
	if snap.State != "Completed" {
		t.Errorf("expected state Completed, got %s", snap.State)
	}
	if c.SessionID() != ip.SessionID() {
		t.Error("session ID should be preserved across transition")
	}
}

func TestCompletedSession_Summarize_TransitionsToSummarized(t *testing.T) {
	c := mustReachCompleted(t)
	notes := "The party defeated the dragon."
	s, err := c.Summarize(notes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := s.Snapshot()
	if snap.State != "Summarized" {
		t.Errorf("expected state Summarized, got %s", snap.State)
	}
	if snap.Notes != notes {
		t.Errorf("expected notes %q, got %q", notes, snap.Notes)
	}
}

func TestCompletedSession_Summarize_RequiresNonEmptyNotes(t *testing.T) {
	c := mustReachCompleted(t)

	for _, notes := range []string{"", "   ", "\t\n"} {
		_, err := c.Summarize(notes)
		if err == nil {
			t.Fatalf("expected error for empty notes %q", notes)
		}
		if err != entity.ErrNotesRequired {
			t.Fatalf("expected ErrNotesRequired, got: %v", err)
		}
	}
}

func TestSummarizedSession_Close_TransitionsToIdle(t *testing.T) {
	s := mustReachSummarized(t)
	idle, err := s.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := idle.Snapshot()
	if snap.State != "Idle" {
		t.Errorf("expected state Idle, got %s", snap.State)
	}
	if idle.SessionID() != s.SessionID() {
		t.Error("session ID should be preserved across transition")
	}
}

func TestIdleSession_PreservesIdentity(t *testing.T) {
	s := mustReachSummarized(t)
	idle, err := s.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idle.SessionID() != s.SessionID() {
		t.Error("session ID should be preserved in idle state")
	}
	if idle.CampaignID() != s.CampaignID() {
		t.Error("campaign ID should be preserved in idle state")
	}
	snap := idle.Snapshot()
	if snap.Notes != s.Snapshot().Notes {
		t.Error("notes should be preserved in idle state")
	}
}
