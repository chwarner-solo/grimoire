# ADR-028 Amendment 001: Schema Corrections — Blocker Resolution

## Status
Accepted

## Date
2026-05-17

## Amends
ADR-028: GraphQL Mutation Schema

---

## Context

Claude Code's readiness review identified three hard blockers where the
schema could not map to the interactor as written, and two content-field
decisions that needed resolution before implementation.

All five are corrected here.

---

## Correction 1 — AddStandingLevelInput: add ordinal

`StandingLevel` is a value object requiring three fields:
`name`, `threshold`, and `ordinal`. The ADR-028 schema omitted `ordinal`,
making it impossible to construct the value object in the resolver.

`ordinal` defines the sort order of standing tiers. The GM supplies it
explicitly so that tiers render in a meaningful sequence regardless of
insertion order.

```graphql
# BEFORE
input AddStandingLevelInput {
  factionId: ID!
  gameId:    ID!
  name:      String!
  threshold: Int!
}

# AFTER
input AddStandingLevelInput {
  factionId: ID!
  gameId:    ID!
  name:      String!
  threshold: Int!
  ordinal:   Int!
}
```

No interactor change. Resolver passes all three scalars to
`factionvalue.NewStandingLevel(name, threshold, ordinal)`.

---

## Correction 2 — AddFactionMemberInput: add rank

`NewFactionMembership(id, factionID, characterID, rank)` requires a
`rank` string — the NPC's role within the faction (e.g. "Captain",
"Spy", "Initiate"). The ADR-028 schema omitted it.

```graphql
# BEFORE
input AddFactionMemberInput {
  factionId: ID!
  gameId:    ID!
  npcId:     ID!
}

# AFTER
input AddFactionMemberInput {
  factionId: ID!
  gameId:    ID!
  npcId:     ID!
  rank:      String!
}
```

No interactor change. Resolver passes `rank` to the interactor request.

---

## Correction 3 — CreateLocationInput: drop description, add locationType

Two problems in one input:

- Schema had `description: String!` — the interactor has no description
  field. Content fields follow the sparse-not-errored principle and are
  set later via a future `updateLocationContent` mutation.

- The interactor requires `LocationType locationvalue.LocationType` which
  the schema was missing entirely. LocationType is an identity-level field
  (not content) — it determines where in the hierarchy the location lives.

```graphql
# BEFORE
input CreateLocationInput {
  gameId:      ID!
  name:        String!
  description: String!
  parentId:    ID
}

# AFTER
enum LocationType {
  WORLD
  REGION
  SETTLEMENT
  BUILDING
  SCENE
}

input CreateLocationInput {
  gameId:       ID!
  name:         String!
  locationType: LocationType!
  parentId:     ID
}
```

No interactor change. `description` removed from schema input.
`locationType` maps to `locationvalue.LocationType` in the resolver.

---

## Correction 4 — CreateCampaignBeatInput: drop description and playerDesc

**Decision:** Align schema to interactor — content fields set via
`updateBeatContent`, not at creation time.

The `CreateCampaignBeat` interactor accepts only `name`. Description and
playerDescription are set in a follow-up call to `updateBeatContent`.
This is consistent with the Beat creation pattern established by
`CreateMasterBeat`.

```graphql
# BEFORE
input CreateCampaignBeatInput {
  campaignId:  ID!
  gameId:      ID!
  name:        String!
  description: String!
  playerDesc:  String!
}

# AFTER
input CreateCampaignBeatInput {
  campaignId: ID!
  gameId:     ID!
  name:       String!
}
```

No interactor change. Content is set via the existing `updateBeatContent`
mutation after creation.

---

## Correction 5 — CreateMacGuffinInput: strip content, add updateMacGuffinContent

**Decision:** Strip content fields from `CreateMacGuffinInput` and add
a new `updateMacGuffinContent` mutation backed by a new interactor.

The `CreateMacGuffin` interactor accepts only `gameId` and `name`.
Content fields (description, playerDesc) had no interactor to receive
them, making the schema unimplementable.

```graphql
# BEFORE
input CreateMacGuffinInput {
  gameId:      ID!
  name:        String!
  description: String!
  playerDesc:  String!
}

# AFTER
input CreateMacGuffinInput {
  gameId: ID!
  name:   String!
}

# NEW mutation added to character.graphql
extend type Mutation {
  """Update content fields on a MacGuffin."""
  updateMacGuffinContent(input: UpdateMacGuffinContentInput!): MutationResult!
}

input UpdateMacGuffinContentInput {
  macGuffinId: ID!
  gameId:      ID!
  name:        String!
  description: String!
  playerDesc:  String!
}
```

The `UpdateMacGuffinContentInteractor` is specified in
ADR-026 Amendment 001.

---

## Revised Total

ADR-028 specified 46 mutations. After corrections:

```
Removed:  0  (no mutations removed)
Added:    1  (updateMacGuffinContent)
Total:    47 mutations
```

---

## Summary of All Changes

| # | Input | Change |
|---|-------|--------|
| 1 | AddStandingLevelInput | Add `ordinal: Int!` |
| 2 | AddFactionMemberInput | Add `rank: String!` |
| 3 | CreateLocationInput | Drop `description`, add `locationType: LocationType!`, add `LocationType` enum |
| 4 | CreateCampaignBeatInput | Drop `description` and `playerDesc` |
| 5 | CreateMacGuffinInput | Drop `description` and `playerDesc`; add `updateMacGuffinContent` mutation + input |