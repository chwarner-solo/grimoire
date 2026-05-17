# ADR-022: Campaign Interactors

## Status
Accepted

## Date
2026-05-17

---

## Context

A Campaign is a table running a Game. The GM creates it, assembles the
party, runs sessions, and eventually completes it. Campaign interactors
cover the full lifecycle.

Authorization: every interactor loads the Game by GameID, checks
CallerID == game.GMID(), then loads the Campaign. Two loads per
write — acceptable. Game documents are small.

---

## Ports

```go
// grimoire-domain/roots/campaign/interactor/ports.go

type GameRepository interface {
    Load(ctx context.Context, id identity.GameID) (gameentity.Game, error)
}

type CampaignRepository interface {
    Save(ctx context.Context, campaign entity.Campaign) error
    Load(ctx context.Context, id identity.CampaignID) (entity.Campaign, error)
}
```

---

## Interactor: CreateCampaign

**File:** `grimoire-domain/roots/campaign/interactor/create_campaign.go`

Creates a new Campaign and links it to the Game via EventBus.
The GameStatusHandler (ADR-011) reacts to CampaignLinked and calls
Game.LinkCampaign() — the interactor does not touch the Game aggregate.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    Name        string
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound if absent
    auth: req.CallerID != game.GMID() → ErrUnauthorized
    entity.CreateCampaign(req.CampaignID, req.Name, req.GameID, req.Source)
      → domain validation errors
    campaignRepo.Save(campaign) → ErrRepositorySaveFailed
    bus.Dispatch(EntityCreated{ entity_type: "campaign" })
    bus.Dispatch(EntityLinked{ relationship: "campaign",
                               entity_a: gameID, entity_b: campaignID })

Result:
    Campaign  entity.NewCampaign
    Events    []event.Event

AggregateType: event.AggregateCampaign
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CreatesCampaign_Succeeds | valid request, game in repo | NewCampaign returned |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| GameNotFound_ReturnsError | empty game repo | ErrGameNotFound |
| EmptyName_ReturnsError | Name="" | domain error, no save |
| SaveFailure_NoDispatch | saveErr set | ErrRepositorySaveFailed, 0 dispatched |
| DispatchesEntityCreatedEvent | valid request | TypeEntityCreated envelope |
| DispatchesEntityLinkedEvent | valid request | TypeEntityLinked envelope |

---

## Interactor: AddCharacterToCampaign

**File:** `grimoire-domain/roots/campaign/interactor/add_character.go`

Adds a PlayerCharacter to a NewCampaign or FormingCampaign.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    CharacterID identity.PlayerCharacterID
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    campaignRepo.Load(req.CampaignID) → ErrCampaignNotFound
    switch campaign state:
        NewCampaign     → nc.AddCharacter(req.CharacterID, req.Source)
        FormingCampaign → fc.AddCharacter(req.CharacterID, req.Source)
        otherwise       → ErrInvalidCampaignState
    campaignRepo.Save(updated) → ErrRepositorySaveFailed
    bus.Dispatch(EntityLinked{ relationship: "participates_in" })

Result:
    Campaign  entity.Campaign
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| AddToNewCampaign_Succeeds | new campaign | campaign saved, EntityLinked dispatched |
| AddToFormingCampaign_Succeeds | forming campaign | campaign saved |
| AddToActiveCampaign_ReturnsError | active campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DuplicateCharacter_ReturnsError | char already added | domain error |
| ZeroCharacterID_ReturnsError | CharacterID zero | domain error, no save |

---

## Interactor: BeginCampaignFormation

**File:** `grimoire-domain/roots/campaign/interactor/begin_formation.go`

Transitions Campaign from New → Forming. GM signals party assembly
has begun.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    campaignRepo.Load(req.CampaignID) → ErrCampaignNotFound
    nc, ok := campaign.(entity.NewCampaign)
      → ErrInvalidCampaignState if not NewCampaign
    forming, events, err := nc.BeginFormation(req.Source)
    campaignRepo.Save(forming) → ErrRepositorySaveFailed
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "forming" })

Result:
    Campaign  entity.FormingCampaign
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| BeginFormation_Succeeds | new campaign | FormingCampaign returned |
| AlreadyForming_ReturnsError | forming campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| SaveFailure_NoDispatch | saveErr set | no events dispatched |

---

## Interactor: StartFirstSession

**File:** `grimoire-domain/roots/campaign/interactor/start_first_session.go`

Transitions FormingCampaign → Active. Requires at least one character
(enforced by domain). Emits SessionStarted which drives Game state
via GameStatusHandler (ADR-011).

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    SessionID   identity.SessionID
    Date        time.Time
    Source      event.Source

Flow:
    gameRepo.Load(req.GameID) → ErrGameNotFound
    auth check → ErrUnauthorized
    campaignRepo.Load(req.CampaignID) → ErrCampaignNotFound
    fc, ok := campaign.(entity.FormingCampaign)
      → ErrInvalidCampaignState if not FormingCampaign
    active, events, err := fc.StartFirstSession(req.SessionID, req.Date)
      → ErrNoCharactersInCampaign if characterIDs empty (domain guard)
    campaignRepo.Save(active) → ErrRepositorySaveFailed
    bus.Dispatch(SessionStarted{ session_id, campaign_id, date })

Result:
    Campaign  entity.ActiveCampaign
    Events    []event.Event

Note: SessionStarted triggers GameStatusHandler → Game transitions
      to Active (ADR-011). Interactor does not touch Game directly.
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| StartFirstSession_Succeeds | forming + 1 character | ActiveCampaign returned |
| NoCharacters_ReturnsError | forming, 0 characters | ErrNoCharactersInCampaign |
| NotFormingCampaign_ReturnsError | new campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| ZeroSessionID_ReturnsError | SessionID zero | domain error |
| SaveFailure_NoDispatch | saveErr set | ErrRepositorySaveFailed, no dispatch |
| DispatchesSessionStartedEvent | succeeds | SessionStarted envelope |

---

## Interactor: StartNewSession

**File:** `grimoire-domain/roots/campaign/interactor/start_new_session.go`

Transitions IdleCampaign → Active for a follow-on session.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    SessionID   identity.SessionID
    Date        time.Time
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    campaignRepo.Load → ErrCampaignNotFound
    ic, ok := campaign.(entity.IdleCampaign)
      → ErrInvalidCampaignState
    active, events, err := ic.StartNewSession(req.SessionID, req.Date)
    campaignRepo.Save(active) → ErrRepositorySaveFailed
    bus.Dispatch(SessionStarted)

Result:
    Campaign  entity.ActiveCampaign
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| StartNewSession_Succeeds | idle campaign | ActiveCampaign |
| NotIdleCampaign_ReturnsError | active campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DispatchesSessionStartedEvent | succeeds | SessionStarted envelope |

---

## Interactor: EndSession

**File:** `grimoire-domain/roots/campaign/interactor/end_session.go`

Transitions ActiveCampaign → Idle. Called when the GM wraps the
session. Emits SessionEnded which drives Game state via
GameStatusHandler.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    SessionID   identity.SessionID
    Notes       string    // optional session summary
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    campaignRepo.Load → ErrCampaignNotFound
    ac, ok := campaign.(entity.ActiveCampaign)
      → ErrInvalidCampaignState
    idle, events, err := ac.NotifySessionSummarized()
    campaignRepo.Save(idle) → ErrRepositorySaveFailed
    bus.Dispatch(SessionEnded{ session_id, notes })

Result:
    Campaign  entity.IdleCampaign
    Events    []event.Event

Note: SessionEnded triggers GameStatusHandler → Game transitions
      to Idle if no other active campaigns (ADR-011).
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| EndSession_Succeeds | active campaign | IdleCampaign returned |
| NotActiveCampaign_ReturnsError | idle campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DispatchesSessionEndedEvent | succeeds | SessionEnded envelope |
| NotesAreOptional | Notes="" | succeeds, SessionEnded.Notes="" |

---

## Interactor: MoveParty

**File:** `grimoire-domain/roots/campaign/interactor/move_party.go`

Updates the party's current location on an ActiveCampaign.
Emits EntityLinked{ relationship: "located_at" } for Neo4j tracking.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    LocationID  identity.LocationID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    campaignRepo.Load → ErrCampaignNotFound
    ac, ok := campaign.(entity.ActiveCampaign)
      → ErrInvalidCampaignState (party can only move mid-session)
    updated, events, err := ac.MoveToLocation(req.LocationID, req.Source)
    campaignRepo.Save(updated) → ErrRepositorySaveFailed
    bus.Dispatch(EntityUpdated{ field: "current_location" })
    bus.Dispatch(EntityLinked{ relationship: "located_at" })

Result:
    Campaign  entity.ActiveCampaign
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| MoveParty_Succeeds | active campaign | location updated |
| NotActiveCampaign_ReturnsError | idle campaign | ErrInvalidCampaignState |
| ZeroLocationID_ReturnsError | LocationID zero | domain error |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DispatchesTwoEvents | succeeds | EntityUpdated + EntityLinked |

---

## Interactor: CompleteCampaign

**File:** `grimoire-domain/roots/campaign/interactor/complete_campaign.go`

Terminal operation. Transitions IdleCampaign → Complete.

```
Request:
    CallerID    string
    CampaignID  identity.CampaignID
    GameID      identity.GameID
    Source      event.Source

Flow:
    gameRepo.Load → auth check
    campaignRepo.Load → ErrCampaignNotFound
    ic, ok := campaign.(entity.IdleCampaign)
      → ErrInvalidCampaignState
    complete, events, err := ic.Complete(req.Source)
    campaignRepo.Save(complete) → ErrRepositorySaveFailed
    bus.Dispatch(EntityUpdated{ field: "status", new_value: "complete" })

Result:
    Campaign  entity.CompleteCampaign
    Events    []event.Event
```

**Test cases:**

| Test | Setup | Expected |
|------|-------|----------|
| CompleteCampaign_Succeeds | idle campaign | CompleteCampaign returned |
| ActiveCampaign_ReturnsError | active campaign | ErrInvalidCampaignState |
| WrongCaller_ReturnsUnauthorized | CallerID != gmID | ErrUnauthorized |
| DispatchesEntityUpdatedEvent | succeeds | field="status", new_value="complete" |

---

## Errors

```go
// grimoire-domain/roots/campaign/interactor/errors.go

var (
    ErrRepositorySaveFailed   = errors.New("interactor: failed to save campaign")
    ErrRepositoryLoadFailed   = errors.New("interactor: failed to load campaign")
    ErrEventDispatchFailed    = errors.New("interactor: failed to dispatch event")
    ErrGameNotFound           = errors.New("interactor: game not found")
    ErrCampaignNotFound       = errors.New("interactor: campaign not found")
    ErrInvalidCampaignState   = errors.New("interactor: campaign is not in required state")
)

// ErrUnauthorized from grimoire-domain/shared/interactor/errors.go
```

---

## File Locations

```
grimoire-domain/roots/campaign/interactor/
    ports.go                  GameRepository, CampaignRepository
    errors.go
    create_campaign.go
    create_campaign_test.go
    add_character.go
    add_character_test.go
    begin_formation.go
    begin_formation_test.go
    start_first_session.go
    start_first_session_test.go
    start_new_session.go
    start_new_session_test.go
    end_session.go
    end_session_test.go
    move_party.go
    move_party_test.go
    complete_campaign.go
    complete_campaign_test.go
```

---

## Consequences

- Every campaign write loads the Game for authorization — two repository
  calls per request. Acceptable: Game is a single Firestore document.
- MoveParty is only valid during an active session — consistent with
  the domain: party movement is a mid-session event.
- CompleteCampaign is terminal — no recovery. Matches domain design.
- Session state (notes, date, recap) is deferred to a future Session ADR.
  EndSession carries optional notes via SessionEnded event payload only.