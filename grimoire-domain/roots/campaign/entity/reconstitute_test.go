package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

func TestReconstituteCampaign_NewState(t *testing.T) {
	snap := CampaignSnapshot{
		ID:     identity.NewCampaignID(),
		Name:   "Test Campaign",
		GameID: identity.NewGameID(),
		State:  "new",
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(NewCampaign); !ok {
		t.Fatal("expected NewCampaign interface")
	}
}

func TestReconstituteCampaign_FormingState(t *testing.T) {
	snap := CampaignSnapshot{
		ID:           identity.NewCampaignID(),
		Name:         "Test Campaign",
		GameID:       identity.NewGameID(),
		State:        "forming",
		CharacterIDs: []identity.CharacterID{identity.NewCharacterID()},
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(FormingCampaign); !ok {
		t.Fatal("expected FormingCampaign interface")
	}
}

func TestReconstituteCampaign_ActiveState(t *testing.T) {
	snap := CampaignSnapshot{
		ID:           identity.NewCampaignID(),
		Name:         "Test Campaign",
		GameID:       identity.NewGameID(),
		State:        "active",
		CharacterIDs: []identity.CharacterID{identity.NewCharacterID()},
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(ActiveCampaign); !ok {
		t.Fatal("expected ActiveCampaign interface")
	}
}

func TestReconstituteCampaign_IdleState(t *testing.T) {
	snap := CampaignSnapshot{
		ID:           identity.NewCampaignID(),
		Name:         "Test Campaign",
		GameID:       identity.NewGameID(),
		State:        "idle",
		CharacterIDs: []identity.CharacterID{identity.NewCharacterID()},
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(IdleCampaign); !ok {
		t.Fatal("expected IdleCampaign interface")
	}
}

func TestReconstituteCampaign_CompleteState(t *testing.T) {
	snap := CampaignSnapshot{
		ID:     identity.NewCampaignID(),
		Name:   "Test Campaign",
		GameID: identity.NewGameID(),
		State:  "complete",
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := c.(CompleteCampaign); !ok {
		t.Fatal("expected CompleteCampaign interface")
	}
}

func TestReconstituteCampaign_UnknownState_ReturnsError(t *testing.T) {
	snap := CampaignSnapshot{
		ID:     identity.NewCampaignID(),
		Name:   "Test Campaign",
		GameID: identity.NewGameID(),
		State:  "bogus",
	}
	_, err := ReconstituteCampaign(snap)
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestReconstituteCampaign_RoundTrip(t *testing.T) {
	// Create → AddCharacter → BeginFormation → snapshot → reconstitute → snapshot → compare
	nc, _, _ := CreateCampaign(identity.NewCampaignID(), "Round Trip Campaign", identity.NewGameID(), event.SourceGrimoire)
	charID := identity.NewCharacterID()
	nc, _, _ = nc.AddCharacter(charID, event.SourceGrimoire)
	fc, _, _ := nc.BeginFormation(event.SourceGrimoire)

	snap1 := fc.Snapshot()

	reconstituted, err := ReconstituteCampaign(snap1)
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
	if snap1.GameID.String() != snap2.GameID.String() {
		t.Fatal("GameID mismatch after round trip")
	}
	if len(snap1.CharacterIDs) != len(snap2.CharacterIDs) {
		t.Fatal("CharacterIDs length mismatch after round trip")
	}
	for i := range snap1.CharacterIDs {
		if snap1.CharacterIDs[i].String() != snap2.CharacterIDs[i].String() {
			t.Fatalf("CharacterID[%d] mismatch after round trip", i)
		}
	}
	if snap1.CurrentLocationID.String() != snap2.CurrentLocationID.String() {
		t.Fatal("CurrentLocationID mismatch after round trip")
	}
}

func TestReconstituteCampaign_WithCurrentLocationID(t *testing.T) {
	locID := identity.NewLocationID()
	snap := CampaignSnapshot{
		ID:                identity.NewCampaignID(),
		Name:              "Campaign",
		GameID:            identity.NewGameID(),
		State:             "active",
		CharacterIDs:      []identity.CharacterID{identity.NewCharacterID()},
		CurrentLocationID: locID,
	}
	c, err := ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.CurrentLocationID().String() != locID.String() {
		t.Fatalf("expected currentLocationID %s, got %s", locID, c.CurrentLocationID())
	}
}

func TestReconstituteCampaign_ReconstitutedForming_CanTransition(t *testing.T) {
	charID := identity.NewCharacterID()
	snap := CampaignSnapshot{
		ID:           identity.NewCampaignID(),
		Name:         "Forming Campaign",
		GameID:       identity.NewGameID(),
		State:        "forming",
		CharacterIDs: []identity.CharacterID{charID},
	}
	c, _ := ReconstituteCampaign(snap)
	fc := c.(FormingCampaign)

	ac, _, err := fc.StartFirstSession(identity.NewSessionID(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ac.Snapshot().State != "active" {
		t.Fatal("expected active state after transition from reconstituted forming")
	}
}
