# ADR-011: Domain Event Chaining via EventBus

## Status
Accepted

## Date
2026-05-09

## Context
Aggregate roots sometimes need to react to events emitted by other
aggregate roots. The canonical example in Grimoire:

```
Campaign.StartSession() emits SessionStarted
Game is interested in SessionStarted:
  "if this is my first active campaign,
   transition from Draft to Active"
Game then emits GameStatusChanged
```

Aggregates must never communicate directly. Game cannot call Campaign.
Campaign cannot call Game. They are independent aggregate roots with
independent lifecycles and boundaries.

The question is how one aggregate reacts to another aggregate's event
without coupling them directly.

## Decision
All inter-aggregate communication flows through the EventBus via
subscribed domain event handlers. Aggregates never reference each other
directly.

**Flow:**
```
StartSession mutation
      ↓
StartSessionCommand dispatched
      ↓
Campaign aggregate loaded from Firestore
Campaign.StartSession() called
SessionStarted event emitted to EventBus
Campaign saved to Firestore
      ↓ (EventBus routes to all subscribers)
      ├── GCSWriter.Handle(SessionStarted)
      │     writes event to GCS ndjson
      │
      ├── Neo4jUpdater.Handle(SessionStarted)
      │     mutates graph
      │
      └── GameStatusHandler.Handle(SessionStarted)
            loads Game aggregate from Firestore
            game.OnCampaignSessionStarted()
            if state changed → GameStatusChanged emitted
            Game saved to Firestore
                  ↓ (EventBus routes again)
                  ├── GCSWriter.Handle(GameStatusChanged)
                  └── Neo4jUpdater.Handle(GameStatusChanged)
```

**Port definitions:**
```go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType EventType, handler EventHandler) error
}

type EventHandler interface {
    Handle(ctx context.Context, event Event) error
}
```

**Local development adapter — in-memory, synchronous:**
```go
type InMemoryEventBus struct {
    mu       sync.RWMutex
    handlers map[EventType][]EventHandler
}
```

Sequential and synchronous. Predictable for testing.

**Production adapter — Cloud Pub/Sub:**
Ordered delivery via ordering keys scoped to aggregate ID.
Ensures events for the same aggregate are processed in sequence.

## Handler Responsibilities
Each handler has exactly one job:

```
GCSWriter           →  append event to GCS ndjson
Neo4jUpdater        →  mutate graph from event
GameStatusHandler   →  react to campaign/session events
                        on behalf of Game aggregate
SyncBroker          →  push to Obsidian or Foundry
                        based on event source
PlayerPushHandler   →  push EntityRevealed to player app
```

## Ordering Constraint
Event chains create ordering dependencies. SessionStarted must be
fully handled by GameStatusHandler before GameStatusChanged fires.

- **In-memory bus:** handlers execute sequentially — no issue
- **Cloud Pub/Sub:** ordering keys scoped to aggregate ID ensure
  events for one aggregate are delivered in order

## Consequences
- Aggregates are fully decoupled from each other
- New reactions to existing events require only a new handler subscription
- Event chains are traceable through the GCS event log
- Handler failures must be handled carefully — a failed GameStatusHandler
  leaves Game in an inconsistent state relative to Campaign
- Dead letter queues required in production for failed handler retries

## Alternatives Considered
**Direct aggregate-to-aggregate calls from command handler** — rejected.
Handler knows too much. Aggregates become coupled. Adding a new reaction
requires modifying the command handler.

**Saga / Process Manager** — considered for complex multi-step flows.
Deferred — current event chains are simple enough that direct handler
subscriptions are sufficient. Revisit if chains become multi-step with
compensation requirements.