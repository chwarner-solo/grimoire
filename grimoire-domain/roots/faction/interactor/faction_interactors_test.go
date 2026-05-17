package interactor

import (
	"context"
	"errors"
	"testing"

	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	factionvalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/value"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// --- Test doubles ---

type stubFactionGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubFactionGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubFactionRepo struct {
	faction    factionentity.Faction
	saved      factionentity.Faction
	membership *factionentity.FactionMembership
	saveErr    error
	loadErr    error
}

func (r *stubFactionRepo) Save(_ context.Context, f factionentity.Faction) error {
	r.saved = f
	return r.saveErr
}

func (r *stubFactionRepo) Load(_ context.Context, _ identity.FactionID) (factionentity.Faction, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.faction, nil
}

func (r *stubFactionRepo) SaveMembership(_ context.Context, m *factionentity.FactionMembership) error {
	r.membership = m
	return r.saveErr
}

func (r *stubFactionRepo) LoadMembership(_ context.Context, _ identity.FactionMembershipID) (*factionentity.FactionMembership, error) {
	return r.membership, r.loadErr
}

type spyFactionBus struct {
	dispatched  []event.EventEnvelope
	dispatchErr error
}

func (b *spyFactionBus) Dispatch(_ context.Context, e event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.dispatched = append(b.dispatched, e)
	return nil
}

func (b *spyFactionBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustFactionGame(t *testing.T) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: "gm-uid", Name: "Test Game", State: "draft",
	})
	if err != nil {
		t.Fatalf("mustFactionGame: %v", err)
	}
	return g, id
}

func mustFaction(t *testing.T, state string, gameID identity.GameID) (factionentity.Faction, identity.FactionID) {
	t.Helper()
	id := identity.NewFactionID()
	f, err := factionentity.ReconstituteFaction(factionentity.FactionSnapshot{
		ID: id, GameID: gameID, Name: "Test Faction", State: state,
	})
	if err != nil {
		t.Fatalf("mustFaction(%s): %v", state, err)
	}
	return f, id
}

func mustStandingLevel(t *testing.T, name string, threshold, ordinal int) factionvalue.StandingLevel {
	t.Helper()
	sl, err := factionvalue.NewStandingLevel(name, threshold, ordinal)
	if err != nil {
		t.Fatalf("mustStandingLevel: %v", err)
	}
	return sl
}

// --- CreateFaction tests ---

func TestCreateFaction_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{}
	bus := &spyFactionBus{}
	i := NewCreateFactionInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), CreateFactionRequest{
		CallerID: "gm-uid", ID: identity.NewFactionID(), GameID: gid,
		Name: "The Iron Circle", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Faction == nil {
		t.Fatal("expected non-nil faction")
	}
	if factionRepo.saved == nil {
		t.Fatal("expected faction to be saved")
	}
}

func TestCreateFaction_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{}
	bus := &spyFactionBus{}
	i := NewCreateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), CreateFactionRequest{
		CallerID: "wrong", ID: identity.NewFactionID(), GameID: gid, Name: "Test", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCreateFaction_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{}
	bus := &spyFactionBus{}
	i := NewCreateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), CreateFactionRequest{
		CallerID: "gm-uid", ID: identity.NewFactionID(), GameID: gid, Name: "", Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if factionRepo.saved != nil {
		t.Fatal("expected no save")
	}
}

func TestCreateFaction_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustFactionGame(t)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{saveErr: errors.New("db error")}
	bus := &spyFactionBus{}
	i := NewCreateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), CreateFactionRequest{
		CallerID: "gm-uid", ID: identity.NewFactionID(), GameID: gid, Name: "Test", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(bus.dispatched))
	}
}

// --- BeginFactionDraft tests ---

func TestBeginFactionDraft_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "new", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewBeginFactionDraftInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), BeginFactionDraftRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.DraftFaction); !ok {
		t.Fatalf("expected DraftFaction, got %T", result.Faction)
	}
}

func TestBeginFactionDraft_AlreadyDraft_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewBeginFactionDraftInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), BeginFactionDraftRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

func TestBeginFactionDraft_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "new", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewBeginFactionDraftInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), BeginFactionDraftRequest{
		CallerID: "wrong", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- ActivateFaction tests ---

func TestActivateFaction_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewActivateFactionInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), ActivateFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", result.Faction)
	}
}

func TestActivateFaction_SparseDraft_StillSucceeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewActivateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ActivateFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("sparse draft should activate: %v", err)
	}
}

func TestActivateFaction_NotDraft_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "new", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewActivateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ActivateFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

func TestActivateFaction_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewActivateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ActivateFactionRequest{
		CallerID: "wrong", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- AddFactionMember tests ---

func TestAddFactionMember_ToDraft_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddFactionMemberInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), AddFactionMemberRequest{
		CallerID:     "gm-uid",
		FactionID:    fid,
		MembershipID: identity.NewFactionMembershipID(),
		NPCID:        identity.NewNarrativeCharacterID(),
		Rank:         "Initiate",
		GameID:       gid,
		Source:       event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) < 2 {
		t.Fatalf("expected at least 2 events (EntityLinked from faction + EntityLinked for NPC), got %d", len(result.Events))
	}
	if factionRepo.membership == nil {
		t.Fatal("expected membership to be saved")
	}
}

func TestAddFactionMember_ToActive_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddFactionMemberInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddFactionMemberRequest{
		CallerID:     "gm-uid",
		FactionID:    fid,
		MembershipID: identity.NewFactionMembershipID(),
		NPCID:        identity.NewNarrativeCharacterID(),
		Rank:         "Member",
		GameID:       gid,
		Source:       event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddFactionMember_ToIdle_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "idle", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddFactionMemberInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddFactionMemberRequest{
		CallerID:     "gm-uid",
		FactionID:    fid,
		MembershipID: identity.NewFactionMembershipID(),
		NPCID:        identity.NewNarrativeCharacterID(),
		Rank:         "Member",
		GameID:       gid,
		Source:       event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

func TestAddFactionMember_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddFactionMemberInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddFactionMemberRequest{
		CallerID: "wrong", FactionID: fid, MembershipID: identity.NewFactionMembershipID(),
		NPCID: identity.NewNarrativeCharacterID(), Rank: "Member", GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- AddStandingLevel tests ---

func TestAddStandingLevel_ToDraft_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddStandingLevelInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddStandingLevelRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid,
		Level: mustStandingLevel(t, "Hostile", 0, 1), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddStandingLevel_ToActive_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddStandingLevelInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddStandingLevelRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid,
		Level: mustStandingLevel(t, "Allied", 2, 5), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddStandingLevel_ToIdle_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "idle", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewAddStandingLevelInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), AddStandingLevelRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid,
		Level: mustStandingLevel(t, "Neutral", 5, 3), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

// --- DeclareAlly tests ---

func TestDeclareAlly_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	allyID := identity.NewFactionID()
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewDeclareAllyInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), DeclareAllyRequest{
		CallerID: "gm-uid", FactionID: fid, AllyID: allyID, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeclareAlly_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewDeclareAllyInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), DeclareAllyRequest{
		CallerID: "wrong", FactionID: fid, AllyID: identity.NewFactionID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- DeclareWar tests ---

func TestDeclareWar_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	enemyID := identity.NewFactionID()
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewDeclareWarInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), DeclareWarRequest{
		CallerID: "gm-uid", FactionID: fid, EnemyID: enemyID, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeclareWar_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewDeclareWarInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), DeclareWarRequest{
		CallerID: "wrong", FactionID: fid, EnemyID: identity.NewFactionID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- MarkFactionDormant tests ---

func TestMarkFactionDormant_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewMarkFactionDormantInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), MarkFactionDormantRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.IdleFaction); !ok {
		t.Fatalf("expected IdleFaction, got %T", result.Faction)
	}
}

func TestMarkFactionDormant_NotActive_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "idle", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewMarkFactionDormantInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), MarkFactionDormantRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

// --- ReactivateFaction tests ---

func TestReactivateFaction_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "idle", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewReactivateFactionInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), ReactivateFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.ActiveFaction); !ok {
		t.Fatalf("expected ActiveFaction, got %T", result.Faction)
	}
}

func TestReactivateFaction_NotIdle_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewReactivateFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ReactivateFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

// --- ArchiveFaction tests ---

func TestArchiveFaction_FromActive_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewArchiveFactionInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", result.Faction)
	}
}

func TestArchiveFaction_FromIdle_Succeeds(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "idle", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewArchiveFactionInteractor(gameRepo, factionRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Faction.(factionentity.ArchivedFaction); !ok {
		t.Fatalf("expected ArchivedFaction, got %T", result.Faction)
	}
}

func TestArchiveFaction_FromDraft_ReturnsError(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "draft", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewArchiveFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveFactionRequest{
		CallerID: "gm-uid", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidFactionState) {
		t.Fatalf("expected ErrInvalidFactionState, got: %v", err)
	}
}

func TestArchiveFaction_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustFactionGame(t)
	faction, fid := mustFaction(t, "active", gid)
	gameRepo := &stubFactionGameRepo{game: game}
	factionRepo := &stubFactionRepo{faction: faction}
	bus := &spyFactionBus{}
	i := NewArchiveFactionInteractor(gameRepo, factionRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveFactionRequest{
		CallerID: "wrong", FactionID: fid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}
