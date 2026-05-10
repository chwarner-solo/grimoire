package entity

import (
	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- NewFaction transitions ---

func (f *newFaction) BeginDraft(source event.Source) (DraftFaction, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "new",
		NewValue: "draft",
		Source:   source,
	}
	return &draftFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}

// --- DraftFaction transitions ---

func (f *draftFaction) AddMember(membershipID identity.FactionMembershipID, source event.Source) (DraftFaction, []event.Event, error) {
	if membershipID.IsZero() {
		return nil, nil, ErrMembershipIDRequired
	}
	if f.hasMember(membershipID) {
		return nil, nil, ErrMemberAlreadyExists
	}
	f.memberIDs = append(f.memberIDs, membershipID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    membershipID.String(),
		Relationship: "member_of",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *draftFaction) AddStandingLevel(level value.StandingLevel, source event.Source) (DraftFaction, []event.Event, error) {
	f.standingLevels = append(f.standingLevels, level)
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "standing_levels",
		OldValue: "",
		NewValue: level.Name(),
		Source:   source,
	}
	return f, []event.Event{evt}, nil
}

func (f *draftFaction) AddAlly(allyID identity.FactionID, source event.Source) (DraftFaction, []event.Event, error) {
	if f.isAtWarWith(allyID) {
		return nil, nil, ErrCannotBeAlliedAndAtWar{FactionID: allyID}
	}
	if f.isAlliedWith(allyID) {
		return nil, nil, ErrAllianceAlreadyExists
	}
	f.alliedWithIDs = append(f.alliedWithIDs, allyID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    allyID.String(),
		Relationship: "allied_with",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *draftFaction) DeclareWar(enemyID identity.FactionID, source event.Source) (DraftFaction, []event.Event, error) {
	if f.isAlliedWith(enemyID) {
		return nil, nil, ErrCannotBeAtWarAndAllied{FactionID: enemyID}
	}
	if f.isAtWarWith(enemyID) {
		return nil, nil, ErrWarAlreadyExists
	}
	f.atWarWithIDs = append(f.atWarWithIDs, enemyID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    enemyID.String(),
		Relationship: "at_war_with",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *draftFaction) Activate(source event.Source) (ActiveFaction, []event.Event, error) {
	if len(f.memberIDs) == 0 {
		return nil, nil, ErrCannotActivateWithoutMembers
	}
	if len(f.standingLevels) == 0 {
		return nil, nil, ErrCannotActivateWithoutStandingLevels
	}
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "draft",
		NewValue: "active",
		Source:   source,
	}
	return &activeFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}

// --- ActiveFaction transitions ---

func (f *activeFaction) AddMember(membershipID identity.FactionMembershipID, source event.Source) (ActiveFaction, []event.Event, error) {
	if membershipID.IsZero() {
		return nil, nil, ErrMembershipIDRequired
	}
	if f.hasMember(membershipID) {
		return nil, nil, ErrMemberAlreadyExists
	}
	f.memberIDs = append(f.memberIDs, membershipID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    membershipID.String(),
		Relationship: "member_of",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *activeFaction) AddAlly(allyID identity.FactionID, source event.Source) (ActiveFaction, []event.Event, error) {
	if f.isAtWarWith(allyID) {
		return nil, nil, ErrCannotBeAlliedAndAtWar{FactionID: allyID}
	}
	if f.isAlliedWith(allyID) {
		return nil, nil, ErrAllianceAlreadyExists
	}
	f.alliedWithIDs = append(f.alliedWithIDs, allyID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    allyID.String(),
		Relationship: "allied_with",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *activeFaction) DeclareWar(enemyID identity.FactionID, source event.Source) (ActiveFaction, []event.Event, error) {
	if f.isAlliedWith(enemyID) {
		return nil, nil, ErrCannotBeAtWarAndAllied{FactionID: enemyID}
	}
	if f.isAtWarWith(enemyID) {
		return nil, nil, ErrWarAlreadyExists
	}
	f.atWarWithIDs = append(f.atWarWithIDs, enemyID)
	evt := event.EntityLinked{
		EntityAID:    f.id.String(),
		EntityBID:    enemyID.String(),
		Relationship: "at_war_with",
		Source:       source,
	}
	return f, []event.Event{evt}, nil
}

func (f *activeFaction) MarkDormant(source event.Source) (IdleFaction, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "idle",
		Source:   source,
	}
	return &idleFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}

func (f *activeFaction) Archive(source event.Source) (ArchivedFaction, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "active",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}

// --- IdleFaction transitions ---

func (f *idleFaction) Reactivate(source event.Source) (ActiveFaction, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "active",
		Source:   source,
	}
	return &activeFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}

func (f *idleFaction) Archive(source event.Source) (ArchivedFaction, []event.Event, error) {
	evt := event.EntityUpdated{
		EntityID: f.id.String(),
		Field:    "status",
		OldValue: "idle",
		NewValue: "archived",
		Source:   source,
	}
	return &archivedFaction{factionCore: f.factionCore}, []event.Event{evt}, nil
}
