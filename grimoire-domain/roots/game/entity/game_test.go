package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Test helpers ---

func mustCreateNewGame(t *testing.T) NewGame {
	t.Helper()
	g, _, err := CreateGame(identity.NewGameID(), "Ashes & Chains", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCreateNewGame: %v", err)
	}
	return g
}

func mustReachDraft(t *testing.T) DraftGame {
	t.Helper()
	ng := mustCreateNewGame(t)
	dg, _, err := ng.AddNarrativeElement(identity.NewNarrativeID(), "The Fall of Kael", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustReachDraft: %v", err)
	}
	return dg
}

func mustReachActive(t *testing.T) ActiveGame {
	t.Helper()
	dg := mustReachDraft(t)
	cid := identity.NewCampaignID()
	dg, _, err := dg.LinkCampaign(cid, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustReachActive link: %v", err)
	}
	ag, err := dg.ActivateFromCampaign(cid)
	if err != nil {
		t.Fatalf("mustReachActive activate: %v", err)
	}
	return ag
}

func mustReachIdle(t *testing.T) IdleGame {
	t.Helper()
	ag := mustReachActive(t)
	snap := ag.Snapshot()
	cid := snap.CampaignIDs[0]
	result, err := ag.NotifyCampaignIdle(cid)
	if err != nil {
		t.Fatalf("mustReachIdle: %v", err)
	}
	ig, ok := result.(IdleGame)
	if !ok {
		t.Fatal("mustReachIdle: expected IdleGame")
	}
	return ig
}

// --- Constructor tests ---

func TestCreateGame_RequiresNonZeroID(t *testing.T) {
	_, _, err := CreateGame(identity.GameID{}, "Ashes & Chains", event.SourceGrimoire)
	if !errors.Is(err, ErrGameIDRequired) {
		t.Fatalf("expected ErrGameIDRequired, got: %v", err)
	}
}

func TestCreateGame_RequiresNonEmptyName(t *testing.T) {
	_, _, err := CreateGame(identity.NewGameID(), "", event.SourceGrimoire)
	if !errors.Is(err, ErrGameNameRequired) {
		t.Fatalf("expected ErrGameNameRequired, got: %v", err)
	}
}

func TestCreateGame_RequiresNonWhitespaceName(t *testing.T) {
	_, _, err := CreateGame(identity.NewGameID(), "   ", event.SourceGrimoire)
	if !errors.Is(err, ErrGameNameRequired) {
		t.Fatalf("expected ErrGameNameRequired, got: %v", err)
	}
}

func TestCreateGame_ValidInputs_ReturnsNewGame(t *testing.T) {
	g := mustCreateNewGame(t)
	if g.GameName() != "Ashes & Chains" {
		t.Fatalf("expected name 'Ashes & Chains', got %q", g.GameName())
	}
	if g.GameID().IsZero() {
		t.Fatal("expected non-zero GameID")
	}
}

// --- NewGame state tests ---

func TestNewGame_AddNarrativeElement_TransitionsToDraft(t *testing.T) {
	ng := mustCreateNewGame(t)
	dg, _, err := ng.AddNarrativeElement(identity.NewNarrativeID(), "The Fall of Kael", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := dg.(DraftGame); !ok {
		t.Fatal("expected DraftGame")
	}
}

func TestNewGame_AddNarrativeElement_RequiresNonZeroID(t *testing.T) {
	ng := mustCreateNewGame(t)
	_, _, err := ng.AddNarrativeElement(identity.NarrativeID{}, "The Fall of Kael", event.SourceGrimoire)
	if !errors.Is(err, ErrNarrativeIDRequired) {
		t.Fatalf("expected ErrNarrativeIDRequired, got: %v", err)
	}
}

func TestNewGame_AddNarrativeElement_RequiresNonEmptyName(t *testing.T) {
	ng := mustCreateNewGame(t)
	_, _, err := ng.AddNarrativeElement(identity.NewNarrativeID(), "", event.SourceGrimoire)
	if !errors.Is(err, ErrNarrativeNameRequired) {
		t.Fatalf("expected ErrNarrativeNameRequired, got: %v", err)
	}
}

// --- DraftGame state tests ---

func TestDraftGame_AddNarrativeElement_StaysDraft(t *testing.T) {
	dg := mustReachDraft(t)
	dg2, _, err := dg.AddNarrativeElement(identity.NewNarrativeID(), "Act II", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dg2.Snapshot().State != "draft" {
		t.Fatal("expected state to remain draft")
	}
}

func TestDraftGame_LinkCampaign_Succeeds(t *testing.T) {
	dg := mustReachDraft(t)
	cid := identity.NewCampaignID()
	dg2, _, err := dg.LinkCampaign(cid, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := dg2.Snapshot()
	if len(snap.CampaignIDs) != 1 {
		t.Fatalf("expected 1 campaign, got %d", len(snap.CampaignIDs))
	}
}

func TestDraftGame_LinkCampaign_RejectsDuplicate(t *testing.T) {
	dg := mustReachDraft(t)
	cid := identity.NewCampaignID()
	dg, _, _ = dg.LinkCampaign(cid, event.SourceGrimoire)
	_, _, err := dg.LinkCampaign(cid, event.SourceGrimoire)
	if !errors.Is(err, ErrCampaignAlreadyLinked) {
		t.Fatalf("expected ErrCampaignAlreadyLinked, got: %v", err)
	}
}

func TestDraftGame_ActivateFromCampaign_TransitionsToActive(t *testing.T) {
	ag := mustReachActive(t)
	if ag.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestDraftGame_ActivateFromCampaign_RequiresLinkedCampaign(t *testing.T) {
	dg := mustReachDraft(t)
	_, err := dg.ActivateFromCampaign(identity.NewCampaignID())
	if !errors.Is(err, ErrCampaignNotLinked) {
		t.Fatalf("expected ErrCampaignNotLinked, got: %v", err)
	}
}

// --- ActiveGame state tests ---

func TestActiveGame_NotifyCampaignIdle_TransitionsToIdle_WhenLastCampaign(t *testing.T) {
	ig := mustReachIdle(t)
	if ig.Snapshot().State != "idle" {
		t.Fatal("expected idle state")
	}
}

func TestActiveGame_NotifyCampaignIdle_StaysActive_WhenOtherCampaignsActive(t *testing.T) {
	dg := mustReachDraft(t)
	c1 := identity.NewCampaignID()
	c2 := identity.NewCampaignID()
	dg, _, _ = dg.LinkCampaign(c1, event.SourceGrimoire)
	dg, _, _ = dg.LinkCampaign(c2, event.SourceGrimoire)
	ag, _ := dg.ActivateFromCampaign(c1)

	snap := ag.Snapshot()
	if snap.ActiveCampaignCount != 1 {
		t.Fatalf("expected active campaign count 1, got %d", snap.ActiveCampaignCount)
	}

	// Notify idle → should transition to idle since count becomes 0
	result, err := ag.NotifyCampaignIdle(c1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(IdleGame); !ok {
		t.Fatal("expected IdleGame when last active campaign goes idle")
	}
}

func TestActiveGame_LinkCampaign_Succeeds(t *testing.T) {
	ag := mustReachActive(t)
	c2 := identity.NewCampaignID()
	ag2, _, err := ag.LinkCampaign(c2, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ag2.Snapshot().CampaignIDs) != 2 {
		t.Fatalf("expected 2 campaigns, got %d", len(ag2.Snapshot().CampaignIDs))
	}
}

func TestActiveGame_NotifyCampaignIdle_RequiresLinkedCampaign(t *testing.T) {
	ag := mustReachActive(t)
	_, err := ag.NotifyCampaignIdle(identity.NewCampaignID())
	if !errors.Is(err, ErrCampaignNotLinked) {
		t.Fatalf("expected ErrCampaignNotLinked, got: %v", err)
	}
}

// --- IdleGame state tests ---

func TestIdleGame_ActivateFromCampaign_TransitionsToActive(t *testing.T) {
	ig := mustReachIdle(t)
	cid := ig.Snapshot().CampaignIDs[0]
	ag, err := ig.ActivateFromCampaign(cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ag.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestIdleGame_Archive_TransitionsToArchived(t *testing.T) {
	ig := mustReachIdle(t)
	ag, _, err := ig.Archive(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ag.Snapshot().State != "archived" {
		t.Fatal("expected archived state")
	}
}

// --- ArchivedGame state tests ---

func TestArchivedGame_PreservesIdentity(t *testing.T) {
	ig := mustReachIdle(t)
	gameID := ig.GameID()
	gameName := ig.GameName()
	ag, _, _ := ig.Archive(event.SourceGrimoire)
	if ag.GameID().String() != gameID.String() {
		t.Fatal("archived game should preserve GameID")
	}
	if ag.GameName() != gameName {
		t.Fatal("archived game should preserve GameName")
	}
}

// --- Event production tests ---

func TestCreateGame_ProducesEntityCreatedEvent(t *testing.T) {
	id := identity.NewGameID()
	_, events, err := CreateGame(id, "Test Game", event.SourceObsidian)
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
	if ec.EntityType != "game" {
		t.Fatalf("expected entity type 'game', got %q", ec.EntityType)
	}
	if ec.Name != "Test Game" {
		t.Fatalf("expected name 'Test Game', got %q", ec.Name)
	}
	if ec.Source != event.SourceObsidian {
		t.Fatalf("expected source obsidian, got %q", ec.Source)
	}
}

func TestAddNarrativeElement_ProducesEntityCreatedEvent(t *testing.T) {
	ng := mustCreateNewGame(t)
	narID := identity.NewNarrativeID()
	_, events, err := ng.AddNarrativeElement(narID, "The Fall of Kael", event.SourceFoundry)
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
	if ec.EntityID != narID.String() {
		t.Fatalf("expected entity ID %s, got %s", narID.String(), ec.EntityID)
	}
	if ec.EntityType != "narrative" {
		t.Fatalf("expected entity type 'narrative', got %q", ec.EntityType)
	}
	if ec.Name != "The Fall of Kael" {
		t.Fatalf("expected name 'The Fall of Kael', got %q", ec.Name)
	}
}

func TestAddNarrativeElement_SourcePassedThrough(t *testing.T) {
	ng := mustCreateNewGame(t)
	_, events, _ := ng.AddNarrativeElement(identity.NewNarrativeID(), "Lore", event.SourceObsidian)
	ec := events[0].(event.EntityCreated)
	if ec.Source != event.SourceObsidian {
		t.Fatalf("expected source obsidian, got %q", ec.Source)
	}
}

func TestLinkCampaign_ProducesEntityLinkedEvent(t *testing.T) {
	dg := mustReachDraft(t)
	campID := identity.NewCampaignID()
	_, events, err := dg.LinkCampaign(campID, event.SourceGrimoire)
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
	if el.EntityAID != dg.GameID().String() {
		t.Fatalf("expected entity A ID %s, got %s", dg.GameID().String(), el.EntityAID)
	}
	if el.EntityBID != campID.String() {
		t.Fatalf("expected entity B ID %s, got %s", campID.String(), el.EntityBID)
	}
	if el.Relationship != "campaign" {
		t.Fatalf("expected relationship 'campaign', got %q", el.Relationship)
	}
}

func TestArchive_ProducesEntityUpdatedEvent(t *testing.T) {
	ig := mustReachIdle(t)
	_, events, err := ig.Archive(event.SourceGrimoire)
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
	if eu.OldValue != "idle" {
		t.Fatalf("expected old value 'idle', got %q", eu.OldValue)
	}
	if eu.NewValue != "archived" {
		t.Fatalf("expected new value 'archived', got %q", eu.NewValue)
	}
}

func TestAddNarrativeElement_Error_ProducesNoEvents(t *testing.T) {
	ng := mustCreateNewGame(t)
	_, events, err := ng.AddNarrativeElement(identity.NarrativeID{}, "Lore", event.SourceGrimoire)
	if err == nil {
		t.Fatal("expected error")
	}
	if events != nil {
		t.Fatalf("expected nil events on error, got %d", len(events))
	}
}

func TestGame_CommandEvents_RoundTrip_ThroughHandle(t *testing.T) {
	gameID := identity.NewGameID()
	narID := identity.NewNarrativeID()
	campID := identity.NewCampaignID()

	// Build state via commands, collecting events
	g, evts, _ := CreateGame(gameID, "Round Trip", event.SourceGrimoire)
	allEvents := evts

	dg, evts, _ := g.AddNarrativeElement(narID, "Lore", event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	dg, evts, _ = dg.LinkCampaign(campID, event.SourceGrimoire)
	allEvents = append(allEvents, evts...)

	// Replay from events
	replayed, err := ReplayGame(allEvents)
	if err != nil {
		t.Fatalf("replay error: %v", err)
	}

	// Compare snapshots
	cmdSnap := dg.Snapshot()
	replaySnap := replayed.Snapshot()

	if cmdSnap.State != replaySnap.State {
		t.Fatalf("state mismatch: command=%s, replay=%s", cmdSnap.State, replaySnap.State)
	}
	if cmdSnap.Name != replaySnap.Name {
		t.Fatalf("name mismatch: command=%s, replay=%s", cmdSnap.Name, replaySnap.Name)
	}
	if len(cmdSnap.CampaignIDs) != len(replaySnap.CampaignIDs) {
		t.Fatalf("campaign IDs length mismatch: command=%d, replay=%d",
			len(cmdSnap.CampaignIDs), len(replaySnap.CampaignIDs))
	}
}
