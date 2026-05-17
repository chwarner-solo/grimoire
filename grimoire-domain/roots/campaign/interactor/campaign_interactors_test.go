package interactor

import (
	"context"
	"errors"
	"testing"
	"time"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// --- Test doubles ---

type stubGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubCampaignRepo struct {
	saved   campaignentity.Campaign
	loadVal campaignentity.Campaign
	saveErr error
	loadErr error
}

func (r *stubCampaignRepo) Save(_ context.Context, c campaignentity.Campaign) error {
	r.saved = c
	return r.saveErr
}

func (r *stubCampaignRepo) Load(_ context.Context, _ identity.CampaignID) (campaignentity.Campaign, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.loadVal, nil
}

type spyCampaignBus struct {
	dispatched  []event.EventEnvelope
	dispatchErr error
}

func (b *spyCampaignBus) Dispatch(_ context.Context, e event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.dispatched = append(b.dispatched, e)
	return nil
}

func (b *spyCampaignBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustGame(t *testing.T, gmID string) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: gmID, Name: "Test Game", State: "draft",
	})
	if err != nil {
		t.Fatalf("mustGame: %v", err)
	}
	return g, id
}

func mustCampaign(t *testing.T, state string) (campaignentity.Campaign, identity.CampaignID, identity.GameID) {
	t.Helper()
	cid := identity.NewCampaignID()
	gid := identity.NewGameID()
	charID := identity.NewPlayerCharacterID()
	snap := campaignentity.CampaignSnapshot{
		ID: cid, Name: "Test Campaign", GameID: gid, State: state,
		CharacterIDs: []identity.PlayerCharacterID{charID},
	}
	c, err := campaignentity.ReconstituteCampaign(snap)
	if err != nil {
		t.Fatalf("mustCampaign(%s): %v", state, err)
	}
	return c, cid, gid
}

// --- CreateCampaign tests ---

func TestCreateCampaign_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyCampaignBus{}
	i := NewCreateCampaignInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), CreateCampaignRequest{
		CallerID:   "gm-uid",
		CampaignID: identity.NewCampaignID(),
		GameID:     gid,
		Name:       "Shattered Realms",
		Source:     event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Campaign == nil {
		t.Fatal("expected non-nil campaign")
	}
	if campaignRepo.saved == nil {
		t.Fatal("expected campaign to be saved")
	}
	if len(bus.dispatched) != 2 {
		t.Fatalf("expected 2 dispatched events (EntityCreated + EntityLinked), got %d", len(bus.dispatched))
	}
}

func TestCreateCampaign_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyCampaignBus{}
	i := NewCreateCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), CreateCampaignRequest{
		CallerID: "wrong", GameID: gid, CampaignID: identity.NewCampaignID(), Name: "Test", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
	if campaignRepo.saved != nil {
		t.Fatal("expected no save")
	}
}

func TestCreateCampaign_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyCampaignBus{}
	i := NewCreateCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), CreateCampaignRequest{
		CallerID: "gm-uid", GameID: gid, CampaignID: identity.NewCampaignID(), Name: "", Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if campaignRepo.saved != nil {
		t.Fatal("expected no save")
	}
}

func TestCreateCampaign_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{saveErr: errors.New("db error")}
	bus := &spyCampaignBus{}
	i := NewCreateCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), CreateCampaignRequest{
		CallerID: "gm-uid", GameID: gid, CampaignID: identity.NewCampaignID(), Name: "Test", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(bus.dispatched))
	}
}

func TestCreateCampaign_DispatchesEntityCreatedAndLinked(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyCampaignBus{}
	i := NewCreateCampaignInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), CreateCampaignRequest{ //nolint:errcheck
		CallerID: "gm-uid", GameID: gid, CampaignID: identity.NewCampaignID(), Name: "Test", Source: event.SourceGrimoire,
	})

	types := make([]event.EventType, 0, len(bus.dispatched))
	for _, e := range bus.dispatched {
		types = append(types, e.Type)
	}
	hasCreated := false
	hasLinked := false
	for _, et := range types {
		if et == event.TypeEntityCreated {
			hasCreated = true
		}
		if et == event.TypeEntityLinked {
			hasLinked = true
		}
	}
	if !hasCreated || !hasLinked {
		t.Fatalf("expected EntityCreated and EntityLinked, got: %v", types)
	}
}

// --- AddCharacterToCampaign tests ---

func TestAddCharacter_ToNewCampaign_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewAddCharacterToCampaignInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), AddCharacterToCampaignRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		CharacterID: identity.NewPlayerCharacterID(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Campaign == nil {
		t.Fatal("expected non-nil campaign")
	}
	if len(bus.dispatched) != 1 {
		t.Fatalf("expected 1 event, got %d", len(bus.dispatched))
	}
}

func TestAddCharacter_ToFormingCampaign_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewAddCharacterToCampaignInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), AddCharacterToCampaignRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		CharacterID: identity.NewPlayerCharacterID(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Campaign == nil {
		t.Fatal("expected non-nil campaign")
	}
}

func TestAddCharacter_ToActiveCampaign_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewAddCharacterToCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), AddCharacterToCampaignRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		CharacterID: identity.NewPlayerCharacterID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestAddCharacter_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewAddCharacterToCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), AddCharacterToCampaignRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid,
		CharacterID: identity.NewPlayerCharacterID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- BeginCampaignFormation tests ---

func TestBeginFormation_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewBeginCampaignFormationInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), BeginCampaignFormationRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Campaign.(campaignentity.FormingCampaign); !ok {
		t.Fatalf("expected FormingCampaign, got %T", result.Campaign)
	}
}

func TestBeginFormation_AlreadyForming_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewBeginCampaignFormationInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), BeginCampaignFormationRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestBeginFormation_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewBeginCampaignFormationInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), BeginCampaignFormationRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestBeginFormation_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign, saveErr: errors.New("db error")}
	bus := &spyCampaignBus{}
	i := NewBeginCampaignFormationInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), BeginCampaignFormationRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(bus.dispatched))
	}
}

// --- StartFirstSession tests ---

func TestStartFirstSession_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartFirstSessionInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), StartFirstSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Campaign.(campaignentity.ActiveCampaign); !ok {
		t.Fatalf("expected ActiveCampaign, got %T", result.Campaign)
	}
	if len(bus.dispatched) != 1 {
		t.Fatalf("expected 1 SessionStarted event, got %d", len(bus.dispatched))
	}
}

func TestStartFirstSession_NotForming_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "new")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartFirstSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), StartFirstSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestStartFirstSession_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartFirstSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), StartFirstSessionRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestStartFirstSession_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign, saveErr: errors.New("db error")}
	bus := &spyCampaignBus{}
	i := NewStartFirstSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), StartFirstSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.dispatched) != 0 {
		t.Fatalf("expected no dispatch, got %d", len(bus.dispatched))
	}
}

func TestStartFirstSession_DispatchesSessionStartedEvent(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "forming")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartFirstSessionInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), StartFirstSessionRequest{ //nolint:errcheck
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})

	if len(bus.dispatched) == 0 || bus.dispatched[0].Type != event.TypeSessionStarted {
		t.Fatalf("expected TypeSessionStarted, got: %v", bus.dispatched)
	}
}

// --- StartNewSession tests ---

func TestStartNewSession_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartNewSessionInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), StartNewSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Campaign.(campaignentity.ActiveCampaign); !ok {
		t.Fatalf("expected ActiveCampaign, got %T", result.Campaign)
	}
}

func TestStartNewSession_NotIdle_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartNewSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), StartNewSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestStartNewSession_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartNewSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), StartNewSessionRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestStartNewSession_DispatchesSessionStartedEvent(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewStartNewSessionInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), StartNewSessionRequest{ //nolint:errcheck
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Date: time.Now(), Source: event.SourceGrimoire,
	})

	if len(bus.dispatched) == 0 || bus.dispatched[0].Type != event.TypeSessionStarted {
		t.Fatalf("expected TypeSessionStarted, got: %v", bus.dispatched)
	}
}

// --- EndSession tests ---

func TestEndSession_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewEndSessionInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), EndSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Notes: "great session", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Campaign.(campaignentity.IdleCampaign); !ok {
		t.Fatalf("expected IdleCampaign, got %T", result.Campaign)
	}
}

func TestEndSession_NotActive_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewEndSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), EndSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestEndSession_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewEndSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), EndSessionRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestEndSession_DispatchesSessionEndedEvent(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewEndSessionInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), EndSessionRequest{ //nolint:errcheck
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Notes: "recap here", Source: event.SourceGrimoire,
	})

	if len(bus.dispatched) == 0 || bus.dispatched[0].Type != event.TypeSessionEnded {
		t.Fatalf("expected TypeSessionEnded, got: %v", bus.dispatched)
	}
}

func TestEndSession_NotesAreOptional(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewEndSessionInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), EndSessionRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		SessionID: identity.NewSessionID(), Notes: "", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- MoveParty tests ---

func TestMoveParty_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewMovePartyInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), MovePartyRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		LocationID: identity.NewLocationID(), Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Campaign == nil {
		t.Fatal("expected non-nil campaign")
	}
}

func TestMoveParty_NotActive_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewMovePartyInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), MovePartyRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		LocationID: identity.NewLocationID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestMoveParty_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewMovePartyInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), MovePartyRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid,
		LocationID: identity.NewLocationID(), Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestMoveParty_DispatchesTwoEvents(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewMovePartyInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), MovePartyRequest{ //nolint:errcheck
		CallerID: "gm-uid", CampaignID: cid, GameID: gid,
		LocationID: identity.NewLocationID(), Source: event.SourceGrimoire,
	})

	if len(bus.dispatched) != 2 {
		t.Fatalf("expected 2 dispatched events (EntityUpdated + EntityLinked), got %d", len(bus.dispatched))
	}
}

// --- CompleteCampaign tests ---

func TestCompleteCampaign_Succeeds(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewCompleteCampaignInteractor(gameRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), CompleteCampaignRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.Campaign.(campaignentity.CompleteCampaign); !ok {
		t.Fatalf("expected CompleteCampaign, got %T", result.Campaign)
	}
}

func TestCompleteCampaign_ActiveCampaign_ReturnsError(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "active")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewCompleteCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), CompleteCampaignRequest{
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidCampaignState) {
		t.Fatalf("expected ErrInvalidCampaignState, got: %v", err)
	}
}

func TestCompleteCampaign_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewCompleteCampaignInteractor(gameRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), CompleteCampaignRequest{
		CallerID: "wrong", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCompleteCampaign_DispatchesEntityUpdatedEvent(t *testing.T) {
	game, gid := mustGame(t, "gm-uid")
	campaign, cid, _ := mustCampaign(t, "idle")
	gameRepo := &stubGameRepo{game: game}
	campaignRepo := &stubCampaignRepo{loadVal: campaign}
	bus := &spyCampaignBus{}
	i := NewCompleteCampaignInteractor(gameRepo, campaignRepo, bus)

	i.Execute(context.Background(), CompleteCampaignRequest{ //nolint:errcheck
		CallerID: "gm-uid", CampaignID: cid, GameID: gid, Source: event.SourceGrimoire,
	})

	if len(bus.dispatched) == 0 || bus.dispatched[0].Type != event.TypeEntityUpdated {
		t.Fatalf("expected TypeEntityUpdated, got: %v", bus.dispatched)
	}
}
