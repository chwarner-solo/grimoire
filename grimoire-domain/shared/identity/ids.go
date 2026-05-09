package identity

type GameID struct{ GrimoireID }
type CampaignID struct{ GrimoireID }
type SessionID struct{ GrimoireID }
type CharacterID struct{ GrimoireID }
type LocationID struct{ GrimoireID }
type NarrativeID struct{ GrimoireID }
type FactionID struct{ GrimoireID }

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

func NewNarrativeID() NarrativeID {
	return NarrativeID{NewGrimoireID()}
}

func ParseNarrativeID(s string) (NarrativeID, error) {
	base, err := ParseGrimoireID(s)
	if err != nil {
		return NarrativeID{}, err
	}
	return NarrativeID{base}, nil
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
