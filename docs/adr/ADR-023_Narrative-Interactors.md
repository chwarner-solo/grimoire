# ADR-023: Narrative Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

The Narrative subsystem has two aggregate roots — MasterNarrative
(world-level, owned by Game) and CampaignNarrative (table-level,
owned by Campaign) — plus Beat as a shared entity (ADR-016).

Both aggregates are purely additive. No state machines. Interactors
append to them; they never transition lifecycle states.

Authorization: every interactor loads the Game by GameID and checks
CallerID == game.GMID() before any domain operation.

MasterNarrative is auto-created by the MasterNarrativeCreatedHandler
when a Game is created (ADR-011). There is no CreateMasterNarrative
interactor — the GM never calls it directly.

CampaignNarrative is auto-created by the CampaignNarrativeCreatedHandler
when a Campaign is created (ADR-011). There is no CreateCampaignNarrative
interactor.

---

## Ports

```go
// grimoire-domain/roots/narrative/interactor/ports.go

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type MasterNarrativeRepository interface {
    Save(ctx context.Context, mn *entity.MasterNarrative) error
    Load(ctx context.Context, id identity.MasterNarrativeID) (*entity.MasterNarrative, error)
    FindByGame(ctx context.Context, id identity.GameID) (*entity.MasterNarrative, error)
}

type CampaignNarrativeRepository interface {
    Save(ctx context.Context, cn *entity.CampaignNarrative) error
    Load(ctx context.Context, id identity.CampaignNarrativeID) (*entity.CampaignNarrative, error)
    FindByCampaign(ctx context.Context, id identity.CampaignID) (*entity.CampaignNarrative, error)
}

type BeatRepository interface {
    Save(ctx context.Context, beat *entity.Beat) error
    Load(ctx context.Context, id identity.BeatID) (*entity.Beat, error)
    LoadPrerequisiteChain(ctx context.Context, id identity.BeatID) ([]*entity.Beat, error)
    FindByGame(ctx context.Context, id identity.GameID) ([]*entity.Beat, error)
}
```

---

## Interactor: CreateMasterBeat

**File:** `grimoire-domain/roots/narrative/interactor/create_master_beat.go`

Creates a Beat scoped to the MasterNarrative. The MasterNarrative
aggregate records the BeatID as a reference.

```
Request:
    CallerID    string
    BeatID      identity.BeatID
    GameID      identity.GameID
    Name        string
    Description string           // GM-only
    PlayerDesc  string           // player-facing, revealed on EntityRevealed
    BeatType    value.BeatType   // required | optional
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    mnRepo.FindByGame(req.GameID) → ErrMasterNarrativeNotFound
    entity.CreateMasterBeat(req.BeatID, req.Name, req.Description,
        req.PlayerDesc, req.BeatType, req.GameID, req.Source)
      → domain validation errors
    beatRepo.Save(beat)
    mn.AddBeat(beat.BeatID())
    mnRepo.Save(mn)
    bus.Dispatch(EntityCreated{ entity_type: "beat" })

Result:
    Beat    *entity.Beat
    Events  []event.Event

Note: Both Beat and MasterNarrative are saved. If Beat save succeeds
      but MasterNarrative save fails, retry is safe — Beat is idempotent
      on re-save, MasterNarrative.AddBeat is idempotent if BeatID already
      recorded (checked before append).
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreateMasterBeat_Succeeds | valid request | Beat saved, MN updated |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | domain error, no saves |
| InvalidBeatType_ReturnsError | BeatType="campaign-specific" | ErrInvalidBeatType |
| BeatSaveFailure_NoMNSave | beatRepo.saveErr set | ErrRepositorySaveFailed |
| DispatchesEntityCreatedEvent | succeeds | entity_type="beat" |

---

## Interactor: UpdateBeatContent

**File:** `grimoire-domain/roots/narrative/interactor/update_beat_content.go`

Updates the narrative content fields on any Beat. Neo4j handler
projects the change as a read-side update.

```
Request:
    CallerID    string
    BeatID      identity.BeatID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    beatRepo.Load(req.BeatID) → ErrBeatNotFound
    beat.UpdateContent(req.Name, req.Description, req.PlayerDesc)
      → ErrBeatNameRequired / ErrBeatDescriptionRequired / etc.
    beatRepo.Save(beat) → ErrRepositorySaveFailed
    bus.Dispatch(EntityUpdated{ field: "content" })

Result:
    Beat    *entity.Beat
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| UpdateContent_Succeeds | valid beat | beat saved, EntityUpdated dispatched |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | ErrBeatNameRequired |
| EmptyDescription_ReturnsError | Description="" | ErrBeatDescriptionRequired |
| BeatNotFound_ReturnsError | empty repo | ErrBeatNotFound |

---

## Interactor: AddBeatPrerequisite

**File:** `grimoire-domain/roots/narrative/interactor/add_beat_prerequisite.go`

Links a prerequisite Beat to a target Beat. Repository loads the full
prerequisite chain; domain validates acyclicity (ADR-016).

```
Request:
    CallerID        string
    BeatID          identity.BeatID      // the beat gaining a prerequisite
    PrerequisiteID  identity.BeatID      // the beat that must be discovered first
    GameID          identity.GameID
    Source          event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    beatRepo.Load(req.BeatID) → ErrBeatNotFound
    chain, err := beatRepo.LoadPrerequisiteChain(req.PrerequisiteID)
      → ErrPrerequisiteChainLoadFailed
    beat.WouldCreateCycle(req.BeatID, chain)
      → ErrCycleDetected
    beat.AddPrerequisite(req.PrerequisiteID, req.Source)
    beatRepo.Save(beat) → ErrRepositorySaveFailed
    bus.Dispatch(EntityLinked{ relationship: "prerequisite_of" })

Result:
    Beat    *entity.Beat
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AddPrerequisite_Succeeds | two beats, no cycle | prerequisite added |
| WouldCreateCycle_ReturnsError | A→B, adding B→A | ErrCycleDetected |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| BeatNotFound_ReturnsError | missing beat | ErrBeatNotFound |
| SelfPrerequisite_ReturnsError | BeatID == PrerequisiteID | ErrCycleDetected |

---

## Interactor: AddActToMasterNarrative

**File:** `grimoire-domain/roots/narrative/interactor/add_act.go`

```
Request:
    CallerID  string
    ActID     identity.ActID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    mnRepo.FindByGame(req.GameID) → ErrMasterNarrativeNotFound
    mn.AddAct(req.ActID)
    mnRepo.Save(mn) → ErrRepositorySaveFailed
    bus.Dispatch(EntityLinked{ relationship: "contains_act",
                               entity_a: mn.ID, entity_b: actID })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AddAct_Succeeds | valid MN | act added, EntityLinked dispatched |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| ZeroActID_ReturnsError | ActID zero | domain error |

---

## Interactor: AddSecretToMasterNarrative

**File:** `grimoire-domain/roots/narrative/interactor/add_secret.go`

Identical shape to AddAct. relationship: "contains_secret".

```
Request:
    CallerID  string
    SecretID  identity.SecretID
    GameID    identity.GameID
    Source    event.Source
```

**Test cases:** mirror AddAct.

---

## Interactor: AddLoreToMasterNarrative

**File:** `grimoire-domain/roots/narrative/interactor/add_lore.go`

Identical shape to AddAct. relationship: "contains_lore".

```
Request:
    CallerID  string
    LoreID    identity.LoreID
    GameID    identity.GameID
    Source    event.Source
```

**Test cases:** mirror AddAct.

---

## Interactor: DiscoverBeat

**File:** `grimoire-domain/roots/narrative/interactor/discover_beat.go`

Called at play time when the party encounters a narrative beat.
O(1) prerequisite check — no DAG traversal (ADR-016).

```
Request:
    CallerID           string
    CampaignNarrativeID identity.CampaignNarrativeID
    BeatID             identity.BeatID
    GameID             identity.GameID
    Source             event.Source

Flow:
    gameRepo.Load → auth check
    cnRepo.Load(req.CampaignNarrativeID) → ErrCampaignNarrativeNotFound
    beatRepo.Load(req.BeatID) → ErrBeatNotFound
    cn.DiscoverBeat(*beat)
      → ErrPrerequisiteNotMet if prerequisites not satisfied
    cnRepo.Save(cn) → ErrRepositorySaveFailed
    bus.Dispatch(EntityLinked{ relationship: "discovered",
                               entity_a: campaignNarrativeID,
                               entity_b: beatID })

Result:
    Events  []event.Event

Note: Neo4j handler writes (:Campaign)-[:DISCOVERED]->(:Beat) edge.
      This is the read-side signal that the beat is found.
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| DiscoverBeat_Succeeds | prerequisites met | EntityLinked dispatched |
| PrerequisiteNotMet_ReturnsError | beat needs undiscovered beat | ErrPrerequisiteNotMet |
| AlreadyDiscovered_Succeeds | idempotent | no error (CN checks before append) |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| BeatNotFound_ReturnsError | missing beat | ErrBeatNotFound |

---

## Interactor: CreateCampaignBeat

**File:** `grimoire-domain/roots/narrative/interactor/create_campaign_beat.go`

GM improvises a beat at the table. Scoped to a specific Campaign.
May later be promoted to MasterNarrative.

```
Request:
    CallerID    string
    BeatID      identity.BeatID
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    cnRepo.FindByCampaign(req.CampaignID) → ErrCampaignNarrativeNotFound
    entity.CreateCampaignBeat(req.BeatID, req.Name, req.Description,
        req.PlayerDesc, req.CampaignID, req.GameID, req.Source)
    beatRepo.Save(beat)
    cn.AddCampaignBeat(beat.BeatID())
    cnRepo.Save(cn)
    bus.Dispatch(EntityCreated{ entity_type: "beat",
                                scope: "campaign" })

Result:
    Beat    *entity.Beat
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreateCampaignBeat_Succeeds | valid request | beat saved, CN updated |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | domain error |

---

## Interactor: PromoteBeatToMaster

**File:** `grimoire-domain/roots/narrative/interactor/promote_beat.go`

Promotes a campaign-scoped Beat to MasterNarrative scope.
Other Campaigns can now discover it (ADR-016).

```
Request:
    CallerID  string
    BeatID    identity.BeatID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    beatRepo.Load(req.BeatID) → ErrBeatNotFound
    beat.PromoteToMaster(req.Source)
      → ErrBeatAlreadyMasterScope if already master
    mnRepo.FindByGame(req.GameID) → ErrMasterNarrativeNotFound
    mn.AddBeat(beat.BeatID())
    beatRepo.Save(beat)
    mnRepo.Save(mn)
    bus.Dispatch(EntityUpdated{ field: "scope",
                                old_value: "campaign:{id}",
                                new_value: "master:{gameID}" })

Result:
    Beat    *entity.Beat
    Events  []event.Event

Note: Neo4j handler moves the Beat node from the campaign graph to
      the master graph on this event.
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| PromoteBeat_Succeeds | campaign beat | scope="master", MN updated |
| AlreadyMaster_ReturnsError | master beat | ErrBeatAlreadyMasterScope |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| BeatNotFound_ReturnsError | missing beat | ErrBeatNotFound |

---

## Errors

```go
// grimoire-domain/roots/narrative/interactor/errors.go

var (
    ErrRepositorySaveFailed         = errors.New("interactor: failed to save")
    ErrRepositoryLoadFailed         = errors.New("interactor: failed to load")
    ErrEventDispatchFailed          = errors.New("interactor: failed to dispatch event")
    ErrGameNotFound                 = errors.New("interactor: game not found")
    ErrBeatNotFound                 = errors.New("interactor: beat not found")
    ErrMasterNarrativeNotFound      = errors.New("interactor: master narrative not found")
    ErrCampaignNarrativeNotFound    = errors.New("interactor: campaign narrative not found")
    ErrPrerequisiteChainLoadFailed  = errors.New("interactor: failed to load prerequisite chain")
    ErrInvalidBeatType              = errors.New("interactor: invalid beat type for this operation")
)
```

---

## File Locations

```
grimoire-domain/roots/narrative/interactor/
    ports.go
    errors.go
    create_master_beat.go
    create_master_beat_test.go
    update_beat_content.go
    update_beat_content_test.go
    add_beat_prerequisite.go
    add_beat_prerequisite_test.go
    add_act.go
    add_act_test.go
    add_secret.go
    add_secret_test.go
    add_lore.go
    add_lore_test.go
    discover_beat.go
    discover_beat_test.go
    create_campaign_beat.go
    create_campaign_beat_test.go
    promote_beat.go
    promote_beat_test.go
```

---

## Consequences

- MasterNarrative and CampaignNarrative are never created by GM —
  they are auto-created by EventBus handlers
- DiscoverBeat is the only play-time narrative interactor — all others
  are world-building (prep time)
- AddAct / AddSecret / AddLore follow identical patterns —
  share test helper infrastructure
- Beat promotion modifies two aggregates (Beat + MasterNarrative) —
  both saves must succeed; Beat is saved first as it carries the
  scope change; MN save failure is retryable