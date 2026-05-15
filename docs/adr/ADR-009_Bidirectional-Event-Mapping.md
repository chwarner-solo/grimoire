# ADR-009: Bidirectional Event Mapping Between Obsidian and Foundry

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire has two primary GM capture surfaces — an Obsidian plugin and a
Foundry VTT module. GMs may work in either surface at any time:

- Creating an NPC in Obsidian should create an Actor in Foundry
- Defeating an NPC in Foundry combat should update the NPC status in Obsidian
- Revealing an entity in Obsidian should update Foundry token visibility
- Creating a journal entry in Foundry should create a lore note in Obsidian

Without bidirectional sync, the GM must manually maintain both surfaces.
This creates drift, duplication, and errors.

## Decision
The Grimoire backend is the single source of truth and the sync broker.
Neither Obsidian nor Foundry communicate directly with each other.
All sync flows through Grimoire events.

**Architecture:**
```
Obsidian Plugin          Foundry Module
(narrative layer)        (mechanical layer)
      │                        │
      │  emit events           │  emit events
      └──────────┬─────────────┘
                 ▼
         Grimoire Backend
         canonical event log
         sync broker
                 │
      ┌──────────┴──────────┐
      ▼                     ▼
 push to Foundry       push to Obsidian
 (if source ≠ foundry) (if source ≠ obsidian)
```

**Infinite loop prevention:**
Every event envelope carries a `source` field:
```
source: obsidian | foundry | grimoire
```

The backend only pushes to a surface if that surface was NOT the event
source. An event from Obsidian is pushed to Foundry but not back to
Obsidian.

**Field ownership — conflict resolution:**
Obsidian and Foundry own different fields on the same entity.
The backend enforces ownership boundaries and never allows one surface
to overwrite fields owned by the other.

```
Obsidian owns:    narrative fields
                  name, description, lore, relationships,
                  faction, player_visible, gm_notes

Foundry owns:     mechanical fields
                  hp, stats, abilities, conditions,
                  token position, combat state
```

Fields do not overlap. Conflict resolution is unnecessary.

**Obsidian → Foundry event mappings:**
```
EntityCreated  { entity_type: "npc" }    →  Create Foundry Actor
EntityCreated  { entity_type: "location" } →  Create Foundry Scene
EntityCreated  { entity_type: "lore" }   →  Create Foundry Journal Entry
EntityUpdated  { field: "status",
                 new_value: "defeated" } →  Set Actor HP to 0
EntityRevealed                           →  Set token/journal visibility
EntityLinked   { relationship:
                 "present_in" }          →  Place token in Scene
```

**Foundry → Obsidian event mappings:**
```
Actor created                →  EntityCreated { entity_type: "npc" }
                                Create note in grimoire/npcs/
Combat ended, actor defeated →  EntityUpdated { field: "status",
                                  new_value: "defeated" }
                                Update status in NPC frontmatter
Journal entry created        →  EntityCreated { entity_type: "lore" }
                                Create note in grimoire/lore/
Scene created                →  EntityCreated { entity_type: "location" }
                                Create note in grimoire/locations/
Token placed on map          →  EntityLinked { relationship: "present_in" }
```

**Sync state in Obsidian frontmatter:**
```yaml
foundry_actor_id: abc123xyz
foundry_sync_status: synced | pending | conflict
foundry_last_sync: 2026-05-09T19:32:00Z
```

**Foundry sync state in Actor flags:**
```json
{
  "flags": {
    "grimoire": {
      "obsidian_path": "grimoire/npcs/inquisitor-korvan.md",
      "grimoire_id": "npc_korvan_001",
      "sync": true,
      "last_sync": "2026-05-09T19:32:00Z"
    }
  }
}
```

## Foundry Push Mechanism
The Grimoire backend pushes to the Foundry module via WebSocket. The
Foundry module maintains a persistent WebSocket connection to the backend.
The backend sends event payloads over this connection when sync is required.

The Foundry REST API (foundryvtt-rest-api module) handles Actor/Scene/
Journal creation and updates initiated by the backend.

## Obsidian Push Mechanism
The Grimoire backend pushes to the Obsidian plugin via webhook. The plugin
registers a local HTTP endpoint that the backend calls when sync is required.
The plugin applies changes to vault files directly — creating notes,
updating frontmatter, writing content.

## Reasoning
Grimoire as sync broker means neither surface needs to know about the other.
Adding a third surface (direct web UI, player app write access) requires
only adding event mappings in the backend — no changes to existing surfaces.

Field ownership eliminates conflict resolution complexity entirely. Each
surface writes what it owns and the backend enforces the boundary.

Source tracking on the event envelope is a simple and reliable mechanism
for loop prevention that requires no coordination between surfaces.

## Consequences
- GM workflow in either surface is fully reflected in the other
- Neither surface requires knowledge of the other
- Field ownership must be strictly maintained — both surfaces must respect
  the boundary
- Foundry requires the foundryvtt-rest-api module installed
- Obsidian plugin must expose a local webhook endpoint
- Network connectivity required for sync — offline changes queue in outbox

## Alternatives Considered
**Direct Obsidian ↔ Foundry communication** — rejected. Creates tight
coupling between surfaces. Adding a third surface requires changes to both
existing surfaces. No canonical source of truth.

**Obsidian as source of truth, Foundry reads from it** — rejected. Foundry
owns mechanical state that Obsidian cannot and should not capture. The
ownership split requires a neutral broker.

**Manual sync by the GM** — rejected. The entire value proposition of
Grimoire is removing this burden from the GM.