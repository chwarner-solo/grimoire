# ADR-026: Character Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

Three aggregate types live in the character package: NPC
(NarrativeCharacter), PlayerCharacter, and MacGuffin (ADR-019).

NPC follows the sealed interface + Handle/Replay pattern (New → Draft →
Active → Idle → Archived). PlayerCharacter uses Snapshot/Reconstitute
(Active → Retired). MacGuffin uses Snapshot/Reconstitute with a
terminal Destroyed flag.

ownerPlayerID on PlayerCharacter is optional — the GM creates all
characters (ADR-019-Amendment-001 / ADR-020).

Authorization: load Game by GameID, check CallerID == game.GMID().

---

## Ports

```go
// grimoire-domain/roots/character/interactor/ports.go

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type NPCRepository interface {
    Save(ctx context.Context, snap entity.NPCSnapshot) error
    Load(ctx context.Context, id identity.NarrativeCharacterID) (entity.NPCSnapshot, error)
}

type PlayerCharacterRepository interface {
    SaveCore(ctx context.Context, snap entity.PlayerCharacterSnapshot) error
    LoadCore(ctx context.Context, id identity.PlayerCharacterID) (entity.PlayerCharacterSnapshot, error)
}

type MacGuffinRepository interface {
    Save(ctx context.Context, snap entity.MacGuffinSnapshot) error
    Load(ctx context.Context, id identity.MacGuffinID) (entity.MacGuffinSnapshot, error)
}
```

---

## ── NPC INTERACTORS ──────────────────────────────────────────────────────

## Interactor: CreateNPC

**File:** `grimoire-domain/roots/character/interactor/create_npc.go`

```
Request:
    CallerID  string
    ID        identity.NarrativeCharacterID
    GameID    identity.GameID
    Name      string
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    entity.CreateNPC(req.ID, req.Name, req.GameID, req.Source)
    npcRepo.Save(npc.Snapshot())
    bus.Dispatch(EntityCreated{ entity_type: "npc" })

Result:
    NPC     entity.NewNPC
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreateNPC_Succeeds | valid request | NewNPC saved |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | ErrNPCNameRequired |
| ZeroID_ReturnsError | ID zero | ErrNPCIDRequired |

---

## Interactor: BeginNPCDraft

**File:** `grimoire-domain/roots/character/interactor/begin_npc_draft.go`

Transitions NewNPC → Draft.

```
Request:
    CallerID  string
    NPCID     identity.NarrativeCharacterID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    snap := npcRepo.Load(req.NPCID) → ErrNPCNotFound
    npc := entity.ReconstituteNPC(snap)
    nf, ok := npc.(entity.NewNPC) → ErrInvalidNPCState
    draft, events, err := nf.BeginDraft(req.Source)
    npcRepo.Save(draft.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "draft" })

Result:
    NPC     entity.DraftNPC
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| BeginDraft_Succeeds | new NPC | DraftNPC returned |
| AlreadyDraft_ReturnsError | draft NPC | ErrInvalidNPCState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: UpdateNPCContent

**File:** `grimoire-domain/roots/character/interactor/update_npc_content.go`

Updates name, description, and playerDescription. Works on Draft or Active.

```
Request:
    CallerID       string
    NPCID          identity.NarrativeCharacterID
    GameID         identity.GameID
    Name           string
    Description    string
    PlayerDesc     string
    Source         event.Source

Flow:
    gameRepo.Load → auth check
    snap := npcRepo.Load → ErrNPCNotFound
    npc := entity.ReconstituteNPC(snap)
    switch state:
        DraftNPC  → dn.UpdateContent(name, desc, playerDesc, source)
        ActiveNPC → an.UpdateContent(name, desc, playerDesc, source)
        otherwise → ErrInvalidNPCState
    npcRepo.Save(updated.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "content" })

Result:
    NPC     entity.NPC
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| UpdateDraftNPC_Succeeds | draft NPC | content updated |
| UpdateActiveNPC_Succeeds | active NPC | content updated |
| UpdateArchivedNPC_ReturnsError | archived NPC | ErrInvalidNPCState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | domain error |

---

## Interactor: ActivateNPC

**File:** `grimoire-domain/roots/character/interactor/activate_npc.go`

Transitions DraftNPC → Active.

```
Request:
    CallerID  string
    NPCID     identity.NarrativeCharacterID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    snap := npcRepo.Load → ErrNPCNotFound
    npc := entity.ReconstituteNPC(snap)
    dn, ok := npc.(entity.DraftNPC) → ErrInvalidNPCState
    active, events, err := dn.Activate(req.Source)
      → ErrNPCNameRequired (domain guard if name still empty)
    npcRepo.Save(active.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "active" })
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| ActivateNPC_Succeeds | draft with name | ActiveNPC |
| EmptyName_ReturnsError | draft, name="" | domain error |
| NotDraft_ReturnsError | new NPC | ErrInvalidNPCState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: ArchiveNPC

**File:** `grimoire-domain/roots/character/interactor/archive_npc.go`

Terminal. Works from Active or Idle. MacGuffin drop handled by
NPCArchivedHandler via EventBus.

```
Request:
    CallerID  string
    NPCID     identity.NarrativeCharacterID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    snap := npcRepo.Load → ErrNPCNotFound
    npc := entity.ReconstituteNPC(snap)
    switch state:
        ActiveNPC → an.Archive(req.Source)
        IdleNPC   → in.Archive(req.Source)
        otherwise → ErrInvalidNPCState
    npcRepo.Save(archived.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "archived" })
    // NPCArchivedHandler drops MacGuffins via EventBus
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| ArchiveActive_Succeeds | active NPC | ArchivedNPC |
| ArchiveIdle_Succeeds | idle NPC | ArchivedNPC |
| ArchiveDraft_ReturnsError | draft NPC | ErrInvalidNPCState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## ── PLAYER CHARACTER INTERACTORS ─────────────────────────────────────────

## Interactor: CreatePlayerCharacter

**File:** `grimoire-domain/roots/character/interactor/create_player_character.go`

```
Request:
    CallerID       string
    ID             identity.PlayerCharacterID
    GameID         identity.GameID
    Name           string
    OwnerPlayerID  string    // optional — may be empty (ADR-019-Amendment-001)
    Source         event.Source

Flow:
    gameRepo.Load → auth check
    entity.CreatePlayerCharacter(req.ID, req.GameID,
        req.Name, req.OwnerPlayerID, req.Source)
    pcRepo.SaveCore(pc.Snapshot())
    bus.Dispatch(EntityCreated{ entity_type: "player_character" })

Result:
    PlayerCharacter  *entity.PlayerCharacter
    Events           []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreatePC_Succeeds | name set, no ownerID | PC saved |
| CreatePC_WithOwner_Succeeds | ownerID set | PC saved, ownerID stored |
| EmptyName_ReturnsError | Name="" | ErrCharacterNameRequired |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| ZeroID_ReturnsError | ID zero | ErrPlayerCharacterIDRequired |

---

## Interactor: UpdatePlayerCharacterContent

**File:** `grimoire-domain/roots/character/interactor/update_player_character_content.go`

```
Request:
    CallerID    string
    ID          identity.PlayerCharacterID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    snap := pcRepo.LoadCore(req.ID) → ErrPlayerCharacterNotFound
    pc := entity.ReconstitutePlayerCharacter(snap)
    updated, events, err := pc.UpdateContent(name, desc, playerDesc, source)
    pcRepo.SaveCore(updated.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "content" })
```

**Test cases:** mirror UpdateNPCContent. Retired PC → domain error.

---

## Interactor: RetirePlayerCharacter

**File:** `grimoire-domain/roots/character/interactor/retire_player_character.go`

Terminal transition for a PlayerCharacter.

```
Request:
    CallerID  string
    ID        identity.PlayerCharacterID
    GameID    identity.GameID
    Source    event.Source

Flow:
    gameRepo.Load → auth check
    snap := pcRepo.LoadCore(req.ID) → ErrPlayerCharacterNotFound
    pc := entity.ReconstitutePlayerCharacter(snap)
    retired, events, err := pc.Retire(req.Source)
      → ErrCharacterAlreadyRetired
    pcRepo.SaveCore(retired.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "retired" })
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| RetirePC_Succeeds | active PC | status="retired" |
| AlreadyRetired_ReturnsError | retired PC | ErrCharacterAlreadyRetired |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## ── MACGUFFIN INTERACTORS ────────────────────────────────────────────────

## Interactor: CreateMacGuffin

**File:** `grimoire-domain/roots/character/interactor/create_macguffin.go`

```
Request:
    CallerID    string
    ID          identity.MacGuffinID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    entity.CreateMacGuffin(req.ID, req.GameID, req.Name,
        req.Description, req.PlayerDesc, req.Source)
    macguffinRepo.Save(mg.Snapshot())
    bus.Dispatch(EntityCreated{ entity_type: "macguffin" })
```

---

## Interactor: AssignMacGuffinToNPC

**File:** `grimoire-domain/roots/character/interactor/assign_macguffin_npc.go`

```
Request:
    CallerID    string
    MacGuffinID identity.MacGuffinID
    NPCID       identity.NarrativeCharacterID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    mgSnap := macguffinRepo.Load → ErrMacGuffinNotFound
    mg := entity.ReconstituteMacGuffin(mgSnap)
      → ErrMacGuffinDestroyed if mg.Destroyed
    npcSnap := npcRepo.Load(req.NPCID) → ErrNPCNotFound
    npc := entity.ReconstituteNPC(npcSnap)
    an, ok := npc.(entity.ActiveNPC) → ErrInvalidNPCState
    updated, events, err := an.AssignMacGuffin(req.MacGuffinID, req.Source)
    mg.SetNPCPossessor(req.NPCID)
    npcRepo.Save(updated.Snapshot())
    macguffinRepo.Save(mg.Snapshot())
    bus.Dispatch(EntityLinked{ relationship: "possesses",
                               entity_a: npcID, entity_b: macguffinID })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AssignToActiveNPC_Succeeds | active NPC, live MG | EntityLinked dispatched |
| NPCNotActive_ReturnsError | draft NPC | ErrInvalidNPCState |
| MacGuffinDestroyed_ReturnsError | destroyed MG | ErrMacGuffinDestroyed |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: AssignMacGuffinToPC

**File:** `grimoire-domain/roots/character/interactor/assign_macguffin_pc.go`

Identical shape to AssignMacGuffinToNPC using PlayerCharacter.
Uses `pc.AssignMacGuffin(req.MacGuffinID, req.Source)`.
Guard: ErrCharacterRetired if PC is retired.

---

## Interactor: DestroyMacGuffin

**File:** `grimoire-domain/roots/character/interactor/destroy_macguffin.go`

Terminal. Clears all possession.

```
Request:
    CallerID    string
    MacGuffinID identity.MacGuffinID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    mgSnap := macguffinRepo.Load → ErrMacGuffinNotFound
    mg := entity.ReconstituteMacGuffin(mgSnap)
    destroyed, events, err := mg.Destroy(req.Source)
      → ErrMacGuffinAlreadyDestroyed
    macguffinRepo.Save(destroyed.Snapshot())
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "destroyed" })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| DestroyMacGuffin_Succeeds | live MG | Destroyed=true |
| AlreadyDestroyed_ReturnsError | destroyed MG | ErrMacGuffinAlreadyDestroyed |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Errors

```go
// grimoire-domain/roots/character/interactor/errors.go

var (
    ErrRepositorySaveFailed       = errors.New("interactor: failed to save")
    ErrRepositoryLoadFailed       = errors.New("interactor: failed to load")
    ErrEventDispatchFailed        = errors.New("interactor: failed to dispatch event")
    ErrGameNotFound               = errors.New("interactor: game not found")
    ErrNPCNotFound                = errors.New("interactor: npc not found")
    ErrPlayerCharacterNotFound    = errors.New("interactor: player character not found")
    ErrMacGuffinNotFound          = errors.New("interactor: macguffin not found")
    ErrMacGuffinDestroyed         = errors.New("interactor: macguffin is destroyed")
    ErrInvalidNPCState            = errors.New("interactor: npc is not in required state")
)
```

---

## File Locations

```
grimoire-domain/roots/character/interactor/
    ports.go
    errors.go
    create_npc.go + _test.go
    begin_npc_draft.go + _test.go
    update_npc_content.go + _test.go
    activate_npc.go + _test.go
    archive_npc.go + _test.go
    create_player_character.go + _test.go
    update_player_character_content.go + _test.go
    retire_player_character.go + _test.go
    create_macguffin.go + _test.go
    assign_macguffin_npc.go + _test.go
    assign_macguffin_pc.go + _test.go
    destroy_macguffin.go + _test.go
```

---

## Consequences

- NPC reconstitution pattern: Load snapshot → ReconstituteNPC → type-assert
  to required state interface → call method → Save snapshot. All NPC
  interactors follow this identical shape.
- PlayerCharacter and MacGuffin use Snapshot/Reconstitute directly —
  no state-interface type assertion needed. Simpler.
- AssignMacGuffin modifies two aggregates (NPC/PC + MacGuffin). NPC/PC
  saved first — MacGuffin save failure is retryable.
- MacGuffin drop on NPC archive is handled by NPCArchivedHandler via
  EventBus — ArchiveNPC interactor does not touch MacGuffins directly.