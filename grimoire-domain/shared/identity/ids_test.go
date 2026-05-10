package identity

import (
	"testing"
)

func TestGameID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewGameID()
	parsed, err := ParseGameID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestGameID_NewIsNotZero(t *testing.T) {
	if NewGameID().IsZero() {
		t.Fatal("new GameID should not be zero")
	}
}

func TestCampaignID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewCampaignID()
	parsed, err := ParseCampaignID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestCampaignID_NewIsNotZero(t *testing.T) {
	if NewCampaignID().IsZero() {
		t.Fatal("new CampaignID should not be zero")
	}
}

func TestSessionID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewSessionID()
	parsed, err := ParseSessionID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestSessionID_NewIsNotZero(t *testing.T) {
	if NewSessionID().IsZero() {
		t.Fatal("new SessionID should not be zero")
	}
}

func TestCharacterID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewCharacterID()
	parsed, err := ParseCharacterID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestCharacterID_NewIsNotZero(t *testing.T) {
	if NewCharacterID().IsZero() {
		t.Fatal("new CharacterID should not be zero")
	}
}

func TestLocationID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewLocationID()
	parsed, err := ParseLocationID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestLocationID_NewIsNotZero(t *testing.T) {
	if NewLocationID().IsZero() {
		t.Fatal("new LocationID should not be zero")
	}
}

func TestFactionID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewFactionID()
	parsed, err := ParseFactionID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestFactionID_NewIsNotZero(t *testing.T) {
	if NewFactionID().IsZero() {
		t.Fatal("new FactionID should not be zero")
	}
}

func TestMasterNarrativeID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewMasterNarrativeID()
	parsed, err := ParseMasterNarrativeID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestMasterNarrativeID_NewIsNotZero(t *testing.T) {
	if NewMasterNarrativeID().IsZero() {
		t.Fatal("new MasterNarrativeID should not be zero")
	}
}

func TestMasterNarrativeID_ZeroValue_IsZero(t *testing.T) {
	var id MasterNarrativeID
	if !id.IsZero() {
		t.Fatal("zero-value MasterNarrativeID should be zero")
	}
}

func TestCampaignNarrativeID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewCampaignNarrativeID()
	parsed, err := ParseCampaignNarrativeID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestCampaignNarrativeID_NewIsNotZero(t *testing.T) {
	if NewCampaignNarrativeID().IsZero() {
		t.Fatal("new CampaignNarrativeID should not be zero")
	}
}

func TestCampaignNarrativeID_ZeroValue_IsZero(t *testing.T) {
	var id CampaignNarrativeID
	if !id.IsZero() {
		t.Fatal("zero-value CampaignNarrativeID should be zero")
	}
}

func TestBeatID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewBeatID()
	parsed, err := ParseBeatID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestBeatID_NewIsNotZero(t *testing.T) {
	if NewBeatID().IsZero() {
		t.Fatal("new BeatID should not be zero")
	}
}

func TestBeatID_ZeroValue_IsZero(t *testing.T) {
	var id BeatID
	if !id.IsZero() {
		t.Fatal("zero-value BeatID should be zero")
	}
}

func TestActID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewActID()
	parsed, err := ParseActID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestActID_NewIsNotZero(t *testing.T) {
	if NewActID().IsZero() {
		t.Fatal("new ActID should not be zero")
	}
}

func TestActID_ZeroValue_IsZero(t *testing.T) {
	var id ActID
	if !id.IsZero() {
		t.Fatal("zero-value ActID should be zero")
	}
}

func TestSecretID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewSecretID()
	parsed, err := ParseSecretID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestSecretID_NewIsNotZero(t *testing.T) {
	if NewSecretID().IsZero() {
		t.Fatal("new SecretID should not be zero")
	}
}

func TestSecretID_ZeroValue_IsZero(t *testing.T) {
	var id SecretID
	if !id.IsZero() {
		t.Fatal("zero-value SecretID should be zero")
	}
}

func TestLoreID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewLoreID()
	parsed, err := ParseLoreID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestLoreID_NewIsNotZero(t *testing.T) {
	if NewLoreID().IsZero() {
		t.Fatal("new LoreID should not be zero")
	}
}

func TestLoreID_ZeroValue_IsZero(t *testing.T) {
	var id LoreID
	if !id.IsZero() {
		t.Fatal("zero-value LoreID should be zero")
	}
}

func TestDecisionID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewDecisionID()
	parsed, err := ParseDecisionID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestDecisionID_NewIsNotZero(t *testing.T) {
	if NewDecisionID().IsZero() {
		t.Fatal("new DecisionID should not be zero")
	}
}

func TestDecisionID_ZeroValue_IsZero(t *testing.T) {
	var id DecisionID
	if !id.IsZero() {
		t.Fatal("zero-value DecisionID should be zero")
	}
}

func TestRevelationID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewRevelationID()
	parsed, err := ParseRevelationID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestRevelationID_NewIsNotZero(t *testing.T) {
	if NewRevelationID().IsZero() {
		t.Fatal("new RevelationID should not be zero")
	}
}

func TestRevelationID_ZeroValue_IsZero(t *testing.T) {
	var id RevelationID
	if !id.IsZero() {
		t.Fatal("zero-value RevelationID should be zero")
	}
}

// --- Parse error tests ---

func TestParseGameID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseGameID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseCampaignID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseCampaignID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseSessionID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseSessionID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseCharacterID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseCharacterID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseLocationID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseLocationID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseFactionID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseFactionID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseMasterNarrativeID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseMasterNarrativeID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseCampaignNarrativeID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseCampaignNarrativeID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseBeatID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseBeatID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseActID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseActID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseSecretID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseSecretID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseLoreID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseLoreID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseDecisionID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseDecisionID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestParseRevelationID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseRevelationID("not-valid")
	if err == nil {
		t.Fatal("expected error for invalid input")
	}
}

// --- EventID tests ---

func TestParseEventID_ValidString_Succeeds(t *testing.T) {
	id, err := ParseEventID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Fatalf("expected '01ARZ3NDEKTSV4RRFFQ69G5FAV', got %q", id.String())
	}
}

func TestParseEventID_EmptyString_ReturnsError(t *testing.T) {
	_, err := ParseEventID("")
	if err == nil {
		t.Fatal("expected error for empty event ID")
	}
}

func TestEventID_IsZero_WhenEmpty(t *testing.T) {
	var id EventID
	if !id.IsZero() {
		t.Fatal("zero-value EventID should be zero")
	}
}

func TestEventID_IsNotZero_WhenParsed(t *testing.T) {
	id, _ := ParseEventID("some-ulid")
	if id.IsZero() {
		t.Fatal("parsed EventID should not be zero")
	}
}
