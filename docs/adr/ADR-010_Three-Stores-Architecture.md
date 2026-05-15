# ADR-010: Three Store Architecture — Firestore, Neo4j, GCS

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire requires persistence for three distinct concerns:

1. **Write side** — loading aggregates fast for command handling and
   invariant enforcement
2. **Read side** — answering rich graph queries for GM and player views
3. **Event log** — immutable, append-only source of truth for all
   state changes

The temptation was to use Neo4j for everything — it holds the graph and
could theoretically store aggregate state. This was evaluated and rejected.
A single store cannot serve all three concerns efficiently.

## Decision
Three stores. Three jobs. No overlap.

```
Firestore   →  aggregate state store (write side)
               one document per aggregate root
               loaded by ID for command handling
               updated atomically after every command

Neo4j       →  read model (read side)
               built from events
               answers graph queries
               never used for command handling

GCS         →  event log (source of truth)
               immutable append-only ndjson
               one file per entity per month
               can rebuild Neo4j from GCS if needed
```

## Firestore as Aggregate Store

Aggregates are documents naturally. Game, Campaign, Session, Character,
Location, Narrative, Faction each map to a Firestore document.

```
Load aggregate:   single document read  O(1)
Save aggregate:   single document write O(1)
Optimistic lock:  Firestore transactions
Cold start:       one round trip to load aggregate
```

Aggregate document size is trivially small. A Game with 50 campaigns
is under 2KB. Firestore's 1MB document limit is never a concern.

## Neo4j as Read Model

Neo4j answers graph queries the write side cannot:

```
"Everything connected to Korvan"
"What does the party currently know"
"Full session timeline for Campaign X"
"All factions and their allegiances"
```

Neo4j is populated by event handlers subscribing to the EventBus.
Every emitted event mutates the graph. The read model is eventually
consistent with the write side.

Neo4j is never written to directly from command handlers.

## GCS as Event Log

GCS holds the canonical, immutable event log as ndjson:

```
/events/{owner_id}/{aggregate_type}/{aggregate_id}/{year}/{month}/events.ndjson
```

GCS is the source of truth. If Neo4j is lost it can be rebuilt by
replaying all events from GCS. If Firestore aggregate state is lost
it can be reconstructed from GCS event replay.

## Consequences
- Each store does exactly one job
- No store is used for a purpose it is not optimised for
- Neo4j is eventually consistent — read model lags write side by one
  event handler execution
- Three stores to operate and monitor in production
- GCS is the disaster recovery source for both other stores

## Alternatives Considered
**Neo4j for everything** — rejected. Neo4j is optimised for relationship
traversal, not key-value aggregate lookup. Using it as an aggregate store
means using a graph database as a key-value store — wrong tool for the
write side problem.

**Firestore for everything** — rejected. Firestore cannot answer graph
queries efficiently. Party knowledge, entity relationships, and session
timelines require graph traversal.

**Single event log only, rebuild on demand** — rejected. Rebuilding
aggregate state from full event replay on every command is expensive
and adds latency to every write operation.