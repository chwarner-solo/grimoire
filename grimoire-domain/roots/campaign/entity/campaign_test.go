package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Test helpers ---

func mustCreateNewCampaign(t *testing.T) NewCampaign {
	t.Helper()
	c, _, err := CreateCampaign(identity.NewCampaignID(), "The Long Night", identity.NewGameID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCreateNewCampaign: %v", err)
	}
	return c
}

func mustReachForming(t *testing.T) FormingCampaign {
	t.Helper()
	nc := mustCreateNewCampaign(t)
	nc, _, err := nc.AddCharacter(identity.NewCharacterID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustReachForming add character: %v", err)
	}
	fc, _, err := nc.BeginFormation(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustReachForming: %v", err)
	}
	return fc
}

func mustReachActive(t *testing.T) ActiveCampaign {
	t.Helper()
	fc := mustReachForming(t)
	ac, _, err := fc.StartFirstSession(identity.NewSessionID(), time.Now())
	if err != nil {
		t.Fatalf("mustReachActive: %v", err)
	}
	return ac
}

func mustReachIdle(t *testing.T) IdleCampaign {
	t.Helper()
	ac := mustReachActive(t)
	ic, err := ac.NotifySessionSummarized()
	if err != nil {
		t.Fatalf("mustReachIdle: %v", err)
	}
	return ic
}

// --- Constructor tests ---

func TestCreateCampaign_RequiresNonZeroID(t *testing.T) {
	_, _, err := CreateCampaign(identity.CampaignID{}, "The Long Night", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignIDRequired) {
		t.Fatalf("expected ErrCampaignIDRequired, got: %v", err)
	}
}

func TestCreateCampaign_RequiresNonEmptyName(t *testing.T) {
	_, _, err := CreateCampaign(identity.NewCampaignID(), "", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignNameRequired) {
		t.Fatalf("expected ErrCampaignNameRequired, got: %v", err)
	}
}

func TestCreateCampaign_RequiresNonZeroGameID(t *testing.T) {
	_, _, err := CreateCampaign(identity.NewCampaignID(), "The Long Night", identity.GameID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got: %v", err)
	}
}

func TestCreateCampaign_ValidInputs_ReturnsNewCampaign(t *testing.T) {
	c := mustCreateNewCampaign(t)
	if c.CampaignName() != "The Long Night" {
		t.Fatalf("expected name 'The Long Night', got %q", c.CampaignName())
	}
	if c.CampaignID().IsZero() {
		t.Fatal("expected non-zero CampaignID")
	}
}

// --- NewCampaign state tests ---

func TestNewCampaign_AddCharacter_Succeeds(t *testing.T) {
	nc := mustCreateNewCampaign(t)
	charID := identity.NewCharacterID()
	nc2, _, err := nc.AddCharacter(charID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := nc2.Snapshot()
	if len(snap.CharacterIDs) != 1 {
		t.Fatalf("expected 1 character, got %d", len(snap.CharacterIDs))
	}
}

func TestNewCampaign_AddCharacter_RejectsDuplicate(t *testing.T) {
	nc := mustCreateNewCampaign(t)
	charID := identity.NewCharacterID()
	nc, _, _ = nc.AddCharacter(charID, event.SourceGrimoire)
	_, _, err := nc.AddCharacter(charID, event.SourceGrimoire)
	if !errors.Is(err, ErrCharacterAlreadyAdded) {
		t.Fatalf("expected ErrCharacterAlreadyAdded, got: %v", err)
	}
}

func TestNewCampaign_BeginFormation_TransitionsToForming(t *testing.T) {
	fc := mustReachForming(t)
	if fc.Snapshot().State != "forming" {
		t.Fatal("expected forming state")
	}
}

// --- FormingCampaign state tests ---

func TestFormingCampaign_AddCharacter_Succeeds(t *testing.T) {
	fc := mustReachForming(t)
	charID := identity.NewCharacterID()
	fc2, _, err := fc.AddCharacter(charID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := fc2.Snapshot()
	if len(snap.CharacterIDs) != 2 {
		t.Fatalf("expected 2 characters, got %d", len(snap.CharacterIDs))
	}
}

func TestFormingCampaign_StartFirstSession_Succeeds(t *testing.T) {
	ac := mustReachActive(t)
	if ac.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestFormingCampaign_StartFirstSession_RequiresAtLeastOneCharacter(t *testing.T) {
	// Use reconstitution to create a forming campaign with no characters
	snap := CampaignSnapshot{
		ID:     identity.NewCampaignID(),
		Name:   "Empty Party",
		GameID: identity.NewGameID(),
		State:  "forming",
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected reconstitution error: %v", err)
	}
	fc := c.(FormingCampaign)
	_, _, err = fc.StartFirstSession(identity.NewSessionID(), time.Now())
	if !errors.Is(err, ErrNoCharacters) {
		t.Fatalf("expected ErrNoCharacters, got: %v", err)
	}
}

// --- ActiveCampaign state tests ---

func TestActiveCampaign_NotifySessionSummarized_TransitionsToIdle(t *testing.T) {
	ic := mustReachIdle(t)
	if ic.Snapshot().State != "idle" {
		t.Fatal("expected idle state")
	}
}

// --- IdleCampaign state tests ---

func TestIdleCampaign_StartNewSession_TransitionsToActive(t *testing.T) {
	ic := mustReachIdle(t)
	ac, _, err := ic.StartNewSession(identity.NewSessionID(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ac.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestIdleCampaign_Complete_TransitionsToComplete(t *testing.T) {
	ic := mustReachIdle(t)
	cc, _, err := ic.Complete(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cc.Snapshot().State != "complete" {
		t.Fatal("expected complete state")
	}
}

// --- CompleteCampaign state tests ---

func TestCompleteCampaign_PreservesIdentity(t *testing.T) {
	ic := mustReachIdle(t)
	campaignID := ic.CampaignID()
	campaignName := ic.CampaignName()
	cc, _, _ := ic.Complete(event.SourceGrimoire)
	if cc.CampaignID().String() != campaignID.String() {
		t.Fatal("complete campaign should preserve CampaignID")
	}
	if cc.CampaignName() != campaignName {
		t.Fatal("complete campaign should preserve CampaignName")
	}
}

// --- Event production tests ---

func TestCreateCampaign_ProducesEntityCreatedEvent(t *testing.T) {
	id := identity.NewCampaignID()
	_, events, err := CreateCampaign(id, "Test Campaign", identity.NewGameID(), event.SourceObsidian)
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
	if ec.EntityID != id.String() {
		t.Fatalf("expected entity ID %s, got %s", id.String(), ec.EntityID)
	}
	if ec.EntityType != "campaign" {
		t.Fatalf("expected entity type 'campaign', got %q", ec.EntityType)
	}
	if ec.Source != event.SourceObsidian {
		t.Fatalf("expected source obsidian, got %q", ec.Source)
	}
}

func TestAddCharacter_ProducesEntityLinkedEvent(t *testing.T) {
	nc := mustCreateNewCampaign(t)
	charID := identity.NewCharacterID()
	_, events, err := nc.AddCharacter(charID, event.SourceFoundry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	el, ok := events[0].(event.EntityLinked)
	if !ok {
		t.Fatalf("expected EntityLinked, got %T", events[0])
	}
	if el.EntityBID != charID.String() {
		t.Fatalf("expected entity B ID %s, got %s", charID.String(), el.EntityBID)
	}
	if el.Relationship != "character" {
		t.Fatalf("expected relationship 'character', got %q", el.Relationship)
	}
	if el.Source != event.SourceFoundry {
		t.Fatalf("expected source foundry, got %q", el.Source)
	}
}

func TestBeginFormation_ProducesEntityUpdatedEvent(t *testing.T) {
	nc := mustCreateNewCampaign(t)
	nc, _, _ = nc.AddCharacter(identity.NewCharacterID(), event.SourceGrimoire)
	_, events, err := nc.BeginFormation(event.SourceGrimoire)
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
	if eu.NewValue != "forming" {
		t.Fatalf("expected new value 'forming', got %q", eu.NewValue)
	}
}

func TestStartFirstSession_ProducesSessionStartedEvent(t *testing.T) {
	fc := mustReachForming(t)
	sessID := identity.NewSessionID()
	date := time.Date(2026, 5, 9, 19, 0, 0, 0, time.UTC)
	_, events, err := fc.StartFirstSession(sessID, date)
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
	if ss.SessionID != sessID.String() {
		t.Fatalf("expected session ID %s, got %s", sessID.String(), ss.SessionID)
	}
	if ss.CampaignID != fc.CampaignID().String() {
		t.Fatalf("expected campaign ID %s, got %s", fc.CampaignID().String(), ss.CampaignID)
	}
	if !ss.Date.Equal(date) {
		t.Fatalf("expected date %v, got %v", date, ss.Date)
	}
}

func TestStartNewSession_ProducesSessionStartedEvent(t *testing.T) {
	ic := mustReachIdle(t)
	sessID := identity.NewSessionID()
	date := time.Date(2026, 5, 10, 19, 0, 0, 0, time.UTC)
	_, events, err := ic.StartNewSession(sessID, date)
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
	if ss.SessionID != sessID.String() {
		t.Fatalf("expected session ID %s, got %s", sessID.String(), ss.SessionID)
	}
}

func TestComplete_ProducesEntityUpdatedEvent(t *testing.T) {
	ic := mustReachIdle(t)
	_, events, err := ic.Complete(event.SourceGrimoire)
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
	if eu.NewValue != "complete" {
		t.Fatalf("expected new value 'complete', got %q", eu.NewValue)
	}
}

func TestAddCharacter_Error_ProducesNoEvents(t *testing.T) {
	nc := mustCreateNewCampaign(t)
	_, events, err := nc.AddCharacter(identity.CharacterID{}, event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error")
	}
	if events != nil {
		t.Fatalf("expected nil events on error, got %d", len(events))
	}
}

func TestCampaign_CommandEvents_RoundTrip_ThroughHandle(t *testing.T) {
	campID := identity.NewCampaignID()
	gameID := identity.NewGameID()
	charID := identity.NewCharacterID()
	sessID := identity.NewSessionID()

	// Build state via commands
	nc, evts, _ := CreateCampaign(campID, "Round Trip", gameID, event.SourceGrimoire)
	allEvents := evts

	nc, evts, _ = nc.AddCharacter(charID, event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	fc, evts, _ := nc.BeginFormation(event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	ac, evts, _ := fc.StartFirstSession(sessID, time.Now())
	allEvents = append(allEvents, evts...)

	// Replay from events
	replayed, err := ReplayCampaign(gameID, allEvents)
	if err != nil {
		t.Fatalf("replay error: %v", err)
	}

	cmdSnap := ac.Snapshot()
	replaySnap := replayed.Snapshot()

	if cmdSnap.State != replaySnap.State {
		t.Fatalf("state mismatch: command=%s, replay=%s", cmdSnap.State, replaySnap.State)
	}
	if cmdSnap.Name != replaySnap.Name {
		t.Fatalf("name mismatch: command=%s, replay=%s", cmdSnap.Name, replaySnap.Name)
	}
	if len(cmdSnap.CharacterIDs) != len(replaySnap.CharacterIDs) {
		t.Fatalf("character IDs length mismatch: command=%d, replay=%d",
			len(cmdSnap.CharacterIDs), len(replaySnap.CharacterIDs))
	}
}
