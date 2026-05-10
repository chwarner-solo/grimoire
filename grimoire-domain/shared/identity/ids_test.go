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

func TestNarrativeID_NewAndParse_Roundtrip(t *testing.T) {
	id := NewNarrativeID()
	parsed, err := ParseNarrativeID(id.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != id.String() {
		t.Fatalf("expected %s, got %s", id, parsed)
	}
}

func TestNarrativeID_NewIsNotZero(t *testing.T) {
	if NewNarrativeID().IsZero() {
		t.Fatal("new NarrativeID should not be zero")
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

func TestParseNarrativeID_InvalidInput_ReturnsError(t *testing.T) {
	_, err := ParseNarrativeID("not-valid")
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
