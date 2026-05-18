# ADR-029: GM Web Application Architecture

## Status
Accepted

## Date
2026-05-18

---

## Context

ADR-008 established GraphQL as the API layer and noted a direct web UI
as a third capture surface alongside Obsidian and Foundry VTT. ADR-028
deferred query design to ADR-030, explicitly pending UI design (this ADR).

The GM needs a web application that covers both operational contexts
defined in the workflow documents:

- **WORKFLOW-001** — World Building & Session Prep
- **WORKFLOW-002** — Session Operation

These are meaningfully different UX contexts. World Building is a
desktop-oriented authoring environment — rich forms, hierarchy navigation,
DAG visualisation. Session Operation is a tablet-oriented command surface —
fast actions, minimal latency, no friction between the GM's intention and
the system's response.

A single deployed application must serve both contexts without compromise
to either.

### Constraints

- **GM-only.** The Player app is a separate future deployment. This
  application has no player-facing views, no role-based access control,
  no information boundary enforcement on the client side.
- **Firebase authentication** is already committed to by ADR-020.
  The application must handle Firebase JWT tokens and attach them to
  every GraphQL request.
- **Mutations return `MutationResult { id, status }` only** (ADR-028).
  All read views are driven by queries (ADR-030). The UI must follow
  the issue-mutation → issue-query pattern, never deriving view state
  from mutation responses alone.
- **Offline resilience is required for Session Operation.** GMs run
  sessions in venues with unreliable wifi. The session context must
  remain usable when connectivity is intermittent. World Building can
  degrade gracefully offline — it is a prep-time activity.
- **No subscriptions in this application.** Real-time push
  (`entityRevealed`, `sessionStarted`) is deferred to the Player app
  ADR. The GM app polls or refetches after mutations.

---

## Decision

### Application Type

Progressive Web Application (PWA) deployed as a static site.

- Installable on desktop and tablet (add to home screen)
- Service worker for offline asset caching and request queuing
- Single binary deployment — no server-side rendering required
- Firebase Hosting is the natural deployment target given existing
  Firebase auth dependency

### Module Location

`grimoire-pwa/` at the workspace root, peer to the four Go modules.

```
grimoire/main/
  go.work
  grimoire-domain/
  grimoire-infrastructure/
  grimoire-testing/
  grimoire-api/
  grimoire-pwa/              ← this ADR
    package.json
    vite.config.ts
    tsconfig.json
    src/
```

`grimoire-pwa` is not a Go module. It has no `go.mod`. The workspace
`go.work` is unchanged.

---

## Technology Stack

### Core

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Language | TypeScript | Type safety end-to-end; codegen outputs TS types |
| UI framework | React 19 | Ecosystem maturity; team familiarity; stable concurrent features |
| Build tool | Vite | Fast dev server; first-class PWA plugin; excellent TS support |
| Styling | Tailwind CSS v4 | Utility-first; consistent design tokens; no runtime overhead |
| PWA | `vite-plugin-pwa` | Workbox-backed; minimal config for asset caching + offline |

### GraphQL

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Client | Apollo Client 3 | Normalised cache; devtools; built for complex mutation flows |
| Code generation | `@graphql-codegen/cli` | Generates typed hooks from `grimoire-api/schema/*.graphql` |

Apollo over urql: the normalised cache is worth the weight here.
After a mutation returns a `MutationResult { id, status }`, the client
issues a follow-up query. Apollo's cache ensures related views update
automatically when the same entity is returned by any query. At 46
mutations across 10 aggregate types, manual cache invalidation would
become a maintenance burden. Apollo handles it structurally.

GraphQL Codegen is non-negotiable. The `grimoire-api` schema is already
fully typed via gqlgen. Codegen bridges that type safety to the frontend.
Writing queries in `.graphql` files and receiving typed React hooks
eliminates an entire category of runtime bugs and removes the need for
manual TypeScript interfaces for API responses.

```ts
// Codegen produces this from a .graphql file:
const { data, loading } = useGetCampaignQuery({
  variables: { id: campaignId }
})
// data.campaign is fully typed — no casting, no any
```

Codegen output lives in `src/graphql/generated/` and is never
hand-edited. It is regenerated whenever `grimoire-api/schema/` changes.

### State Management

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Server state | Apollo Client cache | Queries and mutations managed by Apollo |
| Client state | Zustand | Lightweight; typed stores; no boilerplate |

Zustand is used exclusively for client-side UI state that does not
belong in Apollo's cache:

- Active session context (current sessionID, campaignID in progress)
- UI preferences (sidebar collapsed, active context selection)
- Optimistic local state during in-flight mutations

Redux is rejected. The ceremony-to-value ratio is wrong for this
application. Zustand with explicit typed stores per context provides
the same guarantees without the boilerplate.

### Authentication

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Auth provider | Firebase Auth SDK | Already committed by ADR-020 |
| Token handling | Apollo auth link | Attaches JWT to every GraphQL request transparently |

The Firebase SDK manages token refresh automatically. The Apollo auth
link reads the current token on every request — it never caches a
stale token.

```ts
// grimoire-pwa/src/lib/apollo.ts
const authLink = new ApolloLink(async (operation, forward) => {
  const token = await getIdToken(auth.currentUser)
  operation.setContext({
    headers: { Authorization: `Bearer ${token}` }
  })
  return forward(operation)
})
```

`CallerID` is never supplied by the client. The server resolves it from
the JWT via `CallerIdentityPort.Identify()` (ADR-020). This constraint
is enforced structurally — no mutation input type accepts a caller ID.

---

## Application Structure

### Two Top-Level Contexts

The application has two distinct top-level contexts matching the
workflow documents. These are not just routes — they are separate
UX modes with different layout, interaction density, and performance
requirements.

```
/world-building/*    ← WORKFLOW-001: prep-time authoring
/session/*           ← WORKFLOW-002: at-the-table operation
```

A persistent navigation shell allows switching between contexts.
The active context drives the layout — world building uses a
multi-panel desktop layout; session operation uses a single-panel
tablet layout optimised for touch.

```
grimoire-pwa/src/
  contexts/
    world-building/          ← WORKFLOW-001 screens and components
      pages/
      components/
      stores/                ← Zustand stores scoped to this context
    session-operation/       ← WORKFLOW-002 screens and components
      pages/
      components/
      stores/
  graphql/
    queries/                 ← .graphql files — one per screen/concern
    mutations/               ← .graphql files — mirrors ADR-028 schema
    generated/               ← codegen output, never hand-edited
  lib/
    apollo.ts                ← Apollo client instantiation + auth link
    firebase.ts              ← Firebase app + auth initialisation
  components/                ← shared components used by both contexts
  App.tsx
  main.tsx
```

### Mutation Pattern

All writes follow the same pattern, matching ADR-028's `MutationResult`:

```ts
// 1. Execute mutation — returns id + status only
const [createCampaign] = useCreateCampaignMutation()
const result = await createCampaign({ variables: { input } })
const campaignId = result.data?.createCampaign.id

// 2. Refetch to populate the view — queries drive reads
await getCampaign({ variables: { id: campaignId } })
```

Mutation responses are never used to derive view state directly.
The follow-up query is always issued. This is a hard rule — it
keeps the client aligned with the CQRS split (ADR-003) and ensures
the Apollo cache is populated from the canonical read model (Neo4j),
not from write-side mutation echoes.

### Offline Strategy

Two tiers, matching the two workflow contexts:

**World Building — graceful degradation.**
Service worker caches all static assets. On loss of connectivity,
the app remains open and readable. Writes are blocked with a clear
"You are offline" indicator. No mutation queuing — world building
changes are not time-critical and should not be silently queued for
later replay.

**Session Operation — resilient operation.**
Service worker caches static assets. Apollo's in-memory cache ensures
already-loaded session data remains accessible. Failed mutations are
surfaced immediately with a retry affordance. The GM is never left
with a spinner mid-session.

Workbox strategy: `CacheFirst` for static assets, `NetworkFirst`
for GraphQL requests (falls back to cached response on failure in
session context only).

---

## Codegen Configuration

Codegen reads the existing `grimoire-api/schema/*.graphql` files
directly. No schema duplication.

```yaml
# grimoire-pwa/codegen.ts

schema: "../grimoire-api/schema/*.graphql"

documents: "src/graphql/**/*.graphql"

generates:
  src/graphql/generated/index.ts:
    plugins:
      - typescript
      - typescript-operations
      - typescript-react-apollo
    config:
      withHooks: true
      withComponent: false
      scalars:
        DateTime: string
        Date: string
```

Regenerate after any schema change:

```bash
cd grimoire-pwa && npx graphql-codegen
```

This is a required step before implementing any screen that touches
a new or changed query or mutation. CI enforces that generated files
are up to date.

---

## Consequences

- ADR-030 (query schema) is now unblocked. Screen designs in
  `docs/ui/UI-001` will drive the exact query shapes needed.
- All type safety flows from the Go domain through gqlgen through
  codegen into TypeScript. A field rename in the domain schema
  produces a TypeScript compile error in the frontend.
- The mutation → query pattern adds one round trip per write
  operation. This is acceptable — it is architecturally correct
  and the latency is imperceptible in world building. In session
  operation, the Apollo cache means refetches are often served
  locally.
- No subscriptions means the GM must manually refresh to see
  changes made from Obsidian or Foundry during a session. This
  is acceptable for the GM-only app. The Player app (future) will
  use subscriptions for real-time push.
- `grimoire-pwa` is outside the Go workspace. It has its own
  `package.json`, `tsconfig.json`, and `vite.config.ts`. It is
  built and deployed independently of the Go API.
- The Player app, when built, will be a separate PWA deployment —
  not a view within this application. The GM app has no player-facing
  code paths. Keeping the boundary clean now avoids refactoring later.

---

## Deferred

```
Player app               →  future ADR — separate PWA deployment
                             subscriptions (entityRevealed, sessionStarted)
                             information boundary enforcement on client

Co-GM / collaboration    →  future ADR (ADR-020 deferred this)

Read authorization       →  ADR-030 will define query schema;
                             Neo4j query layer enforces GM-only access
                             for this app (no player-visible filtering needed)

Offline mutation queue   →  deferred; requires conflict resolution
                             strategy for event-sourced backend
```

---

## Related

- ADR-003 — CQRS Split (mutations write, queries read)
- ADR-008 — GraphQL API Layer
- ADR-013 — Four Module Workspace (grimoire-pwa sits alongside)
- ADR-020 — Authentication and Authorization (Firebase JWT)
- ADR-028 — GraphQL Mutation Schema (46 mutations, MutationResult shape)
- ADR-030 — GraphQL Query Schema (next — driven by UI-001)
- WORKFLOW-001 — World Building & Session Prep
- WORKFLOW-002 — Session Operation
- docs/ui/UI-001 — GM Application Screens and Flows (to be written)