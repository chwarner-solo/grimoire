package identity

import (
	"errors"
	"testing"
)

func TestParseGrimoireID_InvalidUUID_ReturnsErrInvalidID(t *testing.T) {
	_, err := ParseGrimoireID("not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
	if !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got: %v", err)
	}
}

func TestParseGrimoireID_InvalidUUID_ErrorContainsInput(t *testing.T) {
	input := "bad-input-value"
	_, err := ParseGrimoireID(input)
	if err == nil {
		t.Fatal("expected error for invalid UUID")
	}
	msg := err.Error()
	if !containsSubstring(msg, input) {
		t.Fatalf("expected error to contain input %q, got: %s", input, msg)
	}
}

func TestParseGrimoireID_ValidUUID_Succeeds(t *testing.T) {
	original := NewGrimoireID()
	parsed, err := ParseGrimoireID(original.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.String() != original.String() {
		t.Fatalf("expected %s, got %s", original, parsed)
	}
}

func TestNewGrimoireID_IsNotZero(t *testing.T) {
	id := NewGrimoireID()
	if id.IsZero() {
		t.Fatal("new GrimoireID should not be zero")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
