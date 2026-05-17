# ADR-020: Ownership, Authentication, and Authorization

## Status
Accepted

## Date
2026-05-17

---

## Context

Grimoire is a GM tool. Every aggregate in the domain is owned by the GM
who created it. No multi-tenancy. No role hierarchy. No shared ownership.

The domain model had no concept of ownership before this ADR. Interactors
had no way to verify that the caller is permitted to act on a given
resource.

Two problems to solve:

1. **Ownership** — which GM created this Game, and therefore owns
   everything beneath it?

2. **Authentication** — is the caller who they say they are? And how
   do we verify this in local development without requiring Firebase?

---

## Ownership Model

### The Rule

```
A GM owns a Game.
All resources scoped to that Game are owned by the same GM.

Game        →  gmID (Firebase UID of the creating GM)
Campaign    →  owned via Game
NPC         →  owned via Game
Location    →  owned via Game
Faction     →  owned via Game
MacGuffin   →  owned via Game
Beat        →  owned via Game
```

There is exactly one owner per Game. Co-GM, player access, and
multi-user collaboration are deferred to a future ADR.

### Game.gmID

`Game` gains one new field at construction:

```go
// gameCore — shared state for all Game states
type gameCore struct {
    id                 identity.GameID
    gmID               string          // ← NEW: Firebase UID of owning GM
    masterNarrativeID  identity.MasterNarrativeID
    campaignIDs        []identity.CampaignID
    activeCampaignCount int
}
```

`gmID` is required at construction. Empty string is rejected:

```go
// CreateGame constructs a new Game aggregate.
// gmID is the Firebase UID of the GM creating the game.
func CreateGame(
    id    identity.GameID,
    gmID  string,
    name  string,
    source event.Source,
) (NewGame, []event.Event, error)
```

New guard:
```
strings.TrimSpace(gmID) == ""  →  ErrGMIDRequired
```

New error:
```go
ErrGMIDRequired = errors.New("game: gm id is required")
```

`gmID` is exposed via getter:
```go
func (g *gameCore) GMID() string { return g.gmID }
```

`GMID()` is added to the `Game` interface and all state interfaces
(NewGame, DraftGame, ActiveGame, IdleGame, ArchivedGame).

`GameSnapshot` gains `GMID string` — round-trips through Firestore.

### Authorization Check in Interactors

Every interactor that loads a Game (directly or via a Game-scoped
resource) performs this check immediately after load:

```go
if req.CallerID != game.GMID() {
    return ErrUnauthorized
}
```

This check happens **before** any domain method is called.

```go
var ErrUnauthorized = errors.New("interactor: caller is not authorized")
```

`ErrUnauthorized` lives in `grimoire-domain/shared/interactor/errors.go`
so all interactor packages can import it without circular dependencies.

---

## Authentication — CallerIdentityPort

### The Port

Authentication is a port. The interactor layer defines the interface.
Infrastructure provides the adapter. The domain never sees tokens.

```go
// grimoire-domain/shared/interactor/caller_identity.go

// CallerIdentityPort verifies a caller token and returns the caller's ID.
// The interactor layer trusts the returned ID completely.
// Authentication has already been performed by the adapter.
type CallerIdentityPort interface {
    Identify(ctx context.Context, token string) (string, error)
}
```

### Every Interactor Request Carries CallerID

The API layer (GraphQL handlers, ADR-008) resolves the caller token,
calls `CallerIdentityPort.Identify()`, and places the result in the
request before passing it to the interactor.

```go
// Example — all interactor requests follow this pattern
type CreateGameRequest struct {
    CallerID string          // verified Firebase UID — set by API layer
    ID       identity.GameID
    Name     string
    Source   event.Source
}
```

The interactor never calls `CallerIdentityPort.Identify()` itself.
The port is used by the API handler layer only. By the time a request
reaches an interactor, `CallerID` is a verified string.

---

## Adapters

### Production — Firebase

```go
// grimoire-infrastructure/auth/firebase_caller_identity.go

// FirebaseCallerIdentityAdapter verifies Firebase JWTs.
// Used in production only.
type FirebaseCallerIdentityAdapter struct {
    client *auth.Client
}

func (a *FirebaseCallerIdentityAdapter) Identify(
    ctx context.Context,
    token string,
) (string, error) {
    t, err := a.client.VerifyIDToken(ctx, token)
    if err != nil {
        return "", fmt.Errorf("firebase auth: %w", err)
    }
    return t.UID, nil
}
```

### Local Development — DevAuth

```go
// grimoire-infrastructure/auth/dev_caller_identity.go

// DevCallerIdentityAdapter trusts the caller token directly as the caller ID.
// If the token is empty, DefaultCallerID is returned.
//
// WARNING: This adapter performs NO verification.
// It MUST NOT be used in production.
// Controlled by GRIMOIRE_ENV environment variable in main.go.
type DevCallerIdentityAdapter struct {
    DefaultCallerID string
}

func (a *DevCallerIdentityAdapter) Identify(
    ctx context.Context,
    token string,
) (string, error) {
    if strings.TrimSpace(token) != "" {
        return token, nil
    }
    return a.DefaultCallerID, nil
}
```

### Wiring in main.go

```go
var authAdapter interactor.CallerIdentityPort

switch os.Getenv("GRIMOIRE_ENV") {
case "production":
    authAdapter = auth.NewFirebaseCallerIdentityAdapter(firebaseApp)
default:
    authAdapter = auth.NewDevCallerIdentityAdapter("dev-gm-uid-001")
    log.Warn("DevCallerIdentityAdapter active — not safe for production")
}
```

---

## Interactor Test Pattern

Tests construct requests with a literal `CallerID`. No Firebase. No mocks.

```go
func TestCreateGame_Succeeds(t *testing.T) {
    repo := newInMemoryGameRepo()
    bus  := newNoOpEventBus()

    i := interactor.NewCreateGameInteractor(repo, bus)

    result, err := i.Execute(ctx, interactor.CreateGameRequest{
        CallerID: "test-gm-uid-001",
        ID:       identity.NewGameID(),
        Name:     "Ashes & Chains",
        Source:   event.SourceGrimoire,
    })

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.Game.GMID() != "test-gm-uid-001" {
        t.Fatal("expected gmID to be set from CallerID")
    }
}

func TestCreateGame_UnauthorizedCaller_Rejected(t *testing.T) {
    // For interactors that load an existing resource and check ownership
    repo := newInMemoryGameRepoWithGame("owner-uid-001")
    bus  := newNoOpEventBus()

    i := interactor.NewSomeGameInteractor(repo, bus)

    _, err := i.Execute(ctx, interactor.SomeGameRequest{
        CallerID: "different-uid-999",
        GameID:   knownGameID,
    })

    if !errors.Is(err, interactor.ErrUnauthorized) {
        t.Fatalf("expected ErrUnauthorized, got %v", err)
    }
}
```

---

## What This ADR Does Not Cover

```
Co-GM / table collaborators    →  future ADR
Player app authorization       →  future ADR (Player context)
Read authorization (queries)   →  future ADR (Neo4j query layer)
Role-based access control      →  future ADR if ever needed
OAuth scopes                   →  future ADR
Firebase Security Rules        →  future ADR (Player app)
```

---

## Breaking Changes

- `CreateGame` signature changes — `gmID string` added as second parameter
- All callers of `CreateGame` must pass `gmID`
- `GameSnapshot` gains `GMID string` — existing snapshots need migration
  (existing documents have no GMID field — read as empty string, which
  will fail the `ErrGMIDRequired` guard on next save; operator migration
  required before deployment)
- All Game state interfaces gain `GMID() string`

---

## Consequences

- Every write operation is authorized with a single string equality check
- Firebase handles authentication — the domain never sees a token
- Local development requires no Firebase setup — `GRIMOIRE_ENV` controls the adapter
- Interactor tests are plain Go — no Firebase, no mocks of Firebase
- The ownership model is simple enough to reason about completely
- Future multi-user or Player app access requires a new ADR — not a refactor of this one

## Alternatives Considered

**AuthorizationPort interface in interactors** — rejected for now.
Adds ceremony without benefit until roles exist. A single equality check
is the entire authorization model. When roles arrive, extract then.

**Firebase UID checked at API layer only** — rejected. The interactor
layer is the domain boundary. Authorization belongs there, not in the
HTTP/GraphQL handler. Handlers should not enforce domain rules.

**Separate ownership store** — rejected. Game already lives in Firestore.
Adding gmID to the Game document costs one field and avoids a second lookup.