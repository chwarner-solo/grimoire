# UI-001: GM Application — Screens & Flows

## Scope

This document covers the **World Building & Session Prep** context
(WORKFLOW-001). Session Operation screens are UI-002.

This document drives ADR-030 (GraphQL Query Schema). Every screen
section calls out the queries it requires.

---

## Navigation Model

### Shell Layout

```
┌─────────────────────────────────────────────────────────────────┐
│  GRIMOIRE          [World Building ▼]          [Avatar / Sign out]│
├──────────┬──────────────────────────────────────────────────────┤
│          │                                                        │
│  Left    │  Main Content Area                                     │
│  Sidebar │                                                        │
│  (240px) │                                                        │
│          │                                                        │
│          │                                                        │
└──────────┴──────────────────────────────────────────────────────┘
```

- **Top bar** — Grimoire wordmark, context switcher (World Building |
  Session), current game name once selected, user avatar.
- **Left sidebar** — collapsible to icon-only (48px). Content is
  context-dependent (see below).
- **Main content area** — all screens render here.

### Sidebar Contents (World Building context, game selected)

```
▼ [Game Name]                       [⋯ game menu]
  ─────────────────────
  Overview
  ─────────────────────
  Narrative
  Locations
  Factions
  Characters
  MacGuffins
  ─────────────────────
  Campaigns
```

When no game is selected, the sidebar shows only the Games list.

---

## Screen Inventory

```
SCR-001  Games Dashboard         — list of all games, entry point
SCR-002  Create Game             — modal / inline form
SCR-003  Game Overview           — hub for a selected game
SCR-004  Locations               — hierarchical location tree
SCR-005  Location Detail         — draft/edit a single location + scenes
SCR-006  Narrative               — master beat DAG + beat list
SCR-007  Beat Detail             — edit beat content + prerequisites
SCR-008  Factions                — faction list with status
SCR-009  Faction Detail          — draft/edit faction, members, relationships
SCR-010  Characters              — NPC list + PC list tabs
SCR-011  NPC Detail              — draft/edit NPC, activate
SCR-012  MacGuffins              — macguffin list
SCR-013  MacGuffin Detail        — edit macguffin
SCR-014  Campaigns               — campaign list under a game
SCR-015  Campaign Setup          — create → form → add characters wizard
```

---

## SCR-001 — Games Dashboard

**Route:** `/`

**Purpose:** Entry point. Lists all games the GM owns.

### Layout

```
┌────────────────────────────────────────────────┐
│  My Games                         [+ New Game] │
├────────────────────────────────────────────────┤
│                                                 │
│  ┌─────────────────┐  ┌─────────────────┐      │
│  │ Ashes & Chains  │  │ The Sunken Keep │      │
│  │ Status: Active  │  │ Status: Draft   │      │
│  │ 2 Campaigns     │  │ 0 Campaigns     │      │
│  │ Last: 3 days ago│  │ Last: 1 week ago│      │
│  └─────────────────┘  └─────────────────┘      │
│                                                 │
│  ┌─────────────────┐                           │
│  │  + New Game     │                           │
│  └─────────────────┘                           │
└────────────────────────────────────────────────┘
```

Game cards show: name, lifecycle state badge, campaign count,
last activity date. Clicking a card navigates to SCR-003.

**Empty state:** Single centred card — "Create your first Game" with
a large `+` and a brief line explaining what a Game is.

**Queries needed (ADR-030):**
```graphql
query MyGames {
  games {
    id
    name
    status
    campaignCount
    lastActivityAt
  }
}
```

---

## SCR-002 — Create Game

**Trigger:** `[+ New Game]` button on SCR-001.

**Pattern:** Inline modal over the dashboard. Not a full page —
creating a game is a two-field operation.

### Layout

```
┌──────────────────────────────────────┐
│  Create New Game              [✕]    │
├──────────────────────────────────────┤
│                                      │
│  Game Name                           │
│  ┌────────────────────────────────┐  │
│  │ Ashes & Chains                 │  │
│  └────────────────────────────────┘  │
│  The name for this campaign setting. │
│                                      │
│              [Cancel]  [Create Game] │
└──────────────────────────────────────┘
```

**On submit:**
1. `createGame({ name })` → returns `{ id, status }`
2. Navigate to SCR-003 for the new game

No loading screen between creation and the game hub — navigate
immediately, the hub handles its own loading state.

**Mutations:**
- `createGame(input: CreateGameInput!)`

---

## SCR-003 — Game Overview

**Route:** `/games/:gameId`

**Purpose:** Hub for a selected game. Shows world health at a glance
and surfaces the next recommended action based on game state.

### Layout

```
┌─────────────────────────────────────────────────────────────┐
│  Ashes & Chains                              Status: Draft  │
├────────────────────┬────────────────────────────────────────┤
│                    │                                         │
│  Sidebar           │  ┌──────────────────────────────────┐  │
│  ─ Overview  ◀     │  │  Next Step                       │  │
│  ─ Narrative       │  │  Author your first story beat    │  │
│  ─ Locations       │  │  to activate the Master Narrative│  │
│  ─ Factions        │  │                [Go to Narrative] │  │
│  ─ Characters      │  └──────────────────────────────────┘  │
│  ─ MacGuffins      │                                         │
│  ─ Campaigns       │  World Summary                          │
│                    │  ┌─────────┐ ┌─────────┐ ┌─────────┐  │
│                    │  │Locations│ │Factions │ │  NPCs   │  │
│                    │  │  3 Draft│ │  0      │ │  0      │  │
│                    │  │  1 Active│ │        │ │         │  │
│                    │  └─────────┘ └─────────┘ └─────────┘  │
│                    │                                         │
│                    │  Campaigns                              │
│                    │  ┌─────────────────────────────────┐   │
│                    │  │  No campaigns yet               │   │
│                    │  │  [Create Campaign]              │   │
│                    │  └─────────────────────────────────┘   │
└────────────────────┴────────────────────────────────────────┘
```

**"Next Step" card** is state-driven. What it shows depends on
the game's lifecycle state and completeness:

| Game State | Condition | Message |
|------------|-----------|---------|
| New | No narrative | "Author your first beat to activate the Master Narrative" |
| Draft | No locations | "Build your first location" |
| Draft | Locations but no campaign | "Create a Campaign to start playing" |
| Draft | Campaign exists | "Set up your party in [Campaign Name]" |
| Active | Session in progress | "Session in progress → switch to Session mode" |
| Idle | Ready | "Ready for your next session" |

**Queries needed:**
```graphql
query GameOverview($id: ID!) {
  game(id: $id) {
    id
    name
    status
    locationSummary { draft active idle }
    factionSummary  { draft active idle }
    npcSummary      { draft active idle }
    campaigns {
      id
      name
      status
    }
  }
}
```

---

## SCR-004 — Locations

**Route:** `/games/:gameId/locations`

**Purpose:** Build and manage the world map. Hierarchical tree view
with inline status badges.

### Layout

```
┌────────────────────────────────────────────────────────────┐
│  Locations                              [+ Add Location]   │
├───────────────────────────┬────────────────────────────────┤
│                           │                                 │
│  TREE (left, ~320px)      │  DETAIL PANEL (right)          │
│                           │                                 │
│  ▼ 🌍 The Known World     │  Select a location to view     │
│     ▼ 📍 Blackburrow      │  or edit it.                   │
│       ▼ 🏘️ The Warrens   │                                 │
│           🚪 Guard Room   │                                 │
│           🚪 The Pit      │                                 │
│       📍 Freeport [Draft] │                                 │
│     📍 The Moors [Draft]  │                                 │
│                           │                                 │
│  [+ Add top-level]        │                                 │
└───────────────────────────┴────────────────────────────────┘
```

**Tree nodes** show: icon by LocationType, name, status badge
(Draft/Active/Idle/Archived). Archived locations are shown collapsed
and dimmed — visible for reference, not editable.

**Inline add:** Hovering a tree node reveals a `+` child button.
Clicking opens a popover: name field + LocationType selector.
Submits `createLocation` with `parentId` set.

**Selecting a node** loads Location Detail in the right panel.
On narrow screens the panel takes full width and the tree collapses.

**Queries needed:**
```graphql
query LocationTree($gameId: ID!) {
  locations(gameId: $gameId) {
    id
    name
    locationType
    status
    parentId
    sceneCount
    childCount
  }
}
```

---

## SCR-005 — Location Detail (right panel of SCR-004)

**Purpose:** Draft and edit a single location. Add scenes.
Connect to other locations. Activate.

### Layout — Draft State

```
┌─────────────────────────────────────────────────┐
│  Blackburrow                        [⋯ menu]    │
│  BUILDING  ·  Draft                             │
├─────────────────────────────────────────────────┤
│                                                  │
│  Name                                           │
│  ┌─────────────────────────────────────────┐   │
│  │ Blackburrow                             │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Scenes                              [+ Scene]  │
│  ┌─────────────────────────────────────────┐   │
│  │  Guard Room                        [✕]  │   │
│  │  The Pit                           [✕]  │   │
│  │  The Warden's Office               [✕]  │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Travel Connections          [+ Connect]        │
│  ┌─────────────────────────────────────────┐   │
│  │  → Freeport (one-way)                   │   │
│  │  ↔ The Warrens (both directions)        │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│                          [Activate Location ▶]  │
└─────────────────────────────────────────────────┘
```

**Activate button** is always visible but shows a tooltip if a
guard would block it ("Add at least one scene first"). Since the
domain removed the scene guard (ADR-018-Amendment-001), the button
is never blocked — this is informational only.

**`[+ Scene]`** opens an inline row in the Scenes list: name text
field + confirm. Submits `addScene`.

**`[+ Connect]`** opens a popover with a location search/picker and
a direction selector: `A → B`, `B → A`, `A ↔ B` (which calls
`connectLocations` twice for symmetric). Direction is a UI concept —
the domain only models directed edges.

**`[⋯ menu]`** contains: Archive Location (with a confirmation
modal warning that this is terminal and cascades to children).

**Layout — Active State**

Same as above but the Activate button is replaced by a status
indicator: `✓ Active`. The Archive option in `[⋯]` adds the
party-presence warning.

**Mutations:**
- `createLocation` (from tree inline add)
- `addScene`
- `connectLocations`
- `activateLocation`
- `archiveLocation`

**Queries needed:**
```graphql
query LocationDetail($id: ID!, $gameId: ID!) {
  location(id: $id, gameId: $gameId) {
    id
    name
    locationType
    status
    scenes { id name }
    connections {
      toLocation { id name }
      direction   # OUTBOUND | INBOUND | BOTH (read-model convenience)
    }
    parentLocation { id name }
  }
}
```

---

## SCR-006 — Narrative

**Route:** `/games/:gameId/narrative`

**Purpose:** Author the Master Narrative. Two views: DAG visualisation
and list view. Toggle between them.

### Layout

```
┌──────────────────────────────────────────────────────────────┐
│  Narrative                   [DAG View] [List View]  [+ Beat]│
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────────────────────────────────────────────┐   │
│  │                                                        │   │
│  │   [REQUIRED]          [REQUIRED]                       │   │
│  │  The Ring Found  →   Inquisitor       [OPTIONAL]      │   │
│  │                       Revealed    →  Road to Freeport  │   │
│  │                           ↓                            │   │
│  │                     [REQUIRED]                         │   │
│  │                    Enter Blackburrow                   │   │
│  │                                                        │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                               │
│  Beat Detail panel slides in from right when beat selected   │
└──────────────────────────────────────────────────────────────┘
```

**DAG view:** Beats rendered as nodes in a directed graph.
- Node colour by BeatType: Required (solid), Optional (outlined),
  Campaign-Specific (dashed)
- Arrows represent prerequisite relationships
- Clicking a node opens the Beat Detail panel (right slide-in)
- `[+ Beat]` opens an inline creation form: name + BeatType selector

**List view:** Flat sortable table — Name, Type, Status, Prerequisite
count. Useful when the DAG becomes complex. Same click → detail panel.

**Mutations:**
- `createMasterBeat`

**Queries needed:**
```graphql
query MasterNarrative($gameId: ID!) {
  masterNarrative(gameId: $gameId) {
    id
    beats {
      id
      name
      beatType
      status
      prerequisites { id name }
    }
  }
}
```

---

## SCR-007 — Beat Detail (slide-in panel from SCR-006)

**Purpose:** Edit beat content. Set prerequisite relationships.

### Layout

```
┌─────────────────────────────────────────────────┐
│  The Inquisitor Revealed          [✕ close]     │
│  REQUIRED  ·  Active                            │
├─────────────────────────────────────────────────┤
│                                                  │
│  Name                                           │
│  ┌─────────────────────────────────────────┐   │
│  │ The Inquisitor Revealed                 │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  GM Description (never shown to players)        │
│  ┌─────────────────────────────────────────┐   │
│  │ Korvan steps from the shadows, his      │   │
│  │ Inquisition badge catching the light... │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Player Description (shown on reveal)           │
│  ┌─────────────────────────────────────────┐   │
│  │ A figure in Inquisition robes steps     │   │
│  │ forward, face obscured...               │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Prerequisites                  [+ Add prereq]  │
│  ┌─────────────────────────────────────────┐   │
│  │  The Ring Found               [✕ remove]│   │
│  └─────────────────────────────────────────┘   │
│  Prerequisites use OR logic across sets,        │
│  AND logic within a set.                        │
│                                                  │
│                                  [Save Changes] │
└─────────────────────────────────────────────────┘
```

**`[+ Add prereq]`** opens a beat picker (search/select from existing
beats). Adding a prerequisite calls `addBeatPrerequisite`. If a cycle
would be created the server returns an error surfaced inline.

**Auto-save vs explicit save:** Explicit save button. Fields are
edited freely, saved as a batch via `updateBeatContent`. This avoids
multiple round trips during active editing.

**Mutations:**
- `updateBeatContent`
- `addBeatPrerequisite`

**Queries needed:**
```graphql
query BeatDetail($id: ID!, $gameId: ID!) {
  beat(id: $id, gameId: $gameId) {
    id
    name
    beatType
    description
    playerDescription
    prerequisites { id name beatType }
  }
}
```

---

## SCR-008 — Factions

**Route:** `/games/:gameId/factions`

**Purpose:** List all factions with status. Add new factions.

### Layout

```
┌──────────────────────────────────────────────────┐
│  Factions                         [+ Add Faction]│
├──────────────────────────────────────────────────┤
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  The Inquisition           Active  [→]   │   │
│  │  3 members  ·  4 standing levels         │   │
│  │  Allied with: The Crown                  │   │
│  │  At war with: The Free Compact           │   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  The Free Compact          Draft  [→]    │   │
│  │  0 members  ·  2 standing levels         │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

Clicking `[→]` or the card navigates to SCR-009.

**`[+ Add Faction]`** opens an inline modal — name only. Immediately
creates the faction and navigates to SCR-009 to continue setup.

**Mutations:**
- `createFaction`

**Queries needed:**
```graphql
query Factions($gameId: ID!) {
  factions(gameId: $gameId) {
    id
    name
    status
    memberCount
    standingLevelCount
    allies { id name }
    enemies { id name }
  }
}
```

---

## SCR-009 — Faction Detail

**Route:** `/games/:gameId/factions/:factionId`

**Purpose:** Draft and complete a faction — members, standing levels,
relationships. Activate when ready.

### Layout

```
┌─────────────────────────────────────────────────────────────┐
│  ← Factions                                                  │
│  The Inquisition                               [⋯ menu]    │
│  Draft                                                       │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Name                                                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ The Inquisition                                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Members                                     [+ Add Member] │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Korvan          Captain                   [✕ remove]│  │
│  │  Sister Maren    Interrogator              [✕ remove]│  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Standing Levels                           [+ Add Level]    │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Ordinal  Name          Threshold                    │  │
│  │  1        Unknown          0                    [✕]  │  │
│  │  2        Noticed         25                    [✕]  │  │
│  │  3        Marked         100                    [✕]  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Relationships                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Allied with:  The Crown               [+ Ally]      │  │
│  │  At war with:  The Free Compact        [+ Enemy]     │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│                                       [Activate Faction ▶] │
└─────────────────────────────────────────────────────────────┘
```

**Members picker:** Search NPCs (from the game's active/draft NPC list)
+ rank text field. Calls `addFactionMember`.

**Standing levels:** Inline editable rows. Each row calls
`addStandingLevel` on confirm. Ordinal is explicit (drag to reorder,
which updates ordinal values).

**Relationships:** Faction search/select picker for Ally and Enemy.
Calls `declareAlly` or `declareWar`.

**Mutations:**
- `beginFactionDraft` (auto-called on creation if not already Draft)
- `addFactionMember`
- `addStandingLevel`
- `declareAlly`
- `declareWar`
- `activateFaction`
- `archiveFaction` (via `[⋯ menu]`, terminal — confirmation required)

**Queries needed:**
```graphql
query FactionDetail($id: ID!, $gameId: ID!) {
  faction(id: $id, gameId: $gameId) {
    id
    name
    status
    members {
      npc { id name }
      rank
    }
    standingLevels { ordinal name threshold }
    allies  { id name }
    enemies { id name }
  }
}
```

---

## SCR-010 — Characters

**Route:** `/games/:gameId/characters`

**Purpose:** List NPCs and Player Characters in two tabs.

### Layout

```
┌──────────────────────────────────────────────────┐
│  Characters                                       │
├──────────────────────────────────────────────────┤
│  [NPCs]  [Player Characters]                      │
├──────────────────────────────────────────────────┤
│                                          [+ NPC] │
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  Korvan             Active  Inquisition  │   │
│  │  Sister Maren       Active  Inquisition  │   │
│  │  The Warden         Draft               │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

Status badges are coloured: Draft (grey), Active (green), Idle (amber),
Archived (dimmed, visually de-emphasised but present for reference).

Clicking a row navigates to SCR-011 (NPC) or SCR-012 variant (PC).

**Mutations:**
- `createNPC`
- `createPlayerCharacter`

**Queries needed:**
```graphql
query Characters($gameId: ID!) {
  npcs(gameId: $gameId) {
    id
    name
    status
    factionMemberships { faction { id name } rank }
  }
  playerCharacters(gameId: $gameId) {
    id
    name
    status
    ownerPlayerId
  }
}
```

---

## SCR-011 — NPC Detail

**Route:** `/games/:gameId/characters/npcs/:npcId`

**Purpose:** Draft, edit, and activate an NPC.

### Layout

```
┌─────────────────────────────────────────────────┐
│  ← Characters                                   │
│  Korvan                             [⋯ menu]    │
│  Active                                         │
├─────────────────────────────────────────────────┤
│                                                  │
│  Name                                           │
│  ┌─────────────────────────────────────────┐   │
│  │ Korvan                                  │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Description (GM-only)                         │
│  ┌─────────────────────────────────────────┐   │
│  │ A former Inquisitor who went rogue...   │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Player Description (shown on reveal)           │
│  ┌─────────────────────────────────────────┐   │
│  │ A tall figure in a weathered cloak...   │   │
│  └─────────────────────────────────────────┘   │
│                                                  │
│  Faction Memberships                            │
│  The Inquisition — Captain (via SCR-009)        │
│  (Membership is managed on the Faction screen)  │
│                                                  │
│                                  [Save Changes] │
│                                  [Activate ▶]   │
└─────────────────────────────────────────────────┘
```

**`[Activate ▶]`** calls `activateNPC`. Not shown when already Active.

**`[⋯ menu]`** contains: Mark Dormant (if Active), Reactivate (if Idle),
Archive NPC (terminal, confirmation required — warns MacGuffins will
drop to last known location).

**Mutations:**
- `beginNPCDraft` (auto-called on creation)
- `updateNPCContent`
- `activateNPC`
- `archiveNPC`

**Queries needed:**
```graphql
query NPCDetail($id: ID!, $gameId: ID!) {
  npc(id: $id, gameId: $gameId) {
    id
    name
    status
    description
    playerDescription
    factionMemberships {
      faction { id name }
      rank
    }
  }
}
```

---

## SCR-012 — MacGuffins

**Route:** `/games/:gameId/macguffins`

**Purpose:** List and create narratively significant items.

### Layout

```
┌──────────────────────────────────────────────────┐
│  MacGuffins                    [+ Add MacGuffin] │
├──────────────────────────────────────────────────┤
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  The Obsidian Ring       Possessed: Korvan│   │
│  │  The Broken Compass      Possessed: None  │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

Clicking a row opens the MacGuffin Detail panel (right slide-in,
same pattern as Beat Detail).

**MacGuffin Detail panel:**
- Name field
- Description (GM-only), Player Description
- Current possessor (NPC or Location — read-only, driven by events)
- `[Save Changes]`

**Mutations:**
- `createMacGuffin`
- `updateMacGuffinContent`

**Queries needed:**
```graphql
query MacGuffins($gameId: ID!) {
  macguffins(gameId: $gameId) {
    id
    name
    possessor {
      ... on NPC      { id name }
      ... on Location { id name }
    }
  }
}
```

---

## SCR-013 — Campaigns

**Route:** `/games/:gameId/campaigns`

**Purpose:** List campaigns under a game. Create new ones.

### Layout

```
┌──────────────────────────────────────────────────┐
│  Campaigns                       [+ New Campaign]│
├──────────────────────────────────────────────────┤
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  Table A — Friday Night                  │   │
│  │  Active  ·  4 characters  ·  12 sessions │   │
│  │                                     [→]  │   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│  ┌──────────────────────────────────────────┐   │
│  │  Table B — Sunday Afternoon              │   │
│  │  Forming  ·  2 characters  ·  0 sessions │   │
│  │                                     [→]  │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

**`[+ New Campaign]`** opens a modal — name field only. On create,
navigates to SCR-014 Campaign Setup wizard.

**Mutations:**
- `createCampaign`

**Queries needed:**
```graphql
query Campaigns($gameId: ID!) {
  campaigns(gameId: $gameId) {
    id
    name
    status
    characterCount
    sessionCount
  }
}
```

---

## SCR-014 — Campaign Setup Wizard

**Route:** `/games/:gameId/campaigns/:campaignId/setup`

**Purpose:** Walk the GM through the New → Forming → ready-to-play
lifecycle for a campaign. Three-step wizard.

### Step 1 — Campaign Created

```
┌─────────────────────────────────────────────────────┐
│  Setting up: Table A — Friday Night                 │
│                                                      │
│  ●───────────────○───────────────○                  │
│  Campaign        Party           Ready               │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │  Campaign Name                               │  │
│  │  ┌────────────────────────────────────────┐ │  │
│  │  │ Table A — Friday Night                 │ │  │
│  │  └────────────────────────────────────────┘ │  │
│  └──────────────────────────────────────────────┘  │
│                                                      │
│                          [Begin Party Formation ▶]  │
└─────────────────────────────────────────────────────┘
```

`[Begin Party Formation ▶]` calls `beginCampaignFormation`. On success,
advances to Step 2. The step indicator updates — first dot fills.

### Step 2 — Party Formation

```
┌─────────────────────────────────────────────────────┐
│  Setting up: Table A — Friday Night                 │
│                                                      │
│  ●───────────────●───────────────○                  │
│  Campaign        Party           Ready               │
│                                                      │
│  Add Player Characters                              │
│                                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │  Elowen Voss                          [✕]    │  │
│  │  Declan Ashford                       [✕]    │  │
│  └──────────────────────────────────────────────┘  │
│                                                      │
│  [+ Add Existing Character]  [+ Create New]         │
│                                                      │
│  Need at least 1 character to start playing.        │
│                                                      │
│  [← Back]                [Continue to Review ▶]    │
└─────────────────────────────────────────────────────┘
```

**`[+ Add Existing Character]`** — picker showing PlayerCharacters from
this game that are not yet assigned to another active campaign.
Calls `addCharacterToCampaign`.

**`[+ Create New]`** — inline form: character name + optional player
name. Calls `createPlayerCharacter` then `addCharacterToCampaign`.

`[Continue ▶]` is enabled once at least one character is linked.

### Step 3 — Ready

```
┌─────────────────────────────────────────────────────┐
│  Setting up: Table A — Friday Night                 │
│                                                      │
│  ●───────────────●───────────────●                  │
│  Campaign        Party           Ready               │
│                                                      │
│  ✓  Campaign created                               │
│  ✓  Party assembled (2 characters)                 │
│  ✓  Ready for Session 1                            │
│                                                      │
│  When you're ready to play, switch to              │
│  Session mode to start the first session.          │
│                                                      │
│  [Go to Campaign Overview]  [Start Session Now ▶]  │
└─────────────────────────────────────────────────────┘
```

`[Start Session Now ▶]` switches the context switcher to Session mode
and navigates to the Session start screen (UI-002).

**Mutations:**
- `beginCampaignFormation`
- `createPlayerCharacter`
- `addCharacterToCampaign`

**Queries needed:**
```graphql
query CampaignSetup($id: ID!, $gameId: ID!) {
  campaign(id: $id, gameId: $gameId) {
    id
    name
    status
    characters {
      id
      name
      ownerPlayerId
    }
  }
}
```

---

## Query Summary for ADR-030

All queries required by this document:

| Screen | Query |
|--------|-------|
| SCR-001 | `games` — id, name, status, campaignCount, lastActivityAt |
| SCR-003 | `game` — overview with aggregate summaries |
| SCR-004 | `locations` — id, name, locationType, status, parentId, counts |
| SCR-005 | `location` — detail with scenes, connections |
| SCR-006 | `masterNarrative` — beats with prerequisites |
| SCR-007 | `beat` — full content + prerequisites |
| SCR-008 | `factions` — list with member/level counts, relationships |
| SCR-009 | `faction` — full detail |
| SCR-010 | `npcs` + `playerCharacters` — list |
| SCR-011 | `npc` — full detail with faction memberships |
| SCR-012 | `macguffins` — list with possessor |
| SCR-013 | `campaigns` — list with counts |
| SCR-014 | `campaign` — setup state with characters |

---

## Related

- WORKFLOW-001 — World Building & Session Prep
- ADR-029 — GM Web Application Architecture
- ADR-030 — GraphQL Query Schema (driven by this document)
- UI-002 — Session Operation Screens (to be written)