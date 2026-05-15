# ADR-019: Character Aggregate Architecture

## Status
Accepted

## Date
2026-05-12

## Context

Characters are the people of the world — NPCs the GM controls and Player
Characters the players embody. They carry narrative identity, participate
in the story DAG, possess narratively significant items, and belong to
factions.

Three bounded contexts exist for a character. The domain filter is:

> Does this thing affect the story? If yes, Grimoire owns it.
> If it only affects mechanics, Foundry owns it.

```
Narrative Context   →  identity, descriptions, visibility, location,
                        faction membership, story beats
                        Grimoire owns — NPC aggregate + PlayerCharacter

Mechanical Context  →  stat blocks, HP, conditions, abilities, spells
                        Foundry owns entirely
                        Grimoire holds one correlation key only

Player Context      →  Grimoire cares only about narrative participation:
                        background beats, personal arc beats, MacGuffins
                        Foundry owns: exact GP, full inventory, spell slots
```

Grimoire does not model HP, inventory, wealth, or any mechanical state.
The one exception is the **MacGuffin** — a narratively significant item
that affects the story. The One Ring. The Seal of the Emperor. The stolen
ledger. Grimoire tracks who possesses a MacGuffin because possession
drives narrative consequences. It does not track torches, rope, or rations.

---

## Breaking Change — Campaign.characterIDs

`Campaign.campaignCore.characterIDs []identity.CharacterID` is retired.
Replace with `[]identity.PlayerCharacterID` throughout:

```
grimoire-domain/roots/campaign/entity/campaign.go
  campaignCore.characterIDs  []identity.CharacterID
                           → []identity.PlayerCharacterID

grimoire-domain/roots/campaign/entity/states.go
  NewCampaign.AddCharacter(id identity.CharacterID, ...)
           → AddCharacter(id identity.PlayerCharacterID, ...)
  FormingCampaign.AddCharacter(id identity.CharacterID, ...)
               → AddCharacter(id identity.PlayerCharacterID, ...)

grimoire-domain/roots/campaign/entity/handle.go
  identity.ParseCharacterID(e.EntityBID)
                          → identity.ParsePlayerCharacterID(e.EntityBID)

grimoire-domain/roots/campaign/entity/reconstitute.go
  CampaignSnapshot.CharacterIDs []identity.CharacterID
                              → []identity.PlayerCharacterID
```

Campaigns track Player Characters. NPCs are world-level entities and
are never members of a Campaign's character list.

Snapshot migration: existing CampaignSnapshot documents in Firestore use
string UUIDs for CharacterIDs. The string values are valid PlayerCharacterIDs
under the new type. Migration is a type-rename only — no data transformation.

---

## New Typed IDs

Add to `grimoire-domain/shared/identity/ids.go`:

```go
type NarrativeCharacterID        struct{ GrimoireID }
type PlayerCharacterID           struct{ GrimoireID }
type PlayerCharacterNarrativeID  struct{ GrimoireID }
type MacGuffinID                 struct{ GrimoireID }
```

Add constructor and parser functions for each following the existing pattern:

```go
func NewNarrativeCharacterID() NarrativeCharacterID
func ParseNarrativeCharacterID(s string) (NarrativeCharacterID, error)

func NewPlayerCharacterID() PlayerCharacterID
func ParsePlayerCharacterID(s string) (PlayerCharacterID, error)

func NewPlayerCharacterNarrativeID() PlayerCharacterNarrativeID
func ParsePlayerCharacterNarrativeID(s string) (PlayerCharacterNarrativeID, error)

func NewMacGuffinID() MacGuffinID
func ParseMacGuffinID(s string) (MacGuffinID, error)
```

`CharacterID` is retired. Remove it from ids.go after updating all call
sites (Campaign is the only current user — covered above).

**Naming rationale**: NarrativeCharacterID is the cross-context correlation
key used by FactionMembership and MacGuffin possession. NPC is the local
interface name within the character package. Both names are correct in
their respective contexts.

---

## ADR-016 Extension — BeatScopeCharacter

Add to `grimoire-domain/roots/narrative/value/beat_values.go`:

```go
const (
    BeatScopeMaster    BeatScope = "master"    // ADR-016 — unchanged
    BeatScopeCampaign  BeatScope = "campaign"  // ADR-016 — unchanged
    BeatScopeCharacter BeatScope = "character" // ADR-019
)

const (
    BeatTypeRequired         BeatType = "required"          // ADR-016 — unchanged
    BeatTypeOptional         BeatType = "optional"          // ADR-016 — unchanged
    BeatTypeCampaignSpecific BeatType = "campaign-specific" // ADR-016 — unchanged
    BeatTypeBackground       BeatType = "background"        // ADR-019 — pre-campaign backstory
    BeatTypePersonalArc      BeatType = "personal-arc"      // ADR-019 — forward-moving character beat
)
```

Add to `grimoire-domain/roots/narrative/entity/beat.go` — Beat struct
gains one new field. All existing fields unchanged.

```go
type Beat struct {
    // existing fields — unchanged
    id                identity.BeatID
    name              string
    description       string
    playerDescription string
    scope             value.BeatScope
    beatType          value.BeatType
    status            value.BeatStatus
    gameID            identity.GameID
    campaignID        identity.CampaignID        // IsZero() unless scope is campaign
    prerequisiteSets  [][]identity.BeatID

    // ADR-019 addition
    characterID       identity.PlayerCharacterID // IsZero() unless scope is character
}
```

Update `BeatSnapshot` and `ReconstituteBeat` — `CharacterID` field must
round-trip. Existing snapshots have zero-value CharacterID; this is valid.

New constructor:

```go
// CreateCharacterBeat constructs a Beat scoped to a specific PlayerCharacter.
// beatType must be BeatTypeBackground or BeatTypePersonalArc.
func CreateCharacterBeat(
    id          identity.BeatID,
    name        string,
    beatType    value.BeatType,
    gameID      identity.GameID,
    characterID identity.PlayerCharacterID,
    source      event.Source,
) (*Beat, []event.Event, error)
```

Guards:
- `id.IsZero()` → `ErrBeatIDRequired`
- `name` empty → `ErrBeatNameRequired`
- `gameID.IsZero()` → `ErrGameIDRequired`
- `characterID.IsZero()` → `ErrCharacterIDRequired`
- `beatType` not `BeatTypeBackground` or `BeatTypePersonalArc` → `ErrInvalidBeatTypeForCharacterScope`

Add `ErrCharacterIDRequired` and `ErrInvalidBeatTypeForCharacterScope`
to `grimoire-domain/roots/narrative/entity/narrative_errors.go`.

Cross-scope prerequisite references are valid. A character beat may have
a world beat as a prerequisite and vice versa. The DAG machinery
(`WouldCreateCycle`, `LoadPrerequisiteChain`) is unchanged.

---

## NPC Aggregate — Sealed Interfaces + Handle/Replay

NPCs follow the Game/Faction/Location state machine pattern exactly.
The type IS the status. No status field on the struct.

### Root Interface

The NPC sealed interface exposes `NPCName()` and `GameID()` getter methods
for parity with `Faction` (`FactionID(), FactionName(), GameID()`) and
`Location` (`LocationID(), LocationName(), GameID()`). Both are accessible
through `Snapshot()` but convenience getters are consistent with the pattern.

```go
// grimoire-domain/roots/character/entity/narrative_character.go

// NPC is the sealed interface for the NPC aggregate root.
// The unexported isNPC() marker prevents external implementations.
// Handle enables event replay — required for Bigtable scaling path (ADR-014).
type NPC interface {
    isNPC()
    NarrativeCharacterID() identity.NarrativeCharacterID
    NPCName()              string
    GameID()               identity.GameID
    Snapshot()             NPCSnapshot
    Handle(evt event.Event) (NPC, error)
}
```

### State Interfaces

All methods that emit events take `source event.Source`.
Methods that do not emit events omit source.

```go
type NewNPC interface {
    NPC
    BeginDraft(source event.Source) (DraftNPC, []event.Event, error)
}

type DraftNPC interface {
    NPC
    Activate(source event.Source) (ActiveNPC, []event.Event, error)
    UpdateContent(name, description, playerDescription string, source event.Source) (DraftNPC, []event.Event, error)
    AssignToLocation(id identity.LocationID, source event.Source) (DraftNPC, []event.Event, error)
    // AssignToLocation emits EntityLinked { relationship: "located_at" }
}

type ActiveNPC interface {
    NPC
    UpdateContent(name, description, playerDescription string, source event.Source) (ActiveNPC, []event.Event, error)
    MoveTo(id identity.LocationID, source event.Source) (ActiveNPC, []event.Event, error)
    // MoveTo emits EntityUpdated { field: "location" }
    //        + EntityLinked      { relationship: "located_at" }
    GoIdle(source event.Source) (IdleNPC, []event.Event, error)
    Archive(source event.Source) (ArchivedNPC, []event.Event, error)
    Reveal(to []string, sessionID identity.SessionID, source event.Source) (ActiveNPC, []event.Event, error)
    AssignMacGuffin(id identity.MacGuffinID, source event.Source) (ActiveNPC, []event.Event, error)
    // AssignMacGuffin emits EntityLinked { relationship: "possesses" }
}

type IdleNPC interface {
    NPC
    Activate(source event.Source) (ActiveNPC, []event.Event, error)
    Archive(source event.Source) (ArchivedNPC, []event.Event, error)
}

type ArchivedNPC interface {
    NPC
    // terminal — no transition methods
}
```

### NPC Transition Guards

```
New    → Draft:    always allowed
Draft  → Active:   GUARD: name not empty
                   GUARD: description not empty
Active → Idle:     always allowed (GM decision)
Idle   → Active:   always allowed (GM reactivates)
Idle   → Archived: always allowed (terminal)
Active → Archived: always allowed (terminal)
Archived → *:      no exit — terminal
```

### NPC Archive — MacGuffin Drop

When an NPC is archived, possessed MacGuffins drop to its last known
location. Handled by `NPCArchivedHandler` via EventBus — aggregates
do not mutate other aggregates.

```
NPCArchivedHandler:
  1. Load all MacGuffins via MacGuffinRepository.FindByNPCPossessor(npc.id)
  2. For each: macguffin.Drop(npc.locationID, source)
  3. Save each MacGuffin
  4. Emits EntityUpdated { entity_type: "macguffin", field: "possessor",
                           old_value: npc_id, new_value: "" }
       EntityLinked { relationship: "located_at",
                      entity_a: macguffin_id, entity_b: npc.locationID }
```

### NPC Core Struct

```go
// unexported — callers hold the sealed interface
type npcCore struct {
    id                 identity.NarrativeCharacterID
    gameID             identity.GameID

    // Content — validated by aggregate, served by Neo4j
    name               string
    description        string   // GM only, NEVER shown to players
    playerDescription  string   // shown on EntityRevealed

    // Visibility
    playerVisible      bool

    // Relationships — references only
    locationID         identity.LocationID
    macGuffinIDs       []identity.MacGuffinID

    // Foundry correlation — the ONLY mechanical coupling in this aggregate
    foundryCharacterID string   // zero for NPCs with no Foundry representation
}
```

Getter implementations on `npcCore` (embedded by all state structs):

```go
func (n *npcCore) NarrativeCharacterID() identity.NarrativeCharacterID { return n.id }
func (n *npcCore) NPCName() string                                      { return n.name }
func (n *npcCore) GameID() identity.GameID                              { return n.gameID }
```

### NPC Constructor

```go
// CreateNPC constructs a new NPC aggregate in the New state.
// name is required at construction — consistent with all other aggregates.
// description and playerDescription are set later via DraftNPC.UpdateContent.
func CreateNPC(
    id     identity.NarrativeCharacterID,
    name   string,
    gameID identity.GameID,
    source event.Source,
) (NewNPC, []event.Event, error) {
    if id.IsZero() {
        return nil, nil, ErrNPCIDRequired
    }
    if strings.TrimSpace(name) == "" {
        return nil, nil, ErrNPCNameRequired
    }
    if gameID.IsZero() {
        return nil, nil, ErrGameIDRequired
    }
    evt := event.EntityCreated{
        EntityID:   id.String(),
        EntityType: "npc",
        Name:       name,
        Source:     source,
    }
    return &newNPC{npcCore: npcCore{
        id:     id,
        name:   name,
        gameID: gameID,
    }}, []event.Event{evt}, nil
}
```

Add `ErrNPCIDRequired` and `ErrNPCNameRequired` to
`narrative_character_errors.go`.

The `Draft → Active` guard on name non-empty stands as a secondary check —
UpdateContent in Draft can change the name, so the guard remains valid.

### Handle and Replay

```go
// grimoire-domain/roots/character/entity/narrative_character_handle.go

// Handle implementations for all five NPC states — follow the Faction pattern.
func (n *newNPC) Handle(evt event.Event) (NPC, error)
func (n *draftNPC) Handle(evt event.Event) (NPC, error)
func (n *activeNPC) Handle(evt event.Event) (NPC, error)
func (n *idleNPC) Handle(evt event.Event) (NPC, error)
func (n *archivedNPC) Handle(evt event.Event) (NPC, error)

// copyCore helper — follow the factionCore.copyCore() pattern.
func (n *npcCore) copyCore() npcCore

// ReplayNPC rebuilds an NPC aggregate from a sequence of events.
// The first event must be an EntityCreated for an npc.
// gameID is required because it is not stored in the event payload —
// consistent with ReplayFaction(gameID, events), ReplayCampaign(gameID, events),
// ReplaySession(campaignID, events).
func ReplayNPC(gameID identity.GameID, events []event.Event) (NPC, error) {
    if len(events) == 0 {
        return nil, errors.New("npc: no events to replay")
    }
    first, ok := events[0].(event.EntityCreated)
    if !ok {
        return nil, errors.New("npc: first event must be EntityCreated")
    }
    npcID, err := identity.ParseNarrativeCharacterID(first.EntityID)
    if err != nil {
        return nil, fmt.Errorf("npc replay: %w", err)
    }
    n, _, err := CreateNPC(npcID, first.Name, gameID, event.SourceGrimoire)
    if err != nil {
        return nil, fmt.Errorf("npc replay: %w", err)
    }
    var current NPC = n
    for _, evt := range events[1:] {
        current, err = current.Handle(evt)
        if err != nil {
            return nil, fmt.Errorf("npc replay: %w", err)
        }
    }
    return current, nil
}
```

### NPCSnapshot / Reconstitute

```go
// grimoire-domain/roots/character/entity/narrative_character_snapshot.go
// ReplayNPC is NOT here — it lives in narrative_character_handle.go.

type NPCSnapshot struct {
    ID                 identity.NarrativeCharacterID
    GameID             identity.GameID
    State              string   // "new" | "draft" | "active" | "idle" | "archived"
    Name               string
    Description        string
    PlayerDescription  string
    PlayerVisible      bool
    LocationID         identity.LocationID
    MacGuffinIDs       []identity.MacGuffinID
    FoundryCharacterID string
}

func (n *npcCore) Snapshot(state string) NPCSnapshot
func ReconstituteNPC(snap NPCSnapshot) (NPC, error)
```

---

## PlayerCharacter — Snapshot/Reconstitute Pattern

PlayerCharacter does NOT use sealed interfaces or Handle/Replay.
The lifecycle (Active → Retired) is too simple to justify the pattern.
PlayerCharacter uses Snapshot/Reconstitute only, consistent with Scene
and CampaignNarrative.

Document this departure with a comment on the struct:

```go
// PlayerCharacter uses Snapshot/Reconstitute, not the Handle/Replay pattern
// of Game/Campaign/Faction/Location/NPC. The Active → Retired lifecycle
// does not require event replay for reconstruction. Intentional — see ADR-019.
```

```go
// grimoire-domain/roots/character/entity/player_character.go

type PlayerCharacterStatus string

const (
    PlayerCharacterStatusActive  PlayerCharacterStatus = "active"
    PlayerCharacterStatusRetired PlayerCharacterStatus = "retired"
)

type PlayerCharacter struct {
    id                  identity.PlayerCharacterID
    gameID              identity.GameID

    name                string
    description         string   // GM only — backstory notes, secrets
    playerDescription   string   // the character's public face

    playerVisible       bool     // false until revealed to other players
    ownerPlayerID       string   // always sees full detail regardless of playerVisible

    status              PlayerCharacterStatus

    campaignIDs         []identity.CampaignID    // campaigns this PC has participated in
    macGuffinIDs        []identity.MacGuffinID

    foundryCharacterID  string   // the ONLY mechanical coupling
}
```

### PlayerCharacter Constructor

```go
func CreatePlayerCharacter(
    id            identity.PlayerCharacterID,
    gameID        identity.GameID,
    name          string,
    ownerPlayerID string,
    source        event.Source,
) (*PlayerCharacter, []event.Event, error)
```

Guards:
- `id.IsZero()` → `ErrPlayerCharacterIDRequired`
- `gameID.IsZero()` → `ErrGameIDRequired`
- `name` empty → `ErrCharacterNameRequired`
- `ownerPlayerID` empty → `ErrOwnerPlayerIDRequired`

PlayerCharacter is Active on creation. No Draft state.

### PlayerCharacter Methods

All methods that emit events take `source event.Source`.

```go
func (pc *PlayerCharacter) Retire(source event.Source) (*PlayerCharacter, []event.Event, error)
// Guard: status == Retired → ErrCharacterAlreadyRetired
// Emits EntityUpdated { field: "status", new_value: "retired" }

func (pc *PlayerCharacter) UpdateContent(name, description, playerDescription string, source event.Source) (*PlayerCharacter, []event.Event, error)
// Emits EntityUpdated { field: "content" }

func (pc *PlayerCharacter) AssignMacGuffin(id identity.MacGuffinID, source event.Source) (*PlayerCharacter, []event.Event, error)
// Guard: status == Retired → ErrCharacterRetired
// Emits EntityLinked { relationship: "possesses" }

func (pc *PlayerCharacter) ReleaseMacGuffin(id identity.MacGuffinID, source event.Source) (*PlayerCharacter, []event.Event, error)
// Guard: id not in macGuffinIDs → ErrMacGuffinNotPossessed
// Emits EntityUpdated { field: "macguffin_released" }

func (pc *PlayerCharacter) JoinCampaign(id identity.CampaignID, source event.Source) (*PlayerCharacter, []event.Event, error)
// Guard: id already in campaignIDs → ErrAlreadyInCampaign
// Emits EntityLinked { relationship: "participates_in" }

func (pc *PlayerCharacter) Reveal(to []string, sessionID identity.SessionID, source event.Source) (*PlayerCharacter, []event.Event, error)
// Sets playerVisible true
// Emits EntityRevealed { entity_type: "player_character" }
```

### PlayerCharacter Snapshot/Reconstitute

```go
// grimoire-domain/roots/character/entity/player_character_snapshot.go

type PlayerCharacterSnapshot struct {
    ID                  identity.PlayerCharacterID
    GameID              identity.GameID
    Name                string
    Description         string
    PlayerDescription   string
    PlayerVisible       bool
    OwnerPlayerID       string
    Status              PlayerCharacterStatus
    CampaignIDs         []identity.CampaignID
    MacGuffinIDs        []identity.MacGuffinID
    FoundryCharacterID  string
}

func (pc *PlayerCharacter) Snapshot() PlayerCharacterSnapshot
func ReconstitutePlayerCharacter(snap PlayerCharacterSnapshot) *PlayerCharacter
```

---

## PlayerCharacterNarrative — Additive Journal

Analogous to CampaignNarrative (ADR-016). No state machine. No Handle/Replay.
Purely additive. Lifecycle follows PlayerCharacter.

```go
// grimoire-domain/roots/character/entity/player_character_narrative.go

// PlayerCharacterNarrative is a purely additive event journal.
// No state machine. No Handle/Replay. Snapshot/Reconstitute only.
// Intentional departure from Game/Campaign/Session pattern —
// consistent with CampaignNarrative (ADR-016). See ADR-019.
type PlayerCharacterNarrative struct {
    id                      identity.PlayerCharacterNarrativeID
    characterID             identity.PlayerCharacterID
    campaignID              identity.CampaignID
    gameID                  identity.GameID

    // Backstory — pre-campaign beats. May be hidden from the owning player.
    backgroundBeatIDs       []identity.BeatID

    // Personal arc — forward-moving beats discovered during play.
    personalBeatIDs         []identity.BeatID

    // Background beats the owning player has been shown.
    // A player may not know their own history at campaign start.
    revealedBackgroundIDs   []identity.BeatID
}
```

### PlayerCharacterNarrative Methods

All methods that emit events take `source event.Source`.

```go
func CreatePlayerCharacterNarrative(
    id          identity.PlayerCharacterNarrativeID,
    characterID identity.PlayerCharacterID,
    campaignID  identity.CampaignID,
    gameID      identity.GameID,
    source      event.Source,
) (*PlayerCharacterNarrative, []event.Event, error)

func (pcn *PlayerCharacterNarrative) AddBackgroundBeat(
    beatID identity.BeatID,
    source event.Source,
) (*PlayerCharacterNarrative, []event.Event, error)
// Guard: beatID already in backgroundBeatIDs → ErrBackgroundBeatAlreadyAdded
// Emits EntityLinked { relationship: "has_background_beat" }

func (pcn *PlayerCharacterNarrative) DiscoverPersonalBeat(
    beatID            identity.BeatID,
    prerequisiteSets  [][]identity.BeatID,
    source            event.Source,
) (*PlayerCharacterNarrative, []event.Event, error)
// prerequisiteSets is Beat.PrerequisiteSets() — the interactor loads the
// Beat and passes the sets in. Cross-aggregate imports are forbidden.
// Guard: prerequisites not met → ErrPrerequisitesNotMet
// Emits EntityLinked { relationship: "discovered_personal_beat" }

func (pcn *PlayerCharacterNarrative) RevealBackgroundBeat(
    beatID    identity.BeatID,
    to        []string,
    sessionID identity.SessionID,
    source    event.Source,
) (*PlayerCharacterNarrative, []event.Event, error)
// Guard: beatID not in backgroundBeatIDs → ErrBackgroundBeatNotFound
// Guard: beatID already in revealedBackgroundIDs → ErrBackgroundBeatAlreadyRevealed
// Appends to revealedBackgroundIDs
// Emits EntityRevealed { entity_type: "character_background_beat", entity_id: beatID }

// allDiscoveredBeatIDs returns union of personalBeatIDs + revealedBackgroundIDs.
// Used internally by prerequisite check. Unexported.
func (pcn *PlayerCharacterNarrative) allDiscoveredBeatIDs() []identity.BeatID
```

### Background Beat Visibility

```
backgroundBeatIDs:        [beat_village_burned, beat_trained_by_velleth]
revealedBackgroundIDs:    []   ← player sees nothing yet

GM plays the confrontation scene:
  RevealBackgroundBeat(beat_village_burned, ...)
  → revealedBackgroundIDs: [beat_village_burned]
  → player now sees this beat in their character history
```

### PlayerCharacterNarrative Snapshot/Reconstitute

```go
// grimoire-domain/roots/character/entity/player_character_narrative_snapshot.go

type PlayerCharacterNarrativeSnapshot struct {
    ID                     identity.PlayerCharacterNarrativeID
    CharacterID            identity.PlayerCharacterID
    CampaignID             identity.CampaignID
    GameID                 identity.GameID
    BackgroundBeatIDs      []identity.BeatID
    PersonalBeatIDs        []identity.BeatID
    RevealedBackgroundIDs  []identity.BeatID
}

func (pcn *PlayerCharacterNarrative) Snapshot() PlayerCharacterNarrativeSnapshot
func ReconstitutePlayerCharacterNarrative(snap PlayerCharacterNarrativeSnapshot) *PlayerCharacterNarrative
```

---

## MacGuffin Aggregate — Snapshot/Reconstitute Pattern

MacGuffin does NOT use sealed interfaces or Handle/Replay.
No state machine. Snapshot/Reconstitute only, consistent with Scene
and PlayerCharacter.

Document this departure with a comment on the struct:

```go
// MacGuffin uses Snapshot/Reconstitute, not Handle/Replay.
// No state machine — possession and location are the only mutable state.
// Intentional — see ADR-019.
```

### Possession Model

A MacGuffin is possessed by an NPC, possessed by a PC, at a location,
or lost (all zero). NPC and PC possession are tracked with separate
typed fields. At most one of the three fields is non-zero.

```go
// grimoire-domain/roots/character/entity/macguffin.go

// MacGuffin is a narratively significant item.
// Grimoire tracks MacGuffins because possession drives story consequences.
// Grimoire does NOT track torches, rope, rations, or non-narrative items.
type MacGuffin struct {
    id                      identity.MacGuffinID
    gameID                  identity.GameID

    name                    string
    description             string   // GM only
    playerDescription       string   // shown on EntityRevealed

    playerVisible           bool
    destroyed               bool     // terminal flag — no full state machine needed

    // Possession — at most one of the three is non-zero.
    // All zero = lost / unknown location.
    npcPossessorID          identity.NarrativeCharacterID
    pcPossessorID           identity.PlayerCharacterID
    locationID              identity.LocationID
}
```

### Possession Invariant

Enforced on every state-mutating method before applying the change:

```go
func (m *MacGuffin) validatePossessionInvariant() error {
    owned := 0
    if !m.npcPossessorID.IsZero() { owned++ }
    if !m.pcPossessorID.IsZero()  { owned++ }
    if !m.locationID.IsZero()     { owned++ }
    if owned > 1 {
        return ErrMacGuffinPossessionConflict
    }
    return nil
}
```

### MacGuffin Constructor

```go
func CreateMacGuffin(
    id     identity.MacGuffinID,
    gameID identity.GameID,
    name   string,
    source event.Source,
) (*MacGuffin, []event.Event, error)
```

Guards:
- `id.IsZero()` → `ErrMacGuffinIDRequired`
- `gameID.IsZero()` → `ErrGameIDRequired`
- `name` empty → `ErrMacGuffinNameRequired`

New MacGuffin starts unowned (all possession fields zero).

### MacGuffin Methods

All methods that emit events take `source event.Source`.
All methods guard against `destroyed == true` → `ErrMacGuffinDestroyed`.

```go
func (m *MacGuffin) PlaceAt(locationID identity.LocationID, source event.Source) (*MacGuffin, []event.Event, error)
// Clears npcPossessorID and pcPossessorID, sets locationID
// Guard: locationID.IsZero() → ErrLocationIDRequired
// Emits EntityLinked { relationship: "located_at" }

func (m *MacGuffin) TransferToNPC(id identity.NarrativeCharacterID, source event.Source) (*MacGuffin, []event.Event, error)
// Clears locationID and pcPossessorID, sets npcPossessorID
// Guard: id.IsZero() → ErrCharacterIDRequired
// Emits EntityLinked { relationship: "possesses",
//                       entity_a: npc_id, entity_b: macguffin_id }

func (m *MacGuffin) TransferToPC(id identity.PlayerCharacterID, source event.Source) (*MacGuffin, []event.Event, error)
// Clears locationID and npcPossessorID, sets pcPossessorID
// Guard: id.IsZero() → ErrCharacterIDRequired
// Emits EntityLinked { relationship: "possesses",
//                       entity_a: pc_id, entity_b: macguffin_id }

func (m *MacGuffin) Drop(locationID identity.LocationID, source event.Source) (*MacGuffin, []event.Event, error)
// Clears npcPossessorID and pcPossessorID, sets locationID
// Separate from PlaceAt for semantic clarity — called by NPCArchivedHandler
// Emits EntityUpdated { field: "possessor", old_value: <id>, new_value: "" }
//       EntityLinked  { relationship: "located_at" }

func (m *MacGuffin) Reveal(to []string, sessionID identity.SessionID, source event.Source) (*MacGuffin, []event.Event, error)
// Sets playerVisible true
// Emits EntityRevealed { entity_type: "macguffin" }

func (m *MacGuffin) UpdateContent(name, description, playerDescription string, source event.Source) (*MacGuffin, []event.Event, error)
// Emits EntityUpdated { field: "content" }

func (m *MacGuffin) Destroy(source event.Source) (*MacGuffin, []event.Event, error)
// Sets destroyed = true. Terminal.
// Guard: already destroyed → ErrMacGuffinAlreadyDestroyed
// Clears all possession and location fields
// Emits EntityUpdated { field: "status", new_value: "destroyed" }
```

### MacGuffin Snapshot/Reconstitute

```go
// grimoire-domain/roots/character/entity/macguffin_snapshot.go

type MacGuffinSnapshot struct {
    ID                  identity.MacGuffinID
    GameID              identity.GameID
    Name                string
    Description         string
    PlayerDescription   string
    PlayerVisible       bool
    Destroyed           bool
    NPCPossessorID      identity.NarrativeCharacterID
    PCPossessorID       identity.PlayerCharacterID
    LocationID          identity.LocationID
}

func (m *MacGuffin) Snapshot() MacGuffinSnapshot
func ReconstituteMacGuffin(snap MacGuffinSnapshot) *MacGuffin
```

---

## Character Types Value Object

```go
// grimoire-domain/roots/character/value/character_type.go

type CharacterType string

const (
    CharacterTypeNPC             CharacterType = "npc"
    CharacterTypePlayerCharacter CharacterType = "player-character"
)
```

---

## Port Definitions

```go
// grimoire-domain/roots/character/port/npc_repository.go
type NPCRepository interface {
    Save(ctx context.Context, snap entity.NPCSnapshot) error
    Load(ctx context.Context, id identity.NarrativeCharacterID) (entity.NPCSnapshot, error)
    FindByLocation(ctx context.Context, id identity.LocationID) ([]entity.NPCSnapshot, error)
    FindByGame(ctx context.Context, id identity.GameID) ([]entity.NPCSnapshot, error)
}
// Note: Save accepts NPCSnapshot. Infrastructure adapter calls npc.Snapshot()
// before passing to repository — consistent with GameRepository pattern.

// grimoire-domain/roots/character/port/player_character_repository.go
type PlayerCharacterRepository interface {
    SaveCore(ctx context.Context, snap entity.PlayerCharacterSnapshot) error
    LoadCore(ctx context.Context, id identity.PlayerCharacterID) (entity.PlayerCharacterSnapshot, error)
    SaveNarrative(ctx context.Context, snap entity.PlayerCharacterNarrativeSnapshot) error
    LoadNarrative(ctx context.Context, id identity.PlayerCharacterNarrativeID) (entity.PlayerCharacterNarrativeSnapshot, error)
    FindNarrativesByCharacter(ctx context.Context, id identity.PlayerCharacterID) ([]entity.PlayerCharacterNarrativeSnapshot, error)
    FindByGame(ctx context.Context, id identity.GameID) ([]entity.PlayerCharacterSnapshot, error)
    FindByCampaign(ctx context.Context, id identity.CampaignID) ([]entity.PlayerCharacterSnapshot, error)
}

// grimoire-domain/roots/character/port/macguffin_repository.go
type MacGuffinRepository interface {
    Save(ctx context.Context, snap entity.MacGuffinSnapshot) error
    Load(ctx context.Context, id identity.MacGuffinID) (entity.MacGuffinSnapshot, error)
    FindByNPCPossessor(ctx context.Context, id identity.NarrativeCharacterID) ([]entity.MacGuffinSnapshot, error)
    FindByPCPossessor(ctx context.Context, id identity.PlayerCharacterID) ([]entity.MacGuffinSnapshot, error)
    FindByLocation(ctx context.Context, id identity.LocationID) ([]entity.MacGuffinSnapshot, error)
    FindByGame(ctx context.Context, id identity.GameID) ([]entity.MacGuffinSnapshot, error)
}
```

---

## Aggregate File Structure

Package stubs (`doc.go`) already exist. Claude Code creates:

```
grimoire-domain/roots/character/
  entity/
    doc.go                                   (exists — do not modify)
    narrative_character.go                   NPC sealed interfaces + npcCore + CreateNPC
    narrative_character_handle.go            Handle for all 5 NPC states + copyCore + ReplayNPC
    narrative_character_errors.go            NPC sentinel errors incl. ErrNPCIDRequired, ErrNPCNameRequired
    narrative_character_test.go              NPC TDD tests
    narrative_character_snapshot.go          NPCSnapshot + Snapshot() + ReconstituteNPC
                                             (ReplayNPC is NOT here — it is in handle.go)
    player_character.go                      PlayerCharacter struct + constructor + methods
    player_character_errors.go
    player_character_test.go
    player_character_snapshot.go             PlayerCharacterSnapshot + Snapshot() + Reconstitute
    player_character_narrative.go            PlayerCharacterNarrative struct + methods
    player_character_narrative_errors.go
    player_character_narrative_test.go
    player_character_narrative_snapshot.go   Snapshot + Reconstitute
    macguffin.go                             MacGuffin struct + constructor + methods
    macguffin_errors.go
    macguffin_test.go
    macguffin_snapshot.go                    MacGuffinSnapshot + Snapshot() + Reconstitute

  value/
    doc.go                                   (exists — do not modify)
    character_type.go                        CharacterType constants

  port/
    doc.go                                   (exists — do not modify)
    npc_repository.go
    player_character_repository.go
    macguffin_repository.go
```

Updates to existing files:

```
grimoire-domain/shared/identity/ids.go
  Add: NarrativeCharacterID, PlayerCharacterID,
       PlayerCharacterNarrativeID, MacGuffinID
       with New* and Parse* for each
  Remove: CharacterID, NewCharacterID, ParseCharacterID

grimoire-domain/roots/campaign/entity/campaign.go
  characterIDs []identity.CharacterID → []identity.PlayerCharacterID

grimoire-domain/roots/campaign/entity/states.go
  AddCharacter signature → takes identity.PlayerCharacterID
  on both NewCampaign and FormingCampaign

grimoire-domain/roots/campaign/entity/handle.go
  identity.ParseCharacterID → identity.ParsePlayerCharacterID

grimoire-domain/roots/campaign/entity/reconstitute.go
  CampaignSnapshot.CharacterIDs → []identity.PlayerCharacterID

grimoire-domain/roots/narrative/value/beat_values.go
  Add: BeatScopeCharacter, BeatTypeBackground, BeatTypePersonalArc

grimoire-domain/roots/narrative/entity/beat.go
  Add: characterID identity.PlayerCharacterID field
  Add: CreateCharacterBeat constructor
  Update: BeatSnapshot + ReconstituteBeat for characterID round-trip

grimoire-domain/roots/narrative/entity/narrative_errors.go
  Add: ErrCharacterIDRequired
  Add: ErrInvalidBeatTypeForCharacterScope
```

---

## Events

```
EntityCreated   { entity_type: "npc" }
EntityCreated   { entity_type: "player_character" }
EntityCreated   { entity_type: "player_character_narrative" }
EntityCreated   { entity_type: "macguffin" }
EntityCreated   { entity_type: "beat", scope: "character" }

EntityUpdated   { entity_type: "npc",              field: "status" }
EntityUpdated   { entity_type: "npc",              field: "content" }
EntityUpdated   { entity_type: "npc",              field: "location" }
EntityUpdated   { entity_type: "player_character", field: "status" }
EntityUpdated   { entity_type: "player_character", field: "content" }
EntityUpdated   { entity_type: "macguffin",        field: "content" }
EntityUpdated   { entity_type: "macguffin",        field: "possessor" }
EntityUpdated   { entity_type: "macguffin",        field: "status", new_value: "destroyed" }

EntityLinked    { relationship: "located_at",
                  entity_a: npc_id,        entity_b: location_id }
EntityLinked    { relationship: "located_at",
                  entity_a: macguffin_id,  entity_b: location_id }
EntityLinked    { relationship: "possesses",
                  entity_a: character_id,  entity_b: macguffin_id }
EntityLinked    { relationship: "participates_in",
                  entity_a: pc_id,         entity_b: campaign_id }
EntityLinked    { relationship: "has_background_beat",
                  entity_a: pcn_id,        entity_b: beat_id }
EntityLinked    { relationship: "discovered_personal_beat",
                  entity_a: pcn_id,        entity_b: beat_id }

EntityRevealed  { entity_type: "npc" }
EntityRevealed  { entity_type: "player_character" }
EntityRevealed  { entity_type: "macguffin" }
EntityRevealed  { entity_type: "character_background_beat", entity_id: beat_id }
```

---

## Store Pattern

Follows ADR-016 and ADR-017 exactly.

```
Firestore:   NPCSnapshot documents
             PlayerCharacterSnapshot documents
             PlayerCharacterNarrativeSnapshot documents
             MacGuffinSnapshot documents

Neo4j:       (:NPC), (:PlayerCharacter), (:MacGuffin) nodes
             all content as node properties
             (:NPC)-[:LOCATED_AT]->(:Location)
             (:PlayerCharacter)-[:PARTICIPATES_IN]->(:Campaign)
             (:PlayerCharacterNarrative)-[:HAS_BACKGROUND]->(:Beat)
               { revealed: bool }
             (:PlayerCharacterNarrative)-[:DISCOVERED]->(:Beat)
             (:NPC)-[:POSSESSES]->(:MacGuffin)
             (:PlayerCharacter)-[:POSSESSES]->(:MacGuffin)
             (:MacGuffin)-[:LOCATED_AT]->(:Location)
             player_visible and destroyed on all nodes

GCS:         full event log permanently
             Neo4j rebuildable from GCS
```

AggregateStore port (ADR-014). Phase 1: Firestore. Phase 3: Bigtable.
Command handlers identical in both phases.

---

## Campaign Guard Update

`FormingCampaign.StartFirstSession` already checks `len(characterIDs) > 0`.
No logic change — only the type changes from `CharacterID` to
`PlayerCharacterID` as part of the breaking change above.

---

## Deferred

```
Character location presence as query   →  GM Planning context ADR
MacGuffin as narrative gate            →  MMORPG ADR
Character-to-character relationships   →  Future ADR
PlayerCharacter across multiple Games  →  Future ADR (MMORPG)
Foundry sync for PlayerCharacter       →  ADR-009 extension
Faction presence at locations          →  GM Planning context ADR (ADR-017)
```

---

## Consequences

- Campaign.characterIDs changes to []PlayerCharacterID — breaking change,
  snapshot migration is type-rename only (string UUIDs unchanged)
- NPC uses Handle/Replay (sealed interfaces) — consistent with Game/Campaign/Faction/Location
- NPC sealed interface has NPCName() and GameID() getters — consistent with
  Faction (FactionName(), GameID()) and Location (LocationName(), GameID())
- PlayerCharacter uses Snapshot/Reconstitute — consistent with Scene/CampaignNarrative;
  departure documented in code comment
- MacGuffin uses Snapshot/Reconstitute — same rationale
- All state-transition methods that emit events take source event.Source
- MacGuffin possession tracked with two typed fields — compile-time safe for both NPC and PC
- Possession invariant: at most one of npcPossessorID, pcPossessorID, locationID is non-zero
- MacGuffin Destroy() is terminal — destroyed flag, no state machine
- DiscoverPersonalBeat takes [][]BeatID — no cross-aggregate imports
- ReplayNPC(gameID, events) — consistent with ReplayFaction, ReplayCampaign, ReplaySession
- ReplayNPC lives in narrative_character_handle.go — consistent with all other Replay* functions
- CreateNPC requires name at construction — consistent with all other aggregates
- BeatScopeCharacter extends ADR-016 without breaking existing beats
- Background beats can be hidden from owning player and revealed during play
- PlayerCharacterNarrative is purely additive — same pattern as CampaignNarrative
- foundryCharacterID is the only mechanical coupling in the entire domain
- CharacterID retired — all references updated in Campaign aggregate

## Alternatives Considered

**Sealed interfaces for PlayerCharacter** — rejected. Active/Retired lifecycle
is too simple. Snapshot/Reconstitute is sufficient and less ceremony.

**Single possessorID field for MacGuffin** — rejected. A single field cannot
be type-safe for both NarrativeCharacterID and PlayerCharacterID. Two typed
fields preserve compile-time safety (ADR-006 principle).

**MacGuffin state machine** — rejected. No meaningful lifecycle states beyond
destroyed. A terminal flag is sufficient.

**Handle/Replay for MacGuffin** — rejected. No state machine means no need
to replay events for reconstruction. Snapshot is the full picture.

**Cross-aggregate import in DiscoverPersonalBeat** — rejected. Violates
the rule that aggregate roots never import other aggregate root packages.
Interactor passes prerequisite data as value types.

**CharacterID as shared type for NPC and PC** — rejected. NarrativeCharacterID
and PlayerCharacterID are meaningfully distinct domain concepts. Collapsing them
loses compile-time safety and blurs the faction/campaign/possession relationships.

**Omitting NPCName() and GameID() from NPC sealed interface** — rejected.
Inconsistency with Faction and Location sealed interfaces. Callers should not
need to deserialise a full snapshot to read the name or gameID of an NPC.