# Grimoire

An event-driven narrative state machine for tabletop RPGs.

Grimoire gives GMs a single source of truth for their world — story beats, lore, factions, locations, and characters — while giving players a curated view of only what they've discovered. Multiple capture surfaces (Obsidian, Foundry VTT, web UI) feed into a canonical event log. A Neo4j graph powers the world model. The GM controls the information boundary.

## How It Works

- **Six canonical events** describe every state change: `EntityCreated`, `EntityUpdated`, `EntityLinked`, `EntityRevealed`, `SessionStarted`, `SessionEnded`
- **Aggregate roots** (Game, Campaign, Character, Location, Narrative, Faction) enforce domain invariants on the write side
- **Three stores** serve three purposes: Firestore for aggregate state, Neo4j for graph queries, GCS for the immutable event log
- **The Narrative DAG** models story beats as a directed acyclic graph with prerequisite relationships — campaigns discover beats along different paths through the same master story

## Project Structure

Four Go modules and a PWA in a workspace with a strict inward dependency rule:

```
grimoire-domain/           Pure domain logic, zero infrastructure imports
grimoire-infrastructure/   Adapters: Firestore, Neo4j, GCS, Pub/Sub
grimoire-testing/          Test doubles and fixtures (never in production builds)
grimoire-api/              Composition root, GraphQL, HTTP
grimoire-pwa/              GM web application — React PWA (not a Go module)
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

## Workflows

Operational sequences showing how the system is used end-to-end from the GM's
perspective. Distinct from ADRs — workflows describe what commands are called
and in what order; ADRs record why design decisions were made.

| Workflow | Title |
|----------|-------|
| [WORKFLOW-001](docs/workflows/Workflow-001_world-building-session-prep.md) | World Building & Session Prep |
| [WORKFLOW-002](docs/workflows/Workflow-002_session-operation.md) | Session Operation |

## Architecture Decision Records

Design decisions are documented as ADRs in [`docs/adr/`](docs/adr/).

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
| [ADR-011-Amendment-001](docs/adr/ADR-011_Amendment-001_Event-Delivery_Guarantees.md) | Event Delivery Guarantees and Projection Semantics |
| [ADR-012](docs/adr/ADR-012_Event-Sequencing.md) | Event Sequencing |
| [ADR-013](docs/adr/ADR-013_Four-Module-Workspace.md) | Four Module Workspace |
| [ADR-014](docs/adr/ADR-014_Scaling_Path_BigTable.md) | Scaling Path — Firestore to Bigtable |
| [ADR-015](docs/adr/ADR-015_Narrative-DAG.md) | ~~Narrative as DAG~~ (superseded by ADR-016) |
| [ADR-016](docs/adr/ADR-016_Narrative-Aggregate-Architecture.md) | Narrative Aggregate Architecture — Authoritative Record |
| [ADR-017](docs/adr/ADR-017_Faction-Aggregate.md) | Faction Aggregate Architecture |
| [ADR-017-Amendment-001](docs/adr/ADR-017_Amendment-001_Sparse-Not-Errored.md) | Sparse Not Errored — Faction |
| [ADR-018](docs/adr/ADR-018_Location-Aggregate.md) | Location Aggregate Architecture |
| [ADR-018-Amendment-001](docs/adr/ADR-018_Amendment-001_Sparse-Not-Errored.md) | Sparse Not Errored — Location |
| [ADR-019](docs/adr/ADR-019_Character-aggregate.md) | Character Aggregate Architecture |
| [ADR-019-Amendment-001](docs/adr/ADR-019_Amendment-001_Sparse-Not-Errored.md) | Sparse Not Errored — NPC |
| [ADR-019-Amendment-002](docs/adr/ADR-019_Amendment-002_POwnerPlayerIDOptional.md) | ownerPlayerID Optional on PlayerCharacter |
| [ADR-020](docs/adr/ADR-020_Ownership_Authentication_Authorization.md) | Ownership, Authentication, and Authorization |
| [ADR-021](docs/adr/ADR-021_Game-Interactors.md) | Game Interactors |
| [ADR-022](docs/adr/ADR-022_CAmpaign-interactors.md) | Campaign Interactors |
| [ADR-023](docs/adr/ADR-023_Narrative-Interactors.md) | Narrative Interactors |
| [ADR-024](docs/adr/ADR-024_Faction-Interactors.md) | Faction Interactors |
| [ADR-025](docs/adr/ADR-015_Location-Interactors.md) | Location Interactors |
| [ADR-026](docs/adr/ADR-026_Character-Interactors.md) | Character Interactors |
| [ADR-026-Amendment-001](docs/adr/ADR-026_Amendment-001_Update-MacGuffin-Content.md) | UpdateMacGuffinContent Interactor |
| [ADR-027](docs/adr/ADR-027_RevealEntity-Interactor.md) | RevealEntity Interactor |
| [ADR-028](docs/adr/ADR-028_Graphql-Mutation-Schema.md) | GraphQL Mutation Schema |
| [ADR-028-Amendment-001](docs/adr/ADR-028_Amendment-001_schema-corrections.md) | GraphQL Schema Corrections |
| [ADR-029](docs/adr/ADR-029_GM-Web-Application-architecture.md) | GM Web Application Architecture |
