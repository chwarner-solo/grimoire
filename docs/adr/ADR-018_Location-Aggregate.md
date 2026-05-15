# ADR-018: Location Aggregate Architecture

## Status
Accepted

## Date
2026-05-10

## Context
Locations are the spatial layer of the narrative. They are where beats
happen, where NPCs exist, where factions operate, and where the party
travels. Locations are recursive — a world contains regions, regions
contain settlements, settlements contain buildings. Every level of the
hierarchy can contain Scenes directly.

Scenes within a Location form a Directed Acyclic Graph (DAG) using the
same aggregate store pattern as narrative beats (ADR-016). The Location
hierarchy itself is also a DAG — cycle prevention uses the same repository
chain loading pattern.

This ADR supersedes the initial ADR-018 draft. Nine gaps identified during
implementation planning are resolved here as first-class decisions.

---

## Future ADR Note — Content as First-Class Entity

Content fields (name, description, playerDescription, etc.) on Location
and Scene aggregates are string fields in this ADR. A future ADR will
introduce Content as a first-class entity — a ContentID reference
replacing raw string fields. This enables map tiles, textures, audio,
handouts, and cutscenes as Content entities in MMO clients, delivered
via CDN with independent visibility and lifecycle. Do not refactor until
that ADR is written and accepted.

---

## Future ADR Note — MMORPG Beat Prerequisites

Scene prerequisites in this ADR are SceneIDs only. The TTRPG context
relies on GM judgment for narrative gates — the GM decides when a scene
is accessible based on story context. A future MMORPG ADR will introduce
beatPrerequisiteSets [][]BeatID on Scene for system-enforced narrative
gates where no GM is present. Do not add this until that ADR is written.

---

## Gap Resolutions

### Gap 1 — Missing SceneID
SceneID added to shared/identity/ids.go:
```go
type SceneID struct{ GrimoireID }
```

Value types added:
```
roots/location/value/location_status.go
roots/location/value/scene_status.go
roots/location/value/location_type.go
```

### Gap 2 — AggregateType for Scene
Scene events route through the Location aggregate type.
Scene is a child entity — not an aggregate root.
Scene events use aggregate_type: "location" and aggregate_id: location_id.
No AggregateScene type needed.

### Gap 3 — CampaignNarrative Extension
currentLocationID does NOT belong on CampaignNarrative.
CampaignNarrative is purely additive — mutable position state
contradicts this design principle.

Party position moves to Campaign aggregate:
```go
// Campaign aggregate addition
type campaignCore struct {
...
currentLocationID LocationID   // where the party is now
}
```

exploredSceneIDs and visitedLocationIDs ARE additive and stay
on CampaignNarrative. This is a breaking change — snapshot and
reconstitute require migration.

### Gap 4 — Scene Status is Per-Campaign
Scene has NO status field on the aggregate. A scene in Campaign A
may be Explored while the same scene in Campaign B is Locked.
Scene status is a Neo4j relationship property:

```cypher
(:Campaign)-[:SCENE_STATUS {
    status:  "explored",
    session: 12
}]->(:Scene)
```

CampaignNarrative.exploredSceneIDs is the command-side truth
for prerequisite checking. Neo4j relationship is the read model.

### Gap 5 — State Pattern Consistency
Location uses the state pattern (sealed interfaces) consistent with
Game and Faction. No status field. The type IS the status.

```go
type NewLocation interface {
BeginDraft() (DraftLocation, error)
}
type DraftLocation interface {
AddScene(id SceneID) (DraftLocation, error)
Activate() (ActiveLocation, error)
}
type ActiveLocation interface {
AddScene(id SceneID) (ActiveLocation, error)
ConnectTo(id LocationID) (ActiveLocation, error)
GoIdle() (IdleLocation, error)
}
type IdleLocation interface {
Activate() (ActiveLocation, error)
Archive() (ArchivedLocation, error)
}
type ArchivedLocation interface {
// terminal — no methods
}
```

### Gap 6 — Archive Guard at Interactor Level
Aggregates do not query other aggregates. The party-presence
guard for archiving is enforced at the interactor level:

```
ArchiveLocationInteractor:
  1. Load all Campaigns for this Game
  2. Check Campaign.currentLocationID
     against location being archived
  3. If any Campaign present → reject with error
  4. Dispatch ArchiveLocation command
  5. Location aggregate enforces structural
     invariants only (no active children)
```

### Gap 7 — Directed Graph Connections
Travel connections are directed (one-way). The interactor creates
both directions when symmetric travel is intended:

```
ConnectLocationsInteractor (symmetric):
  ConnectLocation(a → b) command
  ConnectLocation(b → a) command

Some connections are intentionally one-way:
  "You can fall into the pit"
  "You cannot climb back out"
  GM creates only one direction.
```

### Gap 8 — LocationType Hierarchy Validation
Deferred explicitly. No hierarchy validation enforced now.
GM is responsible for sensible structure. A future ADR or GM
Planning context ADR may add hierarchy rules.

### Gap 9 — Cross-Entity Prerequisites (Beat + Scene)
Scene prerequisites are SceneIDs only in the TTRPG context.
Beat→Scene gates are informational — not structurally enforced.

Scene holds recommendedBeatIDs as a flat slice:
```go
recommendedBeatIDs []BeatID  // informational only
// surfaces in GM planning view
// never enforced by command side
```

GM sees recommended beats in their planning view and makes the
call. The system does not enforce narrative gates in TTRPG context.
See Future ADR Note above for MMORPG enforcement.

---

## Location Lifecycle

Location follows the Game/Faction state machine pattern exactly.
Sealed interfaces — the type IS the status. No status field.

```
New → Draft → Active → Idle → Archived
```

```
New       →  just created
Draft     →  GM building, scenes being defined
             not yet in the narrative world
Active    →  part of the living world
             beats happening, NPCs present
             party can travel here
Idle      →  exists but not currently relevant
             no active beats, party has moved on
             can reactivate
Archived  →  destroyed, abandoned (terminal)
             all children cascade to archived
             becomes discoverable lore
```

### Transition Guards
```
Draft  → Active:   GUARD: at least one Scene defined
                   (aggregate checks sceneIDs not empty)
Active → Archived: GUARD: no child Location Active
                   (interactor loads children, checks states —
                    aggregate cannot query other aggregates)
                   GUARD: party not present
                   (interactor checks Campaign.currentLocationID —
                    same reason)
Archived → *:      terminal
```

### Archive Cascade
```
Location archives (terminal):
  LocationArchivedHandler via EventBus
        ↓
  All child Locations → archive recursively
  All Scenes          → archived
  All active beats at location → historical
  Neo4j nodes flagged archived
  Player app: becomes discoverable lore
```

---

## Recursive Hierarchy

Location contains Location. Every level can have Scenes directly.

```
Norrath               (world,      parentID: nil)
  Eastern Karana      (region,     parentID: norrath)
    scenes[]          ← wilderness scenes at region level
    Blackburrow       (dungeon,    parentID: eastern_karana)
      scenes[]        ← throne room, warrens, antechamber
    Qeynos Hills      (region,     parentID: eastern_karana)
      scenes[]        ← outdoor encounter areas
      Qeynos          (settlement, parentID: qeynos_hills)
        scenes[]      ← city district areas
        The Crow's Pub (building,  parentID: qeynos)
          scenes[]    ← main bar, back room, cellar
```

---

## Aggregate Structures

### Location

```go
// roots/location/entity/location.go
// State pattern — no status field, type IS the status

type locationCore struct {
    // Identity
    id               LocationID
    gameID           GameID
    locationType     LocationType    // set at construction, immutable thereafter
                                    // a region cannot become a settlement

    // Hierarchy — IDs only, DAG validated via repository
    parentLocationID LocationID      // IsZero() if top level
    childLocationIDs []LocationID
    sceneIDs         []SceneID

    // Connections — directed graph
    connectedLocationIDs []LocationID

    // Visibility
    playerVisible    bool

    // Content — validated here, served by Neo4j
    // NOTE: future ADR replaces with []ContentID
    name                 string
    description          string    // GM only, never shown to players
    playerDescription    string    // shown on first reveal
    revealedNotes        string    // GM notes shown to players
                                   // after party has been nearby
}
```

### Scene

```go
// roots/location/entity/scene.go
// Child entity — not an aggregate root
// No status field — status is per-campaign (Gap 4)

type Scene struct {
    // Identity
    id           SceneID
    locationID   LocationID
    gameID       GameID

    // DAG — same pattern as Beat (ADR-016)
    scenePrerequisiteSets [][]SceneID   // outer=OR, inner=AND
                                         // spatial prerequisites only

    // Informational beat references (Gap 9)
    recommendedBeatIDs    []BeatID      // GM planning view only
                                         // never enforced by command side

    // Visibility
    playerVisible bool

    // Content — validated here, served by Neo4j
    // NOTE: future ADR replaces with []ContentID
    name                 string
    description          string    // GM only, never shown to players
    playerDescription    string    // shown on first entry
    exploredDescription  string    // shown after explored
                                   // "You remember this place..."
}
```

---

## Description Fields

### Location
```
name                  visible once Location revealed to players
description           GM only — planning notes, secrets
                      NEVER shown to players under any circumstances
playerDescription     shown when Location first revealed
                      first impression on discovery
revealedNotes         GM notes that ARE shown to players
                      after party has been in the area
```

### Scene
```
name                  visible once Scene revealed
description           GM only — NEVER shown to players
playerDescription     shown on first entry (scene becomes Active)
exploredDescription   shown after Scene status: explored in Neo4j
                      GM can update over time
                      "The throne room where you defeated Korvan.
                       The scorch marks remain."
```

All description fields:
- Validated by aggregate (not empty where required)
- Stored in Firestore as full aggregate state (ADR-016 pattern)
- Served by Neo4j as read projections
- GM-only fields filtered at API layer always
- Never exposed through player-context GraphQL queries

---

## DAG Pattern — Three Consistent Implementations

The same repository chain loading pattern (ADR-016) applies to
both Scene prerequisites and Location hierarchy.

```
Narrative (ADR-016):   BeatRepository.LoadPrerequisiteChain
                       Beat.WouldCreateCycle

Location hierarchy:    LocationRepository.LoadChildChain
                       Location.WouldCreateCycle

Scene prerequisites:   SceneRepository.LoadPrerequisiteChain
                       Scene.WouldCreateCycle
```

Same pattern. Same code shape. Different entity types.

### Location Cycle Detection

```go
// roots/location/entity/location — on locationCore
func (l *locationCore) WouldCreateCycle(
    targetID LocationID,
    chain []locationCore,
) error {
    for _, ancestor := range chain {
        if ancestor.id == targetID {
            return ErrLocationCycleDetected{LocationID: l.id}
        }
    }
    return nil
}
```

### Scene Handle / Replay

Scene has no status field on the aggregate (Gap 4). Scene state
is per-campaign and lives in Neo4j. Therefore Scene has no state
machine, no Handle method, and no state transitions to replay.

Snapshot and Reconstitute are sufficient for Scene. The Scene
aggregate is purely structural — IDs, prerequisites, content fields,
and playerVisible. It is created, updated (content), and archived.
No Handle/Replay pattern needed.

```go
// roots/location/entity/scene.go
func (s *Scene) WouldCreateCycle(
    targetID SceneID,
    chain []Scene,
) error {
    for _, ancestor := range chain {
        if ancestor.id == targetID {
            return ErrSceneCycleDetected{SceneID: s.id}
        }
    }
    return nil
}
```

---

## Location Ports

```go
// roots/location/port/location_repository.go
type LocationRepository interface {
    FindByID(ctx context.Context,
        id LocationID) (Location, error)      // sealed interface
    LoadChildChain(ctx context.Context,
        id LocationID) ([]Location, error)    // sealed interface
    Save(ctx context.Context, location Location) error
}

// roots/location/port/scene_repository.go
type SceneRepository interface {
    FindByID(ctx context.Context,
        id SceneID) (*Scene, error)
    FindByLocation(ctx context.Context,
        id LocationID) ([]Scene, error)
    LoadPrerequisiteChain(ctx context.Context,
        id SceneID) ([]Scene, error)
    Save(ctx context.Context, scene *Scene) error
}
```

---

## CampaignNarrative Extension

exploredSceneIDs and visitedLocationIDs are additive state.
They belong on CampaignNarrative. currentLocationID does NOT
(mutable — belongs on Campaign aggregate).

```go
// ADR-016 CampaignNarrative updated — breaking change
type CampaignNarrative struct {
    id                  CampaignNarrativeID
    campaignID          CampaignID
    gameID              GameID

    // Narrative DAG (ADR-016)
    discoveredBeatIDs   []BeatID
    campaignBeatIDs     []BeatID
    decisionIDs         []DecisionID
    revelationIDs       []RevelationID

    // Spatial DAG (ADR-018) — additive only
    exploredSceneIDs    []SceneID
    visitedLocationIDs  []LocationID
}

// Campaign aggregate addition
type campaignCore struct {
    ...
    currentLocationID LocationID  // mutable party position
}
```

Breaking change: snapshot and reconstitute require migration
for both CampaignNarrative and Campaign aggregates.

### Scene Prerequisite Check

```go
func (cn *CampaignNarrative) CanAccessScene(scene Scene) bool {
    return cn.scenePrerequisitesMet(scene)
    // recommendedBeatIDs intentionally not checked here
    // GM judgment handles narrative gates in TTRPG context
}

func (cn *CampaignNarrative) scenePrerequisitesMet(scene Scene) bool {
    for _, set := range scene.scenePrerequisiteSets {
        if cn.allScenesExplored(set) {
            return true
        }
    }
    return len(scene.scenePrerequisiteSets) == 0
}
```

---

## Travel Connections — Directed Graph

Connections are directed (one-way). Interactor creates both
directions when symmetric travel is intended.

```go
// On ActiveLocation state
func (l *activeLocationState) ConnectTo(
    id LocationID,
) (ActiveLocation, error) {
    if l.core.id == id {
        return nil, ErrCannotConnectToSelf
    }
    if l.core.isConnectedTo(id) {
        return nil, ErrAlreadyConnected{LocationID: id}
    }
    // returns new state with connection added
}
```

Neo4j edge carries travel properties:
```cypher
(:Location {name: "Blackburrow"})
    -[:CONNECTS_TO {
        travel_time:    "immediate",
        difficulty:     "dangerous",
        player_visible: true
    }]->
(:Location {name: "Eastern Karana"})
```

---

## Neo4j Read Model

### Full hierarchy
```cypher
MATCH (root:Location {id: $locationId})
      -[:CONTAINS*]->(child:Location)
OPTIONAL MATCH (child)-[:HAS_SCENE]->(s:Scene)
RETURN root, child, s
```

### Per-campaign scene status
```cypher
MATCH (c:Campaign {id: $campaignId})
      -[status:SCENE_STATUS]->(s:Scene)
WHERE s.locationId = $locationId
RETURN s, status.status, status.session
```

### Available scenes for campaign
```cypher
MATCH (candidate:Scene {locationId: $locationId})
WHERE NOT EXISTS {
    MATCH (c:Campaign {id: $campaignId})
          -[:SCENE_STATUS {status: "explored"}]->(candidate)
}
AND (
    NOT EXISTS((candidate)<-[:UNLOCKED_BY]-(:PrerequisiteSet))
    OR EXISTS {
        MATCH (candidate)<-[:UNLOCKED_BY]-(ps:PrerequisiteSet)
              -[:REQUIRES]->(req:Scene)
        WHERE ALL(r IN collect(req) WHERE EXISTS {
            MATCH (c:Campaign {id: $campaignId})
                  -[:SCENE_STATUS {status: "explored"}]->(r)
        })
    }
)
RETURN candidate
```

### GM planning view — recommended beats
```cypher
MATCH (s:Scene {id: $sceneId})
OPTIONAL MATCH (s)-[:RECOMMENDED_BEFORE]->(b:Beat)
OPTIONAL MATCH (c:Campaign {id: $campaignId})
              -[:DISCOVERED]->(b)
RETURN s,
       b,
       CASE WHEN (c)-[:DISCOVERED]->(b)
            THEN "completed"
            ELSE "pending"
       END as beatStatus
```

---

## Events Location Emits

All six canonical events sufficient:

```
EntityCreated  { entity_type: "location" }
EntityCreated  { entity_type: "scene" }
EntityUpdated  { entity_type: "location",
                 field: "state" }           // lifecycle transition
EntityUpdated  { entity_type: "location",
                 field: "player_visible" }
EntityUpdated  { entity_type: "location",
                 field: "content" }         // description update
EntityLinked   { relationship: "contains" } // parent→child location
EntityLinked   { relationship: "connects_to" }
EntityLinked   { relationship: "has_scene" }
EntityLinked   { relationship: "prerequisite_of" } // scene DAG
EntityLinked   { relationship: "recommended_before",
                 entity_a: beat_id,
                 entity_b: scene_id }       // informational only
EntityLinked   { relationship: "visited",
                 entity_a: campaign_id,
                 entity_b: location_id }
EntityLinked   { relationship: "explored",
                 entity_a: campaign_id,
                 entity_b: scene_id }
EntityRevealed { entity_type: "location" }
EntityRevealed { entity_type: "scene" }
```

---

## Store Pattern

Follows ADR-016 exactly:

```
Firestore:   full aggregate state
             Location state struct (all fields)
             Scene documents (all fields)
             Campaign.currentLocationID

Neo4j:       all content as node properties
             CONTAINS hierarchy edges
             CONNECTS_TO travel edges (directed)
             HAS_SCENE edges
             SCENE_STATUS per-campaign relationships
             EXPLORED / VISITED campaign edges
             prerequisite DAG edges for scenes
             RECOMMENDED_BEFORE informational edges

GCS:         full event log permanently
```

---

## Scaling Transparency

Follows ADR-014. AggregateStore port used throughout.
Phase 1: Firestore. Phase 3: Bigtable snapshot + replay.
Command handlers identical in both phases.

---

## Deferred

```
LocationType hierarchy validation  →  Future ADR or
                                       GM Planning context ADR

Faction presence at locations      →  GM Planning context ADR
Faction symbols at locations       →  GM Planning context ADR

Content as first-class entity      →  Future ADR
                                       replaces string fields
                                       with []ContentID
                                       enables maps, audio,
                                       textures, cutscenes

MMORPG beat prerequisites          →  Future ADR
                                       beatPrerequisiteSets [][]BeatID
                                       system-enforced narrative gates

Gated travel connections         →  Future ADR or GM Planning context
                                       requires_beat on CONNECTS_TO edge
                                       consistent with "GM judgment for
                                       narrative gates" decision
                                       MMORPG may need system enforcement
```

---

## Consequences

- Location hierarchy is recursive — any level can have scenes
- Scene status is per-campaign (Neo4j relationship) not aggregate field
- Scene prerequisites are spatial only (SceneIDs) — GM judgment
  handles narrative gates in TTRPG context
- recommendedBeatIDs is informational — surfaces in GM planning view
- Travel connections are directed — interactor creates both directions
  for symmetric travel
- Archive cascade is recursive via EventBus handlers
- Archive party-presence guard lives at interactor level
- CampaignNarrative gains exploredSceneIDs and visitedLocationIDs
- Campaign gains currentLocationID — breaking change with migration
- State pattern consistent with Game and Faction — no status field
- Three consistent DAG implementations across Narrative and Location

## Alternatives Considered

**Scene prerequisites include BeatIDs** — rejected for TTRPG context.
GM judgment handles narrative gates. Beat prerequisites deferred to
MMORPG ADR where system enforcement is required.

**Scene status as aggregate field** — rejected. Same scene exists
across multiple campaigns in different states. Status is per-campaign
and lives as a Neo4j relationship property.

**currentLocationID on CampaignNarrative** — rejected. Mutable
position state contradicts the purely additive design of
CampaignNarrative. Belongs on Campaign aggregate.

**Bidirectional connections** — rejected. Some connections are
intentionally one-way. Interactor handles symmetric case explicitly.

**Flat Location structure** — rejected. World→Region→Settlement→
Building→Scene is a first-class domain concept in both TTRPG
and MMORPG contexts.