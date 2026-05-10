# Grimoire — Claude Code Context

## Project Overview

Grimoire is an event-driven narrative state machine for tabletop RPGs.
It provides pluggable capture surfaces (Obsidian plugin, Foundry VTT module,
direct web UI) backed by a canonical event log, a Neo4j graph as the world
model, and a player-facing app that shows only what the GM has chosen to reveal.

The GM controls the information boundary. Players see their slice of the world.
Everything is driven by six canonical events.

---

## Workspace Structure

Four Go modules in a workspace. Strict inward dependency rule (ADR-013).

```
grimoire/main/
  go.work                          ← Go workspace root
  CLAUDE.md                        ← you are here
  grimoire-domain/                 ← pure domain, zero infrastructure
    go.mod
    docs/
      adr/                         ← Architecture Decision Records
    roots/
      game/
        entity/
        value/
        port/
      campaign/
        entity/
        value/
        port/
      character/
        entity/
        value/
        port/
      location/
        entity/
        value/
        port/
      narrative/
        entity/
        value/
        port/
      faction/
        entity/
        value/
        port/
    shared/
      identity/                    ← GrimoireID, EventID, all typed IDs
      event/                       ← six canonical event types, envelope
  grimoire-infrastructure/         ← adapters for domain ports
    go.mod                            Firestore, Neo4j, GCS, Pub/Sub
  grimoire-testing/                ← test doubles, builders, fixtures
    go.mod                            never compiled into production
  grimoire-api/                    ← composition root, GraphQL, HTTP
    go.mod
    cmd/
    internal/
```

**Dependency rule — arrows point inward only:**
```
grimoire-domain         →  no internal imports (external: uuid only)
grimoire-infrastructure →  grimoire-domain only
grimoire-testing        →  grimoire-domain only
grimoire-api            →  grimoire-domain + grimoire-infrastructure
                           test files only: grimoire-testing
```

---

## Architecture Principles

### Ports & Adapters (Hexagonal Architecture)
- Domain has zero knowledge of infrastructure
- Ports are interfaces defined in the domain
- Adapters live in grimoire-infrastructure, never grimoire-domain
- grimoire-api is the composition root — wires ports to adapters
- The domain is the boss. Infrastructure serves it.

### Test Driven Development
- **A failing test must exist before any implementation is written. No exceptions.**
- Tests live alongside the code they test
- Test names are sentences describing behaviour
- `TestGame_CannotStartSession_WhenStatusIsNew` not `TestStartSession`

### State Pattern for Aggregates
- Each aggregate state is a Go interface
- Each interface exposes ONLY the methods valid for that state
- Invalid transitions are impossible at compile time — not runtime guards
- The type IS the status. No status field needed.
- Transitions return the next state interface

### CQRS — Strict Split (ADR-003, ADR-010, ADR-014)
- Aggregate roots handle commands (write side) only
- Aggregate roots enforce invariants only
- Aggregate roots hold the minimum data needed to enforce invariants
- **Aggregate roots are never used to answer queries**
- Three stores, three jobs (ADR-010):
  - Firestore: aggregate state (write side, one document per root)
  - Neo4j: read model (graph queries, built from events)
  - GCS: event log (immutable ndjson, source of truth)
- Scaling path (ADR-014): Firestore → Bigtable event sourcing + snapshots
  - Domain and command handlers never change — port abstraction
  - Migration is an adapter swap in grimoire-infrastructure
- Do not add fields to aggregates to satisfy query needs

### Event Driven (ADR-004, ADR-011, ADR-012)
- All state changes produce events
- Events are the source of truth
- Six canonical event types cover everything (see ADR-004)
- Events are immutable once written
- Cross-aggregate reactions flow through EventBus (ADR-011)
- Events carry ULID (global order) + SequenceNumber (per-aggregate order) (ADR-012)

---

## Aggregate Roots

```
Game        →  the world, canonical GM content
Campaign    →  a table running a Game
Character   →  PC or NPC
Location    →  places in the world
Narrative   →  lore, acts, revelations, secrets
Faction     →  groups, allegiances, goals
```

Each root owns its lifecycle. No root reaches into another root.
Cross-root references use typed IDs only — never navigation.

---

## Lifecycles (State Machines)

### Game
```
New → Draft → Active → Idle → Archived
```
- New → Draft: first narrative element added
- Draft → Active: first Campaign becomes Active
- Active → Idle: last active Campaign goes Idle
- Idle → Active: a Campaign becomes Active
- Idle → Archived: GM explicitly archives (terminal)

### Campaign
```
New → Forming → Active → Idle → Complete
```
- New → Forming: GM begins party creation
- Forming → Active: first Session starts (GUARD: at least one Character)
- Active → Idle: Session summarized
- Idle → Active: new Session starts
- Idle → Complete: GM explicitly completes (terminal)

### Session
```
New → InProgress → Completed → Summarized → Idle
```
- New → InProgress: GM starts session (GUARD: Campaign is Active, no other InProgress)
- InProgress → Completed: GM ends session
- Completed → Summarized: recap/notes written
- Summarized → Idle: closed out (terminal)

---

## Coding Rules

### Always
- Unexported fields on all structs — mutations go through methods
- Return errors, never panic in domain code
- Typed IDs everywhere — never raw strings or uuid.UUID
- Constructors validate all invariants at creation time
- One package per aggregate root layer (entity, value, port)

### Never
- Never add infrastructure imports to grimoire-domain
- Never use raw `uuid.UUID` as a field type — always a typed ID
- Never write implementation before a failing test
- Never add fields to an aggregate to satisfy a query
- Never let one aggregate root import another aggregate root's package
- Never use `interface{}` or `any` in domain code

### Errors
- Domain errors are typed, not strings
- `fmt.Errorf("context: %w", err)` for wrapping
- Sentinel errors for known domain violations

---

## The Six Canonical Events

```
EntityCreated    { entity_id, entity_type, name, source }
EntityUpdated    { entity_id, field, old_value, new_value, source }
EntityLinked     { entity_a_id, entity_b_id, relationship, source }
EntityRevealed   { entity_id, revealed_to[], session_id, source }
SessionStarted   { session_id, campaign_id, date }
SessionEnded     { session_id, notes? }
```

Every domain event maps to one of these six. See ADR-004.

---

## Bounded Contexts

```
GM Context      →  sees everything, owns world truth
PC Context      →  sees only revealed entities
                   owns character beliefs and annotations
Table Context   →  mechanical state (combat, dice, conditions)
```

---

## Event Envelope

```go
type EventEnvelope struct {
    ID             EventID        // ULID — globally sortable
    Type           EventType
    AggregateID    GrimoireID     // which aggregate this event belongs to
    AggregateType  AggregateType  // game, campaign, session, etc.
    SequenceNumber uint64         // monotonic counter within aggregate
    CampaignID     CampaignID     // routing context (zero if N/A)
    SessionID      SessionID      // routing context (zero if N/A)
    Source         Source         // obsidian | foundry | grimoire
    ActorID        string
    OccurredAt     time.Time
    Payload        Event          // typed domain event (not raw JSON)
}
```

- Source prevents infinite sync loops between Obsidian, Foundry, and Grimoire
- Payload is a typed `Event` interface in the domain; serialization to
  JSON happens at the infrastructure adapter boundary
- EventID uses ULID (generated by infrastructure), not UUID
- SequenceNumber is set atomically by Firestore transaction (ADR-012)

---

## ADR Index

| ADR | Title |
|-----|-------|
| ADR-001 | Game Holds CampaignIDs Not SessionIDs |
| ADR-002 | State Pattern for Aggregate Lifecycles |
| ADR-003 | CQRS Split — Aggregates Write, Neo4j Reads |
| ADR-004 | Six Canonical Event Types |
| ADR-005 | ndjson Over Parquet for Event Storage |
| ADR-006 | Typed UUIDs for All Identity |
| ADR-007 | GrimoireID Base Struct via Embedding |
| ADR-008 | GraphQL API Layer |
| ADR-009 | Bidirectional Event Mapping |
| ADR-010 | Three Stores Architecture |
| ADR-011 | Domain Event Chaining |
| ADR-012 | Event Sequencing |
| ADR-013 | Four Module Workspace |
| ADR-014 | Scaling Path — Firestore to Bigtable |

See `docs/adr/` for full records.