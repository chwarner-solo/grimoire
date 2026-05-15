# ADR-012: Event Sequencing — ULID and Per-Aggregate Sequence Numbers

## Status
Accepted

## Date
2026-05-09

## Context
Events written to GCS must be orderable for correct replay. Timestamp
alone is insufficient:

- Clock skew between Cloud Run instances
- Two events can share the same millisecond
- GCS write order does not guarantee emit order
- Multiple instances writing concurrently creates ambiguity

A reliable sequencing strategy is required at two levels:
1. Global ordering across all events in the system
2. Per-aggregate ordering within a single aggregate's history

## Decision
Two complementary sequencing mechanisms on every EventEnvelope:

**ULID — global ordering:**
```
Format:   01ARZ3NDEKTSV4RRFFQ69G5FAV
          ├─────────────┤├────────────┤
          48-bit ms timestamp  80-bit random

Sortable:     lexicographic sort = chronological sort
Unique:       80 bits of randomness — collision resistant
No authority: generated client side, no central coordinator
Readable:     Crockford base32, no ambiguous characters
```

**SequenceNumber — per-aggregate ordering:**
```
uint64 monotonic counter per aggregate
stored on the Firestore aggregate document
incremented atomically within the Firestore
transaction that saves the aggregate
"this is the 47th event on Campaign X"
```

**EventEnvelope:**
```go
type EventEnvelope struct {
    ID             ULID            // globally sortable event ID
    Type           EventType
    AggregateID    GrimoireID
    AggregateType  string
    SequenceNumber uint64          // monotonic within aggregate
    CampaignID     CampaignID
    SessionID      SessionID
    Source         Source          // obsidian | foundry | grimoire
    ActorID        string
    OccurredAt     time.Time
    Payload        json.RawMessage
}
```

**Atomic sequence increment in Firestore transaction:**
```go
err = client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
    doc := tx.Get(aggregateRef)
    seq := doc.Data()["sequence"].(uint64) + 1
    event.SequenceNumber = seq
    tx.Set(aggregateRef, map[string]interface{}{
        "sequence": seq,
        "state":    aggregateData,
    })
    return nil
})
```

Sequence number and aggregate state update atomically.
No gap. No duplicate. No race condition.

## What Each Mechanism Solves

```
ULID:             "order all events across the system
                   for full replay from GCS"

SequenceNumber:   "order events for one aggregate
                   detect gaps in aggregate history
                   optimistic concurrency control
                   expected sequence N, got M → conflict"
```

## Optimistic Concurrency Control
SequenceNumber doubles as an optimistic lock:

```go
// Command handler loads aggregate
// Aggregate carries its current sequence number
// Before saving, verify sequence has not advanced
// If it has, another instance handled a command first
// Retry or return conflict error
```

## Consequences
- Every event is globally sortable via ULID without a central authority
- Per-aggregate history is gapless and detectable
- Firestore transaction overhead on every command — acceptable at this scale
- ULID library required: github.com/oklog/ulid/v2
- GCS event replay uses ULID sort for correct global ordering
- Per-aggregate replay uses SequenceNumber for correct aggregate ordering

## Alternatives Considered
**Timestamp only** — rejected. Clock skew and same-millisecond collisions
make timestamp ordering unreliable in a distributed system.

**Global monotonic sequence (Firestore counter)** — rejected. Single
global counter is a write bottleneck. Every event across all aggregates
contends on one document.

**UUID as event ID** — rejected. UUIDs are not sortable. Ordering requires
a separate timestamp field which reintroduces clock skew problems.

**Vector clocks** — rejected. Tracks causality between events accurately
but is significantly more complex to implement and query. Overkill for
this use case where ULID + SequenceNumber provides sufficient ordering
guarantees.

**Hybrid Logical Clock (HLC)** — considered. Handles clock skew elegantly.
ULID chosen over HLC for simplicity and available Go library support.
HLC revisitable if ULID proves insufficient.