# ADR-015 Amendment 001: Narrative Architecture Gaps Resolved

## Status
Superseded by ADR-016

## Date
2026-05-10

## Amends
ADR-015: Narrative as a Directed Acyclic Graph

## Context
During initial implementation planning, eight gaps were identified in
ADR-015 that required explicit decisions before the Narrative aggregate
could be built. This amendment resolves all eight gaps. Three gaps
required deeper discussion and produced decisions that refine or correct
the original ADR-015 text where noted.

---

## Gap 1 — Missing Typed IDs

NarrativeID is retired. It was a placeholder that did not map to either
narrative aggregate specifically.

Eight typed IDs replace it, all following the GrimoireID embedding
pattern (ADR-007):

```go
// shared/identity/ids.go additions
type MasterNarrativeID   struct{ GrimoireID }
type CampaignNarrativeID struct{ GrimoireID }
type BeatID              struct{ GrimoireID }
type ActID               struct{ GrimoireID }
type SecretID            struct{ GrimoireID }
type LoreID              struct{ GrimoireID }
type DecisionID          struct{ GrimoireID }
type RevelationID        struct{ GrimoireID }
```

NarrativeID is removed from ids.go entirely. Any code referencing
NarrativeID is a breaking change and must be updated.

---

## Gap 2 — No Lifecycle / State Machine for Narrative Aggregates

MasterNarrative and CampaignNarrative have no state machine.
This is an explicit and intentional departure from the pattern used
by Game, Campaign, and Session.

Both aggregates are **purely additive event journals**:
- Beats are added, never removed
- Decisions are recorded, never reversed
- Revelations are appended, never retracted

Their lifecycle is implicitly tied to their parent:
```
MasterNarrative   →  follows Game lifecycle
                     exists when Game exists
                     archived when Game is archived

CampaignNarrative →  follows Campaign lifecycle
                     exists when Campaign exists
                     complete when Campaign is complete
```

No state interfaces. No transition methods. No status field.
Document this explicitly with a comment on each struct explaining
the intentional departure from the aggregate state machine pattern.

---

## Gap 3 — Beat Ownership and Aggregate Boundary

Beat lives in `roots/narrative/entity/`.

Both MasterNarrative and CampaignNarrative are in the same narrative
subsystem. No cross-root package import occurs.

Beat is created by a standalone command — not by MasterNarrative directly:

```
CreateMasterBeat command   →  scope: master, gameID
CreateCampaignBeat command →  scope: campaign, campaignID
```

MasterNarrative.AddBeat(id BeatID) records the reference only.
Beat creation is its own command with its own handler.

---

## Gap 4 — Game.AddNarrativeElement() Reconciliation

**Breaking change. Acknowledged and accepted.**

Game.AddNarrativeElement() is retired. The following files require
updates: states.go, transitions.go, handle.go, reconstitute.go,
and all associated tests.

The trigger for Game New → Draft is MasterNarrativeCreated,
delivered via the EventBus (ADR-011):

```
CreateMasterNarrative command dispatched
        ↓
MasterNarrative created
        ↓
MasterNarrativeCreated event emitted
        ↓
GameStatusHandler.Handle(MasterNarrativeCreated)
        ↓
game.OnNarrativeCreated(id MasterNarrativeID)
        ↓
Game transitions New → Draft
        ↓
Game stores MasterNarrativeID as reference
```

Game.OnNarrativeCreated(MasterNarrativeID) replaces
Game.AddNarrativeElement(). The transition is explicit and typed.
NarrativeID removed from identity/ids.go as part of Gap 1.

---

## Gap 5 — CampaignNarrative ↔ Campaign Relationship

CampaignNarrative is created automatically when a Campaign is created,
via the EventBus (ADR-011):

```
CampaignCreated event emitted
        ↓
CampaignNarrativeHandler.Handle(CampaignCreated)
        ↓
CreateCampaignNarrative command dispatched
        ↓
CampaignNarrative created with campaignID + gameID
        ↓
CampaignNarrativeCreated event emitted
```

The reference is one-way:
```
CampaignNarrative holds campaignID   ←  knows its parent
Campaign does NOT hold               ←  no reverse reference
  CampaignNarrativeID                    on the aggregate

Neo4j holds the reverse lookup:
  (:Campaign)-[:HAS_NARRATIVE]->(:CampaignNarrative)
```

---

## Gap 6 — Beat Scope and the Pointer Problem

*CampaignID (nil pointer) is replaced with a BeatScope value type
and a zero-value CampaignID check. Consistent with the typed-everything
approach throughout the domain (ADR-006).

```go
type BeatScope string

const (
    BeatScopeMaster   BeatScope = "master"
    BeatScopeCampaign BeatScope = "campaign"
)

type Beat struct {
    // structural fields
    id               BeatID
    scope            BeatScope
    gameID           GameID
    campaignID       CampaignID    // IsZero() when scope is master
    beatType         BeatType
    prerequisiteSets [][]BeatID
    revealsSecretIDs []SecretID
    status           BeatStatus

    // content fields — see Gap 7
    name              string
    description       string
    playerDescription string
}
```

`campaignID.IsZero()` replaces nil check everywhere.

---

## Gap 7 — Content Updates

**Correction to ADR-015.**

ADR-015 stated "content lives in Neo4j only." This requires clarification:

```
Neo4j is the READ SOURCE for content   ← unchanged
Aggregate validates content            ← new decision
Firestore holds content for validation ← new decision
Neo4j is the write projection          ← unchanged
```

Content updates go through the command side. The Beat aggregate
validates content fields and emits an event. The Neo4j handler
updates the node properties as a projection of that event.

**Flow:**
```
UpdateBeatContent command
        ↓
UpdateBeatContentHandler
  loads Beat from AggregateStore
        ↓
beat.UpdateContent(name, description, playerDescription)
  validates:  name not empty
              description not empty
              playerDescription not empty
  mutates:    b.name, b.description, b.playerDescription
  returns:    error or nil
        ↓
Beat saved to AggregateStore
        ↓
BeatContentUpdated event emitted
        ↓ (EventBus routes to all subscribers)
  Neo4jHandler      →  updates node properties
  GCSHandler        →  appends audit event
  FirestoreHandler  →  ignores
```

**Beat.UpdateContent():**
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

The event owner is the aggregate. The pattern is fully consistent.
No direct Neo4j mutations from the API layer.

---

## Gap 8 — DAG Cycle Prevention

### Two Distinct Operations

Cycle detection only applies during world building. It never applies
during play.

```
World building (GM, prep time):
  AddPrerequisite command
  Cycle detection via repository chain load
  Latency acceptable — GM is not in combat

Play time (real time):
  DiscoverBeat command
  O(1) check against discoveredBeatIDs[]
  No traversal whatsoever
  Append and emit — microseconds
```

At Bigtable scale each Beat load is a microsecond range scan.
A 10-level deep chain costs ~1-2ms total. Only occurs during
GM world building. Latency is not a concern at any scale.

### Repository Loads the Chain, Domain Validates It

**Correction to original amendment draft.**

The handler does not perform cycle detection. The repository loads
the full prerequisite chain. The Beat aggregate validates acyclicity.
Domain logic stays in the domain.

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
// AddPrerequisiteHandler — orchestrates only
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

Handler orchestrates. Repository traverses. Domain validates.
Three clear responsibilities. None overlap.

### Defense in Depth

Neo4j structural constraint catches anything that slips through:
```cypher
CREATE CONSTRAINT prerequisite_acyclic
FOR ()-[r:PREREQUISITE_OF]-()
REQUIRE NOT (startNode(r))-[:PREREQUISITE_OF*]->(startNode(r))
```

Should never fire in practice. Command side catches cycles first.

---

## Summary of All Decisions

```
Gap 1:  NarrativeID retired
        8 typed IDs replace it

Gap 2:  Narrative aggregates have no state machine
        purely additive event journals
        lifecycle follows parent implicitly

Gap 3:  Beat in roots/narrative/entity/
        created by standalone command
        MasterNarrative records reference only

Gap 4:  Breaking change accepted
        Game.AddNarrativeElement() retired
        Game.OnNarrativeCreated(MasterNarrativeID)
        triggered by MasterNarrativeCreated via EventBus

Gap 5:  CampaignNarrative auto-created via EventBus
        on CampaignCreated
        one-way reference — Campaign does not hold
        CampaignNarrativeID

Gap 6:  BeatScope value type
        campaignID.IsZero() replaces *CampaignID nil pointer

Gap 7:  Content updates go through command side
        Beat aggregate validates content fields
        Neo4j handler updates as projection
        "Content lives in Neo4j" means Neo4j is
        the read source — not that aggregate never sees it

Gap 8:  Repository loads prerequisite chain
        Beat aggregate validates acyclicity
        Handler orchestrates only
        Neo4j constraint is defense in depth
```