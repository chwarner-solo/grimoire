# UI-002: GM Application — Session Operation Screens

## Scope

This document covers **Session Operation** (WORKFLOW-002):
- Session Start
- At-Table Command Surface
- End Session
- Post-Session

Session screens are tablet-first, single-panel, optimised for speed.
No split panels, no deep navigation hierarchies, no hover states.
Every command the GM needs during live play is reachable in two taps.

---

## Navigation Model (Session Context)

The session context replaces the world-building shell entirely.
The left sidebar collapses. The top bar shrinks to a status strip.

```
┌─────────────────────────────────────────────────────────────┐
│  ◀ World Building    Table A — Friday Night    Session #7   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Full-width single panel content                            │
│                                                              │
│                                                              │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  [Party]      [Narrative]      [Entities]      [⏹ End]     │
└─────────────────────────────────────────────────────────────┘
```

- **Top strip** — back to World Building, campaign name, session number
- **Bottom tab bar** — persistent across all at-table screens (4 tabs)
- **No left sidebar** — reclaimed for content
- Content area is full width, scrollable vertically

The `[⏹ End]` tab is visually distinct (red-tinted) — it is a
destructive action anchor, not a content destination.

---

## Screen Inventory

```
SCR-015  Session Launcher          — pick campaign, confirm, start
SCR-016  Party Tab                 — location + characters, MoveParty
SCR-017  Narrative Tab             — beats, DiscoverBeat, improvise
SCR-018  Entities Tab              — RevealEntity, NPC/Faction lifecycle
SCR-019  End Session               — notes, confirm end
SCR-020  Post-Session Review       — promote beats, retire PCs, complete
```

---

## SCR-015 — Session Launcher

**Route:** `/games/:gameId/campaigns/:campaignId/session/start`

**Trigger:** "Start Session Now" from SCR-014 (wizard) or
"Start New Session" from the Campaign card in World Building.

**Purpose:** Confirm which campaign, confirm the date, start the
session. Intentionally minimal — GMs start sessions fast.

### Layout — First Session (Forming → Active)

```
┌────────────────────────────────────────────────┐
│                                                 │
│         🎲  Starting Session 1                 │
│                                                 │
│         Table A — Friday Night                 │
│         4 characters  ·  Ashes & Chains        │
│                                                 │
│         Date                                    │
│         ┌───────────────────────────────────┐  │
│         │  Monday 18 May 2026          [📅] │  │
│         └───────────────────────────────────┘  │
│         Defaults to today. Change if logging    │
│         a past session.                         │
│                                                 │
│         Party                                   │
│         ✓  Elowen Voss                         │
│         ✓  Declan Ashford                      │
│         ✓  Mira Tannek                         │
│         ✓  Jorath                              │
│                                                 │
│                                                 │
│         ┌───────────────────────────────────┐  │
│         │    ▶  Begin Session              │  │
│         └───────────────────────────────────┘  │
│                                                 │
└────────────────────────────────────────────────┘
```

### Layout — Subsequent Session (Idle → Active)

Same layout. Header shows "Starting Session 7". Party list shows
current characters. If any characters were retired since last session
they do not appear.

**On `[▶ Begin Session]`:**
1. `startFirstSession` or `startNewSession` depending on campaign state
2. Navigate immediately to SCR-016 (Party Tab)
3. Bottom tab bar appears — session is now live

**Error state:** If the guard fires (`ErrNoCharactersInCampaign`),
show inline error: "Add at least one character in World Building
before starting." Link back to campaign setup.

**Mutations:**
- `startFirstSession(input: StartFirstSessionInput!)`
- `startNewSession(input: StartNewSessionInput!)`

**Queries needed:**
```graphql
query SessionLauncher($campaignId: ID!, $gameId: ID!) {
  campaign(id: $campaignId, gameId: $gameId) {
    id
    name
    status
    sessionCount
    characters { id name status }
  }
}
```

---

## SCR-016 — Party Tab

**Route:** `/games/:gameId/campaigns/:campaignId/session/party`

**Purpose:** Where is the party? Who is in it? Move the party to a
new location. The GM's spatial anchor during play.

### Layout

```
┌──────────────────────────────────────────────────────────┐
│  ◀ World Building    Table A — Session #7    Active      │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  Current Location                                        │
│  ┌────────────────────────────────────────────────────┐ │
│  │                                                    │ │
│  │  📍 Blackburrow — The Warden's Office             │ │
│  │     Building · Active                              │ │
│  │                                                    │ │
│  │  Connected locations:                              │ │
│  │  → The Guard Room                                 │ │
│  │  → The Pit                                        │ │
│  │  → Freeport (one-way out)                         │ │
│  │                                                    │ │
│  └────────────────────────────────────────────────────┘ │
│                                    [Move Party →]        │
│                                                           │
│  Party                                                   │
│  ┌────────────────────────────────────────────────────┐ │
│  │  Elowen Voss          Active                  [⋯] │ │
│  │  Declan Ashford       Active                  [⋯] │ │
│  │  Mira Tannek          Active                  [⋯] │ │
│  │  Jorath               Active                  [⋯] │ │
│  └────────────────────────────────────────────────────┘ │
│                                                           │
├──────────────────────────────────────────────────────────┤
│  [Party] ●    [Narrative]    [Entities]    [⏹ End]      │
└──────────────────────────────────────────────────────────┘
```

**Current Location card** shows the active location with its directly
connected locations. Connected locations are tappable — tapping one
is a shortcut to `MoveParty` with a confirm step.

**`[Move Party →]`** opens a full-screen location picker modal:
search + filtered list of Active locations only. Draft and Idle
locations are greyed out with a tooltip "Activate this location in
World Building first." Selecting a location confirms and calls
`moveParty`.

### Move Party Modal

```
┌──────────────────────────────────────────────────┐
│  Move Party To…                           [✕]   │
├──────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────┐   │
│  │ 🔍  Search locations…                    │   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│  Connected (fastest)                             │
│  ┌──────────────────────────────────────────┐   │
│  │  The Guard Room             Building [→] │   │
│  │  The Pit                    Building [→] │   │
│  │  Freeport                  Settlement [→]│   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│  All Active Locations                            │
│  ┌──────────────────────────────────────────┐   │
│  │  The Moors                    Region [→] │   │
│  │  The Warrens                Building [→] │   │
│  └──────────────────────────────────────────┘   │
└──────────────────────────────────────────────────┘
```

Connected locations listed first — the most common case is the
party moving to an adjacent location. All Active locations below
for when the party teleports, fast-travels, or the GM needs to
jump ahead.

**`[⋯]` on Party members** opens a small action sheet:
- Retire Character (terminal — confirmation required, warns irreversible)

This is the only party-member action available mid-session. All
other character edits belong in World Building.

**Mutations:**
- `moveParty`
- `retirePlayerCharacter` (via character action sheet)

**Queries needed:**
```graphql
query PartyTab($campaignId: ID!, $gameId: ID!) {
  campaign(id: $campaignId, gameId: $gameId) {
    id
    currentLocation {
      id
      name
      locationType
      connections {
        toLocation { id name locationType status }
        direction
      }
    }
    characters { id name status }
  }
  locations(gameId: $gameId, status: ACTIVE) {
    id
    name
    locationType
  }
}
```

---

## SCR-017 — Narrative Tab

**Route:** `/games/:gameId/campaigns/:campaignId/session/narrative`

**Purpose:** Discover beats. Improvise campaign beats. See what the
party has already found. This is the GM's narrative control panel
during play. All operations are O(1) — no DAG traversal mid-session.

### Layout

```
┌──────────────────────────────────────────────────────────┐
│  ◀ World Building    Table A — Session #7    Active      │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  Available Beats                     [+ Improvise Beat]  │
│  ┌────────────────────────────────────────────────────┐ │
│  │  ● The Warden's Secret          REQUIRED     [✓]  │ │
│  │  ● Road to Freeport             OPTIONAL     [✓]  │ │
│  │  ● The Informant                OPTIONAL     [✓]  │ │
│  └────────────────────────────────────────────────────┘ │
│  These beats are unlocked given the party's              │
│  current discoveries.                                    │
│                                                           │
│  Discovered This Campaign              (7 total)  [▼]   │
│  ┌────────────────────────────────────────────────────┐ │
│  │  ✓ The Ring Found               Session 1         │ │
│  │  ✓ Enter Blackburrow            Session 3         │ │
│  │  ✓ The Inquisitor Revealed      Session 5         │ │
│  │  ...                                               │ │
│  └────────────────────────────────────────────────────┘ │
│                                                           │
│  This Session                                  (0 so far)│
│  ┌────────────────────────────────────────────────────┐ │
│  │  Nothing discovered yet this session.              │ │
│  └────────────────────────────────────────────────────┘ │
│                                                           │
├──────────────────────────────────────────────────────────┤
│  [Party]    [Narrative] ●    [Entities]    [⏹ End]      │
└──────────────────────────────────────────────────────────┘
```

**Available Beats** are loaded from Neo4j — beats whose prerequisites
are satisfied by `discoveredBeatIDs`. Each row shows name, type badge,
and a `[✓ Discover]` button.

**`[✓ Discover]` tapped** → confirmation sheet slides up:

```
┌──────────────────────────────────────────────────┐
│  Discover Beat?                                  │
│                                                   │
│  The Warden's Secret                            │
│  REQUIRED                                        │
│                                                   │
│  GM Description:                                 │
│  The Warden has been skimming Inquisition        │
│  funds for three years...                        │
│                                                   │
│  Player Description (shown on reveal):           │
│  Documents hidden in the desk suggest the        │
│  Warden has a secret arrangement...              │
│                                                   │
│  [Cancel]              [✓ Mark Discovered]       │
└──────────────────────────────────────────────────┘
```

Showing both descriptions before confirming lets the GM verify
they're marking the right beat and reminds them what the
players will see when/if it's revealed. Confirms via `discoverBeat`.

On success the beat moves from Available to "This Session" instantly
(optimistic update — Apollo cache update on mutation result).

**Error case — ErrPrerequisiteNotMet:** Shown inline below the beat
row with a warning icon. This should not happen (the available list
is already filtered) but can occur if state is stale. A `[Refresh]`
link refetches the available beats query.

**`[+ Improvise Beat]`** opens a modal:

```
┌──────────────────────────────────────────────────┐
│  Improvise a Beat                         [✕]   │
├──────────────────────────────────────────────────┤
│                                                   │
│  Name                                            │
│  ┌──────────────────────────────────────────┐   │
│  │ The Warden Reveals the Smuggler Route    │   │
│  └──────────────────────────────────────────┘   │
│                                                   │
│  This beat is scoped to this campaign only.      │
│  You can promote it to the Master Narrative      │
│  after the session.                              │
│                                                   │
│  [Cancel]                    [Create Beat]       │
└──────────────────────────────────────────────────┘
```

Name only — sparse is not errored (ADR-019-Amendment-001 principle).
Content is added post-session if the beat warrants it.
Calls `createCampaignBeat`. New beat appears in "This Session" list
with a `[Campaign]` scope badge.

**Discovered This Campaign** is collapsible — full history is
available but not in the way during active play.

**Mutations:**
- `discoverBeat`
- `createCampaignBeat`

**Queries needed:**
```graphql
query NarrativeTab($campaignId: ID!, $campaignNarrativeId: ID!, $gameId: ID!) {
  availableBeats(campaignNarrativeId: $campaignNarrativeId, gameId: $gameId) {
    id
    name
    beatType
    description
    playerDescription
  }
  campaignNarrative(id: $campaignNarrativeId, gameId: $gameId) {
    id
    discoveredBeats {
      id
      name
      beatType
      discoveredInSession
    }
    campaignBeats {
      id
      name
      scope
    }
  }
}
```

---

## SCR-018 — Entities Tab

**Route:** `/games/:gameId/campaigns/:campaignId/session/entities`

**Purpose:** Reveal entities to players. Manage NPC and Faction
lifecycle changes that happen during play. The GM's information
boundary control during live play.

### Layout

```
┌──────────────────────────────────────────────────────────┐
│  ◀ World Building    Table A — Session #7    Active      │
├──────────────────────────────────────────────────────────┤
│  [NPCs]  [Factions]  [MacGuffins]                        │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  NPCs                                    [+ Quick Add]   │
│  ┌────────────────────────────────────────────────────┐ │
│  │  Korvan           Active  👁 Revealed         [⋯] │ │
│  │  Sister Maren     Active  🔒 Hidden            [⋯] │ │
│  │  The Warden       Active  🔒 Hidden            [⋯] │ │
│  │  ──────────────────────────────────────────────    │ │
│  │  Idle NPCs                                (2) [▼]  │ │
│  └────────────────────────────────────────────────────┘ │
│                                                           │
├──────────────────────────────────────────────────────────┤
│  [Party]    [Narrative]    [Entities] ●    [⏹ End]      │
└──────────────────────────────────────────────────────────┘
```

NPCs are sorted: Active first, Idle collapsed at the bottom.
Archived NPCs are not shown — irrelevant during play.

Each NPC row shows: name, status, reveal state (👁 Revealed / 🔒 Hidden).

**`[⋯]` on an NPC row** opens an action sheet:

```
┌──────────────────────────────────────────┐
│  Sister Maren                            │
│  Active  ·  Hidden from players          │
├──────────────────────────────────────────┤
│                                          │
│  👁  Reveal to Players…                 │
│  💤  Mark Dormant                        │
│  🗃️  Archive NPC…                        │
│                                          │
└──────────────────────────────────────────┘
```

**"Reveal to Players…"** opens the Reveal sheet:

```
┌──────────────────────────────────────────┐
│  Reveal Sister Maren                     │
├──────────────────────────────────────────┤
│                                          │
│  Player Description:                    │
│  A composed woman in grey robes, her    │
│  voice precise and unhurried...         │
│                                          │
│  Reveal to:                             │
│  ☑  Elowen Voss                         │
│  ☑  Declan Ashford                      │
│  ☑  Mira Tannek                         │
│  ☑  Jorath                             │
│                                          │
│  All players selected by default.       │
│  Deselect to reveal to only some.       │
│                                          │
│  [Cancel]        [👁 Reveal Now]         │
└──────────────────────────────────────────┘
```

All players are selected by default — the common case is a full-party
reveal. Individual player reveals are available for the edge case
where only some players are present or the GM needs selective reveals.
Calls `revealEntity`.

On success, the NPC row updates instantly: 🔒 Hidden → 👁 Revealed.

**"Mark Dormant"** calls `archiveNPC` with the Mark Dormant variant
(no confirmation — easily reversed). Row moves to Idle section.

**"Archive NPC…"** opens a confirmation modal:

```
┌──────────────────────────────────────────┐
│  Archive Sister Maren?                   │
├──────────────────────────────────────────┤
│  This is permanent. She will no longer   │
│  appear in active play.                  │
│                                          │
│  Any MacGuffins she holds will drop      │
│  to her last known location.             │
│                                          │
│  [Cancel]         [Archive]              │
└──────────────────────────────────────────┘
```

**`[+ Quick Add]`** — create a brand-new NPC on the spot (name only,
immediately activates). For when the GM invents an NPC mid-session.
This is the mobile version of the draft-then-activate pattern
compressed to a single action. Calls `createNPC` + `beginNPCDraft`
+ `activateNPC` in sequence.

**Factions tab** (same screen, different sub-tab):

```
┌────────────────────────────────────────────────────┐
│  The Inquisition    Active  👁 Revealed       [⋯] │
│  The Free Compact   Active  🔒 Hidden          [⋯] │
└────────────────────────────────────────────────────┘
```

`[⋯]` action sheet: Reveal to Players, Mark Dormant, Declare
Ally/War, Archive.

**MacGuffins tab:**

```
┌────────────────────────────────────────────────────┐
│  The Obsidian Ring    Possessed: Korvan       [⋯] │
│  The Broken Compass   Possessed: None         [⋯] │
└────────────────────────────────────────────────────┘
```

`[⋯]` action sheet: Reveal to Players only — MacGuffin possession
changes are tracked via NPC archive (EventBus) not a direct mutation.

**Mutations:**
- `revealEntity`
- `archiveNPC` / `reactivateNPC`
- `archiveFaction` / `reactivateFaction`
- `declareAlly` / `declareWar`
- `createNPC` + `beginNPCDraft` + `activateNPC` (Quick Add)

**Queries needed:**
```graphql
query EntitiesTab($gameId: ID!, $campaignId: ID!) {
  npcs(gameId: $gameId, excludeArchived: true) {
    id
    name
    status
    playerVisible
    playerDescription
  }
  factions(gameId: $gameId, excludeArchived: true) {
    id
    name
    status
    playerVisible
  }
  macguffins(gameId: $gameId) {
    id
    name
    playerVisible
    possessor {
      ... on NPC      { id name }
      ... on Location { id name }
    }
  }
  campaign(id: $campaignId, gameId: $gameId) {
    characters { id name }
  }
}
```

---

## SCR-019 — End Session

**Trigger:** Tapping `[⏹ End]` in the bottom tab bar.

**Purpose:** Wrap the session. Optional notes. Confirm — this
transitions the Campaign from Active → Idle.

### Layout

```
┌──────────────────────────────────────────────────────────┐
│  ◀ World Building    Table A — Session #7    Active      │
├──────────────────────────────────────────────────────────┤
│                                                           │
│         ⏹  End Session #7?                              │
│                                                           │
│         Table A — Friday Night                           │
│         Started: Monday 18 May 2026                      │
│                                                           │
│  ─────────────────────────────────────────────────────   │
│                                                           │
│  This Session                                            │
│  ✓  Beats discovered:   2                               │
│  ✓  Entities revealed:  1                               │
│  ✓  Party moved to:     The Warden's Office             │
│                                                           │
│  ─────────────────────────────────────────────────────   │
│                                                           │
│  Session Notes  (optional — add now or skip)            │
│  ┌─────────────────────────────────────────────────┐   │
│  │ The party discovered the Warden's secret...    │   │
│  │                                                │   │
│  │                                                │   │
│  └─────────────────────────────────────────────────┘   │
│                                                           │
│  [← Keep Playing]              [⏹ End Session]         │
│                                                           │
├──────────────────────────────────────────────────────────┤
│  [Party]    [Narrative]    [Entities]    [⏹ End] ●      │
└──────────────────────────────────────────────────────────┘
```

**Session summary** — a brief recap of what happened this session
(beats discovered, entities revealed, final location). Gives the GM
a moment to confirm they haven't accidentally triggered End Session.

**Notes are optional** — matching the domain (`EndSessionInput.notes`
is nullable). The GM can add notes now or skip entirely. There is
no "summarise later" flow in this app — the Session aggregate's
Summarize/Close transitions are not surfaced in the GM PWA
(those are internal domain states; the PWA treats EndSession
as the terminal GM action).

**`[⏹ End Session]`** calls `endSession`. On success:
- Bottom tab bar disappears
- Context switches back to World Building
- Navigate to SCR-020 (Post-Session Review)

**`[← Keep Playing]`** navigates back to SCR-016 (Party Tab).

**Mutations:**
- `endSession(input: EndSessionInput!)`

**Queries needed:**
```graphql
query SessionSummary($campaignId: ID!, $gameId: ID!, $sessionId: ID!) {
  sessionSummary(sessionId: $sessionId, gameId: $gameId) {
    beatsDiscoveredCount
    entitiesRevealedCount
    finalLocation { id name }
  }
}
```

---

## SCR-020 — Post-Session Review

**Route:** `/games/:gameId/campaigns/:campaignId/post-session`

**Purpose:** Handle the post-session housekeeping that requires
deliberate decisions: promote improvised beats, retire characters,
and optionally complete the campaign.

This screen lives in the **World Building context** — the session
is over, the tab bar is gone, the sidebar is back. The GM has
time and space to think.

### Layout

```
┌────────────────────────────────────────────────────────────┐
│  ← Table A — Friday Night                                  │
│  Post-Session — Session #7                                 │
├────────────────────────────────────────────────────────────┤
│                                                             │
│  Improvised Beats                                          │
│  Review campaign beats from this session and decide        │
│  whether to promote them to the Master Narrative.          │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  The Warden Reveals the Smuggler Route               │ │
│  │  Campaign-specific · Session #7                      │ │
│  │                                                      │ │
│  │  [Keep Campaign-Only]    [↑ Promote to Master]       │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
│  ─────────────────────────────────────────────────────    │
│                                                             │
│  Party                                                     │
│  All characters are currently Active.                      │
│  Retire a character if a player is leaving the campaign.   │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Elowen Voss       Active                  [Retire] │ │
│  │  Declan Ashford    Active                  [Retire] │ │
│  │  Mira Tannek       Active                  [Retire] │ │
│  │  Jorath            Active                  [Retire] │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
│  ─────────────────────────────────────────────────────    │
│                                                             │
│  Campaign                                                  │
│  ┌──────────────────────────────────────────────────────┐ │
│  │  Table A — Friday Night  ·  Idle  ·  7 sessions     │ │
│  │                                                      │ │
│  │  Is the campaign's story complete?                   │ │
│  │  This is a one-way door.                             │ │
│  │                                                      │ │
│  │                    [Complete Campaign…]              │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
│                              [Done — Back to Campaign]     │
└────────────────────────────────────────────────────────────┘
```

### Promote Beat — confirmation sheet

Tapping `[↑ Promote to Master]`:

```
┌──────────────────────────────────────────────────┐
│  Promote to Master Narrative?                    │
├──────────────────────────────────────────────────┤
│  The Warden Reveals the Smuggler Route           │
│                                                   │
│  This beat will become part of the canonical     │
│  world story. Other campaigns running this       │
│  game can now discover it.                       │
│                                                   │
│  You should update the beat's content before     │
│  other campaigns encounter it.                   │
│                                                   │
│  [Cancel]         [↑ Promote to Master]          │
└──────────────────────────────────────────────────┘
```

On confirm: calls `promoteBeatToMaster`. The beat row updates to
show `Master Narrative` scope badge. A note appears: "Update this
beat's content in the Narrative screen."

### Retire Character — confirmation sheet

```
┌──────────────────────────────────────────────────┐
│  Retire Elowen Voss?                             │
├──────────────────────────────────────────────────┤
│  This is permanent. Elowen will no longer        │
│  appear as an active party member.               │
│                                                   │
│  [Cancel]              [Retire]                  │
└──────────────────────────────────────────────────┘
```

Calls `retirePlayerCharacter`. Row updates to Retired badge.

### Complete Campaign — confirmation modal (full-screen)

`[Complete Campaign…]` opens a full-screen modal, not a sheet —
this is the most consequential action in the system:

```
┌──────────────────────────────────────────────────────────┐
│                                                           │
│         ⚠️  Complete Campaign                           │
│                                                           │
│         Table A — Friday Night                           │
│         7 sessions  ·  4 characters                     │
│                                                           │
│         This cannot be undone.                          │
│                                                           │
│         The campaign's story is marked complete.        │
│         No further sessions can be run.                 │
│         The campaign will be preserved for reference.   │
│                                                           │
│         Type the campaign name to confirm:              │
│         ┌───────────────────────────────────────────┐  │
│         │                                           │  │
│         └───────────────────────────────────────────┘  │
│                                                           │
│         [Cancel]                                        │
│         [Complete Campaign]  ← enabled when name matches│
│                                                           │
└──────────────────────────────────────────────────────────┘
```

The confirm button is disabled until the GM types the campaign name
exactly. This is a deliberate friction pattern for a terminal,
irreversible action. Calls `completeCampaign` on match.

On success: navigate to `/games/:gameId` (Game Overview). The
campaign now shows as Complete in the Campaigns list.

**Mutations:**
- `promoteBeatToMaster`
- `retirePlayerCharacter`
- `completeCampaign`

**Queries needed:**
```graphql
query PostSession($campaignId: ID!, $gameId: ID!) {
  campaign(id: $campaignId, gameId: $gameId) {
    id
    name
    status
    sessionCount
    characters { id name status }
  }
  campaignBeatsThisSession(campaignId: $campaignId, gameId: $gameId) {
    id
    name
    scope
    createdInSession
  }
}
```

---

## UX Principles — Session Context

These rules govern every decision on the session screens:

**1. Two taps to any command.** MoveParty: tap tab → tap location.
DiscoverBeat: tap row → tap confirm. RevealEntity: tap `[⋯]` →
tap Reveal. Never deeper.

**2. Destructive actions have friction proportional to severity.**
Mark Dormant: no confirmation (reversible). Archive NPC: single
confirmation. Complete Campaign: name must be typed (irreversible,
game-wide consequence).

**3. No navigation away from the session.** The `[⋯ World Building]`
back link is always present but the GM is never pushed there
involuntarily. Session state persists if the GM switches to World
Building to check something.

**4. Errors are inline, not full-page.** ErrPrerequisiteNotMet,
network failures, stale state — all surface as inline banners with
a single retry action. The GM is never blocked by an error modal
during play.

**5. Optimistic updates where safe.** DiscoverBeat, RevealEntity,
MoveParty all update the UI before the server confirms. Rollback
on error. These are O(1) operations on a session the GM controls
— the error rate is effectively zero for correctly-sequenced actions.

---

## Query Summary for ADR-030 (Session Screens)

| Screen | Query |
|--------|-------|
| SCR-015 | `campaign` — status, sessionCount, characters |
| SCR-016 | `campaign.currentLocation` with connections; `locations(status: ACTIVE)` |
| SCR-017 | `availableBeats(campaignNarrativeId)`; `campaignNarrative.discoveredBeats` |
| SCR-018 | `npcs(excludeArchived)`; `factions(excludeArchived)`; `macguffins`; `campaign.characters` |
| SCR-019 | `sessionSummary` — beats, reveals, finalLocation |
| SCR-020 | `campaign` — status, sessionCount, characters; `campaignBeatsThisSession` |

---

## Related

- WORKFLOW-002 — Session Operation
- ADR-029 — GM Web Application Architecture (tablet-first constraint)
- ADR-027 — RevealEntity Interactor
- ADR-022 — Campaign Interactors
- ADR-023 — Narrative Interactors
- UI-001 — World Building Screens
- ADR-030 — GraphQL Query Schema (driven by UI-001 + UI-002)