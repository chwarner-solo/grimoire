# ADR-005: ndjson Over Parquet for Event Storage

## Status
Accepted

## Date
2026-05-09

## Context
Events are written to GCS as the datalake. The choice of file format
affects write complexity, schema flexibility, BigQuery compatibility,
and long-term storage cost. The two main candidates were ndjson
(newline-delimited JSON) and Parquet (columnar binary format).

## Decision
Use ndjson as the primary event storage format in GCS.

```
/events/{owner_id}/{entity_type}/{year}/{month}/events.ndjson
```

## Reasoning
**Write path simplicity:** Events are written one at a time from thin
clients (phone, tablet, Obsidian plugin). ndjson is append-friendly.
Parquet requires batch encoding — inappropriate for a streaming write path
from a mobile client.

**Schema flexibility:** The event schema will evolve. New entity types and
relationship types will be added. ndjson handles schema evolution naturally.
Parquet requires schema definition upfront and migration tooling for changes.

**BigQuery native support:** BigQuery reads ndjson directly via external
tables. No transformation pipeline required to query the datalake.

**Data volume:** Even heavy users produce thousands of events per year —
not millions. Parquet's columnar compression provides no meaningful benefit
at this scale.

**Human readable:** ndjson is debuggable without tooling. Critical during
development and incident response.

## Consequences
- Write path is simple: append one JSON line per event
- BigQuery external tables work without transformation
- Storage cost is higher than Parquet at large scale
- A Parquet archiver can be added later without changing the write path

## Upgrade Path
If aggregate analytics across large user populations become necessary,
a Cloud Run job can convert ndjson to Parquet on a monthly schedule.
The write path never changes. The BigQuery external table pointer changes.

## Alternatives Considered
**Parquet from day one** — rejected. Requires a transformation pipeline
between the write path and storage. Inappropriate complexity for the
current scale and write pattern.

**Avro** — rejected. Same batch-encoding complexity as Parquet without
the columnar query benefits.