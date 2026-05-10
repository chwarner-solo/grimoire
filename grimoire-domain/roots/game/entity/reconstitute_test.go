package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestReconstituteGame_NewState(t *testing.T) {
	snap := GameSnapshot{
		ID:    identity.NewGameID(),
		Name:  "Test Game",
		State: "new",
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(NewGame); !ok {
		t.Fatal("expected NewGame interface")
	}
}

func TestReconstituteGame_DraftState(t *testing.T) {
	snap := GameSnapshot{
		ID:    identity.NewGameID(),
		Name:  "Test Game",
		State: "draft",
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(DraftGame); !ok {
		t.Fatal("expected DraftGame interface")
	}
}

func TestReconstituteGame_ActiveState(t *testing.T) {
	cid := identity.NewCampaignID()
	snap := GameSnapshot{
		ID:                  identity.NewGameID(),
		Name:                "Test Game",
		State:               "active",
		CampaignIDs:         []identity.CampaignID{cid},
		ActiveCampaignCount: 1,
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(ActiveGame); !ok {
		t.Fatal("expected ActiveGame interface")
	}
}

func TestReconstituteGame_IdleState(t *testing.T) {
	snap := GameSnapshot{
		ID:          identity.NewGameID(),
		Name:        "Test Game",
		State:       "idle",
		CampaignIDs: []identity.CampaignID{identity.NewCampaignID()},
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(IdleGame); !ok {
		t.Fatal("expected IdleGame interface")
	}
}

func TestReconstituteGame_ArchivedState(t *testing.T) {
	snap := GameSnapshot{
		ID:    identity.NewGameID(),
		Name:  "Test Game",
		State: "archived",
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := g.(ArchivedGame); !ok {
		t.Fatal("expected ArchivedGame interface")
	}
}

func TestReconstituteGame_UnknownState_ReturnsError(t *testing.T) {
	snap := GameSnapshot{
		ID:    identity.NewGameID(),
		Name:  "Test Game",
		State: "bogus",
	}
	_, err := ReconstituteGame(snap)
	if !errors.Is(err, ErrInvalidGameState) {
		t.Fatalf("expected ErrInvalidGameState, got: %v", err)
	}
}

func TestReconstituteGame_RoundTrip(t *testing.T) {
	// Create → Draft → snapshot → reconstitute → snapshot → compare
	ng, _, _ := CreateGame(identity.NewGameID(), "Round Trip Game", event.SourceGrimoire)
	dg, _, _ := ng.AddNarrativeElement(identity.NewNarrativeID(), "Chapter 1", event.SourceGrimoire)
	cid := identity.NewCampaignID()
	dg, _, _ = dg.LinkCampaign(cid, event.SourceGrimoire)

	snap1 := dg.Snapshot()

	reconstituted, err := ReconstituteGame(snap1)
	if err != nil {
		t.Fatalf("reconstitute error: %v", err)
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
	if len(snap1.CampaignIDs) != len(snap2.CampaignIDs) {
		t.Fatal("CampaignIDs length mismatch after round trip")
	}
	for i := range snap1.CampaignIDs {
		if snap1.CampaignIDs[i].String() != snap2.CampaignIDs[i].String() {
			t.Fatalf("CampaignID[%d] mismatch after round trip", i)
		}
	}
}

func TestReconstituteGame_ActiveWithMultipleCampaigns_StaysActiveAfterOneIdle(t *testing.T) {
	c1 := identity.NewCampaignID()
	c2 := identity.NewCampaignID()
	snap := GameSnapshot{
		ID:                  identity.NewGameID(),
		Name:                "Multi Campaign",
		State:               "active",
		CampaignIDs:         []identity.CampaignID{c1, c2},
		ActiveCampaignCount: 2,
	}
	g, err := ReconstituteGame(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ag := g.(ActiveGame)
	result, err := ag.NotifyCampaignIdle(c1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(ActiveGame); !ok {
		t.Fatal("expected game to stay ActiveGame with one remaining active campaign")
	}
}

func TestReconstituteGame_ReconstitutedDraft_CanTransition(t *testing.T) {
	cid := identity.NewCampaignID()
	snap := GameSnapshot{
		ID:          identity.NewGameID(),
		Name:        "Draft Game",
		State:       "draft",
		CampaignIDs: []identity.CampaignID{cid},
	}
	g, _ := ReconstituteGame(snap)
	dg := g.(DraftGame)

	ag, err := dg.ActivateFromCampaign(cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ag.Snapshot().State != "active" {
		t.Fatal("expected active state after transition from reconstituted draft")
	}
}
