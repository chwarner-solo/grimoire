# ADR-017 Amendment 001: Faction Draft→Active Guards Removed — Sparse Not Errored

## Status
Accepted

## Date
2026-05-15

## Amends
ADR-017: Faction Aggregate Architecture

---

## Context

ADR-017 defined two guards on the `Draft → Active` transition:

```
Draft → Active:   GUARD: at least one member
                  GUARD: at least one standing level defined
```

These guards were written with the assumption that a Faction should be
"complete" before it operates in the world. During customer journey
analysis, this assumption was found to block play without protecting
any genuine domain invariant.

Two specific problems:

**The member guard blocks secret factions.** A faction operating in
the world before any of its members are known to the GM (or before
the GM has modelled them) is a legitimate narrative state. The
Inquisition exists and affects the world before the party — or the
GM's notes — name a single member. Blocking activation until a member
is recorded prevents the GM from creating the faction when they first
need it at the table.

**The standing level guard blocks mid-session creation.** A GM who
invents a faction mid-session needs to activate it immediately with
a name and nothing else. Standing levels are configuration — they can
be added later when the GM decides standing matters for this faction.
Blocking on their absence forces prep that may never be needed.

Neither guard protects an invariant. Both guard completeness.

The principle established by this amendment:

> **Sparse is not errored.** An aggregate with minimum data is valid.
> Incompleteness is a UI concern, not a domain concern. The domain
> guards identity and logical correctness — not authorial readiness.

---

## Decision

Remove both completeness guards from `Draft → Active`.

The corrected transition table for Faction is:

```
New    → Draft:    always allowed
Draft  → Active:   always allowed
                   (name and ID are the only requirements —
                    enforced at construction, not activation)
Active → Idle:     GM explicitly marks dormant
Idle   → Active:   GM reactivates OR player action triggers
Idle   → Archived: GM explicitly retires (terminal)
Active → Archived: GM explicitly retires (terminal)
Archived → *:      no exit — terminal
```

### What changes

```
roots/faction/entity/transitions.go
  draftFaction.Activate()
    Remove: member count check
    Remove: standing level count check
    Keep:   transition to activeFaction{} unchanged
```

### What does not change

- The ally/war contradiction invariant is untouched — this is logical
  correctness, not completeness. A faction cannot be both allied and
  at war with the same faction. That guard stays.
- `FactionID`, `GameID`, and `name` are still required at construction.
  These are identity guards. They are correct.
- Members and standing levels can still be added at any lifecycle state
  that permits it. The methods exist. Nothing prevents their use.
  They are simply no longer required before activation.
- The store pattern, event model, visibility model, and all other
  decisions in ADR-017 are unchanged.

---

## Consequences

- A Faction can be activated with only a name and ID
- A memberless Active Faction is sparse — not invalid
- A Faction with no standing levels is sparse — standing is simply
  not tracked until levels are defined
- GMs can create and activate factions at the table in seconds
- World building and session management use the same commands —
  the UI surfaces them differently, the domain does not distinguish

## Alternatives Considered

**Keep member guard, remove standing level guard** — rejected. Both
guards fail the same test: they protect completeness, not correctness.
Removing one and keeping the other is inconsistent.

**Move guards to interactor** — rejected. Guards that protect
completeness should not exist anywhere in the stack. The interactor
enforcing what the domain relaxed is the same problem in a different
layer. Sparse is a valid state end to end.