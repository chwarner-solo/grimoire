# ADR-004: Six Canonical Event Types

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire has multiple capture surfaces: an Obsidian plugin, a Foundry VTT
module, and a direct web UI. Each surface needs to emit events to the
Grimoire backend. The initial approach was to define a specific event type
per domain action (NPCRegistered, LocationCreated, NPCDefeated, etc.)
which would have produced dozens of event types across all surfaces.

## Decision
Six canonical event types cover all domain actions. Entity type and
relationship type discriminate within events.

```
EntityCreated    { entity_id, entity_type, name, source }
EntityUpdated    { entity_id, field, old_value, new_value, source }
EntityLinked     { entity_a_id, entity_b_id, relationship, source }
EntityRevealed   { entity_id, revealed_to[], session_id, source }
SessionStarted   { session_id, campaign_id, date }
SessionEnded     { session_id, notes? }
```

All domain events map to one of these six. Examples:

```
NPCRegistered    →  EntityCreated  { entity_type: "npc" }
LocationCreated  →  EntityCreated  { entity_type: "location" }
NPCDefeated      →  EntityUpdated  { field: "status", new_value: "defeated" }
PartyMeetsNPC    →  EntityLinked   { relationship: "encountered" }
LoreRevealed     →  EntityRevealed
```

## Reasoning
Schema bloat is where event systems go to die. A small canonical vocabulary
forces clarity about what an event actually means. `entity_type` and
`relationship` carry the domain specificity without proliferating types.

All three capture surfaces emit identical event shapes. The backend handles
one event schema regardless of source. New entity types and relationship
types can be added without changing the event schema.

## Consequences
- All surfaces speak the same language
- Backend handles one schema
- New domain concepts require no new event types
- Event handlers switch on entity_type/relationship for domain-specific logic
- The six types must be expressive enough for all future domain needs

## Alternatives Considered
**One event type per domain action** — rejected. Dozens of event types
across surfaces creates coordination overhead and schema proliferation.
Each new domain concept requires a new event type.

**Single generic event with arbitrary payload** — rejected. Loses all
structure and makes the event log unqueryable without full deserialization.