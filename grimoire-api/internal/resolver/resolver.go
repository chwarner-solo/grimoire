package resolver

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	campaigninteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/campaign/interactor"
	characterinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/character/interactor"
	factioninteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/faction/interactor"
	gameinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/game/interactor"
	locationinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/location/interactor"
	narrativeinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/roots/narrative/interactor"
	sharedinteractor "github.com/chwarner-solo/grimoire/grimoire-domain/shared/interactor"
)

// Resolver is the composition root for all GraphQL mutation resolvers.
// All 47 GM interactors are wired here at startup.
type Resolver struct {
	auth sharedinteractor.CallerIdentityPort

	// Game (2)
	createGame  *gameinteractor.CreateGameInteractor
	archiveGame *gameinteractor.ArchiveGameInteractor

	// Campaign (8)
	createCampaign         *campaigninteractor.CreateCampaignInteractor
	addCharacterToCampaign *campaigninteractor.AddCharacterToCampaignInteractor
	beginCampaignFormation *campaigninteractor.BeginCampaignFormationInteractor
	startFirstSession      *campaigninteractor.StartFirstSessionInteractor
	startNewSession        *campaigninteractor.StartNewSessionInteractor
	endSession             *campaigninteractor.EndSessionInteractor
	moveParty              *campaigninteractor.MovePartyInteractor
	completeCampaign       *campaigninteractor.CompleteCampaignInteractor

	// Narrative (9)
	createMasterBeat          *narrativeinteractor.CreateMasterBeatInteractor
	updateBeatContent         *narrativeinteractor.UpdateBeatContentInteractor
	addBeatPrerequisite       *narrativeinteractor.AddBeatPrerequisiteInteractor
	addActToMasterNarrative   *narrativeinteractor.AddActToMasterNarrativeInteractor
	addSecretToMasterNarrative *narrativeinteractor.AddSecretToMasterNarrativeInteractor
	addLoreToMasterNarrative   *narrativeinteractor.AddLoreToMasterNarrativeInteractor
	discoverBeat              *narrativeinteractor.DiscoverBeatInteractor
	createCampaignBeat        *narrativeinteractor.CreateCampaignBeatInteractor
	promoteBeatToMaster       *narrativeinteractor.PromoteBeatToMasterInteractor

	// Faction (10)
	createFaction      *factioninteractor.CreateFactionInteractor
	beginFactionDraft  *factioninteractor.BeginFactionDraftInteractor
	activateFaction    *factioninteractor.ActivateFactionInteractor
	addFactionMember   *factioninteractor.AddFactionMemberInteractor
	addStandingLevel   *factioninteractor.AddStandingLevelInteractor
	declareAlly        *factioninteractor.DeclareAllyInteractor
	declareWar         *factioninteractor.DeclareWarInteractor
	markFactionDormant *factioninteractor.MarkFactionDormantInteractor
	reactivateFaction  *factioninteractor.ReactivateFactionInteractor
	archiveFaction     *factioninteractor.ArchiveFactionInteractor

	// Location (5)
	createLocation   *locationinteractor.CreateLocationInteractor
	activateLocation *locationinteractor.ActivateLocationInteractor
	addScene         *locationinteractor.AddSceneInteractor
	connectLocations *locationinteractor.ConnectLocationsInteractor
	archiveLocation  *locationinteractor.ArchiveLocationInteractor

	// Character (13)
	createNPC                    *characterinteractor.CreateNPCInteractor
	beginNPCDraft                *characterinteractor.BeginNPCDraftInteractor
	updateNPCContent             *characterinteractor.UpdateNPCContentInteractor
	activateNPC                  *characterinteractor.ActivateNPCInteractor
	archiveNPC                   *characterinteractor.ArchiveNPCInteractor
	createPlayerCharacter        *characterinteractor.CreatePlayerCharacterInteractor
	updatePlayerCharacterContent *characterinteractor.UpdatePlayerCharacterContentInteractor
	retirePlayerCharacter        *characterinteractor.RetirePlayerCharacterInteractor
	createMacGuffin              *characterinteractor.CreateMacGuffinInteractor
	updateMacGuffinContent       *characterinteractor.UpdateMacGuffinContentInteractor
	assignMacGuffinToNPC         *characterinteractor.AssignMacGuffinToNPCInteractor
	assignMacGuffinToPC          *characterinteractor.AssignMacGuffinToPCInteractor
	destroyMacGuffin             *characterinteractor.DestroyMacGuffinInteractor

	// RevealEntity (1)
	revealEntity *sharedinteractor.RevealEntityInteractor
}

// ResolverConfig holds all dependencies required to construct a Resolver.
type ResolverConfig struct {
	Auth sharedinteractor.CallerIdentityPort

	CreateGame  *gameinteractor.CreateGameInteractor
	ArchiveGame *gameinteractor.ArchiveGameInteractor

	CreateCampaign         *campaigninteractor.CreateCampaignInteractor
	AddCharacterToCampaign *campaigninteractor.AddCharacterToCampaignInteractor
	BeginCampaignFormation *campaigninteractor.BeginCampaignFormationInteractor
	StartFirstSession      *campaigninteractor.StartFirstSessionInteractor
	StartNewSession        *campaigninteractor.StartNewSessionInteractor
	EndSession             *campaigninteractor.EndSessionInteractor
	MoveParty              *campaigninteractor.MovePartyInteractor
	CompleteCampaign       *campaigninteractor.CompleteCampaignInteractor

	CreateMasterBeat           *narrativeinteractor.CreateMasterBeatInteractor
	UpdateBeatContent          *narrativeinteractor.UpdateBeatContentInteractor
	AddBeatPrerequisite        *narrativeinteractor.AddBeatPrerequisiteInteractor
	AddActToMasterNarrative    *narrativeinteractor.AddActToMasterNarrativeInteractor
	AddSecretToMasterNarrative *narrativeinteractor.AddSecretToMasterNarrativeInteractor
	AddLoreToMasterNarrative   *narrativeinteractor.AddLoreToMasterNarrativeInteractor
	DiscoverBeat               *narrativeinteractor.DiscoverBeatInteractor
	CreateCampaignBeat         *narrativeinteractor.CreateCampaignBeatInteractor
	PromoteBeatToMaster        *narrativeinteractor.PromoteBeatToMasterInteractor

	CreateFaction      *factioninteractor.CreateFactionInteractor
	BeginFactionDraft  *factioninteractor.BeginFactionDraftInteractor
	ActivateFaction    *factioninteractor.ActivateFactionInteractor
	AddFactionMember   *factioninteractor.AddFactionMemberInteractor
	AddStandingLevel   *factioninteractor.AddStandingLevelInteractor
	DeclareAlly        *factioninteractor.DeclareAllyInteractor
	DeclareWar         *factioninteractor.DeclareWarInteractor
	MarkFactionDormant *factioninteractor.MarkFactionDormantInteractor
	ReactivateFaction  *factioninteractor.ReactivateFactionInteractor
	ArchiveFaction     *factioninteractor.ArchiveFactionInteractor

	CreateLocation   *locationinteractor.CreateLocationInteractor
	ActivateLocation *locationinteractor.ActivateLocationInteractor
	AddScene         *locationinteractor.AddSceneInteractor
	ConnectLocations *locationinteractor.ConnectLocationsInteractor
	ArchiveLocation  *locationinteractor.ArchiveLocationInteractor

	CreateNPC                    *characterinteractor.CreateNPCInteractor
	BeginNPCDraft                *characterinteractor.BeginNPCDraftInteractor
	UpdateNPCContent             *characterinteractor.UpdateNPCContentInteractor
	ActivateNPC                  *characterinteractor.ActivateNPCInteractor
	ArchiveNPC                   *characterinteractor.ArchiveNPCInteractor
	CreatePlayerCharacter        *characterinteractor.CreatePlayerCharacterInteractor
	UpdatePlayerCharacterContent *characterinteractor.UpdatePlayerCharacterContentInteractor
	RetirePlayerCharacter        *characterinteractor.RetirePlayerCharacterInteractor
	CreateMacGuffin              *characterinteractor.CreateMacGuffinInteractor
	UpdateMacGuffinContent       *characterinteractor.UpdateMacGuffinContentInteractor
	AssignMacGuffinToNPC         *characterinteractor.AssignMacGuffinToNPCInteractor
	AssignMacGuffinToPC          *characterinteractor.AssignMacGuffinToPCInteractor
	DestroyMacGuffin             *characterinteractor.DestroyMacGuffinInteractor

	RevealEntity *sharedinteractor.RevealEntityInteractor
}

// NewResolver wires all dependencies into a Resolver ready for the HTTP server.
func NewResolver(cfg ResolverConfig) *Resolver {
	return &Resolver{
		auth:                         cfg.Auth,
		createGame:                   cfg.CreateGame,
		archiveGame:                  cfg.ArchiveGame,
		createCampaign:               cfg.CreateCampaign,
		addCharacterToCampaign:       cfg.AddCharacterToCampaign,
		beginCampaignFormation:       cfg.BeginCampaignFormation,
		startFirstSession:            cfg.StartFirstSession,
		startNewSession:              cfg.StartNewSession,
		endSession:                   cfg.EndSession,
		moveParty:                    cfg.MoveParty,
		completeCampaign:             cfg.CompleteCampaign,
		createMasterBeat:             cfg.CreateMasterBeat,
		updateBeatContent:            cfg.UpdateBeatContent,
		addBeatPrerequisite:          cfg.AddBeatPrerequisite,
		addActToMasterNarrative:      cfg.AddActToMasterNarrative,
		addSecretToMasterNarrative:   cfg.AddSecretToMasterNarrative,
		addLoreToMasterNarrative:     cfg.AddLoreToMasterNarrative,
		discoverBeat:                 cfg.DiscoverBeat,
		createCampaignBeat:           cfg.CreateCampaignBeat,
		promoteBeatToMaster:          cfg.PromoteBeatToMaster,
		createFaction:                cfg.CreateFaction,
		beginFactionDraft:            cfg.BeginFactionDraft,
		activateFaction:              cfg.ActivateFaction,
		addFactionMember:             cfg.AddFactionMember,
		addStandingLevel:             cfg.AddStandingLevel,
		declareAlly:                  cfg.DeclareAlly,
		declareWar:                   cfg.DeclareWar,
		markFactionDormant:           cfg.MarkFactionDormant,
		reactivateFaction:            cfg.ReactivateFaction,
		archiveFaction:               cfg.ArchiveFaction,
		createLocation:               cfg.CreateLocation,
		activateLocation:             cfg.ActivateLocation,
		addScene:                     cfg.AddScene,
		connectLocations:             cfg.ConnectLocations,
		archiveLocation:              cfg.ArchiveLocation,
		createNPC:                    cfg.CreateNPC,
		beginNPCDraft:                cfg.BeginNPCDraft,
		updateNPCContent:             cfg.UpdateNPCContent,
		activateNPC:                  cfg.ActivateNPC,
		archiveNPC:                   cfg.ArchiveNPC,
		createPlayerCharacter:        cfg.CreatePlayerCharacter,
		updatePlayerCharacterContent: cfg.UpdatePlayerCharacterContent,
		retirePlayerCharacter:        cfg.RetirePlayerCharacter,
		createMacGuffin:              cfg.CreateMacGuffin,
		updateMacGuffinContent:       cfg.UpdateMacGuffinContent,
		assignMacGuffinToNPC:         cfg.AssignMacGuffinToNPC,
		assignMacGuffinToPC:          cfg.AssignMacGuffinToPC,
		destroyMacGuffin:             cfg.DestroyMacGuffin,
		revealEntity:                 cfg.RevealEntity,
	}
}
