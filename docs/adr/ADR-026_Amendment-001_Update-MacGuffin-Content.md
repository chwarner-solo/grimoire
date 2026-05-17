# ADR-026 Amendment 001: UpdateMacGuffinContent Interactor

## Status
Accepted

## Date
2026-05-17

## Amends
ADR-026: Character Interactors

---

## Context

ADR-028 Amendment 001 (Correction 5) stripped content fields from
`CreateMacGuffinInput` because the interactor layer had no receiver for
them. The MacGuffin domain entity already has `UpdateContent()` —
the missing piece was the interactor that orchestrates it.

This amendment adds that interactor.

---

## Interactor: UpdateMacGuffinContent

**File:** `grimoire-domain/roots/character/interactor/update_macguffin_content.go`

Updates the narrative content fields on a MacGuffin.
Follows the identical pattern as `UpdateNPCContent` and `UpdateBeatContent`.
Destroyed MacGuffins cannot be updated — domain guard enforces this.

```
Request:
    CallerID     string
    MacGuffinID  identity.MacGuffinID
    GameID       identity.GameID
    Name         string
    Description  string
    PlayerDesc   string
    Source       event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth: req.CallerID != game.GMID() → ErrUnauthorized
    mgSnap := macguffinRepo.Load(req.MacGuffinID) → ErrMacGuffinNotFound
    mg := entity.ReconstituteMacGuffin(mgSnap)
    updated, events, err := mg.UpdateContent(
        req.Name, req.Description, req.PlayerDesc, req.Source)
      → ErrMacGuffinDestroyed if mg.IsDestroyed()
      → content validation errors from domain
    macguffinRepo.Save(updated.Snapshot()) → ErrRepositorySaveFailed
    bus.Dispatch(EntityUpdated{ field: "content" })

Result:
    Events  []event.Event

AggregateType: event.AggregateCharacter
Event emitted: EntityUpdated { field: "content" }
```

---

## Struct

```go
type UpdateMacGuffinContentRequest struct {
    CallerID    string
    MacGuffinID identity.MacGuffinID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source
}

type UpdateMacGuffinContentResult struct {
    Events []event.Event
}

type UpdateMacGuffinContentInteractor struct {
    gameRepo      GameRepository
    macguffinRepo MacGuffinRepository
    bus           event.EventBus
}

func NewUpdateMacGuffinContentInteractor(
    gameRepo      GameRepository,
    macguffinRepo MacGuffinRepository,
    bus           event.EventBus,
) *UpdateMacGuffinContentInteractor
```

---

## Test Cases

**File:** `grimoire-domain/roots/character/interactor/update_macguffin_content_test.go`

| Test | Setup | Expected |
|------|-------|----------|
| UpdateContent_Succeeds | live MacGuffin, valid fields | EntityUpdated dispatched |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| MacGuffinNotFound_ReturnsError | empty repo | ErrMacGuffinNotFound |
| DestroyedMacGuffin_ReturnsError | destroyed MacGuffin | ErrMacGuffinDestroyed |
| EmptyName_ReturnsError | Name="" | domain error, no save |
| EmptyDescription_ReturnsError | Description="" | domain error, no save |
| EmptyPlayerDesc_ReturnsError | PlayerDesc="" | domain error, no save |
| SaveFailure_NoDispatch | saveErr set | ErrRepositorySaveFailed, 0 dispatched |
| DispatchFailure_ReturnsError | bus.dispatchErr set | ErrEventDispatchFailed |

---

## File Locations

```
grimoire-domain/roots/character/interactor/
    update_macguffin_content.go       ← new file
    update_macguffin_content_test.go  ← new file
```

No changes to `errors.go` — all required error sentinels already exist
in the character interactor package.

---

## Consequences

- MacGuffin content follows the same two-step pattern as NPC and Beat:
  create with name only, enrich with content via a separate call
- Destroyed MacGuffins are immutable — the domain guard prevents updates
- The GraphQL mutation `updateMacGuffinContent` (ADR-028 Amendment 001)
  maps directly to this interactor
- Total character interactor count: 14 → 15