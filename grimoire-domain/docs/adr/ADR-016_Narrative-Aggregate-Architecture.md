# ADR-016: Narrative Aggregate Architecture — Authoritative Record

## Status
Accepted

## Date
2026-05-10

## Supersedes
ADR-015: Narrative as a Directed Acyclic Graph
ADR-015-Amendment-001: Narrative Architecture Gaps Resolved

## Context
The Narrative domain is the most complex aggregate system in Grimoire.
It requires a graph structure for story beat prerequisites, two distinct
narrative layers (master and campaign), shared entities across roots,
and clear separation between the write side (command/aggregate) and the
read side (Neo4j projections).

ADR-015 and its amendment produced contradictions — specifically around
what Firestore holds vs what Neo4j holds. This ADR supersedes both and
establishes a single authoritative record with no contradictions.

---

## Foundational Principle: Write Side vs Read Side

```
Write side (Firestore/Bigtable):
  Full aggregate state.
  Structural fields AND content fields.
  Everything the aggregate needs to enforce invariants
  and reconstruct itself from a snapshot.
  Loaded by the command handler. Never by queries.

Read side (Neo4j):
  All content as node properties.
  All relationships as edges.
  The source clients read from.
  Populated as a projection of events.
  Never written to directly from command handlers.

Event log (GCS/Bigtable):
  Full event payload preserved permanently.
  Source of truth for rebuilding either store.
```

The distinction is **read path vs write path** — not which store
holds content. Firestore holds full aggregate state including content.
Neo4j is what clients read. They are projections of the same truth.

---

## The Narrative is a DAG

Story beats have prerequisite relationships. Beat B cannot be discovered
until Beat A is discovered. Multiple prerequisite paths can lead to the
same beat. Campaigns take different paths through the same master story.

```
Nodes:    NarrativeBeats
Edges:    PREREQUISITE_OF relationships
Traversal: "what beats are available to this campaign right now?"
```

Neo4j handles DAG traversal natively and efficiently.
The command side enforces prerequisites using local aggregate data.

---

## Two Narrative Layers

### MasterNarrative — owned by Game
```
The GM's canonical world truth.
Exists independently of any Campaign.
All Campaigns reference its beats.
Beats marked required or optional.
Archived when Game is archived.
No independent lifecycle.
```

### CampaignNarrative — owned by Campaign
```
This table's path through the master story.
Tracks which beats have been discovered.
Holds campaign-specific beat IDs.
Tracks decisions and revelations unique to this table.
Complete when Campaign is complete.
No independent lifecycle.
```

Both aggregates are **purely additive event journals**.
Neither has a state machine. This is an intentional departure
from the Game/Campaign/Session pattern. Document this with a
comment on each struct in code.

---

## Typed IDs

NarrativeID is retired. Eight typed IDs replace it,
all following the GrimoireID embedding pattern (ADR-007):

```go
type MasterNarrativeID   struct{ GrimoireID }
type CampaignNarrativeID struct{ GrimoireID }
type BeatID              struct{ GrimoireID }
type ActID               struct{ GrimoireID }
type SecretID            struct{ GrimoireID }
type LoreID              struct{ GrimoireID }
type DecisionID          struct{ GrimoireID }
type RevelationID        struct{ GrimoireID }
```

---

## Aggregate Structures

### Beat — shared entity in roots/narrative/entity/

Beat is shared between MasterNarrative and CampaignNarrative.
Neither root owns it. Both hold BeatIDs as references.
Beat knows its own scope.

```go
type BeatScope string

const (
    BeatScopeMaster   BeatScope = "master"
    BeatScopeCampaign BeatScope = "campaign"
)

type BeatType string

const (
    BeatTypeRequired         BeatType = "required"
    BeatTypeOptional         BeatType = "optional"
    BeatTypeCampaignSpecific BeatType = "campaign-specific"
)

type Beat struct {
    // Identity and scope
    id               BeatID
    scope            BeatScope
    beatType         BeatType
    gameID           GameID
    campaignID       CampaignID    // IsZero() when scope is master
    status           BeatStatus

    // Prerequisite structure — enforced by aggregate
    prerequisiteSets [][]BeatID    // outer=OR, inner=AND
    revealsSecretIDs []SecretID

    // Content — validated by aggregate, served by Neo4j
    name              string
    description       string       // GM only, never in player projection
    playerDescription string       // player-facing, controlled by GM
}
```

### MasterNarrative

```go
// Purely additive. No state machine. Lifecycle follows Game.
type MasterNarrative struct {
    id        MasterNarrativeID
    gameID    GameID
    actIDs    []ActID
    beatIDs   []BeatID
    secretIDs []SecretID
    loreIDs   []LoreID
}
```

### CampaignNarrative

```go
// Purely additive. No state machine. Lifecycle follows Campaign.
type CampaignNarrative struct {
    id                  CampaignNarrativeID
    campaignID          CampaignID
    gameID              GameID
    discoveredBeatIDs   []BeatID     // prerequisite enforcement
    campaignBeatIDs     []BeatID     // campaign-specific beats
    decisionIDs         []DecisionID
    revelationIDs       []RevelationID
}
```

---

## Beat Creation

Beat is created by a standalone command — not by MasterNarrative directly:

```
CreateMasterBeat command   →  scope: master,   gameID
CreateCampaignBeat command →  scope: campaign, campaignID
```

MasterNarrative.AddBeat(id BeatID) records the reference only.
Beat creation is its own command with its own handler.

---

## Game Transition Trigger

Game.AddNarrativeElement() is retired. **Breaking change.**

Files requiring update: states.go, transitions.go, handle.go,
reconstitute.go, and all associated tests.

The trigger for Game New → Draft is MasterNarrativeCreated,
delivered via the EventBus (ADR-011):

```
CreateMasterNarrative command
        ↓
MasterNarrative created and saved
        ↓
MasterNarrativeCreated event emitted
        ↓
GameStatusHandler.Handle(MasterNarrativeCreated)
        ↓
game.OnNarrativeCreated(id MasterNarrativeID)
        ↓
Game New → Draft
Game stores MasterNarrativeID as reference
```

---

## CampaignNarrative Creation

CampaignNarrative is created automatically via the EventBus
when a Campaign is created:

```
CampaignCreated event
        ↓
CampaignNarrativeHandler.Handle(CampaignCreated)
        ↓
CreateCampaignNarrative command
        ↓
CampaignNarrative created with campaignID + gameID
        ↓
CampaignNarrativeCreated event emitted
```

Reference is one-way:
```
CampaignNarrative holds campaignID   ←  knows its parent
Campaign does NOT hold CampaignNarrativeID
Neo4j: (:Campaign)-[:HAS_NARRATIVE]->(:CampaignNarrative)
```

---

## Beat Content Updates

Content updates go through the command side.
The Beat aggregate validates content fields and emits an event.
The Neo4j handler updates node properties as a projection.

```go
func (b *Beat) UpdateContent(
    name string,
    description string,
    playerDescription string,
) error {
    if strings.TrimSpace(name) == "" {
        return ErrBeatNameRequired
    }
    if strings.TrimSpace(description) == "" {
        return ErrBeatDescriptionRequired
    }
    if strings.TrimSpace(playerDescription) == "" {
        return ErrBeatPlayerDescriptionRequired
    }
    b.name              = name
    b.description       = description
    b.playerDescription = playerDescription
    return nil
}
```

Event handler responsibilities for BeatContentUpdated:
```
Neo4jHandler      →  updates node properties
GCSHandler        →  appends audit event
FirestoreHandler  →  Beat already saved by AggregateStore
                     no separate action needed
```

---

## Beat Promotion

A GM promotes a campaign-specific beat to Master Narrative:

```
EntityUpdated {
    entity_id:  beat_id
    field:      "scope"
    old_value:  "campaign:{campaign_id}"
    new_value:  "master:{game_id}"
}
```

```
FirestoreHandler  →  updates Beat.scope field
Neo4jHandler      →  moves node into master graph
                     other Campaigns can now discover it
GCSHandler        →  appends event permanently
```

---

## Prerequisite Enforcement on Command Side

CampaignNarrative.DiscoverBeat() checks prerequisites locally.
Neo4j is never queried during command handling.

prerequisiteSets is `[][]BeatID`:
```
outer slice  →  OR  (any complete set satisfies)
inner slice  →  AND (all beats in set required)
```

```go
func (cn *CampaignNarrative) DiscoverBeat(beat Beat) error {
    if !cn.prerequisitesMet(beat) {
        return ErrPrerequisiteNotMet{BeatID: beat.id}
    }
    cn.discoveredBeatIDs = append(cn.discoveredBeatIDs, beat.id)
    return nil
}

func (cn *CampaignNarrative) prerequisitesMet(beat Beat) bool {
    for _, set := range beat.prerequisiteSets {
        if cn.allDiscovered(set) {
            return true
        }
    }
    return len(beat.prerequisiteSets) == 0
}

func (cn *CampaignNarrative) allDiscovered(beatIDs []BeatID) bool {
    for _, required := range beatIDs {
        if !cn.hasDiscovered(required) {
            return false
        }
    }
    return true
}
```

---

## DAG Cycle Prevention

### Two Distinct Operations

```
World building (GM, prep time):
  AddPrerequisite command
  Cycle detection via repository chain load
  Latency acceptable — GM is not in combat

Play time (real time):
  DiscoverBeat command
  O(1) check against discoveredBeatIDs[]
  No traversal — append and emit
```

At Bigtable scale a 10-level chain costs ~1-2ms.
Only occurs during GM world building. Not a concern at any scale.

### Repository Loads Chain, Domain Validates

```go
// roots/narrative/port/beat_repository.go
type BeatRepository interface {
    FindByID(ctx context.Context, id BeatID) (*Beat, error)
    LoadPrerequisiteChain(ctx context.Context, id BeatID) ([]Beat, error)
    Save(ctx context.Context, beat *Beat) error
}
```

```go
// roots/narrative/entity/beat.go
func (b *Beat) WouldCreateCycle(
    targetID BeatID,
    chain []Beat,
) error {
    for _, ancestor := range chain {
        if ancestor.id == targetID {
            return ErrCycleDetected{BeatID: b.id}
        }
    }
    return nil
}
```

```go
// Handler orchestrates only — no domain logic here
func (h *AddPrerequisiteHandler) Handle(
    ctx context.Context,
    cmd AddPrerequisiteCommand,
) error {
    beat, err := h.repo.FindByID(ctx, cmd.BeatID)
    if err != nil {
        return err
    }
    chain, err := h.repo.LoadPrerequisiteChain(ctx, cmd.PrerequisiteID)
    if err != nil {
        return err
    }
    if err := beat.WouldCreateCycle(cmd.BeatID, chain); err != nil {
        return err
    }
    beat.AddPrerequisite(cmd.PrerequisiteID)
    return h.repo.Save(ctx, beat)
}
```

Three responsibilities. None overlap:
```
Repository  →  traverses the chain
Domain      →  validates acyclicity
Handler     →  orchestrates only
```

### Defense in Depth

```cypher
CREATE CONSTRAINT prerequisite_acyclic
FOR ()-[r:PREREQUISITE_OF]-()
REQUIRE NOT (startNode(r))-[:PREREQUISITE_OF*]->(startNode(r))
```

Should never fire. Command side catches cycles first.

---

## Neo4j Read Side

### Available Beats Query
```cypher
MATCH (c:Campaign {id: $campaignId})-[:DISCOVERED]->(discovered:Beat)
MATCH (candidate:Beat)
WHERE NOT (c)-[:DISCOVERED]->(candidate)
AND EXISTS {
    MATCH (candidate)<-[:UNLOCKED_BY]-(ps:PrerequisiteSet)
          -[:REQUIRES]->(req:Beat)
    WHERE ALL(r IN collect(req)
              WHERE (c)-[:DISCOVERED]->(r))
}
RETURN candidate
```

### Multiple Paths to Same Beat
```cypher
(:Beat {name: "Road to Freeport"})
  <-[:UNLOCKED_BY]-(:PrerequisiteSet {id: "path_a"})
    -[:REQUIRES]->(:Beat {name: "Ring Found"})
    -[:REQUIRES]->(:Beat {name: "Inquisition Weakened"})

  <-[:UNLOCKED_BY]-(:PrerequisiteSet {id: "path_b"})
    -[:REQUIRES]->(:Beat {name: "Ring Origin Understood"})
```

---

## Player App Information Boundary

```
Discovered beats:   fully visible — playerDescription fields
Available beats:    title only — "There is something here..."
Locked beats:       hidden entirely
Campaign beats:     visible only to this Campaign's players
```

GM description fields (Beat.description) never leave the server.
Information boundary enforced at the API layer on Neo4j queries.

---

## Events

All six canonical events are sufficient. No new types required.

```
EntityCreated  { entity_type: "beat"|"act"|"lore"|"secret" }
EntityUpdated  { field: "scope" }           // promotion
EntityUpdated  { field: "content" }         // content update
EntityUpdated  { field: "status" }          // status change
EntityLinked   { relationship: "discovered" }
EntityLinked   { relationship: "prerequisite_of" }
EntityRevealed { entity_id: beat_id, revealed_to: [player_ids] }
```

### Event Handler Responsibilities

```
EntityCreated { entity_type: "beat" }:
  AggregateStore  →  saves full Beat struct
                     structural + content fields
  Neo4jHandler    →  creates node with all properties
                     creates relationship edges
  GCSHandler      →  appends full event ndjson

EntityUpdated { field: "content" }:
  AggregateStore  →  Beat already saved by command handler
  Neo4jHandler    →  updates node content properties
  GCSHandler      →  appends audit event

EntityUpdated { field: "scope" }:
  AggregateStore  →  updates Beat.scope
  Neo4jHandler    →  moves node to master graph
  GCSHandler      →  appends event

EntityLinked { relationship: "discovered" }:
  AggregateStore  →  appends BeatID to
                      CampaignNarrative.discoveredBeatIDs
  Neo4jHandler    →  creates DISCOVERED edge
  GCSHandler      →  appends event
```

---

## Scaling Transparency

Command side uses AggregateStore port (ADR-014).
Phase 1: Firestore holds full aggregate state.
Phase 3: Bigtable snapshot + event replay reconstructs same state.
Command handler code is identical in both phases.

---

## Consequences

- Firestore holds full aggregate state including content fields
- Neo4j is the read source for all client queries
- No contradictions between write side and read side storage
- Beat is genuinely shared — not owned by either narrative root
- Narrative aggregates have no state machine — documented intentional departure
- Game aggregate refactor required — breaking change on AddNarrativeElement
- CampaignNarrative auto-created via EventBus on CampaignCreated
- Cycle detection owned by domain via repository-loaded chain
- Neo4j constraint is defense in depth only

## Alternatives Considered

**Firestore structural fields only, Neo4j content only** — rejected.
Produces the contradiction that necessitated this ADR. Aggregate cannot
validate content it does not hold. Content updates would bypass the
command side entirely, breaking the events-from-aggregates rule.

**Single store for everything** — rejected. See ADR-010.

**AI-generated narrative beats** — explicitly deferred. Out of scope.
Architecture supports it without structural changes.