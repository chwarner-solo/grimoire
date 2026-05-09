package entity

import (
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Test helpers ---

func mustCreateNewCampaign(t *testing.T) NewCampaign {
	t.Helper()
	c, err := CreateCampaign(identity.NewCampaignID(), "The Long Night", identity.NewGameID())
	if err != nil {
		t.Fatalf("mustCreateNewCampaign: %v", err)
	}
	return c
}

func mustReachForming(t *testing.T) FormingCampaign {
	t.Helper()
	nc := mustCreateNewCampaign(t)
	nc, err := nc.AddCharacter(identity.NewCharacterID())
	if err != nil {
		t.Fatalf("mustReachForming add character: %v", err)
	}
	fc, err := nc.BeginFormation()
	if err != nil {
		t.Fatalf("mustReachForming: %v", err)
	}
	return fc
}

func mustReachActive(t *testing.T) ActiveCampaign {
	t.Helper()
	fc := mustReachForming(t)
	ac, err := fc.StartFirstSession(identity.NewSessionID())
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
	_, err := CreateCampaign(identity.CampaignID{}, "The Long Night", identity.NewGameID())
	if !errors.Is(err, ErrCampaignIDRequired) {
		t.Fatalf("expected ErrCampaignIDRequired, got: %v", err)
	}
}

func TestCreateCampaign_RequiresNonEmptyName(t *testing.T) {
	_, err := CreateCampaign(identity.NewCampaignID(), "", identity.NewGameID())
	if !errors.Is(err, ErrCampaignNameRequired) {
		t.Fatalf("expected ErrCampaignNameRequired, got: %v", err)
	}
}

func TestCreateCampaign_RequiresNonZeroGameID(t *testing.T) {
	_, err := CreateCampaign(identity.NewCampaignID(), "The Long Night", identity.GameID{})
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
	nc2, err := nc.AddCharacter(charID)
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
	nc, _ = nc.AddCharacter(charID)
	_, err := nc.AddCharacter(charID)
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
	fc2, err := fc.AddCharacter(charID)
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
	_, err = fc.StartFirstSession(identity.NewSessionID())
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
	ac, err := ic.StartNewSession(identity.NewSessionID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ac.Snapshot().State != "active" {
		t.Fatal("expected active state")
	}
}

func TestIdleCampaign_Complete_TransitionsToComplete(t *testing.T) {
	ic := mustReachIdle(t)
	cc, err := ic.Complete()
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
	cc, _ := ic.Complete()
	if cc.CampaignID().String() != campaignID.String() {
		t.Fatal("complete campaign should preserve CampaignID")
	}
	if cc.CampaignName() != campaignName {
		t.Fatal("complete campaign should preserve CampaignName")
	}
}
