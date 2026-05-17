package interactor

import (
	"context"
	"errors"
	"testing"

	"github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

func mustIdleGame(t *testing.T) (entity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	game, err := entity.ReconstituteGame(entity.GameSnapshot{
		ID:    id,
		GMID:  "gm-uid-test",
		Name:  "Test Game",
		State: "idle",
	})
	if err != nil {
		t.Fatalf("failed to build idle game: %v", err)
	}
	return game, id
}

func mustActiveGame(t *testing.T) (entity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	game, err := entity.ReconstituteGame(entity.GameSnapshot{
		ID:                  id,
		GMID:                "gm-uid-test",
		Name:                "Test Game",
		State:               "active",
		ActiveCampaignCount: 1,
	})
	if err != nil {
		t.Fatalf("failed to build active game: %v", err)
	}
	return game, id
}

func mustNewGame(t *testing.T) (entity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	game, err := entity.ReconstituteGame(entity.GameSnapshot{
		ID:    id,
		GMID:  "gm-uid-test",
		Name:  "Test Game",
		State: "new",
	})
	if err != nil {
		t.Fatalf("failed to build new game: %v", err)
	}
	return game, id
}

func TestArchiveGame_ArchivesIdleGame_Succeeds(t *testing.T) {
	idle, id := mustIdleGame(t)
	repo := &stubGameRepository{savedGame: idle}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	result, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Game == nil {
		t.Fatal("expected non-nil ArchivedGame in result")
	}
	if _, ok := result.Game.(entity.ArchivedGame); !ok {
		t.Fatalf("expected ArchivedGame, got %T", result.Game)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
	if len(bus.dispatched) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(bus.dispatched))
	}
}

func TestArchiveGame_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	idle, id := mustIdleGame(t)
	repo := &stubGameRepository{savedGame: idle}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "wrong-caller",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected ErrUnauthorized")
	}
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
	if repo.savedGame != idle {
		t.Fatal("expected no save on unauthorized request")
	}
	if len(bus.dispatched) != 0 {
		t.Fatal("expected no dispatch on unauthorized request")
	}
}

func TestArchiveGame_ActiveGame_ReturnsInvalidState(t *testing.T) {
	active, id := mustActiveGame(t)
	repo := &stubGameRepository{savedGame: active}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for ActiveGame")
	}
	if !errors.Is(err, entity.ErrInvalidGameState) {
		t.Fatalf("expected ErrInvalidGameState, got: %v", err)
	}
}

func TestArchiveGame_NewGame_ReturnsInvalidState(t *testing.T) {
	ng, id := mustNewGame(t)
	repo := &stubGameRepository{savedGame: ng}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for NewGame")
	}
	if !errors.Is(err, entity.ErrInvalidGameState) {
		t.Fatalf("expected ErrInvalidGameState, got: %v", err)
	}
}

func TestArchiveGame_RepositoryLoadFailure(t *testing.T) {
	loadErr := errors.New("db unavailable")
	repo := &stubGameRepository{loadErr: loadErr}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   identity.NewGameID(),
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error on load failure")
	}
	if !errors.Is(err, ErrRepositoryLoadFailed) {
		t.Fatalf("expected ErrRepositoryLoadFailed, got: %v", err)
	}
}

func TestArchiveGame_RepositoryLoadFailure_NoDispatch(t *testing.T) {
	loadErr := errors.New("db unavailable")
	repo := &stubGameRepository{loadErr: loadErr}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	i.Execute(context.Background(), ArchiveGameRequest{ //nolint:errcheck
		CallerID: "gm-uid-test",
		GameID:   identity.NewGameID(),
		Source:   event.SourceGrimoire,
	})

	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch on load failure, got %d", len(bus.dispatched))
	}
}

func TestArchiveGame_SaveFailure_NoDispatch(t *testing.T) {
	idle, id := mustIdleGame(t)
	saveErr := errors.New("write failed")
	repo := &stubGameRepository{savedGame: idle, saveErr: saveErr}
	bus := &spyEventBus{}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error on save failure")
	}
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch when save fails, got %d", len(bus.dispatched))
	}
}

func TestArchiveGame_DispatchFailure_ReturnsError(t *testing.T) {
	idle, id := mustIdleGame(t)
	repo := &stubGameRepository{savedGame: idle}
	bus := &spyEventBus{dispatchErr: errors.New("bus down")}
	i := NewArchiveGameInteractor(repo, bus)

	_, err := i.Execute(context.Background(), ArchiveGameRequest{
		CallerID: "gm-uid-test",
		GameID:   id,
		Source:   event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error on dispatch failure")
	}
	if !errors.Is(err, ErrEventDispatchFailed) {
		t.Fatalf("expected ErrEventDispatchFailed, got: %v", err)
	}
}