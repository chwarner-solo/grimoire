package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// Faction is the sealed interface for the Faction aggregate root.
type Faction interface {
	isFaction()
	FactionID() identity.FactionID
	FactionName() string
	GameID() identity.GameID
	Snapshot() FactionSnapshot
	Handle(evt event.Event) (Faction, error)
}

// NewFaction represents a faction that has just been created.
type NewFaction interface {
	Faction
	BeginDraft(source event.Source) (DraftFaction, []event.Event, error)
}

// DraftFaction represents a faction being built by the GM.
type DraftFaction interface {
	Faction
	AddMember(membershipID identity.FactionMembershipID, source event.Source) (DraftFaction, []event.Event, error)
	AddStandingLevel(level value.StandingLevel, source event.Source) (DraftFaction, []event.Event, error)
	AddAlly(allyID identity.FactionID, source event.Source) (DraftFaction, []event.Event, error)
	DeclareWar(enemyID identity.FactionID, source event.Source) (DraftFaction, []event.Event, error)
	Activate(source event.Source) (ActiveFaction, []event.Event, error)
}

// ActiveFaction represents a faction operating in the world.
type ActiveFaction interface {
	Faction
	AddMember(membershipID identity.FactionMembershipID, source event.Source) (ActiveFaction, []event.Event, error)
	AddStandingLevel(level value.StandingLevel, source event.Source) (ActiveFaction, []event.Event, error)
	AddAlly(allyID identity.FactionID, source event.Source) (ActiveFaction, []event.Event, error)
	DeclareWar(enemyID identity.FactionID, source event.Source) (ActiveFaction, []event.Event, error)
	Reveal(to []string, sessionID identity.SessionID, source event.Source) (ActiveFaction, []event.Event, error)
	MarkDormant(source event.Source) (IdleFaction, []event.Event, error)
	Archive(source event.Source) (ArchivedFaction, []event.Event, error)
}

// IdleFaction represents a dormant faction.
type IdleFaction interface {
	Faction
	Reactivate(source event.Source) (ActiveFaction, []event.Event, error)
	Archive(source event.Source) (ArchivedFaction, []event.Event, error)
}

// ArchivedFaction is the terminal state. No transitions are possible.
type ArchivedFaction interface {
	Faction
}
