package value

import (
	"errors"
	"testing"
)

func TestNewStandingLevel_ValidInputs_Succeeds(t *testing.T) {
	sl, err := NewStandingLevel("Kindly", 500, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl.Name() != "Kindly" {
		t.Fatalf("expected name 'Kindly', got %q", sl.Name())
	}
	if sl.Threshold() != 500 {
		t.Fatalf("expected threshold 500, got %d", sl.Threshold())
	}
	if sl.Ordinal() != 2 {
		t.Fatalf("expected ordinal 2, got %d", sl.Ordinal())
	}
}

func TestNewStandingLevel_EmptyName_ReturnsError(t *testing.T) {
	_, err := NewStandingLevel("", 0, 0)
	if !errors.Is(err, ErrStandingLevelNameRequired) {
		t.Fatalf("expected ErrStandingLevelNameRequired, got %v", err)
	}
}

func TestNewStandingLevel_NegativeThreshold_ReturnsError(t *testing.T) {
	_, err := NewStandingLevel("Bad", -1, 0)
	if !errors.Is(err, ErrStandingLevelThresholdNegative) {
		t.Fatalf("expected ErrStandingLevelThresholdNegative, got %v", err)
	}
}

func TestNewStandingLevel_NegativeOrdinal_ReturnsError(t *testing.T) {
	_, err := NewStandingLevel("Bad", 0, -1)
	if !errors.Is(err, ErrStandingLevelOrdinalNegative) {
		t.Fatalf("expected ErrStandingLevelOrdinalNegative, got %v", err)
	}
}

func TestNewStandingLevel_ZeroThresholdAndOrdinal_Succeeds(t *testing.T) {
	sl, err := NewStandingLevel("Neutral", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sl.Name() != "Neutral" {
		t.Fatalf("expected name 'Neutral', got %q", sl.Name())
	}
}
