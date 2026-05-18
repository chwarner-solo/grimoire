# WORKFLOW-002: Session Operation

## Context

The GM is running **an active session**. Players are at the table.
Time pressure is real. Operations must be fast — O(1) where possible,
never a DAG traversal mid-combat.

The GM controls what players see in real time via the information
boundary (`RevealEntity`). All events flow to Foundry VTT and Obsidian
via the sync broker (ADR-009).

Surfaces available in this context: Foundry VTT, web UI.

---

## Phase 1 — Start the Session

### 1.1 Start First Session (Campaign: Forming → Active)

The first session of a Campaign requires at least one Character linked.

```
StartFirstSession(campaignID, gameID, sessionID, date)
  → Campaign state: Forming → Active  [GUARD: ≥1 Character linked]
  → Session state: New → InProgress
  → Game state: Draft → Active
  → SessionStarted event dispatched
```

### 1.2 Start Subsequent Sessions (Campaign: Idle → Active)

```
StartNewSession(campaignID, gameID, sessionID, date)
  → Campaign state: Idle → Active
  → Session state: New → InProgress
  → Game state: Idle → Active  (if this was the only idle Campaign)
  → SessionStarted event dispatched
```

---

## Phase 2 — During the Session

### 2.1 Move the Party

As the party travels, the GM updates their current location.
Only valid during an active session.

```
MoveParty(campaignID, gameID, locationID)
  → Campaign.currentLocationID updated  [GUARD: Campaign is Active]
  → EntityUpdated event dispatched
  → Neo4j updates party position in graph
```

> Party presence at a Location is a guard on `ArchiveLocation`.
> A Location cannot be archived while the party is present — see ADR-018.

### 2.2 Discover a Beat

When the party encounters a narrative beat, the GM marks it discovered.
This is the real-time, O(1) path — no DAG traversal.

```
DiscoverBeat(campaignNarrativeID, beatID, gameID)
  → CampaignNarrative.DiscoverBeat(beat)
      checks prerequisiteSets against discoveredBeatIDs  [O(1)]
      → ErrPrerequisiteNotMet if not satisfied
  → CampaignNarrative saved
  → EntityLinked { relationship: "discovered" } dispatched
  → Neo4j writes (:Campaign)-[:DISCOVERED]->(:Beat) edge
```

> Prerequisite check is entirely local — CampaignNarrative holds
> `discoveredBeatIDs[]` and Beat holds `prerequisiteSets[][]`.
> Neo4j is never queried on the command path. See ADR-016.

### 2.3 Improvise a Campaign Beat

The party goes off-script. The GM creates a beat specific to this Campaign.

```
CreateCampaignBeat(name, beatType=CAMPAIGN_SPECIFIC, campaignNarrativeID)
  → Beat created, scoped to this Campaign only
  → Not visible to other Campaigns
  → May be promoted to MasterNarrative later (see Phase 3)
```

### 2.4 Reveal an Entity to Players

The GM controls the information boundary. When the party discovers
an NPC, Faction, MacGuffin, or PlayerCharacter, the GM explicitly reveals it.

```
RevealEntity(entityID, entityType, gameID, revealedTo[], sessionID)
  → EntityType: npc | player_character | macguffin | faction
  → Entity must be Active (NPC, Faction)  [GUARD: ErrInvalidEntityState if Draft]
  → EntityRevealed event dispatched
  → PlayerPushHandler  → real-time push to Player app
  → Neo4jHandler       → sets playerVisible: true on the node
  → SyncBroker         → pushes to Foundry / Obsidian (ADR-009)
```

> `revealedTo[]` is a list of player IDs. The Player app filters
> visibility per player — a reveal to one player is not a reveal to all.

### 2.5 NPC Lifecycle During Play

An NPC that was active can go dormant mid-campaign or be retired entirely.

```
MarkNPCDormant(npcID, gameID)
  → NPC state: Active → Idle
  → NPC is no longer relevant but not gone

ReactivateNPC(npcID, gameID)
  → NPC state: Idle → Active

RetireNPC(npcID, gameID)
  → NPC state: Active | Idle → Archived  (terminal)
  → EventBus: MacGuffins held by NPC drop to NPC's last known Location
```

### 2.6 Faction Events During Play

```
MarkFactionDormant(factionID, gameID)
  → Faction state: Active → Idle

ReactivateFaction(factionID, gameID)
  → Faction state: Idle → Active

DeclareAlly / DeclareWar (during play)
  → Faction relationship updated
  → Neo4j edge created/updated
```

---

## Phase 3 — End the Session

### 3.1 End the Session

```
EndSession(campaignID, gameID, sessionID, notes?)
  → Session state: InProgress → Completed
  → Campaign state: Active → Idle
  → Game state: Active → Idle  (if last active Campaign goes Idle)
  → SessionEnded event dispatched
  → Notes are optional at end time
```

---

## Phase 4 — Post-Session

Done at the table or shortly after. Campaign moves to Idle for the
duration between sessions.

### 4.1 Promote a Campaign Beat to Master

An improvised Campaign beat was good enough to be canonical world content.

```
PromoteBeatToMaster(beatID, gameID)
  → EntityUpdated { field: "scope",
                    old_value: "campaign:{campaignID}",
                    new_value: "master:{gameID}" }
  → FirestoreHandler  → updates Beat.scope field
  → Neo4jHandler      → moves node into master graph
  → Beat is now discoverable by all Campaigns
```

### 4.2 Retire a Player Character

A player leaves, or the campaign ends with a character retirement.

```
RetirePlayerCharacter(playerCharacterID, gameID)
  → PlayerCharacter state: Active → Retired  (terminal)
```

### 4.3 Complete a Campaign (Terminal)

The Campaign's story is done. This is a one-way door.

```
CompleteCampaign(campaignID, gameID)
  → Campaign state: Idle → Complete  [GUARD: must be Idle, not Active]
  → Campaign is terminal — no further sessions possible
```

---

## Phase 5 — Archive the World (Optional, Terminal)

When a Game setting is fully retired.

```
ArchiveGame(gameID)
  → Game state: Idle → Archived  [GUARD: must be Idle — no active Campaigns]
  → Terminal — no recovery
```

---

## Session Lifecycle Summary

```
StartFirst/NewSession
        ↓
    [session in progress]
        ↓  MoveParty
        ↓  DiscoverBeat
        ↓  CreateCampaignBeat
        ↓  RevealEntity
        ↓  NPC / Faction lifecycle changes
        ↓
  EndSession
        ↓
    [post-session]
        ↓  PromoteBeatToMaster
        ↓  RetirePlayerCharacter
        ↓  CompleteCampaign (terminal, if done)
```

---

## Performance Contracts

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| DiscoverBeat prerequisite check | O(1) | local slice comparison, no I/O |
| MoveParty | O(1) | single field update |
| RevealEntity | O(1) | single aggregate load + save |
| StartSession | O(1) | two repo loads (Game + Campaign) |
| AddBeatPrerequisite (prep only) | O(depth) | chain load, ~1-2ms at depth 10 |

Beat prerequisite cycle detection is **prep-time only**. Never runs during play.

---

## Information Boundary Rules

```
GM context   →  sees everything — all fields, all states, all beats
PC context   →  sees only entities where playerVisible: true
               sees discovered beats (playerDescription only)
               sees available beats as "There is something here..."
               never sees locked beats
               never sees Beat.description (GM-only field)
```

The boundary is enforced at the Neo4j query layer in grimoire-api.
The domain never filters by player visibility — that is a read-side concern.

---

## Related

- ADR-022 — Campaign Interactors
- ADR-023 — Narrative Interactors
- ADR-027 — RevealEntity Interactor
- ADR-009 — Bidirectional Event Mapping (Obsidian / Foundry sync)
- ADR-016 — Narrative Aggregate Architecture (prerequisite enforcement)
- WORKFLOW-001 — World Building & Session Prep