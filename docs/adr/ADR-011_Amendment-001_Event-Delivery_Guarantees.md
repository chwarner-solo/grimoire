# ADR-011 Amendment 001: Event Delivery Guarantees and Projection Semantics

## Status
Accepted

## Date
2026-05-18

## Amends
ADR-011: Domain Event Chaining via EventBus

---

## Context

ADR-011 established the EventBus port and the InMemoryEventBus /
Cloud Pub/Sub adapter pair. Three things were left implicit that
need to be made explicit before the infrastructure layer is built:

1. **What is the source of truth?** Firestore holds the aggregate
   state. GCS and Neo4j are projections derived from that state.
   Handler failures are therefore recoverable — not catastrophic.

2. **What happens if GCS is unavailable?** An event that never
   reaches GCS is an event that never reaches the log. The outbox
   pattern closes this gap.

3. **What transport do handlers use during the monolith phase?**
   The goroutine bus is the correct local approximation of Pub/Sub
   fire-and-forget semantics. The InMemoryEventBus synchronous
   adapter remains valid for tests.

---

## Decision 1 — Firestore Is the Source of Truth

GCS and Neo4j are **projections**. They are derived from aggregate
state written to Firestore. They can always be rebuilt.

```
Firestore   →  source of truth (aggregate state)
GCS         →  derived  (immutable event log, rebuildable)
Neo4j       →  derived  (read model, rebuildable from GCS)
```

Handler failures are **recoverable**, not catastrophic:

- GCS unavailable   →  outbox holds events until GCS recovers
- Neo4j unavailable →  graph can be rebuilt from GCS event log
- AI handler fails  →  loss is acceptable, retry is optional

The command succeeds when Firestore commits. Everything else
is eventual.

---

## Decision 2 — Transactional Outbox for Durable Handler Delivery

The Firestore aggregate write and the outbox entry are committed
in the **same Firestore transaction**. An event is never lost
because it is durable in Firestore before any handler is invoked.

**Outbox document shape:**

```
/outbox/{eventID}
    eventID        string     ULID — matches EventEnvelope.ID
    eventType      string
    aggregateID    string
    envelope       map        fully serialized EventEnvelope
    createdAt      timestamp
    deliveredTo    []string   handler names that have succeeded
    retryCount     int
    lastAttemptAt  timestamp
```

`deliveredTo` is a per-handler delivery receipt. A handler name
present in `deliveredTo` is skipped on retry. This handles partial
delivery (e.g. GCSWriter succeeded, Neo4jUpdater failed) without
re-running completed work.

---

## Decision 3 — Who Marks deliveredTo

**The DurableHandlerWrapper marks deliveredTo after a successful
Handle() call.**

Durable handlers are constructed with an `OutboxAcknowledger` port.
The wrapper calls `Handle()`, and on success calls
`OutboxAcknowledger.Acknowledge()` which appends the handler name
to `deliveredTo` in Firestore.

Fire-and-forget handlers implement `EventHandler` directly and are
never wrapped. They have no knowledge of the outbox.

```go
// grimoire-domain/shared/event/outbox.go

// OutboxAcknowledger marks a durable handler as delivered for a
// given event. Called by DurableHandlerWrapper after a successful
// Handle().
type OutboxAcknowledger interface {
Acknowledge(
ctx         context.Context,
eventID     string,
handlerName string,
) error
}
```

```go
// grimoire-infrastructure/event/durable_handler_wrapper.go

// DurableHandlerWrapper wraps an EventHandler with outbox
// acknowledgment. Used exclusively for durable handlers.
// Fire-and-forget handlers are not wrapped.
type DurableHandlerWrapper struct {
name   string   // e.g. "GCSWriter", "Neo4jUpdater"
inner  EventHandler
outbox OutboxAcknowledger
logger Logger
}

func (w *DurableHandlerWrapper) Handle(
ctx      context.Context,
envelope EventEnvelope,
) error {
if err := w.inner.Handle(ctx, envelope); err != nil {
w.logger.Error("durable handler failed",
"handler", w.name,
"eventID", envelope.ID,
"error", err,
)
return err
}

if err := w.outbox.Acknowledge(ctx, envelope.ID.String(), w.name); err != nil {
// Acknowledge failure is logged but not fatal.
// The OutboxProcessor will retry Handle() — which must
// be idempotent (see Decision 6).
w.logger.Warn("outbox acknowledge failed",
"handler", w.name,
"eventID", envelope.ID,
"error", err,
)
}
return nil
}
```

---

## Decision 4 — Context Detachment in GoroutineEventBus

The HTTP request context is cancelled when the response returns.
Goroutines must not inherit this context or their writes will be
cancelled mid-flight.

**Durable handlers** use `context.WithoutCancel(ctx)` (Go 1.21+).
This preserves request values (tracing, caller ID) while detaching
the cancellation signal. Writes survive HTTP response return.

**Fire-and-forget handlers** use a fresh `context.Background()` with
a bounded timeout. They need neither request values nor cancellation
survival — a clean context with a timeout is sufficient and simpler.

```go
// grimoire-infrastructure/event/goroutine_bus.go

type GoroutineEventBus struct {
mu       sync.RWMutex
handlers map[EventType][]registeredHandler
logger   Logger
}

type registeredHandler struct {
handler EventHandler
durable bool
}

func (b *GoroutineEventBus) Dispatch(
ctx      context.Context,
envelope EventEnvelope,
) error {
b.mu.RLock()
handlers := b.handlers[envelope.Type]
b.mu.RUnlock()

for _, rh := range handlers {
var hctx context.Context

if rh.durable {
// Detach cancellation — write must complete even
// after the HTTP response has returned.
hctx = context.WithoutCancel(ctx)
} else {
// Fire-and-forget — clean background context
// with a generous timeout.
// cancel is closed over by the goroutine and
// deferred — timer is released whether the handler
// finishes early or times out.
hctx, cancel = context.WithTimeout(
context.Background(),
30*time.Second,
)
}

go func(h EventHandler, c context.Context, cancel context.CancelFunc) {
if cancel != nil {
defer cancel()
}
if err := h.Handle(c, envelope); err != nil {
b.logger.Error("handler error",
"handler", fmt.Sprintf("%T", h),
"eventID", envelope.ID,
"error", err,
)
}
}(rh.handler, hctx, cancel)
}
return nil
}
```

`Subscribe` accepts a `durable bool` to record handler category:

```go
func (b *GoroutineEventBus) Subscribe(
    eventType EventType,
    handler   EventHandler,
    durable   bool,
) error
```

The `EventBus` port interface gains the `durable` parameter on
`Subscribe`. The `InMemoryEventBus` ignores it — all handlers run
synchronously and the context is never cancelled mid-test.

---

## Decision 5 — OutboxProcessor Calls Handlers Directly

The OutboxProcessor does **not** re-dispatch through the EventBus.

Re-dispatching through the bus would re-invoke fire-and-forget
handlers (AI, analytics, PlayerPush) for events they already
received. An AI beat generator must not re-run because GCSWriter
was temporarily unavailable.

The OutboxProcessor maintains its own registry of durable handlers,
independent of the EventBus subscriber list. On each tick:

1. Query Firestore for outbox entries older than a grace period
   (default 5 seconds, tunable via config — allows the goroutine
   bus a chance to succeed before the processor intervenes)
2. For each entry, check `deliveredTo` against the durable handler
   registry
3. Call `Handle()` directly on any handler not yet in `deliveredTo`
4. On success, call `Acknowledge()` to update `deliveredTo`
5. Delete the outbox entry when all durable handlers are in
   `deliveredTo`

```go
// grimoire-infrastructure/event/outbox_processor.go

type OutboxProcessor struct {
    outboxRepo OutboxRepository
    handlers   map[string]EventHandler  // name → durable handler
    logger     Logger
}

func (p *OutboxProcessor) Tick(ctx context.Context) error {
    pending, err := p.outboxRepo.LoadPending(ctx)
    if err != nil {
        return err
    }

    for _, entry := range pending {
        delivered := toSet(entry.DeliveredTo)

        for name, h := range p.handlers {
            if delivered[name] {
                continue  // already succeeded — skip
            }
            if err := h.Handle(ctx, entry.Envelope); err != nil {
                p.logger.Error("outbox retry failed",
                    "handler", name,
                    "eventID", entry.EventID,
                    "error", err,
                )
                continue
            }
            _ = p.outboxRepo.Acknowledge(ctx, entry.EventID, name)
        }

        if p.allDelivered(entry) {
            _ = p.outboxRepo.Delete(ctx, entry.EventID)
        }
    }
    return nil
}
```

In the monolith, OutboxProcessor runs as a background goroutine on
a ticker. In production it becomes a Cloud Run job triggered by
Firestore change streams or Cloud Scheduler.

---

## Decision 6 — Idempotency Is Load-Bearing

Durable handlers run **twice** on failure: once via the goroutine
bus (fails or acknowledge fails), then again via the OutboxProcessor.
Idempotency is not aspirational — it is required for correctness.

**GCSWriter — idempotency via ULID deduplication:**

Before appending to the GCS ndjson log, GCSWriter checks whether
an entry with the same `EventEnvelope.ID` (ULID) already exists.

```
check GCS for eventID  →  found     →  skip, return nil
                       →  not found  →  append, return nil/err
```

At TTRPG scale this check is cheap. At MMORPG scale the GCS object
is partitioned by aggregate — the check scans only that partition.

**Neo4jUpdater — idempotency via upsert:**

All Neo4j writes use `MERGE` not `CREATE`. A node or relationship
that already exists is updated in place, not duplicated.

```cypher
MERGE (n:NPC {id: $id})
SET n.name = $name, n.status = $status
```

A second write of the same event produces the same graph state.
No duplicates. No errors.

---

## Handler Classification

```
DURABLE — outbox backed, wrapped with DurableHandlerWrapper
    GCSWriter       →  append to GCS event log
                       idempotent via ULID deduplication
    Neo4jUpdater    →  project state changes to graph
                       idempotent via MERGE upsert

FIRE-AND-FORGET — no outbox, loss acceptable
    AIBeatGenerator     →  LLM-generated story beats
    ConsistencyChecker  →  LLM cross-beat validation
    AnalyticsHandler    →  usage metrics
    PlayerPushHandler   →  real-time push to player app
                           client polls on miss
    SyncBroker          →  Obsidian/Foundry sync
                           retriable by client
```

---

## Revised Flow

```
Command executes
    └── Firestore transaction (atomic)
            ├── aggregate snapshot updated
            └── outbox entry written
                    ↓
    GoroutineEventBus.Dispatch(envelope)
        ├── DurableHandlerWrapper(GCSWriter)      goroutine, WithoutCancel ctx
        ├── DurableHandlerWrapper(Neo4jUpdater)   goroutine, WithoutCancel ctx
        ├── AIBeatGenerator                        goroutine, bg ctx + timeout
        ├── PlayerPushHandler                      goroutine, bg ctx + timeout
        └── SyncBroker                             goroutine, bg ctx + timeout

    Happy path — both durable handlers succeed:
        DurableHandlerWrapper → Acknowledge() → deliveredTo complete
        OutboxProcessor → finds entry complete → deletes outbox entry

    Unhappy path — GCSWriter fails, Neo4jUpdater succeeds:
        Neo4jUpdater  → Acknowledge() → deliveredTo: ["Neo4jUpdater"]
        GCSWriter     → logs error, returns
        OutboxProcessor tick:
            entry.deliveredTo = ["Neo4jUpdater"]
            calls GCSWriter.Handle() directly — skips Neo4jUpdater
            GCSWriter idempotency check passes (not yet written)
            on success → Acknowledge() → deliveredTo complete → delete
```

---

## Bus Adapters by Phase

```
Phase 1 — Tests
    InMemoryEventBus    synchronous, blocking, errors propagate
                        durable flag ignored — context never cancelled

Phase 2 — Monolith (Docker)
    GoroutineEventBus   fire-and-forget, context detached per category
                        OutboxProcessor as background goroutine ticker

Phase 3 — Production (GCP)
    PubSubEventBus      durable, ordered, distributed
                        OutboxProcessor graduates to Cloud Run job
                        Dead letter topics for persistent failures
```

Handler code does not change across phases.

---

## What Does Not Change From ADR-011

- `EventHandler` interface — unchanged
- All handler implementations — unchanged
- Interactor pattern (ADR-021) — unchanged
- The rule: if Firestore save fails, no dispatch — unchanged
- `EventBus.Subscribe` gains `durable bool` — backwards-compatible
  within this project (no implementations are wired yet), but all
  three EventBus implementors must update their signature

---

## Consequences

- Durable handlers carry one infrastructure dependency: OutboxAcknowledger
- Fire-and-forget handlers have zero infrastructure dependencies
- OutboxProcessor is the sole retry mechanism — no retry logic in handlers
- GoroutineEventBus graduates to PubSubEventBus with zero handler changes
- Neo4j and GCS can be fully rebuilt from Firestore at any time
- Idempotency requirements are explicit per handler — not assumed

---

## Alternatives Considered

**Re-dispatch through EventBus in OutboxProcessor** — rejected.
Would re-invoke fire-and-forget handlers unnecessarily. Direct
handler calls from OutboxProcessor is simpler and safer.

**Synchronous bus in monolith** — rejected. GoroutineEventBus is the
correct local approximation of Pub/Sub semantics. Running synchronous
locally and async in production means different failure modes across
environments.

**Outbox covers all handlers** — rejected. Fire-and-forget handlers
have no correctness requirement. Tracking them adds noise without
benefit.

**Single context strategy for all goroutines** — rejected.
`context.WithoutCancel` preserves request values (tracing, caller ID)
while detaching cancellation — exactly right for durable handlers.
A background context with timeout is simpler and sufficient for
fire-and-forget. Two strategies, two purposes.