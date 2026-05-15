# ADR-019 Amendment 001: NPC Draft→Active Description Guard Removed — Sparse Not Errored

## Status
Accepted

## Date
2026-05-15

## Amends
ADR-019: Character Aggregate Architecture

---

## Context

ADR-019 defined two guards on the NPC `Draft → Active` transition:

```
Draft → Active:   GUARD: name not empty
                  GUARD: description not empty
```

The name guard is an identity guard — it is correct and stays.
You cannot meaningfully refer to an NPC at the table without a name.

The description guard is a completeness guard. During customer journey
analysis, it was found to block a common mid-session pattern: a GM
invents an NPC on the spot — gives them a name, a voice, a purpose —
and needs to activate them immediately. Writing a description takes
time the table does not have. The description will be written later,
if at all. Many NPCs who appear for one scene never need a formal
description in the system.

The description is content. Per ADR-015 and ADR-016, content lives
in Neo4j. The aggregate holds only the minimum structural data needed
to enforce invariants. A description that does not yet exist is not
a violation of any invariant — it is an empty content field. The NPC
is still identifiable, locatable, assignable, and revealable.

The principle established by this amendment:

> **Sparse is not errored.** An aggregate with minimum data is valid.
> Incompleteness is a UI concern, not a domain concern. The domain
> guards identity and logical correctness — not authorial readiness.

This is the same principle established for Faction in
ADR-017-Amendment-001 and for Location in ADR-018-Amendment-001.

---

## Decision

Remove the description completeness guard from `Draft → Active`.
Retain the name identity guard — name is required for activation.

The corrected transition table for NPC is:

```
New    → Draft:    always allowed
Draft  → Active:   GUARD: name not empty
                   (description no longer required)
Active → Idle:     always allowed (GM decision)
Idle   → Active:   always allowed (GM reactivates)
Idle   → Archived: always allowed (terminal)
Active → Archived: always allowed (terminal)
Archived → *:      no exit — terminal
```

### What changes

```
roots/character/entity/narrative_character_errors.go
  ErrNPCDescriptionRequired
    Remove this sentinel error — it is no longer used

roots/character/entity/transitions.go  (or equivalent)
  draftNPC.Activate()
    Remove: description empty check
    Keep:   name empty check
    Keep:   transition to activeNPC{} unchanged
```

### What does not change

- `ErrNPCNameRequired` stays. Name is identity.
- `UpdateContent` is available on both `DraftNPC` and `ActiveNPC`.
  The GM can add or update description at any time. Nothing prevents
  this — it is simply no longer required before activation.
- The MacGuffin drop on archive, the location assignment model, the
  Reveal method, the Handle/Replay pattern — all unchanged.
- `playerDescription` and `description` remain separate fields.
  The content model is unchanged. Only the activation guard is removed.

### Note on content fields at activation

An NPC activated with no description has:
```
name:              "Mira"    ← required, set
description:       ""        ← empty, valid
playerDescription: ""        ← empty, valid
playerVisible:     false     ← default, correct
locationID:        zero      ← unassigned, valid
```

This is a sparse but internally consistent NPC. The GM can fill in
description and playerDescription via `UpdateContent` before or after
the session. The player-facing description only matters when the NPC
is revealed — which requires `playerDescription` to have content for
a good player experience, but the domain does not enforce this.

---

## Consequences

- An NPC can be activated with only a name
- An NPC with no description is sparse — not invalid
- GMs can create and activate NPCs at the table in seconds
- The description warning ("this NPC has no description") is a UI
  concern — surfaces as a nudge in the session management screen,
  not as a domain error
- World building and session management use the same commands —
  the UI surfaces them differently, the domain does not distinguish

## Alternatives Considered

**Remove name guard as well** — rejected. Name is identity, not
content. An unnamed NPC cannot be referenced at the table or in
any event payload meaningfully. The name guard is correct.

**Move description guard to interactor** — rejected. A completeness
guard does not belong anywhere in the stack. The UI layer surfaces
the nudge; the domain does not enforce the constraint.

**Add a separate ActivateSparse() method** — rejected. Two activation
methods for the same transition with different guards fragments the
interface and implies that sparse activation is somehow less valid
than full activation. It is not. There is one Activate().
