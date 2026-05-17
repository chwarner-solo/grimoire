package interactor

import (
	"context"
	"fmt"
	"time"

	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

type AddFactionMemberRequest struct {
	CallerID     string
	FactionID    identity.FactionID
	MembershipID identity.FactionMembershipID
	NPCID        identity.NarrativeCharacterID
	Rank         string
	GameID       identity.GameID
	Source       event.Source
}

type AddFactionMemberResult struct {
	Events []event.Event
}

type AddFactionMemberInteractor struct {
	gameRepo    GameRepository
	factionRepo FactionRepository
	bus         event.EventBus
}

func NewAddFactionMemberInteractor(gameRepo GameRepository, factionRepo FactionRepository, bus event.EventBus) *AddFactionMemberInteractor {
	return &AddFactionMemberInteractor{gameRepo: gameRepo, factionRepo: factionRepo, bus: bus}
}

func (i *AddFactionMemberInteractor) Execute(ctx context.Context, req AddFactionMemberRequest) (AddFactionMemberResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return AddFactionMemberResult{}, fmt.Errorf("%w: %v", ErrGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return AddFactionMemberResult{}, shared.ErrUnauthorized
	}

	faction, err := i.factionRepo.Load(ctx, req.FactionID)
	if err != nil {
		return AddFactionMemberResult{}, fmt.Errorf("%w: %v", ErrFactionNotFound, err)
	}

	var updated factionentity.Faction
	var entityEvents []event.Event

	switch f := faction.(type) {
	case factionentity.DraftFaction:
		updated, entityEvents, err = f.AddMember(req.MembershipID, req.Source)
	case factionentity.ActiveFaction:
		updated, entityEvents, err = f.AddMember(req.MembershipID, req.Source)
	default:
		return AddFactionMemberResult{}, fmt.Errorf("%w: can only add members to Draft or Active factions", ErrInvalidFactionState)
	}
	if err != nil {
		return AddFactionMemberResult{}, err
	}

	if err := i.factionRepo.Save(ctx, updated); err != nil {
		return AddFactionMemberResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	membership, err := factionentity.NewFactionMembership(req.MembershipID, req.FactionID, req.NPCID, req.Rank)
	if err != nil {
		return AddFactionMemberResult{}, err
	}

	if err := i.factionRepo.SaveMembership(ctx, membership); err != nil {
		return AddFactionMemberResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	linkedEvt := event.EntityLinked{
		EntityAID:    req.NPCID.String(),
		EntityBID:    req.FactionID.String(),
		Relationship: "member_of",
		Source:       req.Source,
	}
	allEvents := append(entityEvents, linkedEvt)

	for _, evt := range allEvents {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.FactionID.GrimoireID),
			AggregateType: event.AggregateFaction,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return AddFactionMemberResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return AddFactionMemberResult{Events: allEvents}, nil
}
