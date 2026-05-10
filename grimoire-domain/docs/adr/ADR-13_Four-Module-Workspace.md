# ADR-013: Four Module Workspace Structure

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire requires a clean separation between domain logic, infrastructure
adapters, test support, and the application entry point. A single module
approach would allow infrastructure concerns to leak into the domain and
test doubles to ship in production binaries.

Go workspaces (go.work) allow multiple modules to reference each other
locally during development without publishing to a registry first. This
enables a clean multi-module structure with strict dependency boundaries
enforced by the Go module system itself.

## Decision
Four modules in the Go workspace. Strict inward dependency rule.

```
grimoire/main/
  go.work

  grimoire-domain/
    go.mod    (github.com/chwarner-solo/grimoire/grimoire-domain)

  grimoire-infrastructure/
    go.mod    (github.com/chwarner-solo/grimoire/grimoire-infrastructure)

  grimoire-testing/
    go.mod    (github.com/chwarner-solo/grimoire/grimoire-testing)

  grimoire-api/
    go.mod    (github.com/chwarner-solo/grimoire/grimoire-api)
```

**Dependency rule — arrows point inward only:**
```
grimoire-domain         →  no internal imports
                           external: uuid, ulid only

grimoire-infrastructure →  grimoire-domain only
                           external: Firestore, Neo4j,
                                     GCS, Pub/Sub SDKs

grimoire-testing        →  grimoire-domain only
                           external: testify or stdlib testing

grimoire-api            →  grimoire-domain
                           grimoire-infrastructure
                           external: gqlgen, HTTP server
                           test files only: grimoire-testing
```

## Module Responsibilities

**grimoire-domain**
Pure domain logic. Aggregate roots, value objects, domain events,
port interfaces. Zero infrastructure dependencies. Compilable and
testable with no external services.

**grimoire-infrastructure**
Production adapters implementing domain ports. One package per
external dependency:
```
firestore/   FirestoreGameRepository, FirestoreCampaignRepository
neo4j/       Neo4jEventHandler, Neo4jQueryAdapter
gcs/         GCSEventWriter
pubsub/      PubSubEventBus
```
Never imported by test code directly — test code uses grimoire-testing.

**grimoire-testing**
Test doubles, builders, and fixtures. Never compiled into production
binaries. Only imported in _test.go files.
```
inmemory/    InMemoryEventBus, InMemoryRepository
doubles/     Fake implementations of domain ports
builders/    Fluent test data builders (AGame(), ACampaign())
fixtures/    Known domain states for integration tests
```

**grimoire-api**
Application entry point. Wires domain ports to infrastructure adapters.
GraphQL schema and resolvers. HTTP server. Command handlers. Production
code imports grimoire-infrastructure. Test code imports grimoire-testing.
The only module with a runnable binary.

## go.work Configuration
```
go 1.22

use (
    ./grimoire-domain
    ./grimoire-infrastructure
    ./grimoire-testing
    ./grimoire-api
)
```

## Deployment Boundary
```
Production binary compiles:
  grimoire-domain         ✓
  grimoire-infrastructure ✓
  grimoire-api            ✓
  grimoire-testing        ✗ never compiled in
```

The Go module system enforces this boundary. grimoire-testing is never
imported by production code — only by _test.go files. It cannot
accidentally ship in a production binary.

## Test Builder Pattern
grimoire-testing provides fluent builders for constructing domain
objects in tests without coupling tests to constructor signatures:

```go
// grimoire-testing/builders/game_builder.go
game := AGame().Named("Ashes & Chains").InDraftState().Build()
campaign := ACampaign().ForGame(game).WithCharacter(char).Build()
```

Tests read as sentences describing intent, not construction details.

## Consequences
- Domain is independently compilable and testable with zero infrastructure
- Infrastructure adapters can be swapped without touching domain or API
- Test doubles never ship in production binaries
- Four modules to maintain and version
- go.work ties them together locally — CI must handle module graph correctly
- Adding a new infrastructure adapter requires only grimoire-infrastructure
  changes — domain and API unchanged

## Alternatives Considered
**Single module** — rejected. Infrastructure leaks into domain.
Test doubles ship in production. No enforced boundary.

**Two modules (domain + api)** — rejected. Infrastructure and test
support have no clean home. Test doubles end up in production code
or in the domain module, both wrong.

**grimoire-testing inside grimoire-domain** — rejected. Domain module
ships with test infrastructure attached. Increases production binary
size. Violates the principle that domain has minimal dependencies.

**grimoire-testing inside grimoire-infrastructure** — rejected.
Test doubles are not infrastructure. Conflates production adapters
with test support. Infrastructure module pulls in test dependencies.