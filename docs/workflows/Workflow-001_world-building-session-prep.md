# WORKFLOW-001: World Building & Session Prep

## Context

The GM is working **outside of an active session**. No players are watching.
This is prep time — building the world, authoring the narrative, configuring
the campaign before play begins. The information boundary is fully under the
GM's control. Nothing is visible to players until explicitly revealed.

Surfaces available in this context: Obsidian, Foundry VTT, web UI.

---

## Phase 1 — Create the World

### 1.1 Create a Game

The Game is the container for all canonical world content.
One Game per campaign setting. Multiple Campaigns may run against the same Game.

```
CreateGame
  → GameID assigned
  → Game state: New
```

### 1.2 Attach a Master Narrative

The MasterNarrative is the GM's canonical story DAG — all beats, acts, secrets,
and lore that exist in the world, independent of any individual Campaign's path.

```
OnNarrativeCreated(MasterNarrativeID)
  → Game state: New → Draft
  → MasterNarrative created and linked to Game
```

### 1.3 Build the World Map

Locations are built top-down through the hierarchy:

```
World → Region → Settlement → Building → Scene
```

Each level follows the same lifecycle:

```
CreateLocation(name, locationType, parentID?)
  → Location state: New

BeginDraft
  → Location state: New → Draft
  → GM is building out scenes, connecting children

AddScene(sceneID)         — repeat for each Scene in this Location
AddChild(childLocationID) — repeat for each sub-Location

ConnectLocations(fromID, toID)
  → Directed travel connection (call twice for A↔B symmetric travel)

ActivateLocation
  → Location state: Draft → Active  [GUARD: at least one Scene defined]
  → Location is now part of the living world
  → Party can travel here
```

> **Note:** Location hierarchy validation is GM responsibility.
> The system does not enforce that a Settlement must sit under a Region.
> See ADR-018.

### 1.4 Author the Narrative DAG

Story beats are authored on the MasterNarrative and can be linked as a
directed acyclic graph (DAG) with prerequisite relationships.

```
CreateMasterBeat(name, beatType)
  → BeatType: REQUIRED | OPTIONAL | CAMPAIGN_SPECIFIC
  → Beat created, scoped to MasterNarrative

UpdateBeatContent(name, description, playerDescription)
  → description       — GM-only, never shown to players
  → playerDescription — player-facing, controlled by GM reveal

AddBeatPrerequisite(beatID, prerequisiteID)
  → prerequisiteSets enforced as: outer=OR, inner=AND
  → Cycle detection runs at add time (ADR-016)
  → Latency acceptable — GM is not in combat

AddActToMasterNarrative(actID)
AddSecretToMasterNarrative(secretID)
AddLoreToMasterNarrative(loreID)
```

> **Beat Types:**
> - `REQUIRED` — every Campaign must encounter this beat
> - `OPTIONAL` — some Campaigns will find this, others won't
> - `CAMPAIGN_SPECIFIC` — improvised during play; may be promoted to master later

### 1.5 Create Factions

Factions are groups with goals, allegiances, and presence in the world.

```
CreateFaction(name)
  → Faction state: New

BeginFactionDraft
  → Faction state: New → Draft

ActivateFaction
  → Faction state: Draft → Active  [GUARD: name not empty]

DeclareAlly(factionA, factionB)
DeclareWar(factionA, factionB)
  → Faction relationship edges in Neo4j
```

### 1.6 Create NPCs

NPCs are GM-controlled characters that inhabit the world.

```
CreateNPC(name)
  → NPC state: New

BeginNPCDraft
  → NPC state: New → Draft

ActivateNPC
  → NPC state: Draft → Active  [GUARD: name not empty]
  → NPC is now part of the living world
  → Can be revealed to players mid-session
```

### 1.7 Create MacGuffins

MacGuffins are narratively significant items whose possession drives story.

```
CreateMacGuffin(name)
  → MacGuffin created with no owner (or linked to NPC/Location)
```

---

## Phase 2 — Set Up a Campaign

A Campaign is one table's run through the Game world.

### 2.1 Create the Campaign

```
CreateCampaign(gameID, name)
  → Campaign state: New
  → Campaign linked to Game (EntityLinked event)
  → Game state: Draft → (stays Draft until first Session)
```

### 2.2 Begin Party Formation

GM kicks off the party-creation phase. Players submit their characters.

```
BeginCampaignFormation(campaignID, gameID)
  → Campaign state: New → Forming
```

### 2.3 Add Player Characters

Each player's character is added to the Campaign during the Forming state.

```
CreatePlayerCharacter(name, ownerPlayerID?)
  → PlayerCharacter state: Active (no Draft state — ADR-019)

AddCharacterToCampaign(campaignID, characterID)
  → EntityLinked event
```

> `ownerPlayerID` is optional — see ADR-019-Amendment-002.
> A character can exist without a linked player account.

---

## Phase 3 — Pre-Session Prep

With the world built and campaign created, the GM prepares for the next session.

### 3.1 Review Available Beats

Query the MasterNarrative DAG via Neo4j to see which beats are available
given the Campaign's current discoveredBeatIDs.

```cypher
-- Available beats for a Campaign
MATCH (c:Campaign {id: $campaignId})-[:DISCOVERED]->(discovered:Beat)
MATCH (candidate:Beat)
WHERE NOT (c)-[:DISCOVERED]->(candidate)
AND prerequisites satisfied...
RETURN candidate
```

### 3.2 Activate World Elements

Bring NPCs, Locations, and Factions that will feature in the next session
into Active state if they are still in Draft.

### 3.3 Review Party Knowledge

Query Neo4j for the player-facing information boundary — what the party
currently knows. Nothing with `playerVisible: false` is served to players.

---

## State Summary

| Aggregate | Lifecycle |
|-----------|-----------|
| Game | New → Draft → Active → Idle → Archived |
| Campaign | New → Forming → Active → Idle → Complete |
| Location | New → Draft → Active → Idle → Archived |
| Faction | New → Draft → Active → Idle → Archived |
| NPC | New → Draft → Active → Idle → Archived |
| PlayerCharacter | Active → Retired (no Draft) |

---

## Related

- ADR-015 / ADR-016 — Narrative DAG and Aggregate Architecture
- ADR-017 — Faction Aggregate
- ADR-018 — Location Aggregate
- ADR-019 — Character Aggregate
- ADR-021 — Game Interactors
- ADR-022 — Campaign Interactors
- WORKFLOW-002 — Session Operation