package value

import "testing"

func TestCharacterTypeNPC_HasExpectedValue(t *testing.T) {
	if CharacterTypeNPC != "npc" {
		t.Fatalf("expected 'npc', got %q", CharacterTypeNPC)
	}
}

func TestCharacterTypePlayerCharacter_HasExpectedValue(t *testing.T) {
	if CharacterTypePlayerCharacter != "player-character" {
		t.Fatalf("expected 'player-character', got %q", CharacterTypePlayerCharacter)
	}
}
