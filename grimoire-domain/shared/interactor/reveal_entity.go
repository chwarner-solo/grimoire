package interactor

import (
	"context"
	"fmt"
	"time"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/value"
)

type RevealEntityRequest struct {
	CallerID   string
	EntityID   string
	EntityType value.EntityType
	GameID     identity.GameID
	RevealedTo []string
	SessionID  identity.SessionID
	Source     event.Source
}

type RevealEntityResult struct {
	Events []event.Event
}

type RevealEntityInteractor struct {
	gameRepo      RevealGameRepository
	npcRepo       RevealNPCRepository
	pcRepo        RevealPlayerCharacterRepository
	macguffinRepo RevealMacGuffinRepository
	factionRepo   RevealFactionRepository
	bus           event.EventBus
}

func NewRevealEntityInteractor(
	gameRepo RevealGameRepository,
	npcRepo RevealNPCRepository,
	pcRepo RevealPlayerCharacterRepository,
	macguffinRepo RevealMacGuffinRepository,
	factionRepo RevealFactionRepository,
	bus event.EventBus,
) *RevealEntityInteractor {
	return &RevealEntityInteractor{
		gameRepo:      gameRepo,
		npcRepo:       npcRepo,
		pcRepo:        pcRepo,
		macguffinRepo: macguffinRepo,
		factionRepo:   factionRepo,
		bus:           bus,
	}
}

func (i *RevealEntityInteractor) Execute(ctx context.Context, req RevealEntityRequest) (RevealEntityResult, error) {
	game, err := i.gameRepo.Load(ctx, req.GameID)
	if err != nil {
		return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealGameNotFound, err)
	}

	if req.CallerID != game.GMID() {
		return RevealEntityResult{}, ErrUnauthorized
	}

	if len(req.RevealedTo) == 0 {
		return RevealEntityResult{}, fmt.Errorf("reveal: revealed_to must not be empty")
	}

	if req.SessionID.IsZero() {
		return RevealEntityResult{}, fmt.Errorf("reveal: session_id is required")
	}

	var events []event.Event
	var aggregateID identity.GrimoireID

	switch req.EntityType {
	case value.EntityTypeNPC:
		id, err := identity.ParseNarrativeCharacterID(req.EntityID)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		snap, err := i.npcRepo.Load(ctx, id)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		npc, err := characterentity.ReconstituteNPC(snap)
		if err != nil {
			return RevealEntityResult{}, err
		}
		an, ok := npc.(characterentity.ActiveNPC)
		if !ok {
			return RevealEntityResult{}, fmt.Errorf("%w: npc must be Active to reveal", ErrInvalidEntityState)
		}
		updated, evts, err := an.Reveal(req.RevealedTo, req.SessionID, req.Source)
		if err != nil {
			return RevealEntityResult{}, err
		}
		if err := i.npcRepo.Save(ctx, updated.Snapshot()); err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealRepositorySaveFailed, err)
		}
		events = evts
		aggregateID = identity.GrimoireID(id.GrimoireID)

	case value.EntityTypePlayerCharacter:
		id, err := identity.ParsePlayerCharacterID(req.EntityID)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		snap, err := i.pcRepo.LoadCore(ctx, id)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		pc := characterentity.ReconstitutePlayerCharacter(snap)
		updated, evts, err := pc.Reveal(req.RevealedTo, req.SessionID, req.Source)
		if err != nil {
			return RevealEntityResult{}, err
		}
		if err := i.pcRepo.SaveCore(ctx, updated.Snapshot()); err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealRepositorySaveFailed, err)
		}
		events = evts
		aggregateID = identity.GrimoireID(id.GrimoireID)

	case value.EntityTypeMacGuffin:
		id, err := identity.ParseMacGuffinID(req.EntityID)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		snap, err := i.macguffinRepo.Load(ctx, id)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		mg := characterentity.ReconstituteMacGuffin(snap)
		updated, evts, err := mg.Reveal(req.RevealedTo, req.SessionID, req.Source)
		if err != nil {
			return RevealEntityResult{}, err
		}
		if err := i.macguffinRepo.Save(ctx, updated.Snapshot()); err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealRepositorySaveFailed, err)
		}
		events = evts
		aggregateID = identity.GrimoireID(id.GrimoireID)

	case value.EntityTypeFaction:
		id, err := identity.ParseFactionID(req.EntityID)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		faction, err := i.factionRepo.Load(ctx, id)
		if err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrEntityNotFound, err)
		}
		af, ok := faction.(factionentity.ActiveFaction)
		if !ok {
			return RevealEntityResult{}, fmt.Errorf("%w: faction must be Active to reveal", ErrInvalidEntityState)
		}
		updated, evts, err := af.Reveal(req.RevealedTo, req.SessionID, req.Source)
		if err != nil {
			return RevealEntityResult{}, err
		}
		if err := i.factionRepo.Save(ctx, updated); err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealRepositorySaveFailed, err)
		}
		events = evts
		aggregateID = identity.GrimoireID(id.GrimoireID)

	default:
		return RevealEntityResult{}, fmt.Errorf("%w: %q", ErrUnsupportedEntityType, req.EntityType)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   aggregateID,
			AggregateType: event.AggregateType(req.EntityType),
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return RevealEntityResult{}, fmt.Errorf("%w: %v", ErrRevealEventDispatchFailed, err)
		}
	}

	return RevealEntityResult{Events: events}, nil
}
