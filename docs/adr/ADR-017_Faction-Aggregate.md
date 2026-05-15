# ADR-017: Faction Aggregate Architecture

## Status
Accepted

## Date
2026-05-10

## Context
Factions are groups of entities with shared goals, beliefs, or allegiances.
They drive narrative consequences, group NPCs by allegiance, and track
character standing over time. Factions operate at the world level — they
exist in the Game, not in a specific Campaign — but their effects are felt
at the Campaign and Character level.

Key design tensions resolved in this ADR:

1. Faction visibility is independent of faction state — a faction can be
   actively affecting the world while players know nothing about it.

2. NPC faction membership has its own visibility state — independent of
   both the faction's visibility and the character's visibility.

3. FactionStanding is a Neo4j relationship — not a field on Character
   or Faction aggregates.

4. Faction follows the same store pattern as Narrative (ADR-016):
   Firestore holds full aggregate state, Neo4j is the read source.

---

## Faction Lifecycle

Faction follows the Game lifecycle pattern exactly:

```
New → Draft → Active → Idle → Archived
```

```
New       →  just created, being defined
Draft     →  GM building the faction
             standing levels being defined
             not yet affecting the world

Active    →  operating in the world
             affecting narrative events
             NPCs acting on behalf of faction
             standing changes possible

Idle      →  dormant, not currently acting
             standing still tracked
             could be reactivated by player choices
             or GM decision

Archived  →  defunct, historical
             terminal state
             becomes discoverable lore
             players can uncover their history
             standing frozen at archive time
```

### Transition Guards
```
New    → Draft:    always allowed
Draft  → Active:   GUARD: at least one member
                   GUARD: at least one standing level defined
Active → Idle:     GM explicitly marks dormant
Idle   → Active:   GM reactivates OR player action triggers
Idle   → Archived: GM explicitly retires (terminal)
Active → Archived: GM explicitly retires (terminal)
Archived → *:      no exit — terminal
```

---

## Visibility — Two Independent Levels

Visibility operates at two independent levels.
Both default to false (hidden) on creation.

### Level 1 — Faction Visibility
Does the player know this faction EXISTS?

```
player_visible: false  →  faction operating secretly
                           players unaware it exists
                           GM sees it fully

player_visible: true   →  players know this faction
                           can interact explicitly
                           standing shown on request
```

### Level 2 — Membership Visibility
Does the player know THIS character belongs to THAT faction?
Independent of both faction visibility and character visibility.

```
membership.player_visible: false  →  NPC affiliation hidden
                                      players don't know connection

membership.player_visible: true   →  affiliation known to players
```

### The Four Combinations
```
Faction hidden + Membership hidden:
  Players have no idea the faction exists.
  NPC seems unaffiliated.
  No connection visible.

Faction visible + Membership hidden:
  Players know the faction exists.
  They don't know this NPC is a member.
  "Korvan seems like an Inquisitor zealot
   but we can't confirm he IS one."

Faction visible + Membership visible:
  Full picture. Players know both.

Faction hidden + Membership visible:
  Players see an affiliation symbol
  but don't know what it means.
  "Bears an unfamiliar symbol."
```

### Visibility is Revealed via EntityRevealed

```
// Revealing a faction exists
EntityRevealed {
    entity_id:   faction_id
    entity_type: "faction"
    revealed_to: [player_ids]
    session_id:  session_id
}

// Revealing an NPC's faction membership
EntityRevealed {
    entity_id:   membership_link_id
    entity_type: "faction_membership"
    revealed_to: [player_ids]
    session_id:  session_id
}
```

Membership is a first-class revealable entity.
It has its own ID and its own visibility state.

---

## Aggregate Structure

### Faction

```go
// roots/faction/entity/faction.go
// Purely additive in membership.
// State machine governs lifecycle.

type Faction struct {
    // Identity
    id          FactionID
    gameID      GameID

    // Content — validated by aggregate, served by Neo4j
    name        string
    description string      // GM only, never in player projection
    goals       string      // GM only

    // Lifecycle
    status      FactionStatus   // follows Game pattern

    // Visibility
    playerVisible bool          // does faction existence show to players

    // Structure — references only
    memberIDs       []FactionMembershipID
    alliedWithIDs   []FactionID
    atWarWithIDs    []FactionID
    standingLevels  []StandingLevel   // ordered, defines what standing means
}
```

### FactionMembership — child entity

```go
// roots/faction/entity/faction_membership.go

type FactionMembership struct {
    id            FactionMembershipID
    factionID     FactionID
    characterID   NarrativeCharacterID
    rank          string          // "Inquisitor", "Grunt", "Elder"
    playerVisible bool            // independent visibility
}
```

FactionMembership is its own entity with its own ID.
This allows EntityRevealed to target it directly.

### StandingLevel — value object

```go
// roots/faction/value/standing_level.go

type StandingLevel struct {
    name      string    // "Kindly", "Dubious", "Ally"
    threshold int       // minimum points to reach this level
    ordinal   int       // sort order — 0 is lowest
}
```

Faction defines what standing levels exist and what they mean.
The ordered list is a value object on the Faction aggregate.
Faction enforces that levels are non-overlapping and ordered.

---

## FactionStanding — Neo4j Relationship

FactionStanding is NOT a field on Character or Faction.
It is a Neo4j relationship between NarrativeCharacter and Faction.

```cypher
(:NarrativeCharacter {id: "aria"})
    -[:STANDS_WITH {
        points:         850,
        level:          "Kindly",
        player_visible: true,
        last_changed:   "session_12"
    }]->
(:Faction {id: "freeport_militia"})
```

### Standing is updated by events

Every narrative event that affects standing emits:
```
EntityUpdated {
    entity_id:   character_faction_standing_id
    entity_type: "faction_standing"
    field:       "points"
    old_value:   "700"
    new_value:   "850"
    payload: {
        character_id: aria_id
        faction_id:   militia_id
        delta:        150
        reason:       "completed quest: Rescue the Merchant"
    }
}
```

Neo4j handler updates the STANDS_WITH relationship properties.
Standing level is recalculated from points against Faction.standingLevels.

### Standing time-series tracking

Full history preserved in GCS event log.
Every EntityUpdated { entity_type: "faction_standing" } is appended.
Time-series analytics via BigQuery — deferred to future ADR.

---

## Faction Relationships

Factions relate to other Factions. These are Neo4j edges.
The aggregate holds the IDs. Neo4j holds the traversal.

```cypher
(:Faction {id: "inquisition"})
    -[:AT_WAR_WITH]->
(:Faction {id: "druids_circle"})

(:Faction {id: "combine"})
    -[:ALLIED_WITH]->
(:Faction {id: "inquisition"})
```

### Invariant: no contradictory relationships

```go
func (f *Faction) AddAlly(id FactionID) error {
    if f.isAtWarWith(id) {
        return ErrCannotBeAlliedAndAtWar{FactionID: id}
    }
    f.alliedWithIDs = append(f.alliedWithIDs, id)
    return nil
}

func (f *Faction) DeclareWar(id FactionID) error {
    if f.isAlliedWith(id) {
        return ErrCannotBeAtWarAndAllied{FactionID: id}
    }
    f.atWarWithIDs = append(f.atWarWithIDs, id)
    return nil
}
```

Faction cannot be both allied and at war with the same faction.
Enforced by the aggregate on the command side.

---

## Store Pattern

Follows ADR-016 exactly.

**Firestore holds full aggregate state:**
```
Faction document:
  id, gameID, name, description, goals
  status, playerVisible
  memberIDs[], alliedWithIDs[], atWarWithIDs[]
  standingLevels[]

FactionMembership document:
  id, factionID, characterID, rank, playerVisible
```

**Neo4j holds content and relationships:**
```
(:Faction) nodes with all content properties
(:Faction)-[:AT_WAR_WITH]->(:Faction)
(:Faction)-[:ALLIED_WITH]->(:Faction)
(:NarrativeCharacter)-[:MEMBER_OF {playerVisible}]->(:Faction)
(:NarrativeCharacter)-[:STANDS_WITH {points, level}]->(:Faction)
```

**GCS holds full event log permanently.**

---

## Events Faction Emits

All six canonical events are sufficient:

```
EntityCreated  { entity_type: "faction" }
EntityCreated  { entity_type: "faction_membership" }
EntityUpdated  { field: "status" }           // lifecycle
EntityUpdated  { field: "player_visible" }   // visibility change
EntityUpdated  { field: "standing_levels" }  // GM adds/edits levels
EntityLinked   { relationship: "member_of" } // NPC joins faction
EntityLinked   { relationship: "allied_with" }
EntityLinked   { relationship: "at_war_with" }
EntityLinked   { relationship: "stands_with" } // standing created
EntityUpdated  { entity_type: "faction_standing",
                 field: "points" }            // standing changes
EntityRevealed { entity_type: "faction" }    // faction disclosed
EntityRevealed { entity_type:
                 "faction_membership" }       // membership disclosed
```

### Event Handler Responsibilities

```
EntityCreated { entity_type: "faction" }:
  AggregateStore  →  saves full Faction struct
  Neo4jHandler    →  creates (:Faction) node
                     with all content properties
  GCSHandler      →  appends full event ndjson

EntityLinked { relationship: "member_of" }:
  AggregateStore  →  saves FactionMembership document
                     appends memberID to Faction
  Neo4jHandler    →  creates MEMBER_OF edge
                     with playerVisible property
  GCSHandler      →  appends event

EntityRevealed { entity_type: "faction_membership" }:
  AggregateStore  →  updates FactionMembership.playerVisible
  Neo4jHandler    →  updates MEMBER_OF edge playerVisible
  GCSHandler      →  appends event

EntityUpdated { entity_type: "faction_standing" }:
  AggregateStore  →  no aggregate holds standing
                     (standing is a Neo4j relationship)
  Neo4jHandler    →  updates STANDS_WITH relationship
                     recalculates level from points
  GCSHandler      →  appends audit event
```

---

## Faction Ports

```go
// roots/faction/port/faction_repository.go

type FactionRepository interface {
    FindByID(ctx context.Context, id FactionID) (*Faction, error)
    FindMembership(ctx context.Context,
        id FactionMembershipID) (*FactionMembership, error)
    Save(ctx context.Context, faction *Faction) error
    SaveMembership(ctx context.Context,
        membership *FactionMembership) error
}
```

---

## Scaling Transparency

Follows ADR-014. Command side uses AggregateStore port.
Phase 1: Firestore. Phase 3: Bigtable snapshot + replay.
Command handler code identical in both phases.

---

## Deferred

```
Location-based faction presence    →  Location ADR
Faction symbols at locations       →  GM Planning context ADR
FactionStanding time-series        →  Future ADR
Standing decay over game time      →  Future ADR
Standing effects on NPC behavior   →  GM Planning context ADR
```

---

## Consequences

- Faction visibility and membership visibility are independent
- FactionStanding lives in Neo4j — never on an aggregate
- Standing history fully preserved in GCS event log
- Contradictory faction relationships (allied + at war) caught
  by aggregate on command side
- FactionMembership is a first-class entity with its own ID
  enabling EntityRevealed to target it directly
- Faction content (name, goals, description) validated by
  aggregate, served by Neo4j read projections
- Standing level calculation happens in Neo4j handler
  using Faction.standingLevels from the event payload

## Alternatives Considered

**FactionStanding as field on Character** — rejected. Standing belongs
to the relationship between Character and Faction. Putting it on
Character couples Character to every faction in the world. Neo4j
relationship is the correct model.

**Membership visibility as field on Faction** — rejected. Faction
cannot know which players know which of its members. Visibility
belongs on the membership relationship itself.

**Faction lifecycle simpler than Game** — rejected. Faction affecting
the world vs dormant vs historical are meaningfully distinct states
that drive narrative consequences. The Game lifecycle maps exactly.