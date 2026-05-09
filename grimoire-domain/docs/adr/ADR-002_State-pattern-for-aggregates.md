# ADR-002: State Pattern for Aggregate Lifecycles

## Status
Accepted

## Date
2026-05-09

## Context
Aggregates (Game, Campaign, Session) have well-defined lifecycles with
distinct states. Each state has a specific set of valid operations. The
naive implementation uses a status field and guard clauses:

```go
func (g *Game) StartCampaign() error {
    if g.status != GameStatusDraft {
        return fmt.Errorf("game must be draft to start campaign")
    }
    // ...
}
```

This approach has problems:
- Invalid operations are caught at runtime not compile time
- Every method must check status defensively
- Nothing prevents calling any method in any state
- Adding a new state requires auditing every method

## Decision
Each aggregate lifecycle state is a Go interface. Each interface exposes
only the methods that are valid for that state. State transitions return
the next state interface.

```go
type NewGame interface {
    AddNarrativeElement(name string) (DraftGame, error)
}

type DraftGame interface {
    AddNarrativeElement(name string) (DraftGame, error)
    CreateCampaign(id CampaignID, name string) (DraftGame, error)
    Activate() (ActiveGame, error)
}
```

## Reasoning
Invalid state transitions become impossible at compile time. You cannot
call `Activate()` on a `NewGame` because the interface does not expose it.
The compiler enforces the state machine — not runtime guards.

The type IS the status. No status field is needed. Switch statements on
the interface type give you the current state without checking a field.

## Consequences
- State transitions are structurally enforced by the compiler
- Each state struct holds only the data relevant to that state
- Adding a new state requires defining a new interface and implementing it
- Callers must handle the returned interface type — they always know what state they have
- Persistence requires mapping state interfaces back to a serializable form

## Alternatives Considered
**Status field with guard clauses** — rejected. Runtime errors instead of
compile time errors. Defensive code in every method. Easy to miss a guard
when adding new operations.

**iota enum with exhaustive switch** — rejected. Go does not enforce
exhaustive switches. Same runtime failure risk as guard clauses without
the structural clarity of the State pattern.