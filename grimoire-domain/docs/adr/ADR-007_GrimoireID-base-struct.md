# ADR-007: GrimoireID Base Struct via Embedding

## Status
Accepted

## Date
2026-05-09

## Context
With multiple typed ID types (see ADR-006), each requires consistent
behaviour: String(), IsZero(), MarshalJSON(), UnmarshalJSON(). Without
a shared base, this behaviour must be implemented on every ID type —
introducing repetition and risk of inconsistency.

Go does not have abstract base classes or inheritance. The options were:
an interface all IDs implement, a code generation approach, or embedding
a shared base struct.

## Decision
A single `GrimoireID` struct encapsulates the uuid.UUID value and
implements all shared behaviour. All typed ID types embed GrimoireID
and inherit its behaviour via Go embedding.

```go
type GrimoireID struct {
    value uuid.UUID
}

func (id GrimoireID) String() string         { ... }
func (id GrimoireID) IsZero() bool           { ... }
func (id GrimoireID) MarshalJSON() []byte    { ... }
func (id *GrimoireID) UnmarshalJSON() error  { ... }

// Each typed ID embeds — zero boilerplate
type GameID struct{ GrimoireID }
type SessionID struct{ GrimoireID }
```

Construction per type:

```go
func NewGameID() GameID {
    return GameID{NewGrimoireID()}
}

func ParseGameID(s string) (GameID, error) {
    base, err := ParseGrimoireID(s)
    if err != nil {
        return GameID{}, err
    }
    return GameID{base}, nil
}
```

## Reasoning
Embedding gives each typed ID all shared behaviour with zero repetition.
String(), IsZero(), MarshalJSON(), and UnmarshalJSON() are defined once
on GrimoireID and promoted to every embedding type.

Adding a new ID type requires only the struct definition and two
constructor functions. No behaviour code is repeated.

JSON serialization is consistent across all ID types — they all marshal
as UUID strings.

## Consequences
- All ID behaviour defined once on GrimoireID
- New ID types require minimal code per type
- Consistent JSON format across all IDs
- GrimoireID is an implementation detail — unexported value field
- Typed IDs remain distinct types despite shared base (compiler still
  catches mismatches)

## Alternatives Considered
**Interface with per-type implementation** — rejected. Every ID type
implements String(), IsZero(), MarshalJSON(), UnmarshalJSON() independently.
Repetitive and risks inconsistency.

**go generate / code generation** — rejected. Adds tooling complexity
for a problem embedding solves cleanly.

**Raw uuid.UUID with type alias** — rejected. Type aliases in Go are
transparent to the compiler. `type GameID = uuid.UUID` provides no
type safety — GameID and SessionID are the same type.