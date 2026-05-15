# Grimoire

An event-driven narrative state machine for tabletop RPGs.

Grimoire gives GMs a single source of truth for their world — story beats, lore, factions, locations, and characters — while giving players a curated view of only what they've discovered. Multiple capture surfaces (Obsidian, Foundry VTT, web UI) feed into a canonical event log. A Neo4j graph powers the world model. The GM controls the information boundary.

## How It Works

- **Six canonical events** describe every state change: `EntityCreated`, `EntityUpdated`, `EntityLinked`, `EntityRevealed`, `SessionStarted`, `SessionEnded`
- **Aggregate roots** (Game, Campaign, Character, Location, Narrative, Faction) enforce domain invariants on the write side
- **Three stores** serve three purposes: Firestore for aggregate state, Neo4j for graph queries, GCS for the immutable event log
- **The Narrative DAG** models story beats as a directed acyclic graph with prerequisite relationships — campaigns discover beats along different paths through the same master story

## Project Structure

Four Go modules in a workspace with a strict inward dependency rule:

```
grimoire-domain/           Pure domain logic, zero infrastructure imports
grimoire-infrastructure/   Adapters: Firestore, Neo4j, GCS, Pub/Sub
grimoire-testing/          Test doubles and fixtures (never in production builds)
grimoire-api/              Composition root, GraphQL, HTTP
```

Dependencies point inward only:
- `grimoire-domain` imports nothing internal (external: `uuid` only)
- `grimoire-infrastructure` imports `grimoire-domain`
- `grimoire-testing` imports `grimoire-domain`
- `grimoire-api` imports `grimoire-domain` + `grimoire-infrastructure`

## Requirements

- **Go 1.25.2+** — this project uses a Go workspace (`go.work`)
- **Neo4j** — graph database for the read model
- **Google Cloud SDK** — for Firestore, GCS, and Pub/Sub (local emulators work for development)

## Development

```bash
# Run all tests across the workspace
go test ./...

# Run tests for a specific module
cd grimoire-domain && go test ./...

# Run vet
go vet ./...
```

## Architecture Decision Records

Design decisions are documented as ADRs in [`grimoire-domain/docs/adr/`](docs/adr/).

| ADR | Title |
|-----|-------|
| [ADR-001](docs/adr/ADR-001_Game-Holds-Campaign-Ids.md) | Game Holds CampaignIDs Not SessionIDs |
| [ADR-002](docs/adr/ADR-002_State-pattern-for-aggregates.md) | State Pattern for Aggregate Lifecycles |
| [ADR-003](docs/adr/ADR-003_CQRS-split-read-write.md) | CQRS Split — Aggregates Write, Neo4j Reads |
| [ADR-004](docs/adr/ADR-004_Six-cononical-events.md) | Six Canonical Event Types |
| [ADR-005](docs/adr/ADR-005_NDJSON-over-parquet.md) | ndjson Over Parquet for Event Storage |
| [ADR-006](docs/adr/ADR-006_Typed-UUIDS-for-identity.md) | Typed UUIDs for All Identity |
| [ADR-007](docs/adr/ADR-007_GrimoireID-base-struct.md) | GrimoireID Base Struct via Embedding |
| [ADR-008](docs/adr/ADR-008_GraphQL-API-LAyer.md) | GraphQL API Layer |
| [ADR-009](docs/adr/ADR-009_Bidirectional-Event-Mapping.md) | Bidirectional Event Mapping |
| [ADR-010](docs/adr/ADR-010_Three-Stores-Architecture.md) | Three Stores Architecture |
| [ADR-011](docs/adr/ADR-011_Domain-Event-Chaining.md) | Domain Event Chaining |
| [ADR-012](docs/adr/ADR-012_Event-Sequencing.md) | Event Sequencing |
| [ADR-013](docs/adr/ADR-013_Four-Module-Workspace.md) | Four Module Workspace |
| [ADR-014](docs/adr/ADR-014_Scaling_Path_BigTable.md) | Scaling Path — Firestore to Bigtable |
| [ADR-015](docs/adr/ADR-015_Narrative-DAG.md) | ~~Narrative as DAG~~ (superseded by ADR-016) |
| [ADR-016](docs/adr/ADR-016_Narrative-Aggregate-Architecture.md) | Narrative Aggregate Architecture — Authoritative Record |
| [ADR-017](docs/adr/ADR-017_Faction-Aggregate.md) | Faction Aggregate Architecture |
| [ADR-018](docs/adr/ADR-018_Location-Aggregate.md) | Location Aggregate Architecture |
