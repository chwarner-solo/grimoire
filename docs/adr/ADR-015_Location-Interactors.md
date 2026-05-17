# ADR-025: Location Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

Locations are the spatial layer of the narrative — recursive, DAG-connected,
with scenes inside them (ADR-018, ADR-018-Amendment-001).

Activation requires no scenes — sparse is not errored (Amendment-001).

ArchiveLocation requires two cross-aggregate checks at the interactor level:
- No child Location is still Active (interactor loads children)
- Party is not currently at this Location (interactor checks Campaign)

These are interactor-level guards, not domain guards — aggregates cannot
query other aggregates (ADR-018).

Authorization: load Game by GameID, check CallerID == game.GMID().

---

## Ports

```go
// grimoire-domain/roots/location/interactor/ports.go

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type LocationRepository interface {
    Save(ctx context.Context, location entity.Location) error
    Load(ctx context.Context, id identity.LocationID) (entity.Location, error)
    FindChildren(ctx context.Context, id identity.LocationID) ([]entity.Location, error)
    FindByGame(ctx context.Context, id identity.GameID) ([]entity.Location, error)
}

type CampaignRepository interface {
    FindByGame(ctx context.Context, id identity.GameID) ([]campaignentity.Campaign, error)
}
```

---

## Interactor: CreateLocation

**File:** `grimoire-domain/roots/location/interactor/create_location.go`

```
Request:
    CallerID    string
    ID          identity.LocationID
    GameID      identity.GameID
    Name        string
    Description string
    ParentID    identity.LocationID  // zero if top-level
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    entity.CreateLocation(req.ID, req.Name, req.Description,
        req.GameID, req.ParentID, req.Source)
    locationRepo.Save(location)
    bus.Dispatch(EntityCreated{ entity_type: "location" })
    if !req.ParentID.IsZero():
        bus.Dispatch(EntityLinked{ relationship: "child_of",
                                   entity_a: locationID,
                                   entity_b: parentID })

Result:
    Location  entity.NewLocation
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreateTopLevel_Succeeds | no parentID | 1 event dispatched |
| CreateChild_Succeeds | parentID set | 2 events: Created + Linked |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| EmptyName_ReturnsError | Name="" | domain error |
| SaveFailure_NoDispatch | saveErr set | no events dispatched |

---

## Interactor: ActivateLocation

**File:** `grimoire-domain/roots/location/interactor/activate_location.go`

Transitions DraftLocation → Active. No scene guard (Amendment-001).

```
Request:
    CallerID    string
    LocationID  identity.LocationID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    locationRepo.Load → ErrLocationNotFound
    dl, ok := location.(entity.DraftLocation) → ErrInvalidLocationState
    active, events, err := dl.Activate(req.Source)
    locationRepo.Save(active)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "active" })

Result:
    Location  entity.ActiveLocation
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| Activate_Succeeds | draft, no scenes | ActiveLocation returned |
| Activate_WithScenes_Succeeds | draft, 1+ scenes | ActiveLocation returned |
| NotDraft_ReturnsError | new location | ErrInvalidLocationState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: AddScene

**File:** `grimoire-domain/roots/location/interactor/add_scene.go`

Adds a Scene to a Draft or Active Location.

```
Request:
    CallerID    string
    LocationID  identity.LocationID
    SceneID     identity.SceneID
    GameID      identity.GameID
    Name        string
    Description string
    PlayerDesc  string
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    locationRepo.Load → ErrLocationNotFound
    scene := entity.NewScene(req.SceneID, req.Name,
                 req.Description, req.PlayerDesc, req.Source)
    switch location state:
        DraftLocation  → dl.AddScene(scene, req.Source)
        ActiveLocation → al.AddScene(scene, req.Source)
        otherwise      → ErrInvalidLocationState
    locationRepo.Save(updated)
    bus.Dispatch(EntityCreated{ entity_type: "scene",
                                parent_id: locationID })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AddSceneToDraft_Succeeds | draft location | scene saved |
| AddSceneToActive_Succeeds | active location | scene saved |
| AddSceneToArchived_ReturnsError | archived location | ErrInvalidLocationState |
| EmptySceneName_ReturnsError | Name="" | domain error |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |

---

## Interactor: ConnectLocations

**File:** `grimoire-domain/roots/location/interactor/connect_locations.go`

Creates a directed travel connection between two Locations.
For symmetric connections (A↔B), call twice — the interactor handles
one direction per call. Symmetric is a UI concern, not a domain concern.

```
Request:
    CallerID    string
    FromID      identity.LocationID
    ToID        identity.LocationID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    locationRepo.Load(req.FromID) → ErrLocationNotFound
    locationRepo.Load(req.ToID) → ErrLocationNotFound
    from.AddConnection(req.ToID, req.Source)
      → ErrConnectionAlreadyExists
      → ErrSelfConnection if FromID == ToID
    locationRepo.Save(from)
    bus.Dispatch(EntityLinked{ relationship: "connects_to",
                               entity_a: fromID, entity_b: toID })

Result:
    Events  []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| Connect_Succeeds | two active locations | EntityLinked dispatched |
| SelfConnection_ReturnsError | FromID == ToID | ErrSelfConnection |
| AlreadyConnected_ReturnsError | connection exists | ErrConnectionAlreadyExists |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| ArchivedFrom_ReturnsError | from is archived | ErrInvalidLocationState |

---

## Interactor: ArchiveLocation

**File:** `grimoire-domain/roots/location/interactor/archive_location.go`

Terminal. Cascades to child Locations and Scenes via EventBus handler.
Two cross-aggregate guards at interactor level (ADR-018):
1. No child Location is Active
2. No Campaign's party is currently here

```
Request:
    CallerID    string
    LocationID  identity.LocationID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    locationRepo.Load → ErrLocationNotFound
    switch state: Active | Idle → proceed; else ErrInvalidLocationState

    // Guard 1: no active children
    children, _ := locationRepo.FindChildren(req.LocationID)
    for each child: if child is ActiveLocation → ErrActiveChildrenExist

    // Guard 2: party not present
    campaigns, _ := campaignRepo.FindByGame(req.GameID)
    for each campaign:
        if campaign.CurrentLocationID() == req.LocationID
        AND campaign is ActiveCampaign → ErrPartyPresent

    // Proceed
    archived, events, err := location.Archive(req.Source)
    locationRepo.Save(archived)
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "archived" })

    // Cascade handled by LocationArchivedHandler via EventBus (ADR-018)

Result:
    Location  entity.ArchivedLocation
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| ArchiveActive_Succeeds | no children, no party | ArchivedLocation |
| ArchiveIdle_Succeeds | idle, no party | ArchivedLocation |
| ActiveChildPresent_ReturnsError | child is Active | ErrActiveChildrenExist |
| PartyPresent_ReturnsError | campaign at this location | ErrPartyPresent |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DraftLocation_ReturnsError | draft location | ErrInvalidLocationState |

---

## Errors

```go
// grimoire-domain/roots/location/interactor/errors.go

var (
    ErrRepositorySaveFailed   = errors.New("interactor: failed to save location")
    ErrRepositoryLoadFailed   = errors.New("interactor: failed to load location")
    ErrEventDispatchFailed    = errors.New("interactor: failed to dispatch event")
    ErrGameNotFound           = errors.New("interactor: game not found")
    ErrLocationNotFound       = errors.New("interactor: location not found")
    ErrInvalidLocationState   = errors.New("interactor: location is not in required state")
    ErrActiveChildrenExist    = errors.New("interactor: cannot archive — active child locations exist")
    ErrPartyPresent           = errors.New("interactor: cannot archive — party is currently here")
    ErrConnectionAlreadyExists = errors.New("interactor: connection already exists")
    ErrSelfConnection         = errors.New("interactor: location cannot connect to itself")
)
```

---

## File Locations

```
grimoire-domain/roots/location/interactor/
    ports.go
    errors.go
    create_location.go + _test.go
    activate_location.go + _test.go
    add_scene.go + _test.go
    connect_locations.go + _test.go
    archive_location.go + _test.go
```

---

## Consequences

- ArchiveLocation is the most complex interactor — two cross-aggregate
  guard checks before the domain call. These checks are intentionally
  at the interactor boundary, not in the domain (ADR-018 rationale).
- Archive cascade (children, scenes) is handled by LocationArchivedHandler
  via EventBus — the interactor only archives the root location.
- ConnectLocations is intentionally one-directional — symmetric connections
  are a UI concern. The API layer may expose a "connect both ways" mutation
  that calls this interactor twice.