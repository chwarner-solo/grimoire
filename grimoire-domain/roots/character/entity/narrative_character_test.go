package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- NPC test helpers ---

func mustCreateNewNPC(t *testing.T) NewNPC {
	t.Helper()
	n, _, err := CreateNPC(identity.NewNarrativeCharacterID(), "Velleth", identity.NewGameID(), event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCreateNewNPC: %v", err)
	}
	return n
}

func mustDraftNPC(t *testing.T) DraftNPC {
	t.Helper()
	n := mustCreateNewNPC(t)
	d, _, err := n.BeginDraft(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustDraftNPC: %v", err)
	}
	return d
}

func mustActiveNPC(t *testing.T) ActiveNPC {
	t.Helper()
	d := mustDraftNPC(t)
	a, _, err := d.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustActiveNPC: %v", err)
	}
	return a
}

func mustIdleNPC(t *testing.T) IdleNPC {
	t.Helper()
	a := mustActiveNPC(t)
	i, _, err := a.GoIdle(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustIdleNPC: %v", err)
	}
	return i
}

// --- Constructor tests ---

func TestCreateNPC_ValidInputs(t *testing.T) {
	id := identity.NewNarrativeCharacterID()
	gameID := identity.NewGameID()
	n, events, err := CreateNPC(id, "Velleth", gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.NPCName() != "Velleth" {
		t.Fatalf("expected name 'Velleth', got %q", n.NPCName())
	}
	if n.NarrativeCharacterID().String() != id.String() {
		t.Fatalf("expected ID %s, got %s", id, n.NarrativeCharacterID())
	}
	if n.GameID().String() != gameID.String() {
		t.Fatalf("expected GameID %s, got %s", gameID, n.GameID())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	ec, ok := events[0].(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated, got %T", events[0])
	}
	if ec.EntityType != "npc" {
		t.Fatalf("expected entity type 'npc', got %q", ec.EntityType)
	}
}

func TestCreateNPC_RequiresNonZeroID(t *testing.T) {
	_, _, err := CreateNPC(identity.NarrativeCharacterID{}, "Name", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrNPCIDRequired) {
		t.Fatalf("expected ErrNPCIDRequired, got: %v", err)
	}
}

func TestCreateNPC_RequiresNonEmptyName(t *testing.T) {
	_, _, err := CreateNPC(identity.NewNarrativeCharacterID(), "", identity.NewGameID(), event.SourceGrimoire)
	if !errors.Is(err, ErrNPCNameRequired) {
		t.Fatalf("expected ErrNPCNameRequired, got: %v", err)
	}
}

func TestCreateNPC_RequiresNonZeroGameID(t *testing.T) {
	_, _, err := CreateNPC(identity.NewNarrativeCharacterID(), "Name", identity.GameID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrNPCGameIDRequired) {
		t.Fatalf("expected ErrNPCGameIDRequired, got: %v", err)
	}
}

// --- Transition tests ---

func TestNewNPC_BeginDraft_TransitionsToDraft(t *testing.T) {
	d := mustDraftNPC(t)
	if d.Snapshot().State != "draft" {
		t.Fatal("expected draft state")
	}
}

func TestDraftNPC_UpdateContent_SetsFields(t *testing.T) {
	d := mustDraftNPC(t)
	d, events, err := d.UpdateContent("New Name", "A desc", "Player desc", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.NPCName() != "New Name" {
		t.Fatalf("expected 'New Name', got %q", d.NPCName())
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestDraftNPC_AssignToLocation_SetsLocation(t *testing.T) {
	d := mustDraftNPC(t)
	locID := identity.NewLocationID()
	d, events, err := d.AssignToLocation(locID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := d.Snapshot()
	if snap.LocationID.String() != locID.String() {
		t.Fatalf("expected locationID %s, got %s", locID, snap.LocationID)
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "located_at" {
		t.Fatalf("expected relationship 'located_at', got %q", el.Relationship)
	}
}

func TestDraftNPC_AssignToLocation_RequiresLocationID(t *testing.T) {
	d := mustDraftNPC(t)
	_, _, err := d.AssignToLocation(identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrLocationIDRequired) {
		t.Fatalf("expected ErrLocationIDRequired, got: %v", err)
	}
}

func TestDraftNPC_Activate_Succeeds_WhenSparse(t *testing.T) {
	d := mustDraftNPC(t)
	a, events, err := d.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Snapshot().State != "active" {
		t.Fatalf("expected state 'active', got %q", a.Snapshot().State)
	}
	if a.NPCName() != "Velleth" {
		t.Fatalf("expected name 'Velleth', got %q", a.NPCName())
	}
	if a.Snapshot().Description != "" {
		t.Fatalf("expected empty description, got %q", a.Snapshot().Description)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestActiveNPC_UpdateContent_SetsFields(t *testing.T) {
	a := mustActiveNPC(t)
	a, _, err := a.UpdateContent("Updated", "New desc", "New player desc", event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.NPCName() != "Updated" {
		t.Fatalf("expected 'Updated', got %q", a.NPCName())
	}
}

func TestActiveNPC_MoveTo_SetsLocation(t *testing.T) {
	a := mustActiveNPC(t)
	locID := identity.NewLocationID()
	a, events, err := a.MoveTo(locID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Snapshot().LocationID.String() != locID.String() {
		t.Fatalf("expected location %s, got %s", locID, a.Snapshot().LocationID)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events (EntityUpdated + EntityLinked), got %d", len(events))
	}
}

func TestActiveNPC_MoveTo_RequiresLocationID(t *testing.T) {
	a := mustActiveNPC(t)
	_, _, err := a.MoveTo(identity.LocationID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrLocationIDRequired) {
		t.Fatalf("expected ErrLocationIDRequired, got: %v", err)
	}
}

func TestActiveNPC_GoIdle_TransitionsToIdle(t *testing.T) {
	i := mustIdleNPC(t)
	if i.Snapshot().State != "idle" {
		t.Fatal("expected idle state")
	}
}

func TestActiveNPC_Archive_TransitionsToArchived(t *testing.T) {
	a := mustActiveNPC(t)
	arch, _, err := a.Archive(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arch.Snapshot().State != "archived" {
		t.Fatal("expected archived state")
	}
}

func TestActiveNPC_Reveal_SetsPlayerVisible(t *testing.T) {
	a := mustActiveNPC(t)
	sessID := identity.NewSessionID()
	a, events, err := a.Reveal([]string{"player1"}, sessID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.Snapshot().PlayerVisible {
		t.Fatal("expected playerVisible true")
	}
	er := events[0].(event.EntityRevealed)
	if er.SessionID != sessID.String() {
		t.Fatalf("expected sessionID %s, got %s", sessID, er.SessionID)
	}
}

func TestActiveNPC_Reveal_RequiresSessionID(t *testing.T) {
	a := mustActiveNPC(t)
	_, _, err := a.Reveal([]string{"player1"}, identity.SessionID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrSessionIDRequired) {
		t.Fatalf("expected ErrSessionIDRequired, got: %v", err)
	}
}

func TestActiveNPC_AssignMacGuffin_Succeeds(t *testing.T) {
	a := mustActiveNPC(t)
	mgID := identity.NewMacGuffinID()
	a, events, err := a.AssignMacGuffin(mgID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.Snapshot().MacGuffinIDs) != 1 {
		t.Fatalf("expected 1 macguffin, got %d", len(a.Snapshot().MacGuffinIDs))
	}
	el := events[0].(event.EntityLinked)
	if el.Relationship != "possesses" {
		t.Fatalf("expected relationship 'possesses', got %q", el.Relationship)
	}
}

func TestActiveNPC_AssignMacGuffin_RejectsDuplicate(t *testing.T) {
	a := mustActiveNPC(t)
	mgID := identity.NewMacGuffinID()
	a, _, _ = a.AssignMacGuffin(mgID, event.SourceGrimoire)
	_, _, err := a.AssignMacGuffin(mgID, event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinAlreadyAssigned) {
		t.Fatalf("expected ErrMacGuffinAlreadyAssigned, got: %v", err)
	}
}

func TestActiveNPC_AssignMacGuffin_RequiresID(t *testing.T) {
	a := mustActiveNPC(t)
	_, _, err := a.AssignMacGuffin(identity.MacGuffinID{}, event.SourceGrimoire)
	if !errors.Is(err, ErrMacGuffinIDRequired) {
		t.Fatalf("expected ErrMacGuffinIDRequired, got: %v", err)
	}
}

func TestIdleNPC_Activate_TransitionsToActive(t *testing.T) {
	i := mustIdleNPC(t)
	a, _, err := i.Activate(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestIdleNPC_Archive_TransitionsToArchived(t *testing.T) {
	i := mustIdleNPC(t)
	arch, _, err := i.Archive(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if arch.Snapshot().State != "archived" {
		t.Fatal("expected archived state")
	}
}

// --- Handle tests ---

func TestHandle_NewNPC_EntityUpdated_TransitionsToDraft(t *testing.T) {
	n := mustCreateNewNPC(t)
	result, err := n.Handle(event.EntityUpdated{
		EntityID: n.NarrativeCharacterID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(DraftNPC); !ok {
		t.Fatalf("expected DraftNPC, got %T", result)
	}
}

func TestHandle_NewNPC_UnexpectedEvent_ReturnsError(t *testing.T) {
	n := mustCreateNewNPC(t)
	_, err := n.Handle(event.EntityCreated{EntityID: "x", EntityType: "npc", Name: "x"})
	if !errors.Is(err, ErrNPCUnexpectedEvent) {
		t.Fatalf("expected ErrNPCUnexpectedEvent, got: %v", err)
	}
}

func TestHandle_DraftNPC_Content_UpdatesName(t *testing.T) {
	d := mustDraftNPC(t)
	result, err := d.Handle(event.EntityUpdated{
		EntityID: d.NarrativeCharacterID().String(), Field: "content",
		OldValue: "", NewValue: "Updated Name", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPCName() != "Updated Name" {
		t.Fatalf("expected 'Updated Name', got %q", result.NPCName())
	}
}

func TestHandle_DraftNPC_LocatedAt_SetsLocation(t *testing.T) {
	d := mustDraftNPC(t)
	locID := identity.NewLocationID()
	result, err := d.Handle(event.EntityLinked{
		EntityAID: d.NarrativeCharacterID().String(), EntityBID: locID.String(),
		Relationship: "located_at", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Snapshot().LocationID.String() != locID.String() {
		t.Fatalf("expected locationID %s, got %s", locID, result.Snapshot().LocationID)
	}
}

func TestHandle_DraftNPC_StatusActive_TransitionsToActive(t *testing.T) {
	d := mustDraftNPC(t)
	result, err := d.Handle(event.EntityUpdated{
		EntityID: d.NarrativeCharacterID().String(), Field: "status",
		OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ActiveNPC); !ok {
		t.Fatalf("expected ActiveNPC, got %T", result)
	}
}

func TestHandle_ActiveNPC_Content_UpdatesName(t *testing.T) {
	a := mustActiveNPC(t)
	result, err := a.Handle(event.EntityUpdated{
		EntityID: a.NarrativeCharacterID().String(), Field: "content",
		OldValue: "", NewValue: "New Name", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPCName() != "New Name" {
		t.Fatalf("expected 'New Name', got %q", result.NPCName())
	}
}

func TestHandle_ActiveNPC_Location_UpdatesLocation(t *testing.T) {
	a := mustActiveNPC(t)
	locID := identity.NewLocationID()
	result, err := a.Handle(event.EntityUpdated{
		EntityID: a.NarrativeCharacterID().String(), Field: "location",
		OldValue: "", NewValue: locID.String(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Snapshot().LocationID.String() != locID.String() {
		t.Fatalf("expected locationID %s, got %s", locID, result.Snapshot().LocationID)
	}
}

func TestHandle_ActiveNPC_Possesses_AddsMacGuffin(t *testing.T) {
	a := mustActiveNPC(t)
	mgID := identity.NewMacGuffinID()
	result, err := a.Handle(event.EntityLinked{
		EntityAID: a.NarrativeCharacterID().String(), EntityBID: mgID.String(),
		Relationship: "possesses", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Snapshot().MacGuffinIDs) != 1 {
		t.Fatalf("expected 1 macguffin, got %d", len(result.Snapshot().MacGuffinIDs))
	}
}

func TestHandle_ActiveNPC_Revealed_SetsVisible(t *testing.T) {
	a := mustActiveNPC(t)
	result, err := a.Handle(event.EntityRevealed{
		EntityID: a.NarrativeCharacterID().String(), RevealedTo: []string{"p1"},
		SessionID: identity.NewSessionID().String(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Snapshot().PlayerVisible {
		t.Fatal("expected playerVisible true")
	}
}

func TestHandle_ActiveNPC_StatusIdle_TransitionsToIdle(t *testing.T) {
	a := mustActiveNPC(t)
	result, err := a.Handle(event.EntityUpdated{
		EntityID: a.NarrativeCharacterID().String(), Field: "status",
		OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(IdleNPC); !ok {
		t.Fatalf("expected IdleNPC, got %T", result)
	}
}

func TestHandle_ActiveNPC_StatusArchived_TransitionsToArchived(t *testing.T) {
	a := mustActiveNPC(t)
	result, err := a.Handle(event.EntityUpdated{
		EntityID: a.NarrativeCharacterID().String(), Field: "status",
		OldValue: "active", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ArchivedNPC); !ok {
		t.Fatalf("expected ArchivedNPC, got %T", result)
	}
}

func TestHandle_IdleNPC_StatusActive_TransitionsToActive(t *testing.T) {
	i := mustIdleNPC(t)
	result, err := i.Handle(event.EntityUpdated{
		EntityID: i.NarrativeCharacterID().String(), Field: "status",
		OldValue: "idle", NewValue: "active", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ActiveNPC); !ok {
		t.Fatalf("expected ActiveNPC, got %T", result)
	}
}

func TestHandle_IdleNPC_StatusArchived_TransitionsToArchived(t *testing.T) {
	i := mustIdleNPC(t)
	result, err := i.Handle(event.EntityUpdated{
		EntityID: i.NarrativeCharacterID().String(), Field: "status",
		OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ArchivedNPC); !ok {
		t.Fatalf("expected ArchivedNPC, got %T", result)
	}
}

func TestHandle_ArchivedNPC_AnyEvent_ReturnsError(t *testing.T) {
	a := mustActiveNPC(t)
	arch, _, _ := a.Archive(event.SourceGrimoire)
	_, err := arch.Handle(event.EntityUpdated{
		EntityID: arch.NarrativeCharacterID().String(), Field: "status",
		OldValue: "archived", NewValue: "active", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrNPCUnexpectedEvent) {
		t.Fatalf("expected ErrNPCUnexpectedEvent, got: %v", err)
	}
}

func TestHandle_NPC_Immutability_OriginalUnchanged(t *testing.T) {
	n := mustCreateNewNPC(t)
	snapBefore := n.Snapshot()
	_, err := n.Handle(event.EntityUpdated{
		EntityID: n.NarrativeCharacterID().String(), Field: "status",
		OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatal(err)
	}
	snapAfter := n.Snapshot()
	if snapBefore.State != snapAfter.State {
		t.Fatal("Handle mutated the original entity state")
	}
}

// --- Replay tests ---

func TestReplayNPC_FullLifecycle(t *testing.T) {
	npcID := identity.NewNarrativeCharacterID()
	gameID := identity.NewGameID()
	locID := identity.NewLocationID()
	mgID := identity.NewMacGuffinID()
	sessID := identity.NewSessionID()

	events := []event.Event{
		event.EntityCreated{EntityID: npcID.String(), EntityType: "npc", Name: "Velleth", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: npcID.String(), Field: "status", OldValue: "new", NewValue: "draft", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: npcID.String(), Field: "content", OldValue: "", NewValue: "Velleth", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: npcID.String(), EntityBID: locID.String(), Relationship: "located_at", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: npcID.String(), Field: "status", OldValue: "draft", NewValue: "active", Source: event.SourceGrimoire},
		event.EntityLinked{EntityAID: npcID.String(), EntityBID: mgID.String(), Relationship: "possesses", Source: event.SourceGrimoire},
		event.EntityRevealed{EntityID: npcID.String(), RevealedTo: []string{"player1"}, SessionID: sessID.String(), Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: npcID.String(), Field: "status", OldValue: "active", NewValue: "idle", Source: event.SourceGrimoire},
		event.EntityUpdated{EntityID: npcID.String(), Field: "status", OldValue: "idle", NewValue: "archived", Source: event.SourceGrimoire},
	}

	result, err := ReplayNPC(gameID, events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ArchivedNPC); !ok {
		t.Fatalf("expected ArchivedNPC, got %T", result)
	}
	snap := result.Snapshot()
	if snap.Name != "Velleth" {
		t.Fatalf("expected name 'Velleth', got %q", snap.Name)
	}
	if !snap.PlayerVisible {
		t.Fatal("expected playerVisible true")
	}
	if len(snap.MacGuffinIDs) != 1 {
		t.Fatalf("expected 1 macguffin, got %d", len(snap.MacGuffinIDs))
	}
}

func TestReplayNPC_EmptyEvents_ReturnsError(t *testing.T) {
	_, err := ReplayNPC(identity.NewGameID(), nil)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

// --- Reconstitute tests ---

func TestReconstituteNPC_AllStates(t *testing.T) {
	states := []struct {
		state    string
		checkFn func(NPC) bool
	}{
		{"new", func(n NPC) bool { _, ok := n.(NewNPC); return ok }},
		{"draft", func(n NPC) bool { _, ok := n.(DraftNPC); return ok }},
		{"active", func(n NPC) bool { _, ok := n.(ActiveNPC); return ok }},
		{"idle", func(n NPC) bool { _, ok := n.(IdleNPC); return ok }},
		{"archived", func(n NPC) bool { _, ok := n.(ArchivedNPC); return ok }},
	}
	for _, tc := range states {
		t.Run(tc.state, func(t *testing.T) {
			snap := NPCSnapshot{
				ID:     identity.NewNarrativeCharacterID(),
				GameID: identity.NewGameID(),
				State:  tc.state,
				Name:   "Test NPC",
			}
			n, err := ReconstituteNPC(snap)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.checkFn(n) {
				t.Fatalf("expected %s state interface", tc.state)
			}
		})
	}
}

func TestReconstituteNPC_UnknownState_ReturnsError(t *testing.T) {
	snap := NPCSnapshot{
		ID:     identity.NewNarrativeCharacterID(),
		GameID: identity.NewGameID(),
		State:  "bogus",
		Name:   "Test",
	}
	_, err := ReconstituteNPC(snap)
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestReconstituteNPC_RoundTrip(t *testing.T) {
	a := mustActiveNPC(t)
	locID := identity.NewLocationID()
	a, _, _ = a.MoveTo(locID, event.SourceGrimoire)
	mgID := identity.NewMacGuffinID()
	a, _, _ = a.AssignMacGuffin(mgID, event.SourceGrimoire)

	snap1 := a.Snapshot()
	reconstituted, err := ReconstituteNPC(snap1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap2 := reconstituted.Snapshot()

	if snap1.ID.String() != snap2.ID.String() {
		t.Fatal("ID mismatch after round trip")
	}
	if snap1.Name != snap2.Name {
		t.Fatal("Name mismatch after round trip")
	}
	if snap1.State != snap2.State {
		t.Fatal("State mismatch after round trip")
	}
	if snap1.LocationID.String() != snap2.LocationID.String() {
		t.Fatal("LocationID mismatch after round trip")
	}
	if len(snap1.MacGuffinIDs) != len(snap2.MacGuffinIDs) {
		t.Fatal("MacGuffinIDs length mismatch after round trip")
	}
}

func TestReconstituteNPC_ActiveCanTransition(t *testing.T) {
	snap := NPCSnapshot{
		ID:          identity.NewNarrativeCharacterID(),
		GameID:      identity.NewGameID(),
		State:       "active",
		Name:        "Reconstituted NPC",
		Description: "Has a description",
	}
	n, _ := ReconstituteNPC(snap)
	an := n.(ActiveNPC)
	idle, _, err := an.GoIdle(event.SourceGrimoire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idle.Snapshot().State != "idle" {
		t.Fatal("expected idle state after transition")
	}
}
