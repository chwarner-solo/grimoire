# ADR-019 Amendment 001: PlayerCharacter.ownerPlayerID Is Optional — GM Owns Everything

## Status
Accepted

## Date
2026-05-17

## Amends
ADR-019: Character Aggregate Architecture

---

## Context

ADR-019 defined `ownerPlayerID` as a required field on `PlayerCharacter`,
enforced at construction by `ErrOwnerPlayerIDRequired`.

This was written under the assumption that a PlayerCharacter always
belongs to a real player. During authorization design (ADR-020), this
assumption was found to be wrong.

**Grimoire is a GM tool.** The GM creates and owns all entities — including
PlayerCharacters. A PlayerCharacter is a narrative entity first. The
association to a real player (who will one day have a Player app) is a
future hook, not a current requirement.

Two problems with the required guard:

**It conflates narrative ownership with application authorization.**
The GM owns the PlayerCharacter in the domain sense regardless of whether
a real player has been associated. Requiring a player association before
creation blocks the GM from modelling a party before players have accounts.

**It creates a chicken-and-egg problem.**
A GM runs a session zero. They want to create PlayerCharacters before
any player has signed up. The required guard prevents this entirely.

The principle established by ADR-017-Amendment-001 applies here:

> **Sparse is not errored.** An aggregate with minimum data is valid.
> Incompleteness is a UI concern, not a domain concern.

---

## Decision

`ownerPlayerID` on `PlayerCharacter` is **optional**. Empty string is valid.

### What changes

```
roots/character/entity/player_character.go
  CreatePlayerCharacter()
    Remove: ownerPlayerID empty check
    Remove: ErrOwnerPlayerIDRequired guard

roots/character/entity/player_character_errors.go
    Remove: ErrOwnerPlayerIDRequired

roots/character/entity/player_character_snapshot.go
    No change — OwnerPlayerID string field remains
    Empty string is a valid snapshot value
```

### What does not change

- `ownerPlayerID` field remains on the struct and snapshot
- The field is the future association point for the Player app
- All other construction guards unchanged — ID, GameID, and name
  are still required at construction (identity guards, not completeness)
- All other PlayerCharacter methods unchanged

### Corrected construction signature

```go
// CreatePlayerCharacter constructs a new PlayerCharacter.
// ownerPlayerID is optional — empty string is valid.
// The GM creates PlayerCharacters. Player association is a future concern.
func CreatePlayerCharacter(
    id            identity.PlayerCharacterID,
    gameID        identity.GameID,
    name          string,
    ownerPlayerID string,   // optional — may be empty
    source        event.Source,
) (*PlayerCharacter, []event.Event, error)
```

Guards that remain:
```
id.IsZero()              → ErrPlayerCharacterIDRequired
gameID.IsZero()          → ErrPCGameIDRequired
strings.TrimSpace(name)  → ErrCharacterNameRequired
```

Guard removed:
```
strings.TrimSpace(ownerPlayerID)  → ErrOwnerPlayerIDRequired  ← REMOVED
```

---

## Consequences

- A GM can create a PlayerCharacter with no player association
- `ownerPlayerID` is populated later when a real player claims the character
- Player app authorization design is fully deferred — no current dependency
- All existing snapshots with empty `OwnerPlayerID` are valid — no migration
- `ErrOwnerPlayerIDRequired` is deleted — no callers outside tests

## Alternatives Considered

**Remove ownerPlayerID entirely** — rejected. The field is the correct
future hook for Player app association. Removing it now means adding it
back later with a migration. The field costs nothing as an optional string.

**Keep the guard, provide a GM sentinel value** — rejected. A sentinel
value (e.g. `"gm"`) is a workaround for a constraint that should not
exist. Remove the constraint.