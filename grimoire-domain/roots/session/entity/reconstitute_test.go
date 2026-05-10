package entity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/session/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestReconstituteSession_New(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "New",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(entity.NewSession); !ok {
		t.Fatal("expected NewSession interface")
	}
}

func TestReconstituteSession_InProgress(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "InProgress",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(entity.InProgressSession); !ok {
		t.Fatal("expected InProgressSession interface")
	}
}

func TestReconstituteSession_Completed(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "Completed",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(entity.CompletedSession); !ok {
		t.Fatal("expected CompletedSession interface")
	}
}

func TestReconstituteSession_Summarized(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "Summarized",
		Notes:      "The party rested.",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(entity.SummarizedSession); !ok {
		t.Fatal("expected SummarizedSession interface")
	}
}

func TestReconstituteSession_Idle(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "Idle",
		Notes:      "Final notes.",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := s.(entity.IdleSession); !ok {
		t.Fatal("expected IdleSession interface")
	}
}

func TestReconstituteSession_UnknownState_ReturnsError(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "Bogus",
	}
	_, err := entity.ReconstituteSession(snap)
	if err == nil {
		t.Fatal("expected error for unknown state")
	}
	if !errors.Is(err, entity.ErrInvalidSessionState) {
		t.Fatalf("expected ErrInvalidSessionState, got: %v", err)
	}
}

func TestReconstituteSession_RoundTrip(t *testing.T) {
	// Create a session, transition it, snapshot, reconstitute, verify.
	original, _, err := entity.CreateSession(identity.NewSessionID(), identity.NewCampaignID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ip, _, _ := original.Start(time.Now())
	c, _, _ := ip.End()
	s, _, _ := c.Summarize("Round-trip notes.", event.SourceGrimoire)

	snap := s.Snapshot()
	reconstituted, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reconSnap := reconstituted.Snapshot()
	if reconSnap.ID != snap.ID {
		t.Error("ID mismatch after round-trip")
	}
	if reconSnap.CampaignID != snap.CampaignID {
		t.Error("CampaignID mismatch after round-trip")
	}
	if reconSnap.State != snap.State {
		t.Errorf("State mismatch: expected %s, got %s", snap.State, reconSnap.State)
	}
	if reconSnap.Notes != snap.Notes {
		t.Errorf("Notes mismatch: expected %q, got %q", snap.Notes, reconSnap.Notes)
	}
}

func TestReconstituteSession_CanTransitionAfterReconstitution(t *testing.T) {
	snap := entity.SessionSnapshot{
		ID:         identity.NewSessionID(),
		CampaignID: identity.NewCampaignID(),
		State:      "Completed",
	}
	s, err := entity.ReconstituteSession(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	completed, ok := s.(entity.CompletedSession)
	if !ok {
		t.Fatal("expected CompletedSession interface")
	}

	summarized, _, err := completed.Summarize("Post-reconstitution notes.", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summarized.Snapshot().State != "Summarized" {
		t.Errorf("expected Summarized state, got %s", summarized.Snapshot().State)
	}
}
