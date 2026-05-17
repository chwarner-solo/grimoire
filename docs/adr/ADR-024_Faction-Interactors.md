# ADR-024: Faction Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

Factions are world-level entities owned by a Game. They follow the
New → Draft → Active → Idle → Archived lifecycle (ADR-017,
ADR-017-Amendment-001). Draft→Active requires no guards — sparse
is not errored.

Authorization: load Game by GameID, check CallerID == game.GMID().

---

## Ports

```go
// grimoire-domain/roots/faction/interactor/ports.go

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type FactionRepository interface {
    Save(ctx context.Context, faction entity.Faction) error
    Load(ctx context.Context, id identity.FactionID) (entity.Faction, error)
    SaveMembership(ctx context.Context, m *entity.FactionMembership) error
    LoadMembership(ctx context.Context, id identity.FactionMembershipID) (*entity.FactionMembership, error)
}
```

---

## Interactor: CreateFaction

**File:** `grimoire-domain/roots/faction/interactor/create_faction.go`

```
Request:
    CallerID  string
    ID        identity.FactionID
    GameID    identity.GameID
    Name      string
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    entity.CreateFaction(req.ID, req.Name, req.GameID, req.Source)
      → ErrFactionIDRequired, ErrFactionNameRequired, ErrGameIDRequired
    factionRepo.Save(faction)
    bus.Dispatch(EntityCreated{ entity_type: "faction" })

Result:
    Faction  entity.NewFaction
    Events   []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreateFaction_Succeeds | valid request | NewFaction saved |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | ErrFactionNameRequired |
| ZeroID_ReturnsError | ID zero | ErrFactionIDRequired |
| SaveFailure_NoDispatch | saveErr set | ErrRepositorySaveFailed |

---

## Interactor: BeginFactionDraft

**File:** `grimoire-domain/roots/faction/interactor/begin_faction_draft.go`

Transitions NewFaction → Draft.

```
Request:
    CallerID   string
    FactionID  identity.FactionID
    GameID     identity.GameID
    Source     event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    nf, ok := faction.(entity.NewFaction) → ErrInvalidFactionState
    draft, events, err := nf.BeginDraft(req.Source)
    factionRepo.Save(draft)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "draft" })

Result:
    Faction  entity.DraftFaction
    Events   []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| BeginDraft_Succeeds | new faction | DraftFaction returned |
| AlreadyDraft_ReturnsError | draft faction | ErrInvalidFactionState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: ActivateFaction

**File:** `grimoire-domain/roots/faction/interactor/activate_faction.go`

Transitions DraftFaction → Active. No guards (ADR-017-Amendment-001).

```
Request:
    CallerID   string
    FactionID  identity.FactionID
    GameID     identity.GameID
    Source     event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    df, ok := faction.(entity.DraftFaction) → ErrInvalidFactionState
    active, events, err := df.Activate(req.Source)
    factionRepo.Save(active)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "active" })

Result:
    Faction  entity.ActiveFaction
    Events   []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| ActivateFaction_Succeeds | draft faction | ActiveFaction returned |
| NoMembersOrStandings_StillSucceeds | sparse draft | succeeds (amendment) |
| NotDraftFaction_ReturnsError | new faction | ErrInvalidFactionState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: AddFactionMember

**File:** `grimoire-domain/roots/faction/interactor/add_faction_member.go`

Adds an NPC to a faction. Works on Draft or Active.

```
Request:
    CallerID      string
    FactionID     identity.FactionID
    MembershipID  identity.FactionMembershipID
    NPCID         identity.NarrativeCharacterID
    GameID        identity.GameID
    Source        event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    switch faction state:
        DraftFaction  → df.AddMember(req.MembershipID, req.Source)
        ActiveFaction → af.AddMember(req.MembershipID, req.Source)
        otherwise     → ErrInvalidFactionState
    factionRepo.Save(updated)
    membership := entity.NewFactionMembership(req.MembershipID,
                      req.FactionID, req.NPCID, req.Source)
    factionRepo.SaveMembership(membership)
    bus.Dispatch(EntityCreated{ entity_type: "faction_membership" })
    bus.Dispatch(EntityLinked{ relationship: "member_of",
                               entity_a: npcID, entity_b: factionID })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AddMember_ToDraft_Succeeds | draft faction | membership saved |
| AddMember_ToActive_Succeeds | active faction | membership saved |
| AddMember_ToIdle_ReturnsError | idle faction | ErrInvalidFactionState |
| DuplicateMembership_ReturnsError | already member | domain error |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DispatchesTwoEvents | succeeds | EntityCreated + EntityLinked |

---

## Interactor: AddStandingLevel

**File:** `grimoire-domain/roots/faction/interactor/add_standing_level.go`

Defines a named standing tier (e.g. "Hostile", "Neutral", "Allied").
Works on Draft or Active.

```
Request:
    CallerID   string
    FactionID  identity.FactionID
    GameID     identity.GameID
    Level      value.StandingLevel
    Source     event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    switch state: Draft | Active → addStandingLevel; else ErrInvalidFactionState
    factionRepo.Save(updated)
    bus.Dispatch(EntityUpdated{ field: "standing_levels" })

Result:
    Events  []event.Event
```

**Test cases:** mirror AddFactionMember pattern.

---

## Interactor: DeclareAlly

**File:** `grimoire-domain/roots/faction/interactor/declare_ally.go`

```
Request:
    CallerID  string
    FactionID identity.FactionID
    AllyID    identity.FactionID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load(req.FactionID) → ErrFactionNotFound
    switch state: Draft | Active → declareAlly; else ErrInvalidFactionState
      → ErrAllyAlreadyDeclared if already allied
      → ErrAllyWarConflict if currently at war with same faction
    factionRepo.Save(updated)
    bus.Dispatch(EntityLinked{ relationship: "allied_with" })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| DeclareAlly_Succeeds | draft faction | EntityLinked dispatched |
| AlreadyAllied_ReturnsError | ally already set | ErrAllyAlreadyDeclared |
| AtWarWithAlly_ReturnsError | currently at war | ErrAllyWarConflict |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: DeclareWar

**File:** `grimoire-domain/roots/faction/interactor/declare_war.go`

Identical shape to DeclareAlly. relationship: "at_war_with".
Guard: ErrWarAlreadyDeclared, ErrWarAllyConflict.

---

## Interactor: MarkFactionDormant

**File:** `grimoire-domain/roots/faction/interactor/mark_dormant.go`

Transitions ActiveFaction → Idle.

```
Request:
    CallerID   string
    FactionID  identity.FactionID
    GameID     identity.GameID
    Source     event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    af, ok := faction.(entity.ActiveFaction) → ErrInvalidFactionState
    idle, events, err := af.MarkDormant(req.Source)
    factionRepo.Save(idle)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "idle" })
```

---

## Interactor: ReactivateFaction

**File:** `grimoire-domain/roots/faction/interactor/reactivate_faction.go`

Transitions IdleFaction → Active.

```
Request / Flow: mirror MarkFactionDormant, reversed.
Event: EntityUpdated{ field: "status", new_value: "active" }
```

---

## Interactor: ArchiveFaction

**File:** `grimoire-domain/roots/faction/interactor/archive_faction.go`

Terminal. Works from Active or Idle.

```
Request:
    CallerID   string
    FactionID  identity.FactionID
    GameID     identity.GameID
    Source     event.Source

Flow:
    gameRepo.Load → auth check
    factionRepo.Load → ErrFactionNotFound
    switch state:
        ActiveFaction → af.Archive(req.Source)
        IdleFaction   → if.Archive(req.Source)
        otherwise     → ErrInvalidFactionState
    factionRepo.Save(archived)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "archived" })

Result:
    Faction  entity.ArchivedFaction
    Events   []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| ArchiveFromActive_Succeeds | active faction | ArchivedFaction |
| ArchiveFromIdle_Succeeds | idle faction | ArchivedFaction |
| ArchiveFromDraft_ReturnsError | draft faction | ErrInvalidFactionState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Errors

```go
// grimoire-domain/roots/faction/interactor/errors.go

var (
    ErrRepositorySaveFailed  = errors.New("interactor: failed to save faction")
    ErrRepositoryLoadFailed  = errors.New("interactor: failed to load faction")
    ErrEventDispatchFailed   = errors.New("interactor: failed to dispatch event")
    ErrGameNotFound          = errors.New("interactor: game not found")
    ErrFactionNotFound       = errors.New("interactor: faction not found")
    ErrInvalidFactionState   = errors.New("interactor: faction is not in required state")
)
```

---

## File Locations

```
grimoire-domain/roots/faction/interactor/
    ports.go
    errors.go
    create_faction.go + _test.go
    begin_faction_draft.go + _test.go
    activate_faction.go + _test.go
    add_faction_member.go + _test.go
    add_standing_level.go + _test.go
    declare_ally.go + _test.go
    declare_war.go + _test.go
    mark_dormant.go + _test.go
    reactivate_faction.go + _test.go
    archive_faction.go + _test.go
```