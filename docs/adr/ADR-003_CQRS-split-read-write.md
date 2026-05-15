# ADR-003: CQRS Split — Aggregates Write, Neo4j Reads

## Status
Accepted

## Date
2026-05-09

## Context
The traditional ORM pattern treats the domain model as both the write model
and the read model. This leads to aggregates accumulating fields to satisfy
query needs, navigation properties for reporting, and lazy loading hacks to
avoid N+1 problems. The Grimoire domain model was being pulled in this
direction during initial modeling — specifically around whether Game should
hold full Session objects to satisfy "get game with sessions" queries.

## Decision
Strict CQRS split:

**Write side:** Aggregate roots handle commands and enforce invariants only.
They hold the minimum data required to enforce those invariants. They are
never used to answer queries.

**Read side:** Neo4j answers all queries. BigQuery answers analytics.
The read model is shaped by query needs, not aggregate boundaries.

## Reasoning
Aggregate roots exist to enforce invariants — not to serve screens or
answer questions. The moment you add a field to an aggregate to satisfy
a query, you have coupled your write model to your read model. This
creates pressure to keep aggregates loaded with data they do not need
for invariant enforcement.

Neo4j builds its graph from events. Every event mutates the graph. The
graph can be queried in any shape needed without touching aggregates.
Query shape changes do not affect aggregate design.

## Consequences
- Aggregates are significantly simpler and smaller
- No navigation properties on aggregates
- No lazy loading concerns
- Query shape changes require no aggregate changes
- Two models to maintain (aggregate + graph projection)
- All writes go through aggregates; all reads go through Neo4j

## Alternatives Considered
**Single model (ORM style)** — rejected. Leads to bloated aggregates,
lazy loading, and coupling of write and read concerns. This was the
pattern being escaped.

**Aggregates for simple queries** — rejected. Any exception to the rule
creates pressure for more exceptions. The boundary must be absolute.