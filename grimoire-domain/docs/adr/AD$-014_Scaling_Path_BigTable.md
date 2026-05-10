# ADR-014: Scaling Path — Firestore to Bigtable with Event Sourcing Snapshots

## Status
Accepted

## Date
2026-05-09

## Context
Grimoire's initial persistence strategy uses Firestore as the aggregate
state store (ADR-010). Firestore is appropriate for the initial scale —
thousands of users, hundreds of aggregates per user. However, the long
term vision for Grimoire is MMORPG scale: millions of concurrent users,
millions of entities, millions of events per second.

Firestore does not scale to this level cost-effectively. A scaling path
must be defined that:

1. Does not require changes to the domain layer
2. Does not require changes to command handlers
3. Can be migrated to incrementally without a big-bang rewrite
4. Is proven at MMORPG scale

## Decision
Three phase scaling path. Each phase is independently deployable.
The domain port abstraction enables migration without touching domain
or command handler code.

### Phase 1 — Firestore + GCS (current)
```
Aggregate store:   Firestore (full aggregate document)
Event log:         GCS (ndjson append)
Analytics:         BigQuery external tables over GCS
Scale target:      thousands of users
```

### Phase 2 — Firestore + GCS + BigQuery Analytics
```
Aggregate store:   Firestore (unchanged)
Event log:         GCS (unchanged)
Analytics:         BigQuery scheduled jobs
                   materialized views over event log
                   session timelines, entity histories
Scale target:      tens of thousands of users
```

### Phase 3 — Bigtable Event Log + Snapshot Store
```
Aggregate store:   Bigtable (events + snapshots)
Event log:         Bigtable (replaces GCS)
Snapshot generator: BigQuery batch jobs
Analytics:         BigQuery reads from Bigtable
Scale target:      millions of users — MMORPG scale
```

## Phase 3 Architecture

### Bigtable Row Key Design
```
Event rows:
  {entity_id}#event#{sequence_number_zero_padded}
  Example: npc_korvan_001#event#0000000000004521

Snapshot rows:
  {entity_id}#snapshot#{timestamp_unix_ms}
  Example: npc_korvan_001#snapshot#1746820800000
```

Row key design enables:
- Range scan: all events for entity X after sequence N
- Latest snapshot: reverse scan on snapshot prefix, take first row
- Both operations are single range scans — microsecond latency

### Aggregate Reconstruction on Command
```
1. Load latest snapshot from Bigtable
   row key: {entity_id}#snapshot#*  (reverse scan, limit 1)
   result:  Snapshot{ sequence: 4500, state: {...} }

2. Load events since snapshot
   row key range: {entity_id}#event#0000000004501
              to: {entity_id}#event#9999999999999
   result:  []Event since sequence 4500

3. Replay events onto snapshot state
   in memory, no I/O
   result:  current aggregate state

4. Handle command, enforce invariant
   emit new event
   append to Bigtable
   write new snapshot if threshold reached
```

### Snapshot Generation via BigQuery
```
BigQuery scheduled job (hourly or on-demand):
  reads all events for entity X up to watermark
  replays event log in sequence order
  produces entity state at watermark moment
  writes snapshot row to Bigtable

Command handler benefits:
  cold start reads snapshot    ← milliseconds
  reads events since snapshot  ← microseconds per event
  typically < 100 events to replay between snapshots
```

### Snapshot Threshold
A new snapshot is written when:
- Events since last snapshot exceed N (configurable, default 100)
- Time since last snapshot exceeds T (configurable, default 1 hour)
- Triggered by BigQuery batch job on schedule

## The Port That Enables Zero-Domain-Change Migration

```go
// grimoire-domain/shared/port/aggregate_store.go

type Snapshot struct {
    EntityID        GrimoireID
    AggregateType   string
    SequenceNumber  uint64
    State           json.RawMessage
    SnapshotAt      time.Time
}

type AggregateStore interface {
    // Load latest snapshot for entity
    LoadSnapshot(ctx context.Context, id GrimoireID) (Snapshot, error)

    // Load all events after sequence number
    LoadEventsSince(ctx context.Context, id GrimoireID, seq uint64) ([]Event, error)

    // Save a new snapshot
    SaveSnapshot(ctx context.Context, snap Snapshot) error

    // Append a new event
    AppendEvent(ctx context.Context, event Event) error
}
```

**Phase 1 adapter:** FirestoreAggregateStore
**Phase 3 adapter:** BigtableAggregateStore

Command handlers call the port. They never change between phases.
Swapping the adapter in grimoire-infrastructure is the entire migration.

## Why Bigtable at MMORPG Scale

```
Bigtable:   petabyte scale
            millions of read/write ops per second
            single-digit millisecond latency
            same API as Apache HBase
            used by Google Search, Maps, Gmail
            designed for exactly this access pattern:
            "give me all rows for key X after row Y"

Firestore:  50,000 writes/second per database
            strong consistency
            document model
            excellent for Phase 1
            ceiling too low for MMORPG scale
```

## Migration Path Phase 1 → Phase 3

```
1. Build BigtableAggregateStore adapter
   in grimoire-infrastructure/bigtable/
   implement AggregateStore port
   write integration tests against
   Bigtable emulator

2. Backfill — replay GCS event log
   into Bigtable event rows
   generate initial snapshots via BigQuery

3. Dual write period
   write to both Firestore and Bigtable
   verify consistency

4. Read switchover
   command handlers read from Bigtable
   verify correctness

5. Firestore retired
   Bigtable is sole aggregate store
```

Zero domain changes. Zero command handler changes.
One new adapter. One wire-up change in grimoire-api.

## Consequences
- Domain and command handlers are insulated from infrastructure scaling decisions
- Each phase is independently deployable with no big-bang migration
- Bigtable requires careful row key design — documented above
- Snapshot threshold tuning required at scale
- BigQuery batch jobs add operational complexity in Phase 3
- Bigtable emulator available for local development and CI

## Alternatives Considered
**Firestore forever** — rejected. 50,000 writes/second ceiling.
Cost becomes prohibitive at MMORPG scale. Not designed for
time-series event log access patterns.

**PostgreSQL with event sourcing** — considered. Mature tooling,
familiar operations. Does not scale horizontally to MMORPG level
without significant sharding complexity. Bigtable handles sharding
transparently.

**Spanner** — considered. Globally consistent, horizontally scalable.
Significantly more expensive than Bigtable. Consistency guarantees
exceed what event sourcing requires. Bigtable's eventual consistency
is sufficient for the snapshot + replay pattern.

**Skip snapshots, full replay always** — rejected. At MMORPG scale
an entity may have millions of events. Full replay on every command
is not viable. Snapshots are required at this scale.