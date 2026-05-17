package interactor

import (
	"context"
	"errors"
	"testing"

	campaignentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/entity"
	locationentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/entity"
	locationvalue "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/value"
	gameentity "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/entity"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/event"
	"github.com/chwarner-solo/grimoire/grimoire-domain/shared/identity"
	shared "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// --- Test doubles ---

type stubLocationGameRepo struct {
	game    gameentity.Game
	loadErr error
}

func (r *stubLocationGameRepo) Load(_ context.Context, _ identity.GameID) (gameentity.Game, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.game, nil
}

type stubLocationRepo struct {
	location locationentity.Location
	children []locationentity.Location
	saved    locationentity.Location
	saveErr  error
	loadErr  error
}

func (r *stubLocationRepo) Save(_ context.Context, l locationentity.Location) error {
	r.saved = l
	return r.saveErr
}

func (r *stubLocationRepo) Load(_ context.Context, _ identity.LocationID) (locationentity.Location, error) {
	if r.loadErr != nil {
		return nil, r.loadErr
	}
	return r.location, nil
}

func (r *stubLocationRepo) FindChildren(_ context.Context, _ identity.LocationID) ([]locationentity.Location, error) {
	return r.children, nil
}

func (r *stubLocationRepo) FindByGame(_ context.Context, _ identity.GameID) ([]locationentity.Location, error) {
	return nil, nil
}

type stubMultiLocationRepo struct {
	locations map[string]locationentity.Location
	children  []locationentity.Location
	saved     locationentity.Location
	saveErr   error
}

func (r *stubMultiLocationRepo) Save(_ context.Context, l locationentity.Location) error {
	r.saved = l
	return r.saveErr
}

func (r *stubMultiLocationRepo) Load(_ context.Context, id identity.LocationID) (locationentity.Location, error) {
	l, ok := r.locations[id.String()]
	if !ok {
		return nil, errors.New("not found")
	}
	return l, nil
}

func (r *stubMultiLocationRepo) FindChildren(_ context.Context, _ identity.LocationID) ([]locationentity.Location, error) {
	return r.children, nil
}

func (r *stubMultiLocationRepo) FindByGame(_ context.Context, _ identity.GameID) ([]locationentity.Location, error) {
	return nil, nil
}

type stubCampaignRepo struct {
	campaigns []campaignentity.Campaign
}

func (r *stubCampaignRepo) FindByGame(_ context.Context, _ identity.GameID) ([]campaignentity.Campaign, error) {
	return r.campaigns, nil
}

type spyLocationBus struct {
	envelopes   []event.EventEnvelope
	dispatchErr error
}

func (b *spyLocationBus) Dispatch(_ context.Context, env event.EventEnvelope) error {
	if b.dispatchErr != nil {
		return b.dispatchErr
	}
	b.envelopes = append(b.envelopes, env)
	return nil
}

func (b *spyLocationBus) Subscribe(_ event.EventType, _ event.EventHandler) error { return nil }

// --- Helpers ---

func mustLocationGame(t *testing.T) (gameentity.Game, identity.GameID) {
	t.Helper()
	id := identity.NewGameID()
	g, err := gameentity.ReconstituteGame(gameentity.GameSnapshot{
		ID: id, GMID: "gm-uid", Name: "Test Game", State: "active",
	})
	if err != nil {
		t.Fatalf("mustLocationGame: %v", err)
	}
	return g, id
}

func mustLocation(t *testing.T, state string, gameID identity.GameID) (locationentity.Location, identity.LocationID) {
	t.Helper()
	id := identity.NewLocationID()
	l, err := locationentity.ReconstituteLocation(locationentity.LocationSnapshot{
		ID: id, GameID: gameID, Name: "Test Location",
		LocationType: locationvalue.LocationTypeSettlement, State: state,
	})
	if err != nil {
		t.Fatalf("mustLocation(%s): %v", state, err)
	}
	return l, id
}

func mustActiveCampaign(t *testing.T, gameID identity.GameID, locationID identity.LocationID) campaignentity.Campaign {
	t.Helper()
	c, err := campaignentity.ReconstituteCampaign(campaignentity.CampaignSnapshot{
		ID: identity.NewCampaignID(), GameID: gameID, Name: "Test Campaign",
		State: "active", CurrentLocationID: locationID,
	})
	if err != nil {
		t.Fatalf("mustActiveCampaign: %v", err)
	}
	return c
}

func mustIdleCampaign(t *testing.T, gameID identity.GameID, locationID identity.LocationID) campaignentity.Campaign {
	t.Helper()
	c, err := campaignentity.ReconstituteCampaign(campaignentity.CampaignSnapshot{
		ID: identity.NewCampaignID(), GameID: gameID, Name: "Test Campaign",
		State: "idle", CurrentLocationID: locationID,
	})
	if err != nil {
		t.Fatalf("mustIdleCampaign: %v", err)
	}
	return c
}

// === CreateLocation tests ===

func TestCreateLocation_TopLevel_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{}
	bus := &spyLocationBus{}
	i := NewCreateLocationInteractor(gameRepo, locationRepo, bus)

	result, err := i.Execute(context.Background(), CreateLocationRequest{
		CallerID: "gm-uid", ID: identity.NewLocationID(), GameID: gid,
		Name: "The City of Ardenmoor", LocationType: locationvalue.LocationTypeSettlement,
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected non-nil location")
	}
	if locationRepo.saved == nil {
		t.Fatal("expected location to be saved")
	}
	if len(bus.envelopes) != 1 {
		t.Fatalf("expected 1 event dispatched, got %d", len(bus.envelopes))
	}
}

func TestCreateLocation_Child_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{}
	bus := &spyLocationBus{}
	i := NewCreateLocationInteractor(gameRepo, locationRepo, bus)

	parentID := identity.NewLocationID()
	result, err := i.Execute(context.Background(), CreateLocationRequest{
		CallerID: "gm-uid", ID: identity.NewLocationID(), GameID: gid,
		Name: "The Inn", LocationType: locationvalue.LocationTypeBuilding,
		ParentID: parentID, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected non-nil location")
	}
	if len(bus.envelopes) != 2 {
		t.Fatalf("expected 2 events (Created + Linked), got %d", len(bus.envelopes))
	}
	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events in result, got %d", len(result.Events))
	}
}

func TestCreateLocation_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustLocationGame(t)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{}
	bus := &spyLocationBus{}
	i := NewCreateLocationInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), CreateLocationRequest{
		CallerID: "not-the-gm", ID: identity.NewLocationID(), GameID: gid,
		Name: "Some Place", LocationType: locationvalue.LocationTypeSettlement,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestCreateLocation_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{}
	bus := &spyLocationBus{}
	i := NewCreateLocationInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), CreateLocationRequest{
		CallerID: "gm-uid", ID: identity.NewLocationID(), GameID: gid,
		Name: "", LocationType: locationvalue.LocationTypeSettlement,
		Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateLocation_SaveFailure_NoDispatch(t *testing.T) {
	game, gid := mustLocationGame(t)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{saveErr: errors.New("db error")}
	bus := &spyLocationBus{}
	i := NewCreateLocationInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), CreateLocationRequest{
		CallerID: "gm-uid", ID: identity.NewLocationID(), GameID: gid,
		Name: "Some Place", LocationType: locationvalue.LocationTypeSettlement,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrRepositorySaveFailed) {
		t.Fatalf("expected ErrRepositorySaveFailed, got: %v", err)
	}
	if len(bus.envelopes) != 0 {
		t.Fatalf("expected no events dispatched on save failure, got %d", len(bus.envelopes))
	}
}

// === ActivateLocation tests ===

func TestActivateLocation_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewActivateLocationInteractor(gameRepo, locationRepo, bus)

	result, err := i.Execute(context.Background(), ActivateLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected non-nil ActiveLocation")
	}
	if locationRepo.saved == nil {
		t.Fatal("expected location to be saved")
	}
}

func TestActivateLocation_NotDraft_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "new", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewActivateLocationInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ActivateLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidLocationState) {
		t.Fatalf("expected ErrInvalidLocationState, got: %v", err)
	}
}

func TestActivateLocation_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewActivateLocationInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ActivateLocationRequest{
		CallerID: "wrong", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === AddScene tests ===

func TestAddScene_ToDraft_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewAddSceneInteractor(gameRepo, locationRepo, bus)

	result, err := i.Execute(context.Background(), AddSceneRequest{
		CallerID: "gm-uid", LocationID: lid, SceneID: identity.NewSceneID(),
		GameID: gid, Name: "The Main Hall", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events in result")
	}
	if locationRepo.saved == nil {
		t.Fatal("expected location to be saved")
	}
}

func TestAddScene_ToActive_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewAddSceneInteractor(gameRepo, locationRepo, bus)

	result, err := i.Execute(context.Background(), AddSceneRequest{
		CallerID: "gm-uid", LocationID: lid, SceneID: identity.NewSceneID(),
		GameID: gid, Name: "The Alley", Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events in result")
	}
}

func TestAddScene_ToArchived_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "archived", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewAddSceneInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), AddSceneRequest{
		CallerID: "gm-uid", LocationID: lid, SceneID: identity.NewSceneID(),
		GameID: gid, Name: "Some Scene", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidLocationState) {
		t.Fatalf("expected ErrInvalidLocationState, got: %v", err)
	}
}

func TestAddScene_EmptyName_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewAddSceneInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), AddSceneRequest{
		CallerID: "gm-uid", LocationID: lid, SceneID: identity.NewSceneID(),
		GameID: gid, Name: "", Source: event.SourceGrimoire,
	})
	if err == nil {
		t.Fatal("expected error for empty scene name")
	}
}

func TestAddScene_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	bus := &spyLocationBus{}
	i := NewAddSceneInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), AddSceneRequest{
		CallerID: "wrong", LocationID: lid, SceneID: identity.NewSceneID(),
		GameID: gid, Name: "Some Scene", Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

// === ConnectLocations tests ===

func TestConnectLocations_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	from, fromID := mustLocation(t, "active", gid)
	to, toID := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubMultiLocationRepo{
		locations: map[string]locationentity.Location{
			fromID.String(): from,
			toID.String():   to,
		},
	}
	bus := &spyLocationBus{}
	i := NewConnectLocationsInteractor(gameRepo, locationRepo, bus)

	result, err := i.Execute(context.Background(), ConnectLocationsRequest{
		CallerID: "gm-uid", FromID: fromID, ToID: toID,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Events) == 0 {
		t.Fatal("expected events in result")
	}
	if len(bus.envelopes) == 0 {
		t.Fatal("expected event dispatched")
	}
}

func TestConnectLocations_SelfConnection_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	loc, lid := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubMultiLocationRepo{
		locations: map[string]locationentity.Location{lid.String(): loc},
	}
	bus := &spyLocationBus{}
	i := NewConnectLocationsInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ConnectLocationsRequest{
		CallerID: "gm-uid", FromID: lid, ToID: lid,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrSelfConnection) {
		t.Fatalf("expected ErrSelfConnection, got: %v", err)
	}
}

func TestConnectLocations_AlreadyConnected_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	toID := identity.NewLocationID()
	// Build from-location already connected to toID.
	from, fromID := mustLocation(t, "active", gid)
	al := from.(locationentity.ActiveLocation)
	from, _, _ = al.ConnectTo(toID, event.SourceGrimoire)

	to, _ := mustLocation(t, "active", gid)
	// Use stub that returns the already-connected from for fromID, and to for toID.
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubMultiLocationRepo{
		locations: map[string]locationentity.Location{
			fromID.String(): from,
			toID.String():   to,
		},
	}
	bus := &spyLocationBus{}
	i := NewConnectLocationsInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ConnectLocationsRequest{
		CallerID: "gm-uid", FromID: fromID, ToID: toID,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrConnectionAlreadyExists) {
		t.Fatalf("expected ErrConnectionAlreadyExists, got: %v", err)
	}
}

func TestConnectLocations_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustLocationGame(t)
	from, fromID := mustLocation(t, "active", gid)
	to, toID := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubMultiLocationRepo{
		locations: map[string]locationentity.Location{
			fromID.String(): from,
			toID.String():   to,
		},
	}
	bus := &spyLocationBus{}
	i := NewConnectLocationsInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ConnectLocationsRequest{
		CallerID: "wrong", FromID: fromID, ToID: toID,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}

func TestConnectLocations_ArchivedFrom_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	from, fromID := mustLocation(t, "archived", gid)
	to, toID := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubMultiLocationRepo{
		locations: map[string]locationentity.Location{
			fromID.String(): from,
			toID.String():   to,
		},
	}
	bus := &spyLocationBus{}
	i := NewConnectLocationsInteractor(gameRepo, locationRepo, bus)

	_, err := i.Execute(context.Background(), ConnectLocationsRequest{
		CallerID: "gm-uid", FromID: fromID, ToID: toID,
		GameID: gid, Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidLocationState) {
		t.Fatalf("expected ErrInvalidLocationState, got: %v", err)
	}
}

// === ArchiveLocation tests ===

func TestArchiveLocation_Active_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected ArchivedLocation")
	}
}

func TestArchiveLocation_Idle_Succeeds(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "idle", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected ArchivedLocation")
	}
}

func TestArchiveLocation_ActiveChildPresent_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	child, _ := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{
		location: location,
		children: []locationentity.Location{child},
	}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrActiveChildrenExist) {
		t.Fatalf("expected ErrActiveChildrenExist, got: %v", err)
	}
}

func TestArchiveLocation_PartyPresent_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	activeCampaign := mustActiveCampaign(t, gid, lid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{campaigns: []campaignentity.Campaign{activeCampaign}}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrPartyPresent) {
		t.Fatalf("expected ErrPartyPresent, got: %v", err)
	}
}

func TestArchiveLocation_IdleCampaignPresent_Succeeds(t *testing.T) {
	// Idle campaign at this location should NOT block archival.
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	idleCampaign := mustIdleCampaign(t, gid, lid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{campaigns: []campaignentity.Campaign{idleCampaign}}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	result, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Location == nil {
		t.Fatal("expected ArchivedLocation")
	}
}

func TestArchiveLocation_Draft_ReturnsError(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "draft", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "gm-uid", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, ErrInvalidLocationState) {
		t.Fatalf("expected ErrInvalidLocationState, got: %v", err)
	}
}

func TestArchiveLocation_WrongCaller_ReturnsUnauthorized(t *testing.T) {
	game, gid := mustLocationGame(t)
	location, lid := mustLocation(t, "active", gid)
	gameRepo := &stubLocationGameRepo{game: game}
	locationRepo := &stubLocationRepo{location: location}
	campaignRepo := &stubCampaignRepo{}
	bus := &spyLocationBus{}
	i := NewArchiveLocationInteractor(gameRepo, locationRepo, campaignRepo, bus)

	_, err := i.Execute(context.Background(), ArchiveLocationRequest{
		CallerID: "wrong", LocationID: lid, GameID: gid,
		Source: event.SourceGrimoire,
	})
	if !errors.Is(err, shared.ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got: %v", err)
	}
}
