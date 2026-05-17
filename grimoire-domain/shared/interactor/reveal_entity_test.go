package interactor

import (
	"context"
	"errors"
	"testing"

	characterentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/entity"
	factionentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/value"
)

// --- Test doubles ---

type stubRevealGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubRevealGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubRevealNPCRepo struct {
	snap    characterentity.NPCSnapshot
	saveErr error
	loadErr error
}

func (r *stubRevealNPCRepo) Save(_ context.Context, _ characterentity.NPCSnapshot) error {
	return r.saveErr
}

func (r *stubRevealNPCRepo) Load(_ context.Context, _ identity.NarrativeCharacterID) (characterentity.NPCSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.NPCSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type stubRevealPCRepo struct {
	snap    characterentity.PlayerCharacterSnapshot
	saveErr error
	loadErr error
}

func (r *stubRevealPCRepo) SaveCore(_ context.Context, _ characterentity.PlayerCharacterSnapshot) error {
	return r.saveErr
}

func (r *stubRevealPCRepo) LoadCore(_ context.Context, _ identity.PlayerCharacterID) (characterentity.PlayerCharacterSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.PlayerCharacterSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type stubRevealMGRepo struct {
	snap    characterentity.MacGuffinSnapshot
	saveErr error
	loadErr error
}

func (r *stubRevealMGRepo) Save(_ context.Context, _ characterentity.MacGuffinSnapshot) error {
	return r.saveErr
}

func (r *stubRevealMGRepo) Load(_ context.Context, _ identity.MacGuffinID) (characterentity.MacGuffinSnapshot, error) {
	if r.loadErr != nil {
		return characterentity.MacGuffinSnapshot{}, r.loadErr
	}
	return r.snap, nil
}

type stubRevealFactionRepo struct {
	faction factionentity.Faction
	saveErr error
	loadErr error
}

func (r *stubRevealFactionRepo) Save(_ context.Context, _ factionentity.Faction) error {
	return r.saveErr
}

func (r *stubRevealFactionRepo) Load(_ context.Context, _ identity.FactionID) (factionentity.Faction, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.faction, nil
}

type spyRevealBus struct {
	envelopes   []event.EventEnvelope
	dispatchErr error
}

func (b *spyRevealBus) Dispatch(_ context.Context, env event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.envelopes = append(b.envelopes, env)
	return nil
}

func (b *spyRevealBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustRevealGame(t *testing.T) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: "gm-uid", Name: "Test Game", State: "active",
	})
	if err != nil {
		t.Fatalf("mustRevealGame: %v", err)
	}
	return g, id
}

func mustRevealInteractor(t *testing.T, game gameentity.Game, npcSnap characterentity.NPCSnapshot, pcSnap characterentity.PlayerCharacterSnapshot, mgSnap characterentity.MacGuffinSnapshot, faction factionentity.Faction, bus *spyRevealBus) *RevealEntityInteractor {
	t.Helper()
	return NewRevealEntityInteractor(
		&stubRevealGameRepo{game: game},
		&stubRevealNPCRepo{snap: npcSnap},
		&stubRevealPCRepo{snap: pcSnap},
		&stubRevealMGRepo{snap: mgSnap},
		&stubRevealFactionRepo{faction: faction},
		bus,
	)
}

func mustActiveNPCSnap(t *testing.T, gameID identity.GameID) (characterentity.NPCSnapshot, identity.NarrativeCharacterID) {
	t.Helper()
	id := identity.NewNarrativeCharacterID()
	return characterentity.NPCSnapshot{
		ID: id, GameID: gameID, Name: "Lady Vex", State: "active",
	}, id
}

func mustActivePCSnap(t *testing.T, gameID identity.GameID) (characterentity.PlayerCharacterSnapshot, identity.PlayerCharacterID) {
	t.Helper()
	id := identity.NewPlayerCharacterID()
	return characterentity.PlayerCharacterSnapshot{
		ID: id, GameID: gameID, Name: "Aria",
		Status: characterentity.PlayerCharacterStatusActive,
	}, id
}

func mustLiveMGSnap(t *testing.T, gameID identity.GameID) (characterentity.MacGuffinSnapshot, identity.MacGuffinID) {
	t.Helper()
	id := identity.NewMacGuffinID()
	return characterentity.MacGuffinSnapshot{
		ID: id, GameID: gameID, Name: "The Orb", Destroyed: false,
	}, id
}

func mustActiveFaction(t *testing.T, gameID identity.GameID) (factionentity.Faction, identity.FactionID) {
	t.Helper()
	id := identity.NewFactionID()
	f, err := factionentity.ReconstituteFaction(factionentity.FactionSnapshot{
		ID: id, GameID: gameID, Name: "The Iron Circle", State: "active",
	})
	if err != nil {
		t.Fatalf("mustActiveFaction: %v", err)
	}
	return f, id
}

func validRevealSession() identity.SessionID {
	return identity.NewSessionID()
}

// === RevealEntity tests ===

func TestRevealEntity_NPC_Succeeds(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, nid := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.envelopes) == 0 {
		t.Fatal("expected EntityRevealed dispatched")
	}
}

func TestRevealEntity_PC_Succeeds(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, _ := mustActiveNPCSnap(t, gid)
	pcSnap, pid := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: pid.String(),
		EntityType: value.EntityTypePlayerCharacter, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.envelopes) == 0 {
		t.Fatal("expected EntityRevealed dispatched")
	}
}

func TestRevealEntity_MacGuffin_Succeeds(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, _ := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, mgid := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: mgid.String(),
		EntityType: value.EntityTypeMacGuffin, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.envelopes) == 0 {
		t.Fatal("expected EntityRevealed dispatched")
	}
}

func TestRevealEntity_Faction_Succeeds(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, _ := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, fid := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: fid.String(),
		EntityType: value.EntityTypeFaction, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bus.envelopes) == 0 {
		t.Fatal("expected EntityRevealed dispatched")
	}
}

func TestRevealEntity_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, nid := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "not-gm", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestRevealEntity_UnsupportedEntityType_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, _ := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: "some-id",
		EntityType: "beat", GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrUnsupportedEntityType) {
		t.Fatalf("expected ErrUnsupportedEntityType, got: %v", err)
	}
}

func TestRevealEntity_NPCNotActive_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	// Use draft NPC — not revealable.
	draftSnap, nid := mustActiveNPCSnap(t, gid)
	draftSnap.State = "draft"
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, draftSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidEntityState) {
		t.Fatalf("expected ErrInvalidEntityState, got: %v", err)
	}
}

func TestRevealEntity_FactionNotActive_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, _ := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	// Build a draft faction — not revealable.
	draftFaction, fid, _ := buildDraftFaction(t, gid)
	bus := &spyRevealBus{}
	i := NewRevealEntityInteractor(
		&stubRevealGameRepo{game: game},
		&stubRevealNPCRepo{snap: npcSnap},
		&stubRevealPCRepo{snap: pcSnap},
		&stubRevealMGRepo{snap: mgSnap},
		&stubRevealFactionRepo{faction: draftFaction},
		bus,
	)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: fid.String(),
		EntityType: value.EntityTypeFaction, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidEntityState) {
		t.Fatalf("expected ErrInvalidEntityState, got: %v", err)
	}
}

func buildDraftFaction(t *testing.T, gameID identity.GameID) (factionentity.Faction, identity.FactionID, error) {
	t.Helper()
	id := identity.NewFactionID()
	f, err := factionentity.ReconstituteFaction(factionentity.FactionSnapshot{
		ID: id, GameID: gameID, Name: "Draft Faction", State: "draft",
	})
	return f, id, err
}

func TestRevealEntity_EntityNotFound_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	nid := identity.NewNarrativeCharacterID()
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := NewRevealEntityInteractor(
		&stubRevealGameRepo{game: game},
		&stubRevealNPCRepo{loadErr: errors.New("not found")},
		&stubRevealPCRepo{snap: pcSnap},
		&stubRevealMGRepo{snap: mgSnap},
		&stubRevealFactionRepo{faction: faction},
		bus,
	)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrEntityNotFound) {
		t.Fatalf("expected ErrEntityNotFound, got: %v", err)
	}
}

func TestRevealEntity_EmptyRevealedTo_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, nid := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty RevealedTo")
	}
}

func TestRevealEntity_ZeroSessionID_ReturnsError(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, nid := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := mustRevealInteractor(t, game, npcSnap, pcSnap, mgSnap, faction, bus)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: identity.SessionID{},
		Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for zero SessionID")
	}
}

func TestRevealEntity_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustRevealGame(t)
	npcSnap, nid := mustActiveNPCSnap(t, gid)
	pcSnap, _ := mustActivePCSnap(t, gid)
	mgSnap, _ := mustLiveMGSnap(t, gid)
	faction, _ := mustActiveFaction(t, gid)
	bus := &spyRevealBus{}
	i := NewRevealEntityInteractor(
		&stubRevealGameRepo{game: game},
		&stubRevealNPCRepo{snap: npcSnap, saveErr: errors.New("db error")},
		&stubRevealPCRepo{snap: pcSnap},
		&stubRevealMGRepo{snap: mgSnap},
		&stubRevealFactionRepo{faction: faction},
		bus,
	)

	_, err := i.Execute(context.Background(), RevealEntityRequest{
		CallerID: "gm-uid", EntityID: nid.String(),
		EntityType: value.EntityTypeNPC, GameID: gid,
		RevealedTo: []string{"player-1"}, SessionID: validRevealSession(),
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRevealRepositorySaveFailed) {
		t.Fatalf("expected ErrRevealRepositorySaveFailed, got: %v", err)
	}
	if len(bus.envelopes) != 0 {
		t.Fatalf("expected no dispatched events on save failure, got %d", len(bus.envelopes))
	}
}
