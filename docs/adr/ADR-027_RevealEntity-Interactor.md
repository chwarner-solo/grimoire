# ADR-027: RevealEntity Interactor

## Status
Accepted

## Date
2026-05-17

---

## Context

The GM reveals entities to players mid-session. From the GM's perspective,
this is a single action regardless of entity type: "make this visible."

From the domain's perspective, each entity type has its own aggregate and
its own `.Reveal()` method — but they all emit the same `EntityRevealed`
canonical event (ADR-004).

This is a cross-aggregate interactor. One interactor, one request shape,
dispatches to the correct aggregate by entity type. The EventBus then
routes the `EntityRevealed` event to:
- `PlayerPushHandler` — pushes to the Player app (ADR-011)
- `Neo4jHandler` — updates playerVisible flag on the node
- `SyncBroker` — pushes to Foundry/Obsidian if applicable (ADR-009)

Authorization: load Game by GameID, check CallerID == game.GMID().

---

## Supported Entity Types

```
npc               →  NarrativeCharacter aggregate
player_character  →  PlayerCharacter aggregate
macguffin         →  MacGuffin aggregate
faction           →  Faction aggregate
```

Future entity types (Secret, Lore) follow the same pattern — add a
case to the switch, add the port method. No structural change.

---

## Ports

```go
// grimoire-domain/roots/shared/interactor/reveal_entity_ports.go
// Lives in shared because it crosses aggregate boundaries.

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type NPCRepository interface {
    Save(ctx context.Context, snap characterentity.NPCSnapshot) error
    Load(ctx context.Context, id identity.NarrativeCharacterID) (characterentity.NPCSnapshot, error)
}

type PlayerCharacterRepository interface {
    SaveCore(ctx context.Context, snap characterentity.PlayerCharacterSnapshot) error
    LoadCore(ctx context.Context, id identity.PlayerCharacterID) (characterentity.PlayerCharacterSnapshot, error)
}

type MacGuffinRepository interface {
    Save(ctx context.Context, snap characterentity.MacGuffinSnapshot) error
    Load(ctx context.Context, id identity.MacGuffinID) (characterentity.MacGuffinSnapshot, error)
}

type FactionRepository interface {
    Save(ctx context.Context, faction factionentity.Faction) error
    Load(ctx context.Context, id identity.FactionID) (factionentity.Faction, error)
}
```

---

## Interactor: RevealEntity

**File:** `grimoire-domain/roots/shared/interactor/reveal_entity.go`

```
Request:
    CallerID    string
    EntityID    string              // string — entity type determines parse
    EntityType  value.EntityType   // "npc" | "player_character" |
                                   // "macguffin" | "faction"
    GameID      identity.GameID
    RevealedTo  []string           // player IDs receiving the reveal
    SessionID   identity.SessionID
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth: req.CallerID != game.GMID() → ErrUnauthorized

    switch req.EntityType:

    case "npc":
        id, _ := identity.ParseNarrativeCharacterID(req.EntityID)
        snap := npcRepo.Load(id) → ErrEntityNotFound
        npc := characterentity.ReconstituteNPC(snap)
        an, ok := npc.(characterentity.ActiveNPC) → ErrInvalidEntityState
        updated, events, err := an.Reveal(req.RevealedTo,
                                          req.SessionID, req.Source)
        npcRepo.Save(updated.Snapshot())

    case "player_character":
        id, _ := identity.ParsePlayerCharacterID(req.EntityID)
        snap := pcRepo.LoadCore(id) → ErrEntityNotFound
        pc := characterentity.ReconstitutePlayerCharacter(snap)
        updated, events, err := pc.Reveal(req.RevealedTo,
                                          req.SessionID, req.Source)
        pcRepo.SaveCore(updated.Snapshot())

    case "macguffin":
        id, _ := identity.ParseMacGuffinID(req.EntityID)
        snap := macguffinRepo.Load(id) → ErrEntityNotFound
        mg := characterentity.ReconstituteMacGuffin(snap)
        updated, events, err := mg.Reveal(req.RevealedTo,
                                          req.SessionID, req.Source)
        macguffinRepo.Save(updated.Snapshot())

    case "faction":
        id, _ := identity.ParseFactionID(req.EntityID)
        faction := factionRepo.Load(id) → ErrEntityNotFound
        af, ok := faction.(factionentity.ActiveFaction) → ErrInvalidEntityState
        updated, events, err := af.Reveal(req.RevealedTo,
                                          req.SessionID, req.Source)
        factionRepo.Save(updated)

    default:
        → ErrUnsupportedEntityType

    bus.Dispatch(EntityRevealed{
        entity_id:   req.EntityID,
        entity_type: req.EntityType,
        revealed_to: req.RevealedTo,
        session_id:  req.SessionID,
        source:      req.Source,
    })

Result:
    Events  []event.Event

Event emitted: EntityRevealed (canonical — ADR-004)
```

---

## Test Cases

| Test | Setup | Expected |
|------|-------|----------|
| RevealNPC_Succeeds | active NPC | EntityRevealed dispatched |
| RevealPC_Succeeds | active PC | EntityRevealed dispatched |
| RevealMacGuffin_Succeeds | live MacGuffin | EntityRevealed dispatched |
| RevealFaction_Succeeds | active faction | EntityRevealed dispatched |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| UnsupportedEntityType_ReturnsError | EntityType="beat" | ErrUnsupportedEntityType |
| NPCNotActive_ReturnsError | draft NPC | ErrInvalidEntityState |
| FactionNotActive_ReturnsError | draft faction | ErrInvalidEntityState |
| EntityNotFound_ReturnsError | missing entity | ErrEntityNotFound |
| EmptyRevealedTo_ReturnsError | RevealedTo=[] | domain error |
| ZeroSessionID_ReturnsError | SessionID zero | domain error |
| SaveFailure_NoDispatch | saveErr set | ErrRepositorySaveFailed, no dispatch |

---

## EntityType Value Object

```go
// grimoire-domain/shared/value/entity_type.go

type EntityType string

const (
    EntityTypeNPC             EntityType = "npc"
    EntityTypePlayerCharacter EntityType = "player_character"
    EntityTypeMacGuffin       EntityType = "macguffin"
    EntityTypeFaction         EntityType = "faction"
    // Future: EntityTypeSecret, EntityTypeLore
)
```

---

## Errors

```go
// grimoire-domain/shared/interactor/reveal_entity_errors.go

var (
    ErrRepositorySaveFailed  = errors.New("reveal: failed to save entity")
    ErrRepositoryLoadFailed  = errors.New("reveal: failed to load entity")
    ErrEventDispatchFailed   = errors.New("reveal: failed to dispatch event")
    ErrGameNotFound          = errors.New("reveal: game not found")
    ErrEntityNotFound        = errors.New("reveal: entity not found")
    ErrInvalidEntityState    = errors.New("reveal: entity is not in a revealable state")
    ErrUnsupportedEntityType = errors.New("reveal: unsupported entity type")
)

// ErrUnauthorized from grimoire-domain/shared/interactor/errors.go
```

---

## File Locations

```
grimoire-domain/shared/
    value/
        entity_type.go              EntityType constants

    interactor/
        errors.go                   ErrUnauthorized (already exists)
        reveal_entity_ports.go      all five repository interfaces
        reveal_entity_errors.go     reveal-specific errors
        reveal_entity.go            RevealEntityInteractor
        reveal_entity_test.go       table-driven across all entity types
```

---

## Consequences

- One interactor, one GraphQL mutation (`revealEntity`), one event type
  (`EntityRevealed`) — from API to domain to event bus, the shape is
  consistent regardless of what is being revealed
- Adding a new revealable entity type (Secret, Lore) requires:
    1. A new port method on the relevant repository interface
    2. A new `case` in the switch
    3. New test rows in the table
       No structural change to the interactor
- RevealFaction and RevealNPC require Active state — a Draft entity
  cannot be revealed because it has no player-facing content yet
- RevealPlayerCharacter and RevealMacGuffin have no state restriction —
  consistent with their Snapshot/Reconstitute pattern (no state machine)
- The `RevealedTo` slice maps to player IDs — the Player app uses this
  to filter which entities are visible. The interactor does not validate
  that the player IDs exist — that is a future concern for the Player app ADR
