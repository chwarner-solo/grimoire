package entity

import "github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"

// --- NewCampaign transitions ---

func (c *newCampaign) AddCharacter(id identity.CharacterID) (NewCampaign, error) {
	if err := validateCharacterID(id); err != nil {
		return nil, err
	}
	if c.hasCharacter(id) {
		return nil, ErrCharacterAlreadyAdded
	}
	c.characterIDs = append(c.characterIDs, id)
	return c, nil
}

func (c *newCampaign) BeginFormation() (FormingCampaign, error) {
	return &formingCampaign{campaignCore: c.campaignCore}, nil
}

// --- FormingCampaign transitions ---

func (c *formingCampaign) AddCharacter(id identity.CharacterID) (FormingCampaign, error) {
	if err := validateCharacterID(id); err != nil {
		return nil, err
	}
	if c.hasCharacter(id) {
		return nil, ErrCharacterAlreadyAdded
	}
	c.characterIDs = append(c.characterIDs, id)
	return c, nil
}

func (c *formingCampaign) StartFirstSession(id identity.SessionID) (ActiveCampaign, error) {
	if err := validateSessionID(id); err != nil {
		return nil, err
	}
	if len(c.characterIDs) == 0 {
		return nil, ErrNoCharacters
	}
	return &activeCampaign{campaignCore: c.campaignCore}, nil
}

// --- ActiveCampaign transitions ---

func (c *activeCampaign) NotifySessionSummarized() (IdleCampaign, error) {
	return &idleCampaign{campaignCore: c.campaignCore}, nil
}

// --- IdleCampaign transitions ---

func (c *idleCampaign) StartNewSession(id identity.SessionID) (ActiveCampaign, error) {
	if err := validateSessionID(id); err != nil {
		return nil, err
	}
	return &activeCampaign{campaignCore: c.campaignCore}, nil
}

func (c *idleCampaign) Complete() (CompleteCampaign, error) {
	return &completeCampaign{campaignCore: c.campaignCore}, nil
}

// --- Helpers ---

func validateCharacterID(id identity.CharacterID) error {
	if id.IsZero() {
		return ErrCharacterIDRequired
	}
	return nil
}

func validateSessionID(id identity.SessionID) error {
	if id.IsZero() {
		return ErrSessionIDRequired
	}
	return nil
}
