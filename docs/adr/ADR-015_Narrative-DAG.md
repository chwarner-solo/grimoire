# ADR-015: Narrative as a Directed Acyclic Graph — Neo4j for Relationships and Content, Firestore/Bigtable for Structure

## Status
Superseded by ADR-016

## Date
2026-05-10

## Context
The Narrative aggregate is fundamentally a graph problem. Story beats have
prerequisite relationships — Beat B cannot be discovered until Beat A is
discovered. Multiple prerequisite paths can lead to the same beat. Campaigns
take different paths through the same master story.

This is a Directed Acyclic Graph (DAG) where:
- Nodes are NarrativeBeats
- Edges are PREREQUISITE_OF relationships
- Traversal answers "what beats are available to this campaign right now?"

The architecture must support:
1. Efficient command-side invariant enforcement (prerequisites met?)
2. Efficient read-side graph traversal (available beats, story map)
3. Two narrative layers — Master (Game-owned) and Campaign (Campaign-owned)
4. Campaign-specific beats with no Master Narrative counterpart
5. Promotion of Campaign beats to Master Narrative
6. Scaling from Firestore to Bigtable without domain changes (ADR-014)

## Decision

### Foundational Principle: Content Lives in Neo4j. Structure Lives in Firestore.

**Firestore/Bigtable holds structural fields only:**
```
Minimum data required to enforce invariants.
IDs, relationships between IDs, status, scope, type.
No human-readable content fields.
No name. No description. No player-facing text.
```

**Neo4j holds content fields and relationships:**
```
All human-readable content as node properties.
name, description, playerDescription, lore text.
Relationship edges between nodes.
Everything a human reads comes from Neo4j.
```

**The event carries content to both stores on creation:**
```
EntityCreated event payload contains all fields.
Firestore handler writes structural fields only.
Neo4j handler writes all fields including content.
GCS preserves the full event permanently.
Content is always recoverable from GCS if Neo4j is rebuilt.
```

---

### Beat as a Shared Entity

Beat is shared between MasterNarrative and CampaignNarrative.
Neither narrative root owns Beat content. Both hold Beat IDs as references.
Beat itself knows its scope — master or campaign.

```
MasterNarrative    →  []BeatID  (references only)
CampaignNarrative  →  []BeatID  (references only)
Beat               →  knows gameID, scope, campaignID?
```

---

### Two Narrative Layers

**MasterNarrative** — owned by Game:
```
The GM's canonical world truth.
Exists independently of any Campaign.
All Campaigns reference its beats.
Beats marked required or optional.
```

**CampaignNarrative** — owned by Campaign:
```
This table's path through the master story.
Tracks which beats have been discovered.
Holds campaign-specific beat IDs.
Tracks decisions and revelations unique to this table.
```

---

### Narrative Beat Types
```
required           →  every Campaign must hit this beat
optional           →  some Campaigns will find this
campaign-specific  →  GM improvised during play
                      may be promoted to master
```

---

### Beat Promotion
A GM promotes a campaign-specific beat to Master Narrative:
```
EntityUpdated {
    entity_id:  beat_id
    field:      "scope"
    old_value:  "campaign:{campaign_id}"
    new_value:  "master:{game_id}"
}
```
Neo4j moves the node into the master graph.
Firestore updates scope field only.
Other Campaigns can now discover it.

---

### Store Responsibilities

**Neo4j — content and relationships:**
```cypher
(:Beat {
    id:                 "beat_korvan_encounter",
    name:               "The Inquisitor Revealed",
    description:        "Korvan steps from the shadows...",
    playerDescription:  "A figure in Inquisition robes...",
    beatType:           "required",
    scope:              "master"
})

(:Beat {id: "beat_korvan_encounter"})
    -[:PREREQUISITE_OF]->
(:Beat {id: "beat_ring_found"})

(:Campaign {id: "campaign_table_a"})
    -[:DISCOVERED {session: 12}]->
(:Beat {id: "beat_korvan_encounter"})
```

**Firestore/Bigtable — structural fields only:**
```go
// No content fields on any Firestore struct

type Beat struct {
    id               BeatID
    beatType         BeatType        // required|optional|campaign
    scope            BeatScope       // master|campaign
    gameID           GameID
    campaignID       *CampaignID     // nil if master beat
    prerequisiteSets [][]BeatID      // OR across sets, AND within
    revealsSecretIDs []SecretID
    status           BeatStatus
}

type CampaignNarrative struct {
    id                CampaignNarrativeID
    campaignID        CampaignID
    gameID            GameID
    discoveredBeatIDs []BeatID
    campaignBeatIDs   []BeatID
    decisionIDs       []DecisionID
    revelationIDs     []RevelationID
}

type MasterNarrative struct {
    id        MasterNarrativeID
    gameID    GameID
    actIDs    []ActID
    beatIDs   []BeatID
    secretIDs []SecretID
    loreIDs   []LoreID
}
```

**GCS/Bigtable event log — full event preserved:**
```
beat_korvan#event#0000000000000001
EntityCreated {
    entity_type: "beat"
    payload: {
        name:              "The Inquisitor Revealed"
        description:       "Korvan steps from shadows..."
        playerDescription: "A figure in robes..."
        beatType:          "required"
        scope:             "master"
        gameID:            "game_ashes_chains"
        prerequisiteSets:  [["beat_enter_blackburrow"]]
    }
}
```

Full content preserved in the event log permanently.
Neo4j can be rebuilt entirely from GCS. Content is never lost.

---

### Prerequisite Enforcement on Command Side

CampaignNarrative holds `[]DiscoveredBeatIDs`.
Beat holds `[][]BeatID` prerequisite sets.
Both loaded from Firestore. Invariant checked locally.
Neo4j never queried during command handling.

```go
func (cn *CampaignNarrative) DiscoverBeat(beat Beat) error {
    if !cn.prerequisitesMet(beat) {
        return ErrPrerequisiteNotMet{BeatID: beat.id}
    }
    cn.discoveredBeatIDs = append(cn.discoveredBeatIDs, beat.id)
    return nil
}

func (cn *CampaignNarrative) prerequisitesMet(beat Beat) bool {
    // OR across sets — any complete set satisfies
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

### Multiple Paths to the Same Beat

Beat.PrerequisiteSets is `[][]BeatID`:
```
outer slice  →  OR  (any complete set satisfies)
inner slice  →  AND (all beats in set required)
```

Neo4j models this as PrerequisiteSet nodes:
```cypher
(:Beat {name: "Road to Freeport"})
  <-[:UNLOCKED_BY]-(:PrerequisiteSet {id: "path_a"})
    -[:REQUIRES]->(:Beat {name: "Ring Found"})
    -[:REQUIRES]->(:Beat {name: "Inquisition Weakened"})

  <-[:UNLOCKED_BY]-(:PrerequisiteSet {id: "path_b"})
    -[:REQUIRES]->(:Beat {name: "Ring Origin Understood"})
```

---

### Neo4j Read Query — Available Beats
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

---

### Player App — Information Boundary on the DAG
```
Discovered beats:  fully visible — playerDescription fields
Available beats:   title only — "There is something here..."
Locked beats:      hidden entirely
Campaign beats:    visible only to this Campaign's players
```

GM description fields never leave the server under any circumstances.

---

## Event Handler Responsibilities

```
EntityCreated { entity_type: "beat" }:
  FirestoreHandler  →  writes structural fields only
                        id, beatType, scope, gameID,
                        prerequisiteSets, status
  Neo4jHandler      →  writes ALL fields as node properties
                        plus relationship edges
  GCSHandler        →  appends full event ndjson

EntityLinked { relationship: "discovered" }:
  FirestoreHandler  →  appends BeatID to
                        CampaignNarrative.discoveredBeatIDs
  Neo4jHandler      →  creates DISCOVERED edge
                        Campaign → Beat
  GCSHandler        →  appends full event ndjson

EntityUpdated { field: "scope" }:
  FirestoreHandler  →  updates scope field on Beat
  Neo4jHandler      →  moves node to master graph
  GCSHandler        →  appends full event ndjson
```

---

## Scaling Transparency

Command side uses AggregateStore port (ADR-014).
Phase 1: Firestore serves structural aggregate state.
Phase 3: Bigtable snapshot + event replay serves same state.
Command handler code identical in both phases.
Content fields are irrelevant to the command side in both phases.

---

## Consequences
- Command side is lean — structural fields only, fast reads
- Content changes are Neo4j operations not aggregate commands
- Neo4j is the single source of content truth for read side
- GCS event log preserves full content for disaster recovery
- Neo4j rebuild from GCS restores all content and relationships
- Beat is genuinely shared — not owned by either narrative root
- Player app information boundary enforced at API layer on Neo4j queries

## Alternatives Considered
**Content in Firestore, relationships in Neo4j** — rejected. Splits
content across two stores. Command side loads content it never uses.

**Content in Firestore only, Neo4j holds IDs only** — rejected.
Read side must join across Firestore documents to assemble content.
Loses Neo4j's graph traversal on content properties.

**Single store for everything** — rejected. No single store handles
both fast aggregate key-value reads and efficient graph traversal.
See ADR-010.

**AI-generated narrative beats** — explicitly deferred. Out of scope.
Architecture supports it as a future Beat creation source without
structural changes.