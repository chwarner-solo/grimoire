package interactor

import (
	"context"
	"errors"
	"testing"

	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	narrativeentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/entity"
	narrativevalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/value"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// --- Test doubles ---

type stubNarrativeGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubNarrativeGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubMasterNarrativeRepo struct {
	mn      *narrativeentity.MasterNarrative
	saved   *narrativeentity.MasterNarrative
	saveErr error
	loadErr error
}

func (r *stubMasterNarrativeRepo) Save(_ context.Context, mn *narrativeentity.MasterNarrative) error {
	r.saved = mn
	return r.saveErr
}

func (r *stubMasterNarrativeRepo) Load(_ context.Context, _ identity.MasterNarrativeID) (*narrativeentity.MasterNarrative, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.mn, nil
}

func (r *stubMasterNarrativeRepo) FindByGame(_ context.Context, _ identity.GameID) (*narrativeentity.MasterNarrative, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.mn, nil
}

type stubCampaignNarrativeRepo struct {
	cn      *narrativeentity.CampaignNarrative
	saved   *narrativeentity.CampaignNarrative
	saveErr error
	loadErr error
}

func (r *stubCampaignNarrativeRepo) Save(_ context.Context, cn *narrativeentity.CampaignNarrative) error {
	r.saved = cn
	return r.saveErr
}

func (r *stubCampaignNarrativeRepo) Load(_ context.Context, _ identity.CampaignNarrativeID) (*narrativeentity.CampaignNarrative, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.cn, nil
}

func (r *stubCampaignNarrativeRepo) FindByCampaign(_ context.Context, _ identity.CampaignID) (*narrativeentity.CampaignNarrative, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.cn, nil
}

type stubBeatRepo struct {
	beat    *narrativeentity.Beat
	saved   *narrativeentity.Beat
	chain   []*narrativeentity.Beat
	saveErr error
	loadErr error
}

func (r *stubBeatRepo) Save(_ context.Context, beat *narrativeentity.Beat) error {
	r.saved = beat
	return r.saveErr
}

func (r *stubBeatRepo) Load(_ context.Context, _ identity.BeatID) (*narrativeentity.Beat, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.beat, nil
}

func (r *stubBeatRepo) LoadPrerequisiteChain(_ context.Context, _ identity.BeatID) ([]*narrativeentity.Beat, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.chain, nil
}

func (r *stubBeatRepo) FindByGame(_ context.Context, _ identity.GameID) ([]*narrativeentity.Beat, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return nil, nil
}

type spyNarrativeBus struct {
	dispatched  []event.EventEnvelope
	dispatchErr error
}

func (b *spyNarrativeBus) Dispatch(_ context.Context, e event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.dispatched = append(b.dispatched, e)
	return nil
}

func (b *spyNarrativeBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustNarrativeGame(t *testing.T) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: "gm-uid", Name: "Test Game", State: "draft",
	})
	if err != nil {
		t.Fatalf("mustNarrativeGame: %v", err)
	}
	return g, id
}

func mustMasterNarrative(t *testing.T, gameID identity.GameID) *narrativeentity.MasterNarrative {
	t.Helper()
	mn, _, err := narrativeentity.CreateMasterNarrative(identity.NewMasterNarrativeID(), gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustMasterNarrative: %v", err)
	}
	return mn
}

func mustCampaignNarrative(t *testing.T, campaignID identity.CampaignID, gameID identity.GameID) *narrativeentity.CampaignNarrative {
	t.Helper()
	cn, _, err := narrativeentity.CreateCampaignNarrative(identity.NewCampaignNarrativeID(), campaignID, gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCampaignNarrative: %v", err)
	}
	return cn
}

func mustMasterBeat(t *testing.T, gameID identity.GameID) *narrativeentity.Beat {
	t.Helper()
	beat, _, err := narrativeentity.CreateMasterBeat(identity.NewBeatID(), "Test Beat", narrativevalue.BeatTypeRequired, gameID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustMasterBeat: %v", err)
	}
	return beat
}

func mustCampaignBeat(t *testing.T, gameID identity.GameID, campaignID identity.CampaignID) *narrativeentity.Beat {
	t.Helper()
	beat, _, err := narrativeentity.CreateCampaignBeat(identity.NewBeatID(), "Campaign Beat", gameID, campaignID, event.SourceGrimoire)
	if err != nil {
		t.Fatalf("mustCampaignBeat: %v", err)
	}
	return beat
}

// --- CreateMasterBeat tests ---

func TestCreateMasterBeat_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{}
	bus := &spyNarrativeBus{}
	i := NewCreateMasterBeatInteractor(gameRepo, mnRepo, beatRepo, bus)

	result, err := i.Execute(context.Background(), CreateMasterBeatRequest{
		CallerID: "gm-uid", BeatID: identity.NewBeatID(), GameID: gid,
		Name: "The Dark Ritual", BeatType: narrativevalue.BeatTypeRequired, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Beat == nil {
		t.Fatal("expected non-nil beat")
	}
	if beatRepo.saved == nil {
		t.Fatal("expected beat to be saved")
	}
	if mnRepo.saved == nil {
		t.Fatal("expected master narrative to be saved")
	}
}

func TestCreateMasterBeat_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{}
	bus := &spyNarrativeBus{}
	i := NewCreateMasterBeatInteractor(gameRepo, mnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), CreateMasterBeatRequest{
		CallerID: "wrong", BeatID: identity.NewBeatID(), GameID: gid,
		Name: "Test", BeatType: narrativevalue.BeatTypeRequired, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCreateMasterBeat_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{}
	bus := &spyNarrativeBus{}
	i := NewCreateMasterBeatInteractor(gameRepo, mnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), CreateMasterBeatRequest{
		CallerID: "gm-uid", BeatID: identity.NewBeatID(), GameID: gid,
		Name: "", BeatType: narrativevalue.BeatTypeRequired, Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if beatRepo.saved != nil {
		t.Fatal("expected no save")
	}
}

func TestCreateMasterBeat_BeatSaveFailure_NoMNSave(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{saveErr: errors.New("db error")}
	bus := &spyNarrativeBus{}
	i := NewCreateMasterBeatInteractor(gameRepo, mnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), CreateMasterBeatRequest{
		CallerID: "gm-uid", BeatID: identity.NewBeatID(), GameID: gid,
		Name: "Test", BeatType: narrativevalue.BeatTypeRequired, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if mnRepo.saved != nil {
		t.Fatal("expected no MN save when beat save fails")
	}
}

// --- UpdateBeatContent tests ---

func TestUpdateBeatContent_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewUpdateBeatContentInteractor(gameRepo, beatRepo, bus)

	result, err := i.Execute(context.Background(), UpdateBeatContentRequest{
		CallerID: "gm-uid", BeatID: beat.BeatID(), GameID: gid,
		Name: "Updated", Description: "desc", PlayerDesc: "player desc", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Beat == nil {
		t.Fatal("expected non-nil beat")
	}
}

func TestUpdateBeatContent_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewUpdateBeatContentInteractor(gameRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), UpdateBeatContentRequest{
		CallerID: "wrong", BeatID: beat.BeatID(), GameID: gid,
		Name: "Updated", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestUpdateBeatContent_BeatNotFound_ReturnsError(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	gameRepo := &stubNarrativeGameRepo{game: game}
	beatRepo := &stubBeatRepo{loadErr: errors.New("not found")}
	bus := &spyNarrativeBus{}
	i := NewUpdateBeatContentInteractor(gameRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), UpdateBeatContentRequest{
		CallerID: "gm-uid", BeatID: identity.NewBeatID(), GameID: gid,
		Name: "Updated", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrBeatNotFound) {
		t.Fatalf("expected ErrBeatNotFound, got: %v", err)
	}
}

// --- AddBeatPrerequisite tests ---

func TestAddBeatPrerequisite_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	prereq := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	beatRepo := &stubBeatRepo{beat: beat, chain: []*narrativeentity.Beat{prereq}}
	bus := &spyNarrativeBus{}
	i := NewAddBeatPrerequisiteInteractor(gameRepo, beatRepo, bus)

	result, err := i.Execute(context.Background(), AddBeatPrerequisiteRequest{
		CallerID: "gm-uid", BeatID: beat.BeatID(), PrerequisiteID: prereq.BeatID(),
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
}

func TestAddBeatPrerequisite_WouldCreateCycle_ReturnsError(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	// cycle: chain contains the beat itself
	beatRepo := &stubBeatRepo{beat: beat, chain: []*narrativeentity.Beat{beat}}
	bus := &spyNarrativeBus{}
	i := NewAddBeatPrerequisiteInteractor(gameRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), AddBeatPrerequisiteRequest{
		CallerID: "gm-uid", BeatID: beat.BeatID(), PrerequisiteID: beat.BeatID(),
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestAddBeatPrerequisite_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewAddBeatPrerequisiteInteractor(gameRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), AddBeatPrerequisiteRequest{
		CallerID: "wrong", BeatID: beat.BeatID(), PrerequisiteID: identity.NewBeatID(),
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- AddAct tests ---

func TestAddActToMasterNarrative_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	bus := &spyNarrativeBus{}
	i := NewAddActToMasterNarrativeInteractor(gameRepo, mnRepo, bus)

	result, err := i.Execute(context.Background(), AddActToMasterNarrativeRequest{
		CallerID: "gm-uid", ActID: identity.NewActID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
	if mnRepo.saved == nil {
		t.Fatal("expected MN to be saved")
	}
}

func TestAddActToMasterNarrative_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	bus := &spyNarrativeBus{}
	i := NewAddActToMasterNarrativeInteractor(gameRepo, mnRepo, bus)

	_, err := i.Execute(context.Background(), AddActToMasterNarrativeRequest{
		CallerID: "wrong", ActID: identity.NewActID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- AddSecret tests ---

func TestAddSecretToMasterNarrative_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	bus := &spyNarrativeBus{}
	i := NewAddSecretToMasterNarrativeInteractor(gameRepo, mnRepo, bus)

	result, err := i.Execute(context.Background(), AddSecretToMasterNarrativeRequest{
		CallerID: "gm-uid", SecretID: identity.NewSecretID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
}

// --- AddLore tests ---

func TestAddLoreToMasterNarrative_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	bus := &spyNarrativeBus{}
	i := NewAddLoreToMasterNarrativeInteractor(gameRepo, mnRepo, bus)

	result, err := i.Execute(context.Background(), AddLoreToMasterNarrativeRequest{
		CallerID: "gm-uid", LoreID: identity.NewLoreID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
}

// --- DiscoverBeat tests ---

func TestDiscoverBeat_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	cn := mustCampaignNarrative(t, cid, gid)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	cnRepo := &stubCampaignNarrativeRepo{cn: cn}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewDiscoverBeatInteractor(gameRepo, cnRepo, beatRepo, bus)

	cnID := identity.NewCampaignNarrativeID()
	result, err := i.Execute(context.Background(), DiscoverBeatRequest{
		CallerID: "gm-uid", CampaignNarrativeID: cnID, BeatID: beat.BeatID(),
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events from DiscoverBeat")
	}
}

func TestDiscoverBeat_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	cn := mustCampaignNarrative(t, cid, gid)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	cnRepo := &stubCampaignNarrativeRepo{cn: cn}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewDiscoverBeatInteractor(gameRepo, cnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), DiscoverBeatRequest{
		CallerID: "wrong", CampaignNarrativeID: identity.NewCampaignNarrativeID(),
		BeatID: beat.BeatID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestDiscoverBeat_BeatNotFound_ReturnsError(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	cn := mustCampaignNarrative(t, cid, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	cnRepo := &stubCampaignNarrativeRepo{cn: cn}
	beatRepo := &stubBeatRepo{loadErr: errors.New("not found")}
	bus := &spyNarrativeBus{}
	i := NewDiscoverBeatInteractor(gameRepo, cnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), DiscoverBeatRequest{
		CallerID: "gm-uid", CampaignNarrativeID: identity.NewCampaignNarrativeID(),
		BeatID: identity.NewBeatID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrBeatNotFound) {
		t.Fatalf("expected ErrBeatNotFound, got: %v", err)
	}
}

// --- CreateCampaignBeat tests ---

func TestCreateCampaignBeat_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	cn := mustCampaignNarrative(t, cid, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	cnRepo := &stubCampaignNarrativeRepo{cn: cn}
	beatRepo := &stubBeatRepo{}
	bus := &spyNarrativeBus{}
	i := NewCreateCampaignBeatInteractor(gameRepo, cnRepo, beatRepo, bus)

	result, err := i.Execute(context.Background(), CreateCampaignBeatRequest{
		CallerID: "gm-uid", BeatID: identity.NewBeatID(), CampaignID: cid,
		GameID: gid, Name: "Ambush at the Gate", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Beat == nil {
		t.Fatal("expected non-nil beat")
	}
	if beatRepo.saved == nil {
		t.Fatal("expected beat to be saved")
	}
	if cnRepo.saved == nil {
		t.Fatal("expected CN to be saved")
	}
}

func TestCreateCampaignBeat_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	cn := mustCampaignNarrative(t, cid, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	cnRepo := &stubCampaignNarrativeRepo{cn: cn}
	beatRepo := &stubBeatRepo{}
	bus := &spyNarrativeBus{}
	i := NewCreateCampaignBeatInteractor(gameRepo, cnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), CreateCampaignBeatRequest{
		CallerID: "wrong", BeatID: identity.NewBeatID(), CampaignID: cid,
		GameID: gid, Name: "Test", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// --- PromoteBeatToMaster tests ---

func TestPromoteBeatToMaster_Succeeds(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	cid := identity.NewCampaignID()
	mn := mustMasterNarrative(t, gid)
	beat := mustCampaignBeat(t, gid, cid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewPromoteBeatToMasterInteractor(gameRepo, mnRepo, beatRepo, bus)

	result, err := i.Execute(context.Background(), PromoteBeatToMasterRequest{
		CallerID: "gm-uid", BeatID: beat.BeatID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result.Events))
	}
	if mnRepo.saved == nil {
		t.Fatal("expected MN to be saved")
	}
}

func TestPromoteBeatToMaster_AlreadyMaster_ReturnsError(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	mn := mustMasterNarrative(t, gid)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{mn: mn}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewPromoteBeatToMasterInteractor(gameRepo, mnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), PromoteBeatToMasterRequest{
		CallerID: "gm-uid", BeatID: beat.BeatID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, narrativeentity.ErrBeatAlreadyMasterScope) {
		t.Fatalf("expected ErrBeatAlreadyMasterScope, got: %v", err)
	}
}

func TestPromoteBeatToMaster_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustNarrativeGame(t)
	beat := mustMasterBeat(t, gid)
	gameRepo := &stubNarrativeGameRepo{game: game}
	mnRepo := &stubMasterNarrativeRepo{}
	beatRepo := &stubBeatRepo{beat: beat}
	bus := &spyNarrativeBus{}
	i := NewPromoteBeatToMasterInteractor(gameRepo, mnRepo, beatRepo, bus)

	_, err := i.Execute(context.Background(), PromoteBeatToMasterRequest{
		CallerID: "wrong", BeatID: beat.BeatID(), GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}
