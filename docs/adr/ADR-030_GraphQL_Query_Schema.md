# ADR-030: GraphQL Query Schema

## Status
Accepted

## Date
2026-05-18

---

## Context

ADR-008 established GraphQL as the API layer.
ADR-028 defined all 46 mutations covering the GM write surface.
ADR-028 explicitly deferred queries pending UI design (ADR-029).

ADR-029 specified the GM PWA — React, Apollo Client, GraphQL Codegen,
two workflow contexts (World Building, Session Operation).

UI-001 and UI-002 defined 20 screens across both contexts. Every screen
calls out the queries it requires. This ADR translates those requirements
into a complete, implemented query schema.

### Architectural constraints carried forward

**All queries read from Neo4j — never from Firestore (ADR-003).**
Aggregates are write-side only. No aggregate is loaded to answer a query.
The Neo4j graph is the read model, populated by event handlers.

**GM-only scope (ADR-029).** No player visibility filtering in this
schema. The information boundary (playerVisible fields, playerDescription
vs description) is enforced at the Neo4j query layer for the Player app
(future). The GM sees everything.

**Queries never return aggregate state.** Return types are read model
projections shaped for the screen that consumes them, not reflections
of the aggregate's internal structure.

**N+1 is prevented by query design, not dataloaders (for now).**
Neo4j traversals return full subgraphs in a single round trip. Each
screen issues one query, which fetches all required nodes and edges in
one Cypher statement. Dataloaders are deferred to when query complexity
warrants them.

---

## Decision

### Resolver Pattern

Every query resolver follows the same invariant as mutations — identity
is verified first, then the read model is queried:

```go
// grimoire-api/internal/resolver/query_resolver.go

func (r *queryResolver) Games(ctx context.Context) ([]*GameSummary, error) {
    callerID, err := r.auth.Identify(ctx, tokenFromContext(ctx))
    if err != nil {
        return nil, ErrUnauthenticated
    }
    return r.neo4j.FindGamesByGMID(ctx, callerID)
}
```

The `callerID` is always the filter for top-level queries — the GM
only sees their own games. Neo4j holds `gmID` on every Game node,
set when the `EntityCreated` event is projected.

### Query Port

Queries have their own port in `grimoire-api`, separate from the
interactor ports in `grimoire-domain`. Queries do not cross the domain
boundary — they talk directly to Neo4j via the query adapter.

```go
// grimoire-api/internal/query/ports.go

type GameQueryPort interface {
    FindByGMID(ctx context.Context, gmID string) ([]*GameSummary, error)
    FindByID(ctx context.Context, id string, gmID string) (*GameDetail, error)
}

type CampaignQueryPort interface {
    FindByGame(ctx context.Context, gameID string, gmID string) ([]*CampaignSummary, error)
    FindByID(ctx context.Context, id string, gmID string) (*CampaignDetail, error)
    FindSetupState(ctx context.Context, id string, gmID string) (*CampaignSetup, error)
    FindSessionState(ctx context.Context, id string, gmID string) (*CampaignSessionState, error)
}

// ... one port per aggregate area
```

Query ports live in `grimoire-api`, not `grimoire-domain`. Queries
are infrastructure concerns — they are shaped by client needs, not
domain invariants. The domain does not know queries exist.

---

## Schema

### Scalars (existing — ADR-028)

```graphql
scalar DateTime
scalar Date
```

### Shared Enums (existing — ADR-028)

```graphql
enum LocationType { WORLD REGION SETTLEMENT BUILDING SCENE }
enum BeatType     { REQUIRED OPTIONAL CAMPAIGN_SPECIFIC }
```

### New Status Enum

Used across all entity summaries. String status values from the domain
state machine map to this enum at the resolver layer.

```graphql
enum EntityStatus {
  NEW
  DRAFT
  FORMING
  ACTIVE
  IDLE
  COMPLETE
  ARCHIVED
  RETIRED
}
```

---

## Schema — Game Queries

```graphql
# grimoire-api/schema/query.graphql

type Query {

  """All games owned by the authenticated GM."""
  games: [GameSummary!]!

  """A single game by ID. Returns null if not found or not owned by caller."""
  game(id: ID!): GameDetail
}

"""Used in SCR-001 — Games Dashboard card."""
type GameSummary {
  id:             ID!
  name:           String!
  status:         EntityStatus!
  campaignCount:  Int!
  lastActivityAt: DateTime
}

"""Used in SCR-003 — Game Overview hub."""
type GameDetail {
  id:     ID!
  name:   String!
  status: EntityStatus!

  locationSummary: EntityCountSummary!
  factionSummary:  EntityCountSummary!
  npcSummary:      EntityCountSummary!

  campaigns: [CampaignSummary!]!
}

"""Reusable aggregate count by status tier."""
type EntityCountSummary {
  draft:    Int!
  active:   Int!
  idle:     Int!
  archived: Int!
}
```

---

## Schema — Campaign Queries

```graphql
"""All campaigns for a game."""
campaigns(gameId: ID!): [CampaignSummary!]!

"""A single campaign. Used for setup wizard and post-session."""
campaign(id: ID!, gameId: ID!): CampaignDetail

"""
Session-time state for the Party Tab.
Separate from campaign detail — includes currentLocation with connections.
"""
campaignSessionState(id: ID!, gameId: ID!): CampaignSessionState

"""
Campaign beats created during sessions (scope: campaign).
Used in SCR-020 Post-Session Review.
"""
campaignBeats(campaignId: ID!, gameId: ID!): [BeatSummary!]!
```

```graphql
type CampaignSummary {
  id:           ID!
  name:         String!
  status:       EntityStatus!
  characterCount: Int!
  sessionCount:   Int!
}

type CampaignDetail {
  id:           ID!
  name:         String!
  status:       EntityStatus!
  sessionCount: Int!
  characters:   [PlayerCharacterSummary!]!
}

"""Full session-time state. Powers SCR-016 Party Tab."""
type CampaignSessionState {
  id:     ID!
  name:   String!
  status: EntityStatus!
  sessionId: ID!

  currentLocation: LocationWithConnections

  characters: [PlayerCharacterSummary!]!
}

type LocationWithConnections {
  id:           ID!
  name:         String!
  locationType: LocationType!
  connections:  [LocationConnection!]!
}

type LocationConnection {
  toLocation: LocationSummary!
  direction:  ConnectionDirection!
}

enum ConnectionDirection {
  OUTBOUND   # A → B only
  INBOUND    # B → A only (reverse)
  BOTH       # A ↔ B
}
```

---

## Schema — Location Queries

```graphql
"""
All locations for a game as a flat list with parentId.
Client builds the tree. Used in SCR-004 Location tree.
"""
locations(gameId: ID!): [LocationSummary!]!

"""
Single location with scenes and connections.
Used in SCR-005 Location Detail panel.
"""
location(id: ID!, gameId: ID!): LocationDetail

"""
Active locations only. Used in MoveParty modal (SCR-016).
"""
activeLocations(gameId: ID!): [LocationSummary!]!
```

```graphql
type LocationSummary {
  id:           ID!
  name:         String!
  locationType: LocationType!
  status:       EntityStatus!
  parentId:     ID
  sceneCount:   Int!
  childCount:   Int!
}

type LocationDetail {
  id:           ID!
  name:         String!
  locationType: LocationType!
  status:       EntityStatus!
  parentLocation: LocationSummary

  scenes:      [SceneSummary!]!
  connections: [LocationConnection!]!
}

type SceneSummary {
  id:   ID!
  name: String!
}
```

---

## Schema — Narrative Queries

```graphql
"""
All master beats for a game.
Used in SCR-006 Narrative DAG view and list view.
"""
masterNarrative(gameId: ID!): MasterNarrativeView

"""
Single beat with full content and prerequisites.
Used in SCR-007 Beat Detail panel.
"""
beat(id: ID!, gameId: ID!): BeatDetail

"""
Beats available to a campaign given its current discoveredBeatIDs.
Neo4j prerequisite traversal — the only graph query in this schema.
Used in SCR-017 Narrative Tab (session).
"""
availableBeats(campaignNarrativeId: ID!, gameId: ID!): [BeatSummary!]!

"""
All beats discovered by a campaign, with session attribution.
Used in SCR-017 Narrative Tab discovered history.
"""
discoveredBeats(campaignNarrativeId: ID!, gameId: ID!): [DiscoveredBeat!]!
```

```graphql
type MasterNarrativeView {
  id:    ID!
  beats: [BeatSummary!]!
}

type BeatSummary {
  id:       ID!
  name:     String!
  beatType: BeatType!
  scope:    BeatScope!
  prerequisiteCount: Int!
}

enum BeatScope {
  MASTER
  CAMPAIGN
}

type BeatDetail {
  id:                ID!
  name:              String!
  beatType:          BeatType!
  scope:             BeatScope!
  description:       String!
  playerDescription: String!
  prerequisites:     [BeatSummary!]!
}

type DiscoveredBeat {
  id:                  ID!
  name:                String!
  beatType:            BeatType!
  discoveredInSession: Int!
}
```

> **Note on `availableBeats`:** This is the only query that triggers
> a DAG traversal in Neo4j. The Cypher query checks prerequisite sets
> against the campaign's `discoveredBeatIDs`. It runs on the read side
> only — the command-side prerequisite check (O(1) local slice
> comparison) is separate and is never called from a query resolver.
> See ADR-016.

---

## Schema — Faction Queries

```graphql
"""All factions for a game."""
factions(gameId: ID!): [FactionSummary!]!

"""Single faction with full detail."""
faction(id: ID!, gameId: ID!): FactionDetail
```

```graphql
type FactionSummary {
  id:               ID!
  name:             String!
  status:           EntityStatus!
  playerVisible:    Boolean!
  memberCount:      Int!
  standingLevelCount: Int!
  allies:           [FactionRef!]!
  enemies:          [FactionRef!]!
}

type FactionRef {
  id:   ID!
  name: String!
}

type FactionDetail {
  id:     ID!
  name:   String!
  status: EntityStatus!
  playerVisible: Boolean!

  members: [FactionMember!]!
  standingLevels: [StandingLevel!]!
  allies:  [FactionRef!]!
  enemies: [FactionRef!]!
}

type FactionMember {
  npc:  NPCRef!
  rank: String!
}

type NPCRef {
  id:   ID!
  name: String!
}

type StandingLevel {
  ordinal:   Int!
  name:      String!
  threshold: Int!
}
```

---

## Schema — Character Queries

```graphql
"""
All NPCs for a game. excludeArchived filters the list
for session screens where archived NPCs are irrelevant.
"""
npcs(gameId: ID!, excludeArchived: Boolean): [NPCSummary!]!

"""Single NPC with full content and faction memberships."""
npc(id: ID!, gameId: ID!): NPCDetail

"""All player characters for a game."""
playerCharacters(gameId: ID!): [PlayerCharacterSummary!]!
```

```graphql
type NPCSummary {
  id:            ID!
  name:          String!
  status:        EntityStatus!
  playerVisible: Boolean!
}

type NPCDetail {
  id:                ID!
  name:              String!
  status:            EntityStatus!
  playerVisible:     Boolean!
  description:       String!
  playerDescription: String!
  factionMemberships: [FactionMembership!]!
}

type FactionMembership {
  faction: FactionRef!
  rank:    String!
}

type PlayerCharacterSummary {
  id:            ID!
  name:          String!
  status:        EntityStatus!
  ownerPlayerId: String
}
```

---

## Schema — MacGuffin Queries

```graphql
"""All MacGuffins for a game."""
macguffins(gameId: ID!): [MacGuffinSummary!]!
```

```graphql
type MacGuffinSummary {
  id:            ID!
  name:          String!
  playerVisible: Boolean!
  possessor:     MacGuffinPossessor
}

"""
A MacGuffin can be held by an NPC, a PlayerCharacter, or a Location.
Union type — client uses inline fragments to handle each case.
"""
union MacGuffinPossessor = NPCRef | PlayerCharacterRef | LocationRef

type PlayerCharacterRef {
  id:   ID!
  name: String!
}

type LocationRef {
  id:   ID!
  name: String!
}
```

---

## Schema — Session Queries

```graphql
"""
Pre-session launch state. Powers SCR-015 Session Launcher.
"""
sessionLauncher(campaignId: ID!, gameId: ID!): SessionLauncherState

"""
End-of-session summary. Powers SCR-019 End Session screen.
Called after endSession mutation returns — reads from Neo4j
projection of events from the current session.
"""
sessionSummary(sessionId: ID!, gameId: ID!): SessionSummary
```

```graphql
type SessionLauncherState {
  campaignId:   ID!
  campaignName: String!
  status:       EntityStatus!
  sessionCount: Int!
  characters:   [PlayerCharacterSummary!]!
}

type SessionSummary {
  sessionId:              ID!
  beatsDiscoveredCount:   Int!
  entitiesRevealedCount:  Int!
  finalLocation:          LocationRef
}
```

---

## File Layout

All query schema in a single new file alongside the mutation schemas:

```
grimoire-api/
  schema/
    shared.graphql       MutationResult, scalars (existing)
    game.graphql         mutations (existing)
    campaign.graphql     mutations (existing)
    narrative.graphql    mutations (existing)
    faction.graphql      mutations (existing)
    location.graphql     mutations (existing)
    character.graphql    mutations (existing)
    reveal.graphql       mutations (existing)
    query.graphql        ← ALL queries (this ADR)

  internal/
    query/
      ports.go           query port interfaces
      models.go          Go structs matching GraphQL return types
    resolver/
      query_resolver.go  ← NEW: all query resolver implementations
      mutation_resolver.go  (existing)
      errors.go             (existing)
```

All queries in one file. This keeps the split between mutations
(commands) and queries (reads) explicit at the file level, matching
the CQRS split in the domain.

---

## Codegen Impact (`grimoire-pwa`)

Once `query.graphql` is added to `grimoire-api/schema/`, the codegen
configuration in `grimoire-pwa/codegen.ts` picks it up automatically —
it reads `../grimoire-api/schema/*.graphql`. No config change required.

Every `.graphql` file in `grimoire-pwa/src/graphql/queries/` that
references these types will produce fully typed hooks after the next
`npx graphql-codegen` run.

---

## Consequences

- All 20 screens in UI-001 and UI-002 have their queries defined.
  ADR-029's codegen configuration can now generate typed hooks for
  every screen from day one.
- `availableBeats` is the only query triggering a Neo4j DAG traversal.
  It is isolated, named, and documented. Its complexity is bounded
  by the size of the beat graph, not the number of campaigns.
- `MacGuffinPossessor` is a union type — the only union in the query
  schema. It accurately models the domain (a MacGuffin can be held by
  an NPC, PC, or Location) without requiring nullable fields for each
  possessor type.
- `ConnectionDirection` (OUTBOUND / INBOUND / BOTH) is a read-model
  convenience. The domain models only directed edges. Neo4j resolves
  bidirectionality at query time by checking both directions between
  two nodes and returning the combined result.
- Queries do not expose aggregate internals. `CampaignSessionState`
  is shaped for SCR-016 (Party Tab), not for the Campaign aggregate
  structure. A future screen needing different campaign fields gets
  a new query type — the aggregate does not change.
- The `excludeArchived` argument on `npcs` is the only query-time
  filter in the schema. All other filtering is implicit in the
  query's purpose (e.g. `activeLocations` only returns Active).
  This keeps the schema honest about what each query is for.

## Deferred

```
Player app queries         →  future ADR (Player app PWA)
                               playerVisible filtering enforced here
                               playerDescription only, never description
                               availableBeats filtered to player-visible only

Analytics queries          →  BigQuery, not Neo4j (ADR-003)
                               session timelines, beat completion rates

Pagination                 →  deferred; GM datasets are small
                               (one GM, one game at a time)
                               revisit at Player app scale

Subscriptions              →  deferred to Player app ADR
                               entityRevealed, sessionStarted
```

---

## Related

- ADR-003 — CQRS Split (queries from Neo4j only)
- ADR-008 — GraphQL API Layer (query resolver is a port adapter)
- ADR-016 — Narrative Aggregate (availableBeats DAG traversal)
- ADR-028 — GraphQL Mutation Schema
- ADR-029 — GM Web Application Architecture
- UI-001 — World Building Screens (query requirements source)
- UI-002 — Session Operation Screens (query requirements source)