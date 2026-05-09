# ADR-008: GraphQL as API Layer — Mutations as Commands, Queries as Read Model

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire has three clients with three distinct field needs on the same
entities:

- Obsidian plugin — narrative fields only
- Foundry VTT module — mechanical fields only
- Player app — revealed fields only, scoped by GM information boundary

REST was evaluated and rejected — it thinks in resources, Grimoire thinks
in relationships. The data is a graph. REST would cause over-fetching or
under-fetching and require multiple round trips for related entities.

JSON-RPC was evaluated as a write-side-only solution but is not a complete
API strategy.

## Decision
GraphQL is the API layer for Grimoire.

**Mutations map to CQRS commands (write side):**
```graphql
type Mutation {
    createGame(input: CreateGameInput!): NewGame!
    addNarrativeElement(gameId: ID!, name: String!): DraftGame!
    createCampaign(gameId: ID!, input: CreateCampaignInput!): NewCampaign!
    startSession(campaignId: ID!, date: Date!): InProgressSession!
    endSession(sessionId: ID!, notes: String): CompletedSession!
    revealEntity(entityId: ID!, revealedTo: [ID!]!): EntityRevealed!
}
```

**Queries map to the read model (Neo4j):**
```graphql
type Query {
    game(id: ID!): GameState!
    partyKnowledge(campaignId: ID!): PartyKnowledge!
    campaignState(campaignId: ID!): CampaignState!
    entityRelationships(entityId: ID!): EntityGraph!
    sessionTimeline(campaignId: ID!): [Session!]!
}
```

**Subscriptions handle real-time push to the player app:**
```graphql
type Subscription {
    entityRevealed(campaignId: ID!): EntityRevealed!
    sessionStarted(campaignId: ID!): InProgressSession!
}
```

**The state machine surfaces in the schema via unions:**
```graphql
union GameState = NewGame | DraftGame | ActiveGame | IdleGame | ArchivedGame
```

## Implementation
`gqlgen` (99designs/gqlgen) — schema-first code generation for Go.
Write the GraphQL schema. gqlgen generates typed resolver interfaces.
Resolvers are implemented as adapters in grimoire-api. The resolver
IS a port in Ports & Adapters terms.

## Mutation Flow
```
Mutation received
    → GraphQL resolver
    → Command dispatched
    → Aggregate loaded from repository
    → Invariants enforced
    → Event emitted
    → GCS written (ndjson)
    → Neo4j updated
    → Read model projection returned to caller
```

Mutations never return aggregate state. They return a Neo4j read model
projection of the resulting state shaped for the caller's needs.

## Reasoning
GraphQL's client-specified field selection solves the three-client problem
cleanly. Each client requests exactly the fields it needs. One schema
serves all surfaces without over or under-fetching.

Mutations as commands and queries as read model is a natural expression
of CQRS at the API boundary. The GraphQL schema makes the CQRS split
visible and explicit to API consumers.

GraphQL subscriptions replace the need for a separate WebSocket protocol
for real-time player app updates.

## Consequences
- One API serves all three surfaces with appropriate field scoping
- CQRS split is explicit in the schema
- State machine states surface as GraphQL union types
- gqlgen generates typed resolver interfaces — compiler catches mismatches
- N+1 query problem requires dataloader implementation for Neo4j queries
- Schema changes require regenerating gqlgen bindings

## Alternatives Considered
**REST** — rejected. Thinks in resources not relationships. Multiple round
trips for related entities. Poor fit for a graph data model.

**JSON-RPC** — rejected. No standard read side. Sparse tooling. Not a
complete API solution though natural for commands.

**REST for commands, GraphQL for queries** — rejected. Two API protocols
to maintain. GraphQL mutations serve the command role cleanly.