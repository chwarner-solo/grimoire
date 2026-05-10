package interactor

import (
	"context"
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
)

// --- Test doubles ---

type stubGameRepository struct {
	savedGame entity.Game
	saveErr   error
}

func (r *stubGameRepository) Save(_ context.Context, game entity.Game) error {
	r.savedGame = game
	return r.saveErr
}

func (r *stubGameRepository) Load(_ context.Context, _ identity.GameID) (entity.Game, error) {
	return r.savedGame, nil
}

type spyEventBus struct {
	dispatched []event.EventEnvelope
	dispatchErr error
}

func (b *spyEventBus) Dispatch(_ context.Context, envelope event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.dispatched = append(b.dispatched, envelope)
	return nil
}

func (b *spyEventBus) Subscribe(_ event.EventType, _ event.EventHandler) error {
	return nil
}

// --- Tests ---

func TestCreateGame_SavesGameToRepository(t *testing.T) {
	repo := &stubGameRepository{}
	bus := &spyEventBus{}
	interactor := NewCreateGameInteractor(repo, bus)

	gameID := identity.NewGameID()
	req := CreateGameRequest{
		ID:     gameID,
		Name:   "Ashes & Chains",
		Source: event.SourceGrimoire,
	}

	_, err := interactor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.savedGame == nil {
		t.Fatal("expected game to be saved to repository")
	}
}

func TestCreateGame_DispatchesEntityCreatedEvent(t *testing.T) {
	repo := &stubGameRepository{}
	bus := &spyEventBus{}
	interactor := NewCreateGameInteractor(repo, bus)

	gameID := identity.NewGameID()
	req := CreateGameRequest{
		ID:     gameID,
		Name:   "Ashes & Chains",
		Source: event.SourceObsidian,
	}

	_, err := interactor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bus.dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(bus.dispatched))
	}

	envelope := bus.dispatched[0]
	if envelope.Type != event.TypeEntityCreated {
		t.Fatalf("expected event type %s, got %s", event.TypeEntityCreated, envelope.Type)
	}
	if envelope.AggregateID.String() != gameID.String() {
		t.Fatalf("expected aggregate ID %s, got %s", gameID.String(), envelope.AggregateID.String())
	}
	if envelope.AggregateType != event.AggregateGame {
		t.Fatalf("expected aggregate type %s, got %s", event.AggregateGame, envelope.AggregateType)
	}
	if envelope.Source != event.SourceObsidian {
		t.Fatalf("expected source %s, got %s", event.SourceObsidian, envelope.Source)
	}

	ec, ok := envelope.Payload.(event.EntityCreated)
	if !ok {
		t.Fatalf("expected EntityCreated payload, got %T", envelope.Payload)
	}
	if ec.EntityID != gameID.String() {
		t.Fatalf("expected entity ID %s, got %s", gameID.String(), ec.EntityID)
	}
	if ec.Name != "Ashes & Chains" {
		t.Fatalf("expected name 'Ashes & Chains', got %q", ec.Name)
	}
}

func TestCreateGame_ReturnsGameAndEvents(t *testing.T) {
	repo := &stubGameRepository{}
	bus := &spyEventBus{}
	interactor := NewCreateGameInteractor(repo, bus)

	gameID := identity.NewGameID()
	req := CreateGameRequest{
		ID:     gameID,
		Name:   "Ashes & Chains",
		Source: event.SourceGrimoire,
	}

	result, err := interactor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Game == nil {
		t.Fatal("expected non-nil Game in result")
	}
	if result.Game.GameID().String() != gameID.String() {
		t.Fatalf("expected game ID %s, got %s", gameID.String(), result.Game.GameID().String())
	}
	if result.Game.GameName() != "Ashes & Chains" {
		t.Fatalf("expected game name 'Ashes & Chains', got %q", result.Game.GameName())
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
}

func TestCreateGame_InvalidName_ReturnsError(t *testing.T) {
	repo := &stubGameRepository{}
	bus := &spyEventBus{}
	interactor := NewCreateGameInteractor(repo, bus)

	req := CreateGameRequest{
		ID:     identity.NewGameID(),
		Name:   "",
		Source: event.SourceGrimoire,
	}

	_, err := interactor.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	if repo.savedGame != nil {
		t.Fatal("expected no save when validation fails")
	}
	if len(bus.dispatched) != 0 {
		t.Fatal("expected no dispatch when validation fails")
	}
}

func TestCreateGame_RepositoryFailure_ReturnsError(t *testing.T) {
	saveErr := errors.New("connection refused")
	repo := &stubGameRepository{saveErr: saveErr}
	bus := &spyEventBus{}
	interactor := NewCreateGameInteractor(repo, bus)

	req := CreateGameRequest{
		ID:     identity.NewGameID(),
		Name:   "Ashes & Chains",
		Source: event.SourceGrimoire,
	}

	_, err := interactor.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on repository failure")
	}
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatal("expected no dispatch when save fails")
	}
}

func TestCreateGame_DispatchFailure_ReturnsError(t *testing.T) {
	repo := &stubGameRepository{}
	dispatchErr := errors.New("bus unavailable")
	bus := &spyEventBus{dispatchErr: dispatchErr}
	interactor := NewCreateGameInteractor(repo, bus)

	req := CreateGameRequest{
		ID:     identity.NewGameID(),
		Name:   "Ashes & Chains",
		Source: event.SourceGrimoire,
	}

	_, err := interactor.Execute(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on dispatch failure")
	}
	if !errors.Is(err, ErrEventDispatchFailed) {
		t.Fatalf("expected ErrEventDispatchFailed, got: %v", err)
	}
}
