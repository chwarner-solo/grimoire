# ADR-006: Typed UUIDs for All Identity

## Status
Accepted

## Date
2026-05-09

## Context
Aggregate roots and entities require stable, globally unique identifiers.
The choices were: raw strings, raw uuid.UUID, integer sequences, or typed
wrappers over UUID.

Sequential integer IDs were eliminated immediately — they require a central
authority, do not work with local-first event generation, and leak record
counts.

Raw strings and raw uuid.UUID both allow passing any ID where any other
ID is expected. A function accepting a GameID can accidentally receive a
SessionID with no compile-time protection.

## Decision
All identity types are distinct Go structs wrapping GrimoireID (see ADR-007).

```go
type GameID     struct{ GrimoireID }
type CampaignID struct{ GrimoireID }
type SessionID  struct{ GrimoireID }
type CharacterID struct{ GrimoireID }
type LocationID struct{ GrimoireID }
type NarrativeID struct{ GrimoireID }
type FactionID  struct{ GrimoireID }
```

All ID types live in `grimoire-domain/shared/identity/`.

## Reasoning
Distinct types make ID mismatches a compile error, not a runtime error.
Passing a SessionID where a GameID is expected fails at compile time.
This is especially valuable in a domain with many aggregate roots each
with their own ID type.

UUIDs are generated client-side — no central authority required. This
supports the local-first, offline-capable architecture where events are
generated on device before sync.

## Consequences
- ID mismatches caught at compile time
- No central ID generation authority needed
- Consistent JSON serialization via GrimoireID base (see ADR-007)
- All IDs are comparable and usable as map keys
- IsZero() available on all ID types for construction validation

## Alternatives Considered
**Raw uuid.UUID** — rejected. No type safety. Any UUID accepted anywhere.

**Raw string** — rejected. No type safety and no format validation.

**Integer sequences** — rejected. Requires central authority. Incompatible
with local-first event generation. Leaks record counts.