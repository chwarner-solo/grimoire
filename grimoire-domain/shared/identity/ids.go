package identity

import (
	"encoding/json"
	"fmt"
)

// EventID identifies an event using a ULID string.
// Unlike aggregate IDs (UUID-based), EventIDs use ULID for global sortability.
// The domain defines the type; infrastructure generates ULID values.
type EventID struct {
	value string
}

func ParseEventID(s string) (EventID, error) {
	if s == "" {
		return EventID{}, fmt.Errorf("%w: empty event id", ErrInvalidID)
	}
	return EventID{value: s}, nil
}

func (id EventID) String() string { return id.value }
func (id EventID) IsZero() bool   { return id.value == "" }

func (id EventID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.value)
}

func (id *EventID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	id.value = s
	return nil
}

type GameID struct{ GrimoireID }
type CampaignID struct{ GrimoireID }
type SessionID struct{ GrimoireID }
type CharacterID struct{ GrimoireID }
type LocationID struct{ GrimoireID }
type FactionID struct{ GrimoireID }
type MasterNarrativeID struct{ GrimoireID }
type CampaignNarrativeID struct{ GrimoireID }
type BeatID struct{ GrimoireID }
type ActID struct{ GrimoireID }
type SecretID struct{ GrimoireID }
type LoreID struct{ GrimoireID }
type DecisionID struct{ GrimoireID }
type RevelationID struct{ GrimoireID }

func NewGameID() GameID {
	return GameID{NewGrimoireID()}
}

func ParseGameID(s string) (GameID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return GameID{}, err
	}
	return GameID{base}, nil
}

func NewCampaignID() CampaignID {
	return CampaignID{NewGrimoireID()}
}

func ParseCampaignID(s string) (CampaignID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return CampaignID{}, err
	}
	return CampaignID{base}, nil
}

func NewSessionID() SessionID {
	return SessionID{NewGrimoireID()}
}

func ParseSessionID(s string) (SessionID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return SessionID{}, err
	}
	return SessionID{base}, nil
}

func NewCharacterID() CharacterID {
	return CharacterID{NewGrimoireID()}
}

func ParseCharacterID(s string) (CharacterID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return CharacterID{}, err
	}
	return CharacterID{base}, nil
}

func NewLocationID() LocationID {
	return LocationID{NewGrimoireID()}
}

func ParseLocationID(s string) (LocationID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return LocationID{}, err
	}
	return LocationID{base}, nil
}

func NewFactionID() FactionID {
	return FactionID{NewGrimoireID()}
}

func ParseFactionID(s string) (FactionID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return FactionID{}, err
	}
	return FactionID{base}, nil
}

func NewMasterNarrativeID() MasterNarrativeID {
	return MasterNarrativeID{NewGrimoireID()}
}

func ParseMasterNarrativeID(s string) (MasterNarrativeID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return MasterNarrativeID{}, err
	}
	return MasterNarrativeID{base}, nil
}

func NewCampaignNarrativeID() CampaignNarrativeID {
	return CampaignNarrativeID{NewGrimoireID()}
}

func ParseCampaignNarrativeID(s string) (CampaignNarrativeID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return CampaignNarrativeID{}, err
	}
	return CampaignNarrativeID{base}, nil
}

func NewBeatID() BeatID {
	return BeatID{NewGrimoireID()}
}

func ParseBeatID(s string) (BeatID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return BeatID{}, err
	}
	return BeatID{base}, nil
}

func NewActID() ActID {
	return ActID{NewGrimoireID()}
}

func ParseActID(s string) (ActID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return ActID{}, err
	}
	return ActID{base}, nil
}

func NewSecretID() SecretID {
	return SecretID{NewGrimoireID()}
}

func ParseSecretID(s string) (SecretID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return SecretID{}, err
	}
	return SecretID{base}, nil
}

func NewLoreID() LoreID {
	return LoreID{NewGrimoireID()}
}

func ParseLoreID(s string) (LoreID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return LoreID{}, err
	}
	return LoreID{base}, nil
}

func NewDecisionID() DecisionID {
	return DecisionID{NewGrimoireID()}
}

func ParseDecisionID(s string) (DecisionID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return DecisionID{}, err
	}
	return DecisionID{base}, nil
}

func NewRevelationID() RevelationID {
	return RevelationID{NewGrimoireID()}
}

func ParseRevelationID(s string) (RevelationID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return RevelationID{}, err
	}
	return RevelationID{base}, nil
}
