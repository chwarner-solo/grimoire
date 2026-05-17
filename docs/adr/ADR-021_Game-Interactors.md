# ADR-021: Game Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

The Game aggregate is the root of everything. Every other aggregate is
owned by a Game. Two GM-triggered operations exist at this level:
creating a Game, and archiving it when the campaign is over.

`CreateGame` was implemented before this ADR existed. This ADR documents
it canonically and adds `ArchiveGame` alongside it.

All other Game state transitions (New→Draft, Draft→Active, Active→Idle)
are driven by EventBus handlers reacting to downstream events (ADR-011).
They are not interactors.

---

## Standard Interactor Pattern

All interactors in this project follow this invariant:

```
1. Call domain constructor / load aggregate
2. Validate — domain errors surface here
3. Authorization check — CallerID must match game.GMID()
4. Save aggregate — if save fails, stop. No events dispatched.
5. Dispatch events — if dispatch fails, return error.
   Caller decides retry strategy.
```

This pattern is established by `CreateGame` and repeated without
variation across all 27 interactors in ADR-021 through ADR-027.

---

## Ports (local interface aliases)

```go
// grimoire-domain/roots/game/interactor/ports.go

// GameRepository is the persistence port for the Game aggregate.
// Defined locally to avoid circular imports with the port package.
type GameRepository interface {
    Save(ctx context.Context, game entity.Game) error
    Load(ctx context.Context, id identity.GameID) (entity.Game, error)
}
```

`event.EventBus` is imported directly from `grimoire-domain/shared/event`.

---

## Interactor: CreateGame

**File:** `grimoire-domain/roots/game/interactor/create_game.go`

Already implemented. Documented here for completeness.

```
Request:
    CallerID  string           // verified Firebase UID — set by API layer
    ID        identity.GameID
    Name      string
    Source    event.Source

Flow:
    entity.CreateGame(req.ID, req.CallerID, req.Name, req.Source)
      → ErrGMIDRequired if CallerID empty
      → ErrGameIDRequired if ID zero
      → ErrGameNameRequired if Name empty/whitespace
    repo.Save(game)
      → ErrRepositorySaveFailed on failure — no events dispatched
    bus.Dispatch(EntityCreated envelope)
      → ErrEventDispatchFailed on failure

Result:
    Game    entity.NewGame
    Events  []event.Event

AggregateType: event.AggregateGame
```

**Test cases** (`create_game_test.go` — already exists):

| Test | Setup | Expected |
|------|-------|----------|
| SavesGameToRepository | valid request | repo.savedGame non-nil |
| DispatchesEntityCreatedEvent | valid request | 1 envelope, TypeEntityCreated |
| ReturnsGameAndEvents | valid request | result.Game non-nil, 1 event |
| InvalidName_ReturnsError | Name="" | error, no save, no dispatch |
| EmptyCallerID_ReturnsError | CallerID="" | ErrGMIDRequired, no save |
| SetsGMIDFromCallerID | CallerID="uid-001" | result.Game.GMID()=="uid-001" |
| RepositoryFailure_ReturnsError | repo.saveErr set | ErrRepositorySaveFailed, no dispatch |
| DispatchFailure_ReturnsError | bus.dispatchErr set | ErrEventDispatchFailed |

---

## Interactor: ArchiveGame

**File:** `grimoire-domain/roots/game/interactor/archive_game.go`

The GM archives a completed Game. Only an IdleGame can be archived.
An ActiveGame cannot — at least one campaign is still running.

```
Request:
    CallerID  string
    GameID    identity.GameID
    Source    event.Source

Flow:
    repo.Load(req.GameID)
      → ErrRepositoryLoadFailed if not found
    auth check: req.CallerID != game.GMID() → ErrUnauthorized
    idle, ok := game.(entity.IdleGame)
      → ErrInvalidGameState if not IdleGame
    archived, events, err := idle.Archive(req.Source)
      → domain error if transition fails
    repo.Save(archived)
      → ErrRepositorySaveFailed on failure
    bus.Dispatch(EntityUpdated envelope) for each event
      → ErrEventDispatchFailed on failure

Result:
    Game    entity.ArchivedGame
    Events  []event.Event

AggregateType: event.AggregateGame
Event emitted: EntityUpdated { field: "status", new_value: "archived" }
```

**Test cases** (`archive_game_test.go`):

| Test | Setup | Expected |
|------|-------|----------|
| ArchivesIdleGame_Succeeds | idle game in repo | ArchivedGame returned, 1 event |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized, no save |
| ActiveGame_ReturnsInvalidState | active game in repo | ErrInvalidGameState |
| NewGame_ReturnsInvalidState | new game in repo | ErrInvalidGameState |
| RepositoryLoadFailure | repo.loadErr set | ErrRepositoryLoadFailed |
| RepositoryLoadFailure_NoDispatch | repo.loadErr set | no events dispatched |
| SaveFailure_NoDispatch | repo.saveErr set | ErrRepositorySaveFailed, 0 dispatched |
| DispatchFailure_ReturnsError | bus.dispatchErr set | ErrEventDispatchFailed |

---

## Errors

```go
// grimoire-domain/roots/game/interactor/errors.go

var (
    ErrRepositorySaveFailed  = errors.New("interactor: failed to save game")
    ErrRepositoryLoadFailed  = errors.New("interactor: failed to load game")
    ErrEventDispatchFailed   = errors.New("interactor: failed to dispatch event")
)

// ErrUnauthorized is imported from grimoire-domain/shared/interactor/errors.go
// ErrInvalidGameState is imported from grimoire-domain/roots/game/entity
```

---

## File Locations

```
grimoire-domain/roots/game/interactor/
    ports.go              GameRepository interface
    errors.go             ErrRepositorySaveFailed, ErrRepositoryLoadFailed,
                          ErrEventDispatchFailed
    create_game.go        CreateGameInteractor (exists)
    create_game_test.go   (exists — add missing cases above)
    archive_game.go       ArchiveGameInteractor
    archive_game_test.go  table-driven tests
```

---

## Consequences

- CreateGame is the only interactor that does not perform an auth load —
  it establishes ownership rather than verifying it
- ArchiveGame is the only terminal Game operation — no recovery after archive
- All other Game transitions are event-driven, not interactor-driven