package entity

import "github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"

// --- NewGame transitions ---

func (g *newGame) AddNarrativeElement(id identity.NarrativeID, name string) (DraftGame, error) {
	if err := validateNarrative(id, name); err != nil {
		return nil, err
	}
	return &draftGame{gameCore: g.gameCore}, nil
}

// --- DraftGame transitions ---

func (g *draftGame) AddNarrativeElement(id identity.NarrativeID, name string) (DraftGame, error) {
	if err := validateNarrative(id, name); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *draftGame) LinkCampaign(id identity.CampaignID) (DraftGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if g.hasCampaign(id) {
		return nil, ErrCampaignAlreadyLinked
	}
	g.campaignIDs = append(g.campaignIDs, id)
	return g, nil
}

func (g *draftGame) ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount = 1
	return &activeGame{gameCore: g.gameCore}, nil
}

// --- ActiveGame transitions ---

func (g *activeGame) LinkCampaign(id identity.CampaignID) (ActiveGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if g.hasCampaign(id) {
		return nil, ErrCampaignAlreadyLinked
	}
	g.campaignIDs = append(g.campaignIDs, id)
	return g, nil
}

func (g *activeGame) NotifyCampaignIdle(id identity.CampaignID) (Game, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount--
	if g.activeCampaignCount <= 0 {
		g.activeCampaignCount = 0
		return &idleGame{gameCore: g.gameCore}, nil
	}
	return g, nil
}

// --- IdleGame transitions ---

func (g *idleGame) ActivateFromCampaign(id identity.CampaignID) (ActiveGame, error) {
	if err := validateCampaignID(id); err != nil {
		return nil, err
	}
	if !g.hasCampaign(id) {
		return nil, ErrCampaignNotLinked
	}
	g.activeCampaignCount = 1
	return &activeGame{gameCore: g.gameCore}, nil
}

func (g *idleGame) Archive() (ArchivedGame, error) {
	return &archivedGame{gameCore: g.gameCore}, nil
}

// --- Helpers ---

func validateNarrative(id identity.NarrativeID, name string) error {
	if id.IsZero() {
		return ErrNarrativeIDRequired
	}
	if len(name) == 0 {
		return ErrNarrativeNameRequired
	}
	return nil
}

func validateCampaignID(id identity.CampaignID) error {
	if id.IsZero() {
		return ErrCampaignIDRequired
	}
	return nil
}
