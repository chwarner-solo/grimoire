# ADR-018 Amendment 001: Location Draft→Active Guard Removed — Sparse Not Errored

## Status
Accepted

## Date
2026-05-15

## Amends
ADR-018: Location Aggregate Architecture

---

## Context

ADR-018 defined one guard on the `Draft → Active` transition:

```
Draft → Active:   GUARD: at least one Scene defined
                  (aggregate checks sceneIDs not empty)
```

This guard was written with the assumption that a Location should have
at least some navigable content before the party can travel there.
During customer journey analysis, this assumption was found to block
play without protecting any genuine domain invariant.

The specific problem: a GM who invents a location mid-session — a
tavern, a crossroads, an unexpected cave entrance — needs to activate
it and move the party there immediately. Defining scenes takes time
the table does not have. The guard forces prep that does not yet exist
and may never be needed in that form.

A location with no scenes is not broken. It is a place the party can
stand. The GM narrates what is there. Scenes are authored later — or
never, if the location turns out to be a pass-through. The domain has
no stake in which outcome occurs.

The principle established by this amendment:

> **Sparse is not errored.** An aggregate with minimum data is valid.
> Incompleteness is a UI concern, not a domain concern. The domain
> guards identity and logical correctness — not authorial readiness.

This is the same principle established for Faction in
ADR-017-Amendment-001.

---

## Decision

Remove the scene completeness guard from `Draft → Active`.

The corrected transition table for Location is:

```
New      → Draft:    always allowed
Draft    → Active:   always allowed
                     (name and ID are the only requirements —
                      enforced at construction, not activation)
Active   → Idle:     always allowed (GM marks dormant)
Idle     → Active:   always allowed (GM reactivates)
Active   → Archived: GUARD: no child Location Active (interactor)
                     GUARD: party not present (interactor)
Idle     → Archived: GUARD: party not present (interactor)
Archived → *:        terminal
```

### What changes

```
roots/location/entity/transitions.go
  draftLocation.Activate()
    Remove: scene count check (sceneIDs not empty)
    Keep:   transition to activeLocation{} unchanged
```

### What does not change

- The archive guards are untouched. Archiving a location while the
  party is standing in it, or while child locations are still active,
  is genuinely unsafe. Those guards protect correctness, not
  completeness. They stay.
- `LocationID`, `GameID`, and `name` are still required at construction.
- Scenes can still be added to a Draft or Active location at any time.
  `DraftLocation.AddScene` and `ActiveLocation.AddScene` are unchanged.
  Scenes are simply no longer required before activation.
- The recursive hierarchy, DAG connection model, archive cascade,
  store pattern, and all other decisions in ADR-018 are unchanged.
- `recommendedBeatIDs` on Scene remains informational only — the
  TTRPG context relies on GM judgment for narrative gates. Unchanged.

---

## Consequences

- A Location can be activated with only a name and ID
- A sceneless Active Location is sparse — not invalid
- The party can travel to a location the GM just invented
- Scenes are authored when the GM is ready — before, during,
  or after the session in which the location is first visited
- World building and session management use the same commands —
  the UI surfaces them differently, the domain does not distinguish

## Alternatives Considered

**Keep guard, provide a fast-path scene creation** — rejected. The
guard itself is the problem. A fast path that creates a placeholder
scene to satisfy the guard is working around a constraint that
should not exist. Remove the constraint.

**Move guard to interactor** — rejected. Same reasoning as
ADR-017-Amendment-001. A completeness guard does not belong
anywhere in the stack.