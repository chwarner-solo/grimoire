package entity_test

import (
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/session/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- helpers ---

func mustCreateNewSession(t *testing.T) entity.NewSession {
	t.Helper()
	s, _, err := entity.CreateSession(identity.NewSessionID(), identity.NewCampaignID())
	if err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}
	return s
}

func mustReachInProgress(t *testing.T) entity.InProgressSession {
	t.Helper()
	s := mustCreateNewSession(t)
	ip, _, err := s.Start(time.Now())
	if err != nil {
		t.Fatalf("unexpected error starting session: %v", err)
	}
	return ip
}

func mustReachCompleted(t *testing.T) entity.CompletedSession {
	t.Helper()
	ip := mustReachInProgress(t)
	c, _, err := ip.End()
	if err != nil {
		t.Fatalf("unexpected error ending session: %v", err)
	}
	return c
}

func mustReachSummarized(t *testing.T) entity.SummarizedSession {
	t.Helper()
	c := mustReachCompleted(t)
	s, _, err := c.Summarize("The party defeated the dragon.", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error summarizing session: %v", err)
	}
	return s
}

// --- creation tests ---

func TestCreateSession_RequiresNonZeroID(t *testing.T) {
	_, _, err := entity.CreateSession(identity.SessionID{}, identity.NewCampaignID())
	if err == nil {
		t.Fatal("expected error for zero session ID")
	}
	if err != entity.ErrSessionIDRequired {
		t.Fatalf("expected ErrSessionIDRequired, got: %v", err)
	}
}

func TestCreateSession_RequiresNonZeroCampaignID(t *testing.T) {
	_, _, err := entity.CreateSession(identity.NewSessionID(), identity.CampaignID{})
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
	s, _, err := entity.CreateSession(sid, cid)
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
	ip, _, err := s.Start(time.Now())
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
	c, _, err := ip.End()
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
	s, _, err := c.Summarize(notes, event.SourceGrimoire)
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
		_, _, err := c.Summarize(notes, event.SourceGrimoire)
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
	idle, _, err := s.Close(event.SourceGrimoire)
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
	idle, _, err := s.Close(event.SourceGrimoire)
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

// --- Event production tests ---

func TestCreateSession_ProducesEntityCreatedEvent(t *testing.T) {
	sid := identity.NewSessionID()
	_, events, err := entity.CreateSession(sid, identity.NewCampaignID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityID != sid.String() {
		t.Fatalf("expected entity ID %s, got %s", sid.String(), ec.EntityID)
	}
	if ec.EntityType != "session" {
		t.Fatalf("expected entity type 'session', got %q", ec.EntityType)
	}
	if ec.Source != event.SourceGrimoire {
		t.Fatalf("expected source grimoire, got %q", ec.Source)
	}
}

func TestStart_ProducesSessionStartedEvent(t *testing.T) {
	s := mustCreateNewSession(t)
	date := time.Date(2026, 5, 9, 19, 0, 0, 0, time.UTC)
	_, events, err := s.Start(date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ss, ok := events[0].(event.SessionStarted)
	if !ok {
		t.Fatalf("expected SessionStarted, got %T", events[0])
	}
	if ss.SessionID != s.SessionID().String() {
		t.Fatalf("expected session ID %s, got %s", s.SessionID().String(), ss.SessionID)
	}
	if ss.CampaignID != s.CampaignID().String() {
		t.Fatalf("expected campaign ID %s, got %s", s.CampaignID().String(), ss.CampaignID)
	}
	if !ss.Date.Equal(date) {
		t.Fatalf("expected date %v, got %v", date, ss.Date)
	}
}

func TestEnd_ProducesSessionEndedEvent(t *testing.T) {
	ip := mustReachInProgress(t)
	_, events, err := ip.End()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	se, ok := events[0].(event.SessionEnded)
	if !ok {
		t.Fatalf("expected SessionEnded, got %T", events[0])
	}
	if se.SessionID != ip.SessionID().String() {
		t.Fatalf("expected session ID %s, got %s", ip.SessionID().String(), se.SessionID)
	}
}

func TestSummarize_ProducesEntityUpdatedEvent(t *testing.T) {
	c := mustReachCompleted(t)
	notes := "Great session!"
	_, events, err := c.Summarize(notes, event.SourceObsidian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	eu, ok := events[0].(event.EntityUpdated)
	if !ok {
		t.Fatalf("expected EntityUpdated, got %T", events[0])
	}
	if eu.Field != "notes" {
		t.Fatalf("expected field 'notes', got %q", eu.Field)
	}
	if eu.NewValue != notes {
		t.Fatalf("expected new value %q, got %q", notes, eu.NewValue)
	}
	if eu.Source != event.SourceObsidian {
		t.Fatalf("expected source obsidian, got %q", eu.Source)
	}
}

func TestClose_ProducesEntityUpdatedEvent(t *testing.T) {
	s := mustReachSummarized(t)
	_, events, err := s.Close(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	eu, ok := events[0].(event.EntityUpdated)
	if !ok {
		t.Fatalf("expected EntityUpdated, got %T", events[0])
	}
	if eu.Field != "status" {
		t.Fatalf("expected field 'status', got %q", eu.Field)
	}
	if eu.NewValue != "idle" {
		t.Fatalf("expected new value 'idle', got %q", eu.NewValue)
	}
}

func TestSummarize_Error_ProducesNoEvents(t *testing.T) {
	c := mustReachCompleted(t)
	_, events, err := c.Summarize("", event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error")
	}
	if events != nil {
		t.Fatalf("expected nil events on error, got %d", len(events))
	}
}

func TestSession_CommandEvents_RoundTrip_ThroughHandle(t *testing.T) {
	sessID := identity.NewSessionID()
	campID := identity.NewCampaignID()

	// Build state via commands
	s, evts, _ := entity.CreateSession(sessID, campID)
	allEvents := evts

	ip, evts, _ := s.Start(time.Now())
	allEvents = append(allEvents, evts...)

	c, evts, _ := ip.End()
	allEvents = append(allEvents, evts...)

	sum, evts, _ := c.Summarize("Good session", event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	idle, evts, _ := sum.Close(event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	// Replay from events
	replayed, err := entity.ReplaySession(campID, allEvents)
	if err != nil {
		t.Fatalf("replay error: %v", err)
	}

	cmdSnap := idle.Snapshot()
	replaySnap := replayed.Snapshot()

	if cmdSnap.State != replaySnap.State {
		t.Fatalf("state mismatch: command=%s, replay=%s", cmdSnap.State, replaySnap.State)
	}
	if cmdSnap.Notes != replaySnap.Notes {
		t.Fatalf("notes mismatch: command=%q, replay=%q", cmdSnap.Notes, replaySnap.Notes)
	}
}
