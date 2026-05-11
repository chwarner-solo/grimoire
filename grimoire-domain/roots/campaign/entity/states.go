package entity

import (
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Campaign is the sealed interface for the Campaign aggregate root.
// The unexported isCampaign() marker prevents external implementations.
type Campaign interface {
	isCampaign()
	CampaignID() identity.CampaignID
	CampaignName() string
	CurrentLocationID() identity.LocationID
	Snapshot() CampaignSnapshot
	Handle(evt event.Event) (Campaign, error)
}

// NewCampaign represents a campaign that has just been created.
// Characters can be added and the GM can begin formation.
type NewCampaign interface {
	Campaign
	AddCharacter(id identity.CharacterID, source event.Source) (NewCampaign, []event.Event, error)
	BeginFormation(source event.Source) (FormingCampaign, []event.Event, error)
}

// FormingCampaign represents a campaign where the GM is assembling the party.
// Characters can still be added, and the first session can start if at least one character exists.
type FormingCampaign interface {
	Campaign
	AddCharacter(id identity.CharacterID, source event.Source) (FormingCampaign, []event.Event, error)
	StartFirstSession(id identity.SessionID, date time.Time) (ActiveCampaign, []event.Event, error)
}

// ActiveCampaign represents a campaign with an in-progress session.
type ActiveCampaign interface {
	Campaign
	NotifySessionSummarized() (IdleCampaign, error)
	MoveToLocation(locationID identity.LocationID, source event.Source) (ActiveCampaign, []event.Event, error)
}

// IdleCampaign represents a campaign between sessions.
type IdleCampaign interface {
	Campaign
	StartNewSession(id identity.SessionID, date time.Time) (ActiveCampaign, []event.Event, error)
	Complete(source event.Source) (CompleteCampaign, []event.Event, error)
}

// CompleteCampaign is the terminal state. No transitions are possible.
type CompleteCampaign interface {
	Campaign
}
