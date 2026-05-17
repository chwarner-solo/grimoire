package interactor

import (
	"context"
	"errors"
	"testing"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// --- Test doubles ---

type stubCharGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubCharGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubNPCRepo struct {
	snap    characterentity.NPCSnapshot
	saved   characterentity.NPCSnapshot
	saveErr error
	loadErr error
}

func (r *stubNPCRepo) Save(_ context.Context, s characterentity.NPCSnapshot) error {
	r.saved = s
	return r.saveErr
}

func (r *stubNPCRepo) Load(_ context.Context, _ identity.NarrativeCharacterID) (characterentity.NPCSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.NPCSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type stubPCRepo struct {
	snap    characterentity.PlayerCharacterSnapshot
	saved   characterentity.PlayerCharacterSnapshot
	saveErr error
	loadErr error
}

func (r *stubPCRepo) SaveCore(_ context.Context, s characterentity.PlayerCharacterSnapshot) error {
	r.saved = s
	return r.saveErr
}

func (r *stubPCRepo) LoadCore(_ context.Context, _ identity.PlayerCharacterID) (characterentity.PlayerCharacterSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.PlayerCharacterSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type stubMGRepo struct {
	snap    characterentity.MacGuffinSnapshot
	saved   characterentity.MacGuffinSnapshot
	saveErr error
	loadErr error
}

func (r *stubMGRepo) Save(_ context.Context, s characterentity.MacGuffinSnapshot) error {
	r.saved = s
	return r.saveErr
}

func (r *stubMGRepo) Load(_ context.Context, _ identity.MacGuffinID) (characterentity.MacGuffinSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.MacGuffinSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type spyCharBus struct {
	envelopes   []event.EventEnvelope
	dispatchErr error
}

func (b *spyCharBus) Dispatch(_ context.Context, env event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.envelopes = append(b.envelopes, env)
	return nil
}

func (b *spyCharBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustCharGame(t *testing.T) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: "gm-uid", Name: "Test Game", State: "active",
	})
	if err != nil {
		t.Fatalf("mustCharGame: %v", err)
	}
	return g, id
}

func mustNPCSnap(t *testing.T, state string, gameID identity.GameID) (characterentity.NPCSnapshot, identity.NarrativeCharacterID) {
	t.Helper()
	id := identity.NewNarrativeCharacterID()
	return characterentity.NPCSnapshot{
		ID: id, GameID: gameID, Name: "Test NPC", State: state,
	}, id
}

func mustPCSnap(t *testing.T, status characterentity.PlayerCharacterStatus, gameID identity.GameID) (characterentity.PlayerCharacterSnapshot, identity.PlayerCharacterID) {
	t.Helper()
	id := identity.NewPlayerCharacterID()
	return characterentity.PlayerCharacterSnapshot{
		ID: id, GameID: gameID, Name: "Test PC", Status: status,
	}, id
}

func mustMGSnap(t *testing.T, destroyed bool, gameID identity.GameID) (characterentity.MacGuffinSnapshot, identity.MacGuffinID) {
	t.Helper()
	id := identity.NewMacGuffinID()
	return characterentity.MacGuffinSnapshot{
		ID: id, GameID: gameID, Name: "The Orb", Destroyed: destroyed,
	}, id
}

// === CreateNPC tests ===

func TestCreateNPC_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{}
	bus := &spyCharBus{}
	i := NewCreateNPCInteractor(gameRepo, npcRepo, bus)

	result, err := i.Execute(context.Background(), CreateNPCRequest{
		CallerID: "gm-uid", ID: identity.NewNarrativeCharacterID(),
		GameID: gid, Name: "Lady Vex", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPC == nil {
		t.Fatal("expected non-nil NPC")
	}
	if npcRepo.saved.ID.IsZero() {
		t.Fatal("expected NPC to be saved")
	}
	if len(bus.envelopes) != 1 {
		t.Fatalf("expected 1 event dispatched, got %d", len(bus.envelopes))
	}
}

func TestCreateNPC_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreateNPCInteractor(gameRepo, &stubNPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreateNPCRequest{
		CallerID: "not-gm", ID: identity.NewNarrativeCharacterID(),
		GameID: gid, Name: "Some NPC", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCreateNPC_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreateNPCInteractor(gameRepo, &stubNPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreateNPCRequest{
		CallerID: "gm-uid", ID: identity.NewNarrativeCharacterID(),
		GameID: gid, Name: "", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrNPCNameRequired) {
		t.Fatalf("expected ErrNPCNameRequired, got: %v", err)
	}
}

func TestCreateNPC_ZeroID_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreateNPCInteractor(gameRepo, &stubNPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreateNPCRequest{
		CallerID: "gm-uid", ID: identity.NarrativeCharacterID{},
		GameID: gid, Name: "Some NPC", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrNPCIDRequired) {
		t.Fatalf("expected ErrNPCIDRequired, got: %v", err)
	}
}

// === BeginNPCDraft tests ===

func TestBeginNPCDraft_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "new", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	bus := &spyCharBus{}
	i := NewBeginNPCDraftInteractor(gameRepo, npcRepo, bus)

	result, err := i.Execute(context.Background(), BeginNPCDraftRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPC == nil {
		t.Fatal("expected non-nil DraftNPC")
	}
	if npcRepo.saved.State != "draft" {
		t.Fatalf("expected saved state 'draft', got %q", npcRepo.saved.State)
	}
}

func TestBeginNPCDraft_AlreadyDraft_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewBeginNPCDraftInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), BeginNPCDraftRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestBeginNPCDraft_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "new", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewBeginNPCDraftInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), BeginNPCDraftRequest{
		CallerID: "wrong", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === UpdateNPCContent tests ===

func TestUpdateNPCContent_Draft_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	bus := &spyCharBus{}
	i := NewUpdateNPCContentInteractor(gameRepo, npcRepo, bus)

	_, err := i.Execute(context.Background(), UpdateNPCContentRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid,
		Name: "Updated Name", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if npcRepo.saved.Name != "Updated Name" {
		t.Fatalf("expected saved name 'Updated Name', got %q", npcRepo.saved.Name)
	}
}

func TestUpdateNPCContent_Active_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewUpdateNPCContentInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateNPCContentRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid,
		Name: "New Name", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateNPCContent_Archived_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "archived", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewUpdateNPCContentInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateNPCContentRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid,
		Name: "Some Name", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestUpdateNPCContent_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewUpdateNPCContentInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateNPCContentRequest{
		CallerID: "wrong", NPCID: nid, GameID: gid,
		Name: "Some Name", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === ActivateNPC tests ===

func TestActivateNPC_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	snap.Name = "Lady Vex"
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	bus := &spyCharBus{}
	i := NewActivateNPCInteractor(gameRepo, npcRepo, bus)

	result, err := i.Execute(context.Background(), ActivateNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPC == nil {
		t.Fatal("expected non-nil ActiveNPC")
	}
	if npcRepo.saved.State != "active" {
		t.Fatalf("expected saved state 'active', got %q", npcRepo.saved.State)
	}
}

func TestActivateNPC_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	snap.Name = ""
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewActivateNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), ActivateNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name on activation")
	}
}

func TestActivateNPC_NotDraft_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "new", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewActivateNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), ActivateNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestActivateNPC_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewActivateNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), ActivateNPCRequest{
		CallerID: "wrong", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === ArchiveNPC tests ===

func TestArchiveNPC_Active_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	bus := &spyCharBus{}
	i := NewArchiveNPCInteractor(gameRepo, npcRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPC == nil {
		t.Fatal("expected non-nil ArchivedNPC")
	}
	if npcRepo.saved.State != "archived" {
		t.Fatalf("expected saved state 'archived', got %q", npcRepo.saved.State)
	}
}

func TestArchiveNPC_Idle_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "idle", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewArchiveNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	result, err := i.Execute(context.Background(), ArchiveNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NPC == nil {
		t.Fatal("expected non-nil ArchivedNPC")
	}
}

func TestArchiveNPC_Draft_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewArchiveNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), ArchiveNPCRequest{
		CallerID: "gm-uid", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestArchiveNPC_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: snap}
	i := NewArchiveNPCInteractor(gameRepo, npcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), ArchiveNPCRequest{
		CallerID: "wrong", NPCID: nid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === CreatePlayerCharacter tests ===

func TestCreatePlayerCharacter_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{}
	bus := &spyCharBus{}
	i := NewCreatePlayerCharacterInteractor(gameRepo, pcRepo, bus)

	result, err := i.Execute(context.Background(), CreatePlayerCharacterRequest{
		CallerID: "gm-uid", ID: identity.NewPlayerCharacterID(),
		GameID: gid, Name: "Aria", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlayerCharacter == nil {
		t.Fatal("expected non-nil PlayerCharacter")
	}
	if pcRepo.saved.ID.IsZero() {
		t.Fatal("expected PC to be saved")
	}
}

func TestCreatePlayerCharacter_WithOwner_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{}
	i := NewCreatePlayerCharacterInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreatePlayerCharacterRequest{
		CallerID: "gm-uid", ID: identity.NewPlayerCharacterID(),
		GameID: gid, Name: "Aria", OwnerPlayerID: "player-42",
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pcRepo.saved.OwnerPlayerID != "player-42" {
		t.Fatalf("expected ownerPlayerID='player-42', got %q", pcRepo.saved.OwnerPlayerID)
	}
}

func TestCreatePlayerCharacter_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreatePlayerCharacterInteractor(gameRepo, &stubPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreatePlayerCharacterRequest{
		CallerID: "gm-uid", ID: identity.NewPlayerCharacterID(),
		GameID: gid, Name: "", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrCharacterNameRequired) {
		t.Fatalf("expected ErrCharacterNameRequired, got: %v", err)
	}
}

func TestCreatePlayerCharacter_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreatePlayerCharacterInteractor(gameRepo, &stubPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreatePlayerCharacterRequest{
		CallerID: "wrong", ID: identity.NewPlayerCharacterID(),
		GameID: gid, Name: "Aria", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCreatePlayerCharacter_ZeroID_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreatePlayerCharacterInteractor(gameRepo, &stubPCRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreatePlayerCharacterRequest{
		CallerID: "gm-uid", ID: identity.PlayerCharacterID{},
		GameID: gid, Name: "Aria", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrPlayerCharacterIDRequired) {
		t.Fatalf("expected ErrPlayerCharacterIDRequired, got: %v", err)
	}
}

// === UpdatePlayerCharacterContent tests ===

func TestUpdatePlayerCharacterContent_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	i := NewUpdatePlayerCharacterContentInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdatePlayerCharacterContentRequest{
		CallerID: "gm-uid", ID: pid, GameID: gid,
		Name: "Updated Aria", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdatePlayerCharacterContent_Retired_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusRetired, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	i := NewUpdatePlayerCharacterContentInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdatePlayerCharacterContentRequest{
		CallerID: "gm-uid", ID: pid, GameID: gid,
		Name: "Some Name", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrCharacterRetired) {
		t.Fatalf("expected ErrCharacterRetired, got: %v", err)
	}
}

func TestUpdatePlayerCharacterContent_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	i := NewUpdatePlayerCharacterContentInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdatePlayerCharacterContentRequest{
		CallerID: "wrong", ID: pid, GameID: gid,
		Name: "Some Name", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === RetirePlayerCharacter tests ===

func TestRetirePlayerCharacter_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	bus := &spyCharBus{}
	i := NewRetirePlayerCharacterInteractor(gameRepo, pcRepo, bus)

	result, err := i.Execute(context.Background(), RetirePlayerCharacterRequest{
		CallerID: "gm-uid", ID: pid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PlayerCharacter == nil {
		t.Fatal("expected non-nil result")
	}
	if pcRepo.saved.Status != characterentity.PlayerCharacterStatusRetired {
		t.Fatalf("expected status 'retired', got %q", pcRepo.saved.Status)
	}
}

func TestRetirePlayerCharacter_AlreadyRetired_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusRetired, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	i := NewRetirePlayerCharacterInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), RetirePlayerCharacterRequest{
		CallerID: "gm-uid", ID: pid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrCharacterAlreadyRetired) {
		t.Fatalf("expected ErrCharacterAlreadyRetired, got: %v", err)
	}
}

func TestRetirePlayerCharacter_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	snap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: snap}
	i := NewRetirePlayerCharacterInteractor(gameRepo, pcRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), RetirePlayerCharacterRequest{
		CallerID: "wrong", ID: pid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === CreateMacGuffin tests ===

func TestCreateMacGuffin_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{}
	bus := &spyCharBus{}
	i := NewCreateMacGuffinInteractor(gameRepo, mgRepo, bus)

	result, err := i.Execute(context.Background(), CreateMacGuffinRequest{
		CallerID: "gm-uid", ID: identity.NewMacGuffinID(),
		GameID: gid, Name: "The Orb of Power", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MacGuffin == nil {
		t.Fatal("expected non-nil MacGuffin")
	}
	if mgRepo.saved.ID.IsZero() {
		t.Fatal("expected MacGuffin to be saved")
	}
}

func TestCreateMacGuffin_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreateMacGuffinInteractor(gameRepo, &stubMGRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreateMacGuffinRequest{
		CallerID: "gm-uid", ID: identity.NewMacGuffinID(),
		GameID: gid, Name: "", Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateMacGuffin_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	i := NewCreateMacGuffinInteractor(gameRepo, &stubMGRepo{}, &spyCharBus{})

	_, err := i.Execute(context.Background(), CreateMacGuffinRequest{
		CallerID: "wrong", ID: identity.NewMacGuffinID(),
		GameID: gid, Name: "The Orb", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === AssignMacGuffinToNPC tests ===

func TestAssignMacGuffinToNPC_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	npcSnap, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: npcSnap}
	mgRepo := &stubMGRepo{snap: mgSnap}
	bus := &spyCharBus{}
	i := NewAssignMacGuffinToNPCInteractor(gameRepo, npcRepo, mgRepo, bus)

	result, err := i.Execute(context.Background(), AssignMacGuffinToNPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, NPCID: nid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events")
	}
}

func TestAssignMacGuffinToNPC_NPCNotActive_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	npcSnap, nid := mustNPCSnap(t, "draft", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: npcSnap}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewAssignMacGuffinToNPCInteractor(gameRepo, npcRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), AssignMacGuffinToNPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, NPCID: nid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidNPCState) {
		t.Fatalf("expected ErrInvalidNPCState, got: %v", err)
	}
}

func TestAssignMacGuffinToNPC_MacGuffinDestroyed_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, true, gid)
	_, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewAssignMacGuffinToNPCInteractor(gameRepo, &stubNPCRepo{}, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), AssignMacGuffinToNPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, NPCID: nid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("expected ErrMacGuffinDestroyed, got: %v", err)
	}
}

func TestAssignMacGuffinToNPC_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	npcSnap, nid := mustNPCSnap(t, "active", gid)
	gameRepo := &stubCharGameRepo{game: game}
	npcRepo := &stubNPCRepo{snap: npcSnap}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewAssignMacGuffinToNPCInteractor(gameRepo, npcRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), AssignMacGuffinToNPCRequest{
		CallerID: "wrong", MacGuffinID: mgid, NPCID: nid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === AssignMacGuffinToPC tests ===

func TestAssignMacGuffinToPC_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	pcSnap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: pcSnap}
	mgRepo := &stubMGRepo{snap: mgSnap}
	bus := &spyCharBus{}
	i := NewAssignMacGuffinToPCInteractor(gameRepo, pcRepo, mgRepo, bus)

	result, err := i.Execute(context.Background(), AssignMacGuffinToPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, PCID: pid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events")
	}
}

func TestAssignMacGuffinToPC_PCRetired_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	pcSnap, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusRetired, gid)
	gameRepo := &stubCharGameRepo{game: game}
	pcRepo := &stubPCRepo{snap: pcSnap}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewAssignMacGuffinToPCInteractor(gameRepo, pcRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), AssignMacGuffinToPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, PCID: pid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrCharacterRetired) {
		t.Fatalf("expected ErrCharacterRetired, got: %v", err)
	}
}

func TestAssignMacGuffinToPC_MacGuffinDestroyed_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, true, gid)
	_, pid := mustPCSnap(t, characterentity.PlayerCharacterStatusActive, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewAssignMacGuffinToPCInteractor(gameRepo, &stubPCRepo{}, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), AssignMacGuffinToPCRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, PCID: pid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrMacGuffinDestroyed) {
		t.Fatalf("expected ErrMacGuffinDestroyed, got: %v", err)
	}
}

// === DestroyMacGuffin tests ===

func TestDestroyMacGuffin_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	bus := &spyCharBus{}
	i := NewDestroyMacGuffinInteractor(gameRepo, mgRepo, bus)

	result, err := i.Execute(context.Background(), DestroyMacGuffinRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events")
	}
	if !mgRepo.saved.Destroyed {
		t.Fatal("expected MacGuffin to be marked destroyed")
	}
}

func TestDestroyMacGuffin_AlreadyDestroyed_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, true, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewDestroyMacGuffinInteractor(gameRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), DestroyMacGuffinRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, characterentity.ErrMacGuffinAlreadyDestroyed) {
		t.Fatalf("expected ErrMacGuffinAlreadyDestroyed, got: %v", err)
	}
}

func TestDestroyMacGuffin_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewDestroyMacGuffinInteractor(gameRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), DestroyMacGuffinRequest{
		CallerID: "wrong", MacGuffinID: mgid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === UpdateMacGuffinContent tests ===

func TestUpdateMacGuffinContent_Succeeds(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	bus := &spyCharBus{}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, bus)

	result, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid,
		Name: "The Orb of Doom", Description: "A dark sphere.", PlayerDesc: "It hums.",
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events to be returned")
	}
	if len(bus.envelopes) != 1 {
		t.Fatalf("expected 1 dispatched event, got %d", len(bus.envelopes))
	}
}

func TestUpdateMacGuffinContent_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "wrong", MacGuffinID: mgid, GameID: gid,
		Name: "The Orb", Description: "desc", PlayerDesc: "pdesc",
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestUpdateMacGuffinContent_MacGuffinNotFound_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{loadErr: errors.New("not found")}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "gm-uid", MacGuffinID: identity.NewMacGuffinID(), GameID: gid,
		Name: "The Orb", Description: "desc", PlayerDesc: "pdesc",
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrMacGuffinNotFound) {
		t.Fatalf("expected ErrMacGuffinNotFound, got: %v", err)
	}
}

func TestUpdateMacGuffinContent_DestroyedMacGuffin_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, true, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, &spyCharBus{})

	_, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid,
		Name: "The Orb", Description: "desc", PlayerDesc: "pdesc",
		Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for destroyed MacGuffin")
	}
}

func TestUpdateMacGuffinContent_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap, saveErr: errors.New("db down")}
	bus := &spyCharBus{}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, bus)

	_, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid,
		Name: "The Orb", Description: "desc", PlayerDesc: "pdesc",
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.envelopes) != 0 {
		t.Fatalf("expected 0 dispatched events after save failure, got %d", len(bus.envelopes))
	}
}

func TestUpdateMacGuffinContent_DispatchFailure_ReturnsError(t *testing.T) {
	game, gid := mustCharGame(t)
	mgSnap, mgid := mustMGSnap(t, false, gid)
	gameRepo := &stubCharGameRepo{game: game}
	mgRepo := &stubMGRepo{snap: mgSnap}
	bus := &spyCharBus{dispatchErr: errors.New("bus down")}
	i := NewUpdateMacGuffinContentInteractor(gameRepo, mgRepo, bus)

	_, err := i.Execute(context.Background(), UpdateMacGuffinContentRequest{
		CallerID: "gm-uid", MacGuffinID: mgid, GameID: gid,
		Name: "The Orb", Description: "desc", PlayerDesc: "pdesc",
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrEventDispatchFailed) {
		t.Fatalf("expected ErrEventDispatchFailed, got: %v", err)
	}
}
