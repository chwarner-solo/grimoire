# ADR-001: Game Holds CampaignIDs Not SessionIDs

## Status
Accepted

## Date
2026-05-09

## Context
When modeling the Game aggregate root we needed to decide what child
references Game holds. The initial instinct was Game → Sessions, because
that is how an ORM-driven model would naturally express it via foreign keys.

A Session represents a single night of play at a specific table. Multiple
tables can run the same Game simultaneously, each with their own session
history. This means Sessions do not belong to a Game directly — they belong
to the table running the Game.

## Decision
Game holds `[]CampaignID` only.
Campaign holds `[]SessionID`.
Session belongs to exactly one Campaign.

## Reasoning
A Campaign represents a specific table running a Game. The invariant
"only one session in_progress at a time" belongs to Campaign, not Game.
Game has no business enforcing session state — it has no way of knowing
which table a session belongs to.

Placing SessionIDs on Game would blur the Campaign boundary and force Game
to reason about things it does not own.

## Consequences
- Session queries are always scoped through Campaign
- Game aggregate is significantly simpler
- The read model (Neo4j) handles Game+Sessions queries without touching aggregates
- Campaign is a proper aggregate root with its own lifecycle

## Alternatives Considered
**Game holding []SessionID directly** — rejected. Game cannot enforce session
invariants across multiple tables. The relationship Game → Session skips a
meaningful domain concept (the table/campaign).

**Flat structure with no Campaign** — rejected. Loses the concept of multiple
tables running the same Game, which is a first-class domain concept.