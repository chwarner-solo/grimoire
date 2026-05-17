package interactor

import (
	"context"
	"fmt"
	"time"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// GameRepository is the persistence port for the Game aggregate.
// Defined locally to avoid importing the port package.
type GameRepository interface {
	Save(ctx context.Context, game entity.Game) error
	Load(ctx context.Context, id identity.GameID) (entity.Game, error)
}

// CreateGameRequest holds the input for creating a new game.
type CreateGameRequest struct {
	CallerID string          // verified Firebase UID — set by API layer
	ID       identity.GameID
	Name     string
	Source   event.Source
}

// CreateGameResult holds the output of a successful game creation.
type CreateGameResult struct {
	Game   entity.NewGame
	Events []event.Event
}

// CreateGameInteractor orchestrates the creation of a new Game aggregate.
type CreateGameInteractor struct {
	repo GameRepository
	bus  event.EventBus
}

// NewCreateGameInteractor constructs a CreateGameInteractor with its dependencies.
func NewCreateGameInteractor(repo GameRepository, bus event.EventBus) *CreateGameInteractor {
	return &CreateGameInteractor{
		repo: repo,
		bus:  bus,
	}
}

// Execute validates input, creates the Game aggregate, persists it, and dispatches events.
func (i *CreateGameInteractor) Execute(ctx context.Context, req CreateGameRequest) (CreateGameResult, error) {
	game, events, err := entity.CreateGame(req.ID, req.CallerID, req.Name, req.Source)
	if err != nil {
		return CreateGameResult{}, err
	}

	if err := i.repo.Save(ctx, game); err != nil {
		return CreateGameResult{}, fmt.Errorf("%w: %v", ErrRepositorySaveFailed, err)
	}

	for _, evt := range events {
		envelope := event.EventEnvelope{
			Type:          evt.EventType(),
			AggregateID:   identity.GrimoireID(req.ID.GrimoireID),
			AggregateType: event.AggregateGame,
			Source:        req.Source,
			OccurredAt:    time.Now(),
			Payload:       evt,
		}
		if err := i.bus.Dispatch(ctx, envelope); err != nil {
			return CreateGameResult{}, fmt.Errorf("%w: %v", ErrEventDispatchFailed, err)
		}
	}

	return CreateGameResult{
		Game:   game,
		Events: events,
	}, nil
}
