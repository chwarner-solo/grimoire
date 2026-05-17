package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// ArchiveGameRequest holds the input for archiving a Game.
type ArchiveGameRequest struct {
	CallerID string
	GameID   identity.GameID
	Source   event.Source
}

// ArchiveGameResult holds the output of a successful game archive.
type ArchiveGameResult struct {
	Game   entity.ArchivedGame
	Events []event.Event
}

// ArchiveGameInteractor archives an IdleGame.
type ArchiveGameInteractor struct {
	repo GameRepository
	bus  event.EventBus
}

// NewArchiveGameInteractor constructs an ArchiveGameInteractor with its dependencies.
func NewArchiveGameInteractor(repo GameRepository, bus event.EventBus) *ArchiveGameInteractor {
	return &ArchiveGameInteractor{repo: repo, bus: bus}
}

// Execute loads the game, checks authorization, archives it, and dispatches events.
func (i *ArchiveGameInteractor) Execute(ctx context.Context, req ArchiveGameRequest) (ArchiveGameResult, error) {
	game, err := i.repo.Load(ctx, req.GameID)
	if err != nil {
		return ArchiveGameResult{}, fmt.Errorf("%w: %v", ErrRepositoryLoadFailed, err)
	}

	if req.CallerID != game.GMID() {
		return ArchiveGameResult{}, shared.ErrUnauthorized
	}

	idle, ok := game.(entity.IdleGame)
	if !ok {
		return ArchiveGameResult{}, fmt.Errorf("%w: game must be idle to archive", entity.ErrInvalidGameState)
	}

	archived, events, err := idle.Archive(req.Source)
	if err != nil {
		return ArchiveGameResult{}, err
	}

	if err := i.repo.Save(ctx, archived); err != nil {
		return ArchiveGameResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.GameID.GrimoireID),
			AggregateType: event.AggregateGame,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return ArchiveGameResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return ArchiveGameResult{
		Game:   archived,
		Events: events,
	}, nil
}