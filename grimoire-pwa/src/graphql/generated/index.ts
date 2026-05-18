import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Date: { input: string; output: string; }
  DateTime: { input: string; output: string; }
};

export type AddActInput = {
  actId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type AddBeatPrerequisiteInput = {
  beatId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  prerequisiteId: Scalars['ID']['input'];
};

export type AddCharacterToCampaignInput = {
  campaignId: Scalars['ID']['input'];
  characterId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type AddFactionMemberInput = {
  factionId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  npcId: Scalars['ID']['input'];
  rank: Scalars['String']['input'];
};

export type AddLoreInput = {
  gameId: Scalars['ID']['input'];
  loreId: Scalars['ID']['input'];
};

export type AddSceneInput = {
  gameId: Scalars['ID']['input'];
  locationId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type AddSecretInput = {
  gameId: Scalars['ID']['input'];
  secretId: Scalars['ID']['input'];
};

export type AddStandingLevelInput = {
  factionId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  ordinal: Scalars['Int']['input'];
  threshold: Scalars['Int']['input'];
};

export type ArchiveGameInput = {
  gameId: Scalars['ID']['input'];
};

export type AssignMacGuffinToNpcInput = {
  gameId: Scalars['ID']['input'];
  macGuffinId: Scalars['ID']['input'];
  npcId: Scalars['ID']['input'];
};

export type AssignMacGuffinToPcInput = {
  characterId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  macGuffinId: Scalars['ID']['input'];
};

export type BeatDetail = {
  __typename?: 'BeatDetail';
  beatType: BeatType;
  description: Scalars['String']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  playerDescription: Scalars['String']['output'];
  prerequisites: Array<BeatSummary>;
  scope: BeatScope;
};

export enum BeatScope {
  Campaign = 'CAMPAIGN',
  Master = 'MASTER'
}

export type BeatSummary = {
  __typename?: 'BeatSummary';
  beatType: BeatType;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  prerequisiteCount: Scalars['Int']['output'];
  scope: BeatScope;
};

export enum BeatType {
  CampaignSpecific = 'CAMPAIGN_SPECIFIC',
  Optional = 'OPTIONAL',
  Required = 'REQUIRED'
}

export type BeginCampaignFormationInput = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type CampaignDetail = {
  __typename?: 'CampaignDetail';
  characters: Array<PlayerCharacterSummary>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  sessionCount: Scalars['Int']['output'];
  status: EntityStatus;
};

export type CampaignSessionState = {
  __typename?: 'CampaignSessionState';
  characters: Array<PlayerCharacterSummary>;
  currentLocation?: Maybe<LocationWithConnections>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  sessionId: Scalars['ID']['output'];
  status: EntityStatus;
};

export type CampaignSummary = {
  __typename?: 'CampaignSummary';
  characterCount: Scalars['Int']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  sessionCount: Scalars['Int']['output'];
  status: EntityStatus;
};

export type CompleteCampaignInput = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type ConnectLocationsInput = {
  fromId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  toId: Scalars['ID']['input'];
};

export enum ConnectionDirection {
  Both = 'BOTH',
  Inbound = 'INBOUND',
  Outbound = 'OUTBOUND'
}

export type CreateCampaignBeatInput = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateCampaignInput = {
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateFactionInput = {
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateGameInput = {
  name: Scalars['String']['input'];
};

export type CreateLocationInput = {
  gameId: Scalars['ID']['input'];
  locationType: LocationType;
  name: Scalars['String']['input'];
  parentId?: InputMaybe<Scalars['ID']['input']>;
};

export type CreateMacGuffinInput = {
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateMasterBeatInput = {
  beatType: BeatType;
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreateNpcInput = {
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
};

export type CreatePlayerCharacterInput = {
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  ownerPlayerId?: InputMaybe<Scalars['ID']['input']>;
};

export type DiscoverBeatInput = {
  beatId: Scalars['ID']['input'];
  campaignNarrativeId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type DiscoveredBeat = {
  __typename?: 'DiscoveredBeat';
  beatType: BeatType;
  discoveredInSession: Scalars['Int']['output'];
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type EndSessionInput = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  notes?: InputMaybe<Scalars['String']['input']>;
  sessionId: Scalars['ID']['input'];
};

export type EntityCountSummary = {
  __typename?: 'EntityCountSummary';
  active: Scalars['Int']['output'];
  archived: Scalars['Int']['output'];
  draft: Scalars['Int']['output'];
  idle: Scalars['Int']['output'];
};

export enum EntityStatus {
  Active = 'ACTIVE',
  Archived = 'ARCHIVED',
  Complete = 'COMPLETE',
  Draft = 'DRAFT',
  Forming = 'FORMING',
  Idle = 'IDLE',
  New = 'NEW',
  Retired = 'RETIRED'
}

export type FactionDetail = {
  __typename?: 'FactionDetail';
  allies: Array<FactionRef>;
  enemies: Array<FactionRef>;
  id: Scalars['ID']['output'];
  members: Array<FactionMember>;
  name: Scalars['String']['output'];
  playerVisible: Scalars['Boolean']['output'];
  standingLevels: Array<StandingLevel>;
  status: EntityStatus;
};

export type FactionLifecycleInput = {
  factionId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type FactionMember = {
  __typename?: 'FactionMember';
  npc: NpcRef;
  rank: Scalars['String']['output'];
};

export type FactionMembership = {
  __typename?: 'FactionMembership';
  faction: FactionRef;
  rank: Scalars['String']['output'];
};

export type FactionRef = {
  __typename?: 'FactionRef';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type FactionRelationshipInput = {
  factionId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  targetId: Scalars['ID']['input'];
};

export type FactionSummary = {
  __typename?: 'FactionSummary';
  allies: Array<FactionRef>;
  enemies: Array<FactionRef>;
  id: Scalars['ID']['output'];
  memberCount: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  playerVisible: Scalars['Boolean']['output'];
  standingLevelCount: Scalars['Int']['output'];
  status: EntityStatus;
};

export type GameDetail = {
  __typename?: 'GameDetail';
  campaigns: Array<CampaignSummary>;
  factionSummary: EntityCountSummary;
  id: Scalars['ID']['output'];
  locationSummary: EntityCountSummary;
  name: Scalars['String']['output'];
  npcSummary: EntityCountSummary;
  status: EntityStatus;
};

export type GameSummary = {
  __typename?: 'GameSummary';
  campaignCount: Scalars['Int']['output'];
  id: Scalars['ID']['output'];
  lastActivityAt?: Maybe<Scalars['DateTime']['output']>;
  name: Scalars['String']['output'];
  status: EntityStatus;
};

export type LocationConnection = {
  __typename?: 'LocationConnection';
  direction: ConnectionDirection;
  toLocation: LocationSummary;
};

export type LocationDetail = {
  __typename?: 'LocationDetail';
  connections: Array<LocationConnection>;
  id: Scalars['ID']['output'];
  locationType: LocationType;
  name: Scalars['String']['output'];
  parentLocation?: Maybe<LocationSummary>;
  scenes: Array<SceneSummary>;
  status: EntityStatus;
};

export type LocationLifecycleInput = {
  gameId: Scalars['ID']['input'];
  locationId: Scalars['ID']['input'];
};

export type LocationRef = {
  __typename?: 'LocationRef';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type LocationSummary = {
  __typename?: 'LocationSummary';
  childCount: Scalars['Int']['output'];
  id: Scalars['ID']['output'];
  locationType: LocationType;
  name: Scalars['String']['output'];
  parentId?: Maybe<Scalars['ID']['output']>;
  sceneCount: Scalars['Int']['output'];
  status: EntityStatus;
};

export enum LocationType {
  Building = 'BUILDING',
  Region = 'REGION',
  Scene = 'SCENE',
  Settlement = 'SETTLEMENT',
  World = 'WORLD'
}

export type LocationWithConnections = {
  __typename?: 'LocationWithConnections';
  connections: Array<LocationConnection>;
  id: Scalars['ID']['output'];
  locationType: LocationType;
  name: Scalars['String']['output'];
};

export type MacGuffinLifecycleInput = {
  gameId: Scalars['ID']['input'];
  macGuffinId: Scalars['ID']['input'];
};

export type MacGuffinPossessor = LocationRef | NpcRef | PlayerCharacterRef;

export type MacGuffinSummary = {
  __typename?: 'MacGuffinSummary';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  playerVisible: Scalars['Boolean']['output'];
  possessor?: Maybe<MacGuffinPossessor>;
};

export type MasterNarrativeView = {
  __typename?: 'MasterNarrativeView';
  beats: Array<BeatSummary>;
  id: Scalars['ID']['output'];
};

export type MovePartyInput = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
  locationId: Scalars['ID']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Activate a Faction. Transitions Draft → Active. */
  activateFaction: MutationResult;
  /** Activate a Location. Transitions Draft → Active. */
  activateLocation: MutationResult;
  /** Activate an NPC. Transitions Draft → Active. */
  activateNPC: MutationResult;
  /** Add an Act reference to the MasterNarrative. */
  addActToMasterNarrative: MutationResult;
  /** Add a prerequisite relationship between two Beats. */
  addBeatPrerequisite: MutationResult;
  /** Add a PlayerCharacter to a New or Forming Campaign. */
  addCharacterToCampaign: MutationResult;
  /** Add an NPC member to a Draft or Active Faction. */
  addFactionMember: MutationResult;
  /** Add a Lore reference to the MasterNarrative. */
  addLoreToMasterNarrative: MutationResult;
  /** Add a Scene to a Draft or Active Location. */
  addScene: MutationResult;
  /** Add a Secret reference to the MasterNarrative. */
  addSecretToMasterNarrative: MutationResult;
  /** Add a standing level definition to a Faction. */
  addStandingLevel: MutationResult;
  /** Archive a Faction. Terminal. Works from Active or Idle. */
  archiveFaction: MutationResult;
  /** Archive a completed Game. Game must be Idle. */
  archiveGame: MutationResult;
  /** Archive a Location. Terminal. Cascades to children. */
  archiveLocation: MutationResult;
  /** Archive an NPC. Terminal. Works from Active or Idle. */
  archiveNPC: MutationResult;
  /** Assign a MacGuffin to an NPC. */
  assignMacGuffinToNPC: MutationResult;
  /** Assign a MacGuffin to a PlayerCharacter. */
  assignMacGuffinToPC: MutationResult;
  /** Begin party formation. Campaign transitions New → Forming. */
  beginCampaignFormation: MutationResult;
  /** Begin drafting a Faction. Transitions New → Draft. */
  beginFactionDraft: MutationResult;
  /** Begin drafting an NPC. Transitions New → Draft. */
  beginNPCDraft: MutationResult;
  /** Complete a Campaign. Campaign transitions Idle → Complete. Terminal. */
  completeCampaign: MutationResult;
  /**
   * Create a directed travel connection from one Location to another.
   * For symmetric (A↔B) connections call twice.
   */
  connectLocations: MutationResult;
  /** Create a Campaign linked to a Game. */
  createCampaign: MutationResult;
  /** Create a campaign-specific Beat improvised at the table. */
  createCampaignBeat: MutationResult;
  /** Create a new Faction. */
  createFaction: MutationResult;
  /** Create a new Game. The GM owns everything beneath it. */
  createGame: MutationResult;
  /** Create a Location. parentId is optional — omit for top-level. */
  createLocation: MutationResult;
  /** Create a narratively significant item. */
  createMacGuffin: MutationResult;
  /** Create a Beat scoped to the MasterNarrative. */
  createMasterBeat: MutationResult;
  /** Create a new NPC. */
  createNPC: MutationResult;
  /** Create a PlayerCharacter. ownerPlayerId is optional. */
  createPlayerCharacter: MutationResult;
  /** Declare an ally relationship between two Factions. */
  declareAlly: MutationResult;
  /** Declare a war relationship between two Factions. */
  declareWar: MutationResult;
  /** Destroy a MacGuffin. Terminal. */
  destroyMacGuffin: MutationResult;
  /** Discover a Beat for a Campaign. Called at the table. */
  discoverBeat: MutationResult;
  /** End the current session. Campaign transitions Active → Idle. */
  endSession: MutationResult;
  /** Mark a Faction dormant. Transitions Active → Idle. */
  markFactionDormant: MutationResult;
  /** Move the party to a new Location. Campaign must be Active. */
  moveParty: MutationResult;
  /** Promote a campaign-specific Beat to MasterNarrative scope. */
  promoteBeatToMaster: MutationResult;
  /** Reactivate a dormant Faction. Transitions Idle → Active. */
  reactivateFaction: MutationResult;
  /** Retire a PlayerCharacter. Terminal. */
  retirePlayerCharacter: MutationResult;
  /**
   * Reveal an entity to one or more players mid-session.
   * entityType must be one of: NPC, PLAYER_CHARACTER, MACGUFFIN, FACTION.
   */
  revealEntity: MutationResult;
  /** Start the first session. Campaign transitions Forming → Active. */
  startFirstSession: MutationResult;
  /** Start a follow-on session. Campaign transitions Idle → Active. */
  startNewSession: MutationResult;
  /** Update content fields on any Beat. */
  updateBeatContent: MutationResult;
  /** Update content fields on a MacGuffin. */
  updateMacGuffinContent: MutationResult;
  /** Update content fields on a Draft or Active NPC. */
  updateNPCContent: MutationResult;
  /** Update content fields on a PlayerCharacter. */
  updatePlayerCharacterContent: MutationResult;
};


export type MutationActivateFactionArgs = {
  input: FactionLifecycleInput;
};


export type MutationActivateLocationArgs = {
  input: LocationLifecycleInput;
};


export type MutationActivateNpcArgs = {
  input: NpcLifecycleInput;
};


export type MutationAddActToMasterNarrativeArgs = {
  input: AddActInput;
};


export type MutationAddBeatPrerequisiteArgs = {
  input: AddBeatPrerequisiteInput;
};


export type MutationAddCharacterToCampaignArgs = {
  input: AddCharacterToCampaignInput;
};


export type MutationAddFactionMemberArgs = {
  input: AddFactionMemberInput;
};


export type MutationAddLoreToMasterNarrativeArgs = {
  input: AddLoreInput;
};


export type MutationAddSceneArgs = {
  input: AddSceneInput;
};


export type MutationAddSecretToMasterNarrativeArgs = {
  input: AddSecretInput;
};


export type MutationAddStandingLevelArgs = {
  input: AddStandingLevelInput;
};


export type MutationArchiveFactionArgs = {
  input: FactionLifecycleInput;
};


export type MutationArchiveGameArgs = {
  input: ArchiveGameInput;
};


export type MutationArchiveLocationArgs = {
  input: LocationLifecycleInput;
};


export type MutationArchiveNpcArgs = {
  input: NpcLifecycleInput;
};


export type MutationAssignMacGuffinToNpcArgs = {
  input: AssignMacGuffinToNpcInput;
};


export type MutationAssignMacGuffinToPcArgs = {
  input: AssignMacGuffinToPcInput;
};


export type MutationBeginCampaignFormationArgs = {
  input: BeginCampaignFormationInput;
};


export type MutationBeginFactionDraftArgs = {
  input: FactionLifecycleInput;
};


export type MutationBeginNpcDraftArgs = {
  input: NpcLifecycleInput;
};


export type MutationCompleteCampaignArgs = {
  input: CompleteCampaignInput;
};


export type MutationConnectLocationsArgs = {
  input: ConnectLocationsInput;
};


export type MutationCreateCampaignArgs = {
  input: CreateCampaignInput;
};


export type MutationCreateCampaignBeatArgs = {
  input: CreateCampaignBeatInput;
};


export type MutationCreateFactionArgs = {
  input: CreateFactionInput;
};


export type MutationCreateGameArgs = {
  input: CreateGameInput;
};


export type MutationCreateLocationArgs = {
  input: CreateLocationInput;
};


export type MutationCreateMacGuffinArgs = {
  input: CreateMacGuffinInput;
};


export type MutationCreateMasterBeatArgs = {
  input: CreateMasterBeatInput;
};


export type MutationCreateNpcArgs = {
  input: CreateNpcInput;
};


export type MutationCreatePlayerCharacterArgs = {
  input: CreatePlayerCharacterInput;
};


export type MutationDeclareAllyArgs = {
  input: FactionRelationshipInput;
};


export type MutationDeclareWarArgs = {
  input: FactionRelationshipInput;
};


export type MutationDestroyMacGuffinArgs = {
  input: MacGuffinLifecycleInput;
};


export type MutationDiscoverBeatArgs = {
  input: DiscoverBeatInput;
};


export type MutationEndSessionArgs = {
  input: EndSessionInput;
};


export type MutationMarkFactionDormantArgs = {
  input: FactionLifecycleInput;
};


export type MutationMovePartyArgs = {
  input: MovePartyInput;
};


export type MutationPromoteBeatToMasterArgs = {
  input: PromoteBeatInput;
};


export type MutationReactivateFactionArgs = {
  input: FactionLifecycleInput;
};


export type MutationRetirePlayerCharacterArgs = {
  input: PcLifecycleInput;
};


export type MutationRevealEntityArgs = {
  input: RevealEntityInput;
};


export type MutationStartFirstSessionArgs = {
  input: StartFirstSessionInput;
};


export type MutationStartNewSessionArgs = {
  input: StartNewSessionInput;
};


export type MutationUpdateBeatContentArgs = {
  input: UpdateBeatContentInput;
};


export type MutationUpdateMacGuffinContentArgs = {
  input: UpdateMacGuffinContentInput;
};


export type MutationUpdateNpcContentArgs = {
  input: UpdateNpcContentInput;
};


export type MutationUpdatePlayerCharacterContentArgs = {
  input: UpdatePcContentInput;
};

/**
 * Returned by every mutation. Callers use the ID to issue a follow-up
 * query if they need rich data about the resulting state.
 */
export type MutationResult = {
  __typename?: 'MutationResult';
  id: Scalars['ID']['output'];
  status: Scalars['String']['output'];
};

export type NpcDetail = {
  __typename?: 'NPCDetail';
  description: Scalars['String']['output'];
  factionMemberships: Array<FactionMembership>;
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  playerDescription: Scalars['String']['output'];
  playerVisible: Scalars['Boolean']['output'];
  status: EntityStatus;
};

export type NpcLifecycleInput = {
  gameId: Scalars['ID']['input'];
  npcId: Scalars['ID']['input'];
};

export type NpcRef = {
  __typename?: 'NPCRef';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type NpcSummary = {
  __typename?: 'NPCSummary';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  playerVisible: Scalars['Boolean']['output'];
  status: EntityStatus;
};

export type PcLifecycleInput = {
  characterId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type PlayerCharacterRef = {
  __typename?: 'PlayerCharacterRef';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type PlayerCharacterSummary = {
  __typename?: 'PlayerCharacterSummary';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
  ownerPlayerId?: Maybe<Scalars['String']['output']>;
  status: EntityStatus;
};

export type PromoteBeatInput = {
  beatId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};

export type Query = {
  __typename?: 'Query';
  /** Active locations only. Used in MoveParty modal. */
  activeLocations: Array<LocationSummary>;
  /** Beats available to a campaign given its current discoveredBeatIDs. */
  availableBeats: Array<BeatSummary>;
  /** Single beat with full content and prerequisites. */
  beat?: Maybe<BeatDetail>;
  /** A single campaign. */
  campaign?: Maybe<CampaignDetail>;
  /** Campaign beats created during sessions (scope: campaign). */
  campaignBeats: Array<BeatSummary>;
  /** Session-time state for the Party Tab. */
  campaignSessionState?: Maybe<CampaignSessionState>;
  /** All campaigns for a game. */
  campaigns: Array<CampaignSummary>;
  /** All beats discovered by a campaign, with session attribution. */
  discoveredBeats: Array<DiscoveredBeat>;
  /** Single faction with full detail. */
  faction?: Maybe<FactionDetail>;
  /** All factions for a game. */
  factions: Array<FactionSummary>;
  /** A single game by ID. Returns null if not found or not owned by caller. */
  game?: Maybe<GameDetail>;
  /** All games owned by the authenticated GM. */
  games: Array<GameSummary>;
  /** Single location with scenes and connections. */
  location?: Maybe<LocationDetail>;
  /** All locations for a game as a flat list with parentId. */
  locations: Array<LocationSummary>;
  /** All MacGuffins for a game. */
  macguffins: Array<MacGuffinSummary>;
  /** All master beats for a game. */
  masterNarrative?: Maybe<MasterNarrativeView>;
  /** Single NPC with full content and faction memberships. */
  npc?: Maybe<NpcDetail>;
  /** All NPCs for a game. */
  npcs: Array<NpcSummary>;
  /** All player characters for a game. */
  playerCharacters: Array<PlayerCharacterSummary>;
  /** Pre-session launch state. */
  sessionLauncher?: Maybe<SessionLauncherState>;
  /** End-of-session summary. */
  sessionSummary?: Maybe<SessionSummary>;
};


export type QueryActiveLocationsArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryAvailableBeatsArgs = {
  campaignNarrativeId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};


export type QueryBeatArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryCampaignArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryCampaignBeatsArgs = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};


export type QueryCampaignSessionStateArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryCampaignsArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryDiscoveredBeatsArgs = {
  campaignNarrativeId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};


export type QueryFactionArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryFactionsArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryGameArgs = {
  id: Scalars['ID']['input'];
};


export type QueryLocationArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryLocationsArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryMacguffinsArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryMasterNarrativeArgs = {
  gameId: Scalars['ID']['input'];
};


export type QueryNpcArgs = {
  gameId: Scalars['ID']['input'];
  id: Scalars['ID']['input'];
};


export type QueryNpcsArgs = {
  excludeArchived?: InputMaybe<Scalars['Boolean']['input']>;
  gameId: Scalars['ID']['input'];
};


export type QueryPlayerCharactersArgs = {
  gameId: Scalars['ID']['input'];
};


export type QuerySessionLauncherArgs = {
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
};


export type QuerySessionSummaryArgs = {
  gameId: Scalars['ID']['input'];
  sessionId: Scalars['ID']['input'];
};

export type RevealEntityInput = {
  entityId: Scalars['ID']['input'];
  entityType: RevealableEntityType;
  gameId: Scalars['ID']['input'];
  revealedTo: Array<Scalars['ID']['input']>;
  sessionId: Scalars['ID']['input'];
};

export enum RevealableEntityType {
  Faction = 'FACTION',
  Macguffin = 'MACGUFFIN',
  Npc = 'NPC',
  PlayerCharacter = 'PLAYER_CHARACTER'
}

export type SceneSummary = {
  __typename?: 'SceneSummary';
  id: Scalars['ID']['output'];
  name: Scalars['String']['output'];
};

export type SessionLauncherState = {
  __typename?: 'SessionLauncherState';
  campaignId: Scalars['ID']['output'];
  campaignName: Scalars['String']['output'];
  characters: Array<PlayerCharacterSummary>;
  sessionCount: Scalars['Int']['output'];
  status: EntityStatus;
};

export type SessionSummary = {
  __typename?: 'SessionSummary';
  beatsDiscoveredCount: Scalars['Int']['output'];
  entitiesRevealedCount: Scalars['Int']['output'];
  finalLocation?: Maybe<LocationRef>;
  sessionId: Scalars['ID']['output'];
};

export type StandingLevel = {
  __typename?: 'StandingLevel';
  name: Scalars['String']['output'];
  ordinal: Scalars['Int']['output'];
  threshold: Scalars['Int']['output'];
};

export type StartFirstSessionInput = {
  campaignId: Scalars['ID']['input'];
  date: Scalars['Date']['input'];
  gameId: Scalars['ID']['input'];
};

export type StartNewSessionInput = {
  campaignId: Scalars['ID']['input'];
  date: Scalars['Date']['input'];
  gameId: Scalars['ID']['input'];
};

export type UpdateBeatContentInput = {
  beatId: Scalars['ID']['input'];
  description: Scalars['String']['input'];
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  playerDesc: Scalars['String']['input'];
};

export type UpdateMacGuffinContentInput = {
  description: Scalars['String']['input'];
  gameId: Scalars['ID']['input'];
  macGuffinId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  playerDesc: Scalars['String']['input'];
};

export type UpdateNpcContentInput = {
  description: Scalars['String']['input'];
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  npcId: Scalars['ID']['input'];
  playerDesc: Scalars['String']['input'];
};

export type UpdatePcContentInput = {
  characterId: Scalars['ID']['input'];
  description: Scalars['String']['input'];
  gameId: Scalars['ID']['input'];
  name: Scalars['String']['input'];
  playerDesc: Scalars['String']['input'];
};

export type CreateCampaignMutationVariables = Exact<{
  input: CreateCampaignInput;
}>;


export type CreateCampaignMutation = { __typename?: 'Mutation', createCampaign: { __typename?: 'MutationResult', id: string, status: string } };

export type BeginCampaignFormationMutationVariables = Exact<{
  input: BeginCampaignFormationInput;
}>;


export type BeginCampaignFormationMutation = { __typename?: 'Mutation', beginCampaignFormation: { __typename?: 'MutationResult', id: string, status: string } };

export type AddCharacterToCampaignMutationVariables = Exact<{
  input: AddCharacterToCampaignInput;
}>;


export type AddCharacterToCampaignMutation = { __typename?: 'Mutation', addCharacterToCampaign: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateNpcMutationVariables = Exact<{
  input: CreateNpcInput;
}>;


export type CreateNpcMutation = { __typename?: 'Mutation', createNPC: { __typename?: 'MutationResult', id: string, status: string } };

export type UpdateNpcContentMutationVariables = Exact<{
  input: UpdateNpcContentInput;
}>;


export type UpdateNpcContentMutation = { __typename?: 'Mutation', updateNPCContent: { __typename?: 'MutationResult', id: string, status: string } };

export type ActivateNpcMutationVariables = Exact<{
  input: NpcLifecycleInput;
}>;


export type ActivateNpcMutation = { __typename?: 'Mutation', activateNPC: { __typename?: 'MutationResult', id: string, status: string } };

export type ArchiveNpcMutationVariables = Exact<{
  input: NpcLifecycleInput;
}>;


export type ArchiveNpcMutation = { __typename?: 'Mutation', archiveNPC: { __typename?: 'MutationResult', id: string, status: string } };

export type CreatePlayerCharacterMutationVariables = Exact<{
  input: CreatePlayerCharacterInput;
}>;


export type CreatePlayerCharacterMutation = { __typename?: 'Mutation', createPlayerCharacter: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateMacGuffinMutationVariables = Exact<{
  input: CreateMacGuffinInput;
}>;


export type CreateMacGuffinMutation = { __typename?: 'Mutation', createMacGuffin: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateFactionMutationVariables = Exact<{
  input: CreateFactionInput;
}>;


export type CreateFactionMutation = { __typename?: 'Mutation', createFaction: { __typename?: 'MutationResult', id: string, status: string } };

export type ActivateFactionMutationVariables = Exact<{
  input: FactionLifecycleInput;
}>;


export type ActivateFactionMutation = { __typename?: 'Mutation', activateFaction: { __typename?: 'MutationResult', id: string, status: string } };

export type AddFactionMemberMutationVariables = Exact<{
  input: AddFactionMemberInput;
}>;


export type AddFactionMemberMutation = { __typename?: 'Mutation', addFactionMember: { __typename?: 'MutationResult', id: string, status: string } };

export type AddStandingLevelMutationVariables = Exact<{
  input: AddStandingLevelInput;
}>;


export type AddStandingLevelMutation = { __typename?: 'Mutation', addStandingLevel: { __typename?: 'MutationResult', id: string, status: string } };

export type DeclareAllyMutationVariables = Exact<{
  input: FactionRelationshipInput;
}>;


export type DeclareAllyMutation = { __typename?: 'Mutation', declareAlly: { __typename?: 'MutationResult', id: string, status: string } };

export type DeclareWarMutationVariables = Exact<{
  input: FactionRelationshipInput;
}>;


export type DeclareWarMutation = { __typename?: 'Mutation', declareWar: { __typename?: 'MutationResult', id: string, status: string } };

export type ArchiveFactionMutationVariables = Exact<{
  input: FactionLifecycleInput;
}>;


export type ArchiveFactionMutation = { __typename?: 'Mutation', archiveFaction: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateGameMutationVariables = Exact<{
  input: CreateGameInput;
}>;


export type CreateGameMutation = { __typename?: 'Mutation', createGame: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateLocationMutationVariables = Exact<{
  input: CreateLocationInput;
}>;


export type CreateLocationMutation = { __typename?: 'Mutation', createLocation: { __typename?: 'MutationResult', id: string, status: string } };

export type ActivateLocationMutationVariables = Exact<{
  input: LocationLifecycleInput;
}>;


export type ActivateLocationMutation = { __typename?: 'Mutation', activateLocation: { __typename?: 'MutationResult', id: string, status: string } };

export type AddSceneMutationVariables = Exact<{
  input: AddSceneInput;
}>;


export type AddSceneMutation = { __typename?: 'Mutation', addScene: { __typename?: 'MutationResult', id: string, status: string } };

export type ConnectLocationsMutationVariables = Exact<{
  input: ConnectLocationsInput;
}>;


export type ConnectLocationsMutation = { __typename?: 'Mutation', connectLocations: { __typename?: 'MutationResult', id: string, status: string } };

export type ArchiveLocationMutationVariables = Exact<{
  input: LocationLifecycleInput;
}>;


export type ArchiveLocationMutation = { __typename?: 'Mutation', archiveLocation: { __typename?: 'MutationResult', id: string, status: string } };

export type CreateMasterBeatMutationVariables = Exact<{
  input: CreateMasterBeatInput;
}>;


export type CreateMasterBeatMutation = { __typename?: 'Mutation', createMasterBeat: { __typename?: 'MutationResult', id: string, status: string } };

export type UpdateBeatContentMutationVariables = Exact<{
  input: UpdateBeatContentInput;
}>;


export type UpdateBeatContentMutation = { __typename?: 'Mutation', updateBeatContent: { __typename?: 'MutationResult', id: string, status: string } };

export type AddBeatPrerequisiteMutationVariables = Exact<{
  input: AddBeatPrerequisiteInput;
}>;


export type AddBeatPrerequisiteMutation = { __typename?: 'Mutation', addBeatPrerequisite: { __typename?: 'MutationResult', id: string, status: string } };

export type CampaignsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type CampaignsQuery = { __typename?: 'Query', campaigns: Array<{ __typename?: 'CampaignSummary', id: string, name: string, status: EntityStatus, characterCount: number, sessionCount: number }> };

export type CampaignQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type CampaignQuery = { __typename?: 'Query', campaign?: { __typename?: 'CampaignDetail', id: string, name: string, status: EntityStatus, sessionCount: number, characters: Array<{ __typename?: 'PlayerCharacterSummary', id: string, name: string, status: EntityStatus, ownerPlayerId?: string | null }> } | null };

export type CampaignSessionStateQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type CampaignSessionStateQuery = { __typename?: 'Query', campaignSessionState?: { __typename?: 'CampaignSessionState', id: string, name: string, status: EntityStatus, sessionId: string, currentLocation?: { __typename?: 'LocationWithConnections', id: string, name: string, locationType: LocationType, connections: Array<{ __typename?: 'LocationConnection', direction: ConnectionDirection, toLocation: { __typename?: 'LocationSummary', id: string, name: string, locationType: LocationType, status: EntityStatus, parentId?: string | null, sceneCount: number, childCount: number } }> } | null, characters: Array<{ __typename?: 'PlayerCharacterSummary', id: string, name: string, status: EntityStatus, ownerPlayerId?: string | null }> } | null };

export type SessionLauncherQueryVariables = Exact<{
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type SessionLauncherQuery = { __typename?: 'Query', sessionLauncher?: { __typename?: 'SessionLauncherState', campaignId: string, campaignName: string, status: EntityStatus, sessionCount: number, characters: Array<{ __typename?: 'PlayerCharacterSummary', id: string, name: string, status: EntityStatus, ownerPlayerId?: string | null }> } | null };

export type SessionSummaryQueryVariables = Exact<{
  sessionId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type SessionSummaryQuery = { __typename?: 'Query', sessionSummary?: { __typename?: 'SessionSummary', sessionId: string, beatsDiscoveredCount: number, entitiesRevealedCount: number, finalLocation?: { __typename?: 'LocationRef', id: string, name: string } | null } | null };

export type NpcsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
  excludeArchived?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type NpcsQuery = { __typename?: 'Query', npcs: Array<{ __typename?: 'NPCSummary', id: string, name: string, status: EntityStatus, playerVisible: boolean }> };

export type NpcQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type NpcQuery = { __typename?: 'Query', npc?: { __typename?: 'NPCDetail', id: string, name: string, status: EntityStatus, playerVisible: boolean, description: string, playerDescription: string, factionMemberships: Array<{ __typename?: 'FactionMembership', rank: string, faction: { __typename?: 'FactionRef', id: string, name: string } }> } | null };

export type PlayerCharactersQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type PlayerCharactersQuery = { __typename?: 'Query', playerCharacters: Array<{ __typename?: 'PlayerCharacterSummary', id: string, name: string, status: EntityStatus, ownerPlayerId?: string | null }> };

export type FactionsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type FactionsQuery = { __typename?: 'Query', factions: Array<{ __typename?: 'FactionSummary', id: string, name: string, status: EntityStatus, playerVisible: boolean, memberCount: number, standingLevelCount: number, allies: Array<{ __typename?: 'FactionRef', id: string, name: string }>, enemies: Array<{ __typename?: 'FactionRef', id: string, name: string }> }> };

export type FactionQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type FactionQuery = { __typename?: 'Query', faction?: { __typename?: 'FactionDetail', id: string, name: string, status: EntityStatus, playerVisible: boolean, members: Array<{ __typename?: 'FactionMember', rank: string, npc: { __typename?: 'NPCRef', id: string, name: string } }>, standingLevels: Array<{ __typename?: 'StandingLevel', ordinal: number, name: string, threshold: number }>, allies: Array<{ __typename?: 'FactionRef', id: string, name: string }>, enemies: Array<{ __typename?: 'FactionRef', id: string, name: string }> } | null };

export type GamesQueryVariables = Exact<{ [key: string]: never; }>;


export type GamesQuery = { __typename?: 'Query', games: Array<{ __typename?: 'GameSummary', id: string, name: string, status: EntityStatus, campaignCount: number, lastActivityAt?: string | null }> };

export type GameQueryVariables = Exact<{
  id: Scalars['ID']['input'];
}>;


export type GameQuery = { __typename?: 'Query', game?: { __typename?: 'GameDetail', id: string, name: string, status: EntityStatus, locationSummary: { __typename?: 'EntityCountSummary', draft: number, active: number, idle: number, archived: number }, factionSummary: { __typename?: 'EntityCountSummary', draft: number, active: number, idle: number, archived: number }, npcSummary: { __typename?: 'EntityCountSummary', draft: number, active: number, idle: number, archived: number }, campaigns: Array<{ __typename?: 'CampaignSummary', id: string, name: string, status: EntityStatus, characterCount: number, sessionCount: number }> } | null };

export type LocationsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type LocationsQuery = { __typename?: 'Query', locations: Array<{ __typename?: 'LocationSummary', id: string, name: string, locationType: LocationType, status: EntityStatus, parentId?: string | null, sceneCount: number, childCount: number }> };

export type LocationQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type LocationQuery = { __typename?: 'Query', location?: { __typename?: 'LocationDetail', id: string, name: string, locationType: LocationType, status: EntityStatus, parentLocation?: { __typename?: 'LocationSummary', id: string, name: string, locationType: LocationType, status: EntityStatus, parentId?: string | null, sceneCount: number, childCount: number } | null, scenes: Array<{ __typename?: 'SceneSummary', id: string, name: string }>, connections: Array<{ __typename?: 'LocationConnection', direction: ConnectionDirection, toLocation: { __typename?: 'LocationSummary', id: string, name: string, locationType: LocationType, status: EntityStatus, parentId?: string | null, sceneCount: number, childCount: number } }> } | null };

export type ActiveLocationsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type ActiveLocationsQuery = { __typename?: 'Query', activeLocations: Array<{ __typename?: 'LocationSummary', id: string, name: string, locationType: LocationType, status: EntityStatus, parentId?: string | null, sceneCount: number, childCount: number }> };

export type MacGuffinsQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type MacGuffinsQuery = { __typename?: 'Query', macguffins: Array<{ __typename?: 'MacGuffinSummary', id: string, name: string, playerVisible: boolean, possessor?: { __typename?: 'LocationRef', id: string, name: string } | { __typename?: 'NPCRef', id: string, name: string } | { __typename?: 'PlayerCharacterRef', id: string, name: string } | null }> };

export type MasterNarrativeQueryVariables = Exact<{
  gameId: Scalars['ID']['input'];
}>;


export type MasterNarrativeQuery = { __typename?: 'Query', masterNarrative?: { __typename?: 'MasterNarrativeView', id: string, beats: Array<{ __typename?: 'BeatSummary', id: string, name: string, beatType: BeatType, scope: BeatScope, prerequisiteCount: number }> } | null };

export type BeatQueryVariables = Exact<{
  id: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type BeatQuery = { __typename?: 'Query', beat?: { __typename?: 'BeatDetail', id: string, name: string, beatType: BeatType, scope: BeatScope, description: string, playerDescription: string, prerequisites: Array<{ __typename?: 'BeatSummary', id: string, name: string, beatType: BeatType, scope: BeatScope, prerequisiteCount: number }> } | null };

export type AvailableBeatsQueryVariables = Exact<{
  campaignNarrativeId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type AvailableBeatsQuery = { __typename?: 'Query', availableBeats: Array<{ __typename?: 'BeatSummary', id: string, name: string, beatType: BeatType, scope: BeatScope, prerequisiteCount: number }> };

export type DiscoveredBeatsQueryVariables = Exact<{
  campaignNarrativeId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type DiscoveredBeatsQuery = { __typename?: 'Query', discoveredBeats: Array<{ __typename?: 'DiscoveredBeat', id: string, name: string, beatType: BeatType, discoveredInSession: number }> };

export type CampaignBeatsQueryVariables = Exact<{
  campaignId: Scalars['ID']['input'];
  gameId: Scalars['ID']['input'];
}>;


export type CampaignBeatsQuery = { __typename?: 'Query', campaignBeats: Array<{ __typename?: 'BeatSummary', id: string, name: string, beatType: BeatType, scope: BeatScope, prerequisiteCount: number }> };


export const CreateCampaignDocument = gql`
    mutation CreateCampaign($input: CreateCampaignInput!) {
  createCampaign(input: $input) {
    id
    status
  }
}
    `;
export type CreateCampaignMutationFn = Apollo.MutationFunction<CreateCampaignMutation, CreateCampaignMutationVariables>;

/**
 * __useCreateCampaignMutation__
 *
 * To run a mutation, you first call `useCreateCampaignMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateCampaignMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createCampaignMutation, { data, loading, error }] = useCreateCampaignMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateCampaignMutation(baseOptions?: Apollo.MutationHookOptions<CreateCampaignMutation, CreateCampaignMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateCampaignMutation, CreateCampaignMutationVariables>(CreateCampaignDocument, options);
      }
export type CreateCampaignMutationHookResult = ReturnType<typeof useCreateCampaignMutation>;
export type CreateCampaignMutationResult = Apollo.MutationResult<CreateCampaignMutation>;
export type CreateCampaignMutationOptions = Apollo.BaseMutationOptions<CreateCampaignMutation, CreateCampaignMutationVariables>;
export const BeginCampaignFormationDocument = gql`
    mutation BeginCampaignFormation($input: BeginCampaignFormationInput!) {
  beginCampaignFormation(input: $input) {
    id
    status
  }
}
    `;
export type BeginCampaignFormationMutationFn = Apollo.MutationFunction<BeginCampaignFormationMutation, BeginCampaignFormationMutationVariables>;

/**
 * __useBeginCampaignFormationMutation__
 *
 * To run a mutation, you first call `useBeginCampaignFormationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useBeginCampaignFormationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [beginCampaignFormationMutation, { data, loading, error }] = useBeginCampaignFormationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useBeginCampaignFormationMutation(baseOptions?: Apollo.MutationHookOptions<BeginCampaignFormationMutation, BeginCampaignFormationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<BeginCampaignFormationMutation, BeginCampaignFormationMutationVariables>(BeginCampaignFormationDocument, options);
      }
export type BeginCampaignFormationMutationHookResult = ReturnType<typeof useBeginCampaignFormationMutation>;
export type BeginCampaignFormationMutationResult = Apollo.MutationResult<BeginCampaignFormationMutation>;
export type BeginCampaignFormationMutationOptions = Apollo.BaseMutationOptions<BeginCampaignFormationMutation, BeginCampaignFormationMutationVariables>;
export const AddCharacterToCampaignDocument = gql`
    mutation AddCharacterToCampaign($input: AddCharacterToCampaignInput!) {
  addCharacterToCampaign(input: $input) {
    id
    status
  }
}
    `;
export type AddCharacterToCampaignMutationFn = Apollo.MutationFunction<AddCharacterToCampaignMutation, AddCharacterToCampaignMutationVariables>;

/**
 * __useAddCharacterToCampaignMutation__
 *
 * To run a mutation, you first call `useAddCharacterToCampaignMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddCharacterToCampaignMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addCharacterToCampaignMutation, { data, loading, error }] = useAddCharacterToCampaignMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddCharacterToCampaignMutation(baseOptions?: Apollo.MutationHookOptions<AddCharacterToCampaignMutation, AddCharacterToCampaignMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddCharacterToCampaignMutation, AddCharacterToCampaignMutationVariables>(AddCharacterToCampaignDocument, options);
      }
export type AddCharacterToCampaignMutationHookResult = ReturnType<typeof useAddCharacterToCampaignMutation>;
export type AddCharacterToCampaignMutationResult = Apollo.MutationResult<AddCharacterToCampaignMutation>;
export type AddCharacterToCampaignMutationOptions = Apollo.BaseMutationOptions<AddCharacterToCampaignMutation, AddCharacterToCampaignMutationVariables>;
export const CreateNpcDocument = gql`
    mutation CreateNPC($input: CreateNPCInput!) {
  createNPC(input: $input) {
    id
    status
  }
}
    `;
export type CreateNpcMutationFn = Apollo.MutationFunction<CreateNpcMutation, CreateNpcMutationVariables>;

/**
 * __useCreateNpcMutation__
 *
 * To run a mutation, you first call `useCreateNpcMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateNpcMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createNpcMutation, { data, loading, error }] = useCreateNpcMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateNpcMutation(baseOptions?: Apollo.MutationHookOptions<CreateNpcMutation, CreateNpcMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateNpcMutation, CreateNpcMutationVariables>(CreateNpcDocument, options);
      }
export type CreateNpcMutationHookResult = ReturnType<typeof useCreateNpcMutation>;
export type CreateNpcMutationResult = Apollo.MutationResult<CreateNpcMutation>;
export type CreateNpcMutationOptions = Apollo.BaseMutationOptions<CreateNpcMutation, CreateNpcMutationVariables>;
export const UpdateNpcContentDocument = gql`
    mutation UpdateNPCContent($input: UpdateNPCContentInput!) {
  updateNPCContent(input: $input) {
    id
    status
  }
}
    `;
export type UpdateNpcContentMutationFn = Apollo.MutationFunction<UpdateNpcContentMutation, UpdateNpcContentMutationVariables>;

/**
 * __useUpdateNpcContentMutation__
 *
 * To run a mutation, you first call `useUpdateNpcContentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateNpcContentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateNpcContentMutation, { data, loading, error }] = useUpdateNpcContentMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateNpcContentMutation(baseOptions?: Apollo.MutationHookOptions<UpdateNpcContentMutation, UpdateNpcContentMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateNpcContentMutation, UpdateNpcContentMutationVariables>(UpdateNpcContentDocument, options);
      }
export type UpdateNpcContentMutationHookResult = ReturnType<typeof useUpdateNpcContentMutation>;
export type UpdateNpcContentMutationResult = Apollo.MutationResult<UpdateNpcContentMutation>;
export type UpdateNpcContentMutationOptions = Apollo.BaseMutationOptions<UpdateNpcContentMutation, UpdateNpcContentMutationVariables>;
export const ActivateNpcDocument = gql`
    mutation ActivateNPC($input: NPCLifecycleInput!) {
  activateNPC(input: $input) {
    id
    status
  }
}
    `;
export type ActivateNpcMutationFn = Apollo.MutationFunction<ActivateNpcMutation, ActivateNpcMutationVariables>;

/**
 * __useActivateNpcMutation__
 *
 * To run a mutation, you first call `useActivateNpcMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useActivateNpcMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [activateNpcMutation, { data, loading, error }] = useActivateNpcMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useActivateNpcMutation(baseOptions?: Apollo.MutationHookOptions<ActivateNpcMutation, ActivateNpcMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ActivateNpcMutation, ActivateNpcMutationVariables>(ActivateNpcDocument, options);
      }
export type ActivateNpcMutationHookResult = ReturnType<typeof useActivateNpcMutation>;
export type ActivateNpcMutationResult = Apollo.MutationResult<ActivateNpcMutation>;
export type ActivateNpcMutationOptions = Apollo.BaseMutationOptions<ActivateNpcMutation, ActivateNpcMutationVariables>;
export const ArchiveNpcDocument = gql`
    mutation ArchiveNPC($input: NPCLifecycleInput!) {
  archiveNPC(input: $input) {
    id
    status
  }
}
    `;
export type ArchiveNpcMutationFn = Apollo.MutationFunction<ArchiveNpcMutation, ArchiveNpcMutationVariables>;

/**
 * __useArchiveNpcMutation__
 *
 * To run a mutation, you first call `useArchiveNpcMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useArchiveNpcMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [archiveNpcMutation, { data, loading, error }] = useArchiveNpcMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useArchiveNpcMutation(baseOptions?: Apollo.MutationHookOptions<ArchiveNpcMutation, ArchiveNpcMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ArchiveNpcMutation, ArchiveNpcMutationVariables>(ArchiveNpcDocument, options);
      }
export type ArchiveNpcMutationHookResult = ReturnType<typeof useArchiveNpcMutation>;
export type ArchiveNpcMutationResult = Apollo.MutationResult<ArchiveNpcMutation>;
export type ArchiveNpcMutationOptions = Apollo.BaseMutationOptions<ArchiveNpcMutation, ArchiveNpcMutationVariables>;
export const CreatePlayerCharacterDocument = gql`
    mutation CreatePlayerCharacter($input: CreatePlayerCharacterInput!) {
  createPlayerCharacter(input: $input) {
    id
    status
  }
}
    `;
export type CreatePlayerCharacterMutationFn = Apollo.MutationFunction<CreatePlayerCharacterMutation, CreatePlayerCharacterMutationVariables>;

/**
 * __useCreatePlayerCharacterMutation__
 *
 * To run a mutation, you first call `useCreatePlayerCharacterMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreatePlayerCharacterMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createPlayerCharacterMutation, { data, loading, error }] = useCreatePlayerCharacterMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreatePlayerCharacterMutation(baseOptions?: Apollo.MutationHookOptions<CreatePlayerCharacterMutation, CreatePlayerCharacterMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreatePlayerCharacterMutation, CreatePlayerCharacterMutationVariables>(CreatePlayerCharacterDocument, options);
      }
export type CreatePlayerCharacterMutationHookResult = ReturnType<typeof useCreatePlayerCharacterMutation>;
export type CreatePlayerCharacterMutationResult = Apollo.MutationResult<CreatePlayerCharacterMutation>;
export type CreatePlayerCharacterMutationOptions = Apollo.BaseMutationOptions<CreatePlayerCharacterMutation, CreatePlayerCharacterMutationVariables>;
export const CreateMacGuffinDocument = gql`
    mutation CreateMacGuffin($input: CreateMacGuffinInput!) {
  createMacGuffin(input: $input) {
    id
    status
  }
}
    `;
export type CreateMacGuffinMutationFn = Apollo.MutationFunction<CreateMacGuffinMutation, CreateMacGuffinMutationVariables>;

/**
 * __useCreateMacGuffinMutation__
 *
 * To run a mutation, you first call `useCreateMacGuffinMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateMacGuffinMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createMacGuffinMutation, { data, loading, error }] = useCreateMacGuffinMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateMacGuffinMutation(baseOptions?: Apollo.MutationHookOptions<CreateMacGuffinMutation, CreateMacGuffinMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateMacGuffinMutation, CreateMacGuffinMutationVariables>(CreateMacGuffinDocument, options);
      }
export type CreateMacGuffinMutationHookResult = ReturnType<typeof useCreateMacGuffinMutation>;
export type CreateMacGuffinMutationResult = Apollo.MutationResult<CreateMacGuffinMutation>;
export type CreateMacGuffinMutationOptions = Apollo.BaseMutationOptions<CreateMacGuffinMutation, CreateMacGuffinMutationVariables>;
export const CreateFactionDocument = gql`
    mutation CreateFaction($input: CreateFactionInput!) {
  createFaction(input: $input) {
    id
    status
  }
}
    `;
export type CreateFactionMutationFn = Apollo.MutationFunction<CreateFactionMutation, CreateFactionMutationVariables>;

/**
 * __useCreateFactionMutation__
 *
 * To run a mutation, you first call `useCreateFactionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateFactionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createFactionMutation, { data, loading, error }] = useCreateFactionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateFactionMutation(baseOptions?: Apollo.MutationHookOptions<CreateFactionMutation, CreateFactionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateFactionMutation, CreateFactionMutationVariables>(CreateFactionDocument, options);
      }
export type CreateFactionMutationHookResult = ReturnType<typeof useCreateFactionMutation>;
export type CreateFactionMutationResult = Apollo.MutationResult<CreateFactionMutation>;
export type CreateFactionMutationOptions = Apollo.BaseMutationOptions<CreateFactionMutation, CreateFactionMutationVariables>;
export const ActivateFactionDocument = gql`
    mutation ActivateFaction($input: FactionLifecycleInput!) {
  activateFaction(input: $input) {
    id
    status
  }
}
    `;
export type ActivateFactionMutationFn = Apollo.MutationFunction<ActivateFactionMutation, ActivateFactionMutationVariables>;

/**
 * __useActivateFactionMutation__
 *
 * To run a mutation, you first call `useActivateFactionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useActivateFactionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [activateFactionMutation, { data, loading, error }] = useActivateFactionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useActivateFactionMutation(baseOptions?: Apollo.MutationHookOptions<ActivateFactionMutation, ActivateFactionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ActivateFactionMutation, ActivateFactionMutationVariables>(ActivateFactionDocument, options);
      }
export type ActivateFactionMutationHookResult = ReturnType<typeof useActivateFactionMutation>;
export type ActivateFactionMutationResult = Apollo.MutationResult<ActivateFactionMutation>;
export type ActivateFactionMutationOptions = Apollo.BaseMutationOptions<ActivateFactionMutation, ActivateFactionMutationVariables>;
export const AddFactionMemberDocument = gql`
    mutation AddFactionMember($input: AddFactionMemberInput!) {
  addFactionMember(input: $input) {
    id
    status
  }
}
    `;
export type AddFactionMemberMutationFn = Apollo.MutationFunction<AddFactionMemberMutation, AddFactionMemberMutationVariables>;

/**
 * __useAddFactionMemberMutation__
 *
 * To run a mutation, you first call `useAddFactionMemberMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddFactionMemberMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addFactionMemberMutation, { data, loading, error }] = useAddFactionMemberMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddFactionMemberMutation(baseOptions?: Apollo.MutationHookOptions<AddFactionMemberMutation, AddFactionMemberMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddFactionMemberMutation, AddFactionMemberMutationVariables>(AddFactionMemberDocument, options);
      }
export type AddFactionMemberMutationHookResult = ReturnType<typeof useAddFactionMemberMutation>;
export type AddFactionMemberMutationResult = Apollo.MutationResult<AddFactionMemberMutation>;
export type AddFactionMemberMutationOptions = Apollo.BaseMutationOptions<AddFactionMemberMutation, AddFactionMemberMutationVariables>;
export const AddStandingLevelDocument = gql`
    mutation AddStandingLevel($input: AddStandingLevelInput!) {
  addStandingLevel(input: $input) {
    id
    status
  }
}
    `;
export type AddStandingLevelMutationFn = Apollo.MutationFunction<AddStandingLevelMutation, AddStandingLevelMutationVariables>;

/**
 * __useAddStandingLevelMutation__
 *
 * To run a mutation, you first call `useAddStandingLevelMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddStandingLevelMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addStandingLevelMutation, { data, loading, error }] = useAddStandingLevelMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddStandingLevelMutation(baseOptions?: Apollo.MutationHookOptions<AddStandingLevelMutation, AddStandingLevelMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddStandingLevelMutation, AddStandingLevelMutationVariables>(AddStandingLevelDocument, options);
      }
export type AddStandingLevelMutationHookResult = ReturnType<typeof useAddStandingLevelMutation>;
export type AddStandingLevelMutationResult = Apollo.MutationResult<AddStandingLevelMutation>;
export type AddStandingLevelMutationOptions = Apollo.BaseMutationOptions<AddStandingLevelMutation, AddStandingLevelMutationVariables>;
export const DeclareAllyDocument = gql`
    mutation DeclareAlly($input: FactionRelationshipInput!) {
  declareAlly(input: $input) {
    id
    status
  }
}
    `;
export type DeclareAllyMutationFn = Apollo.MutationFunction<DeclareAllyMutation, DeclareAllyMutationVariables>;

/**
 * __useDeclareAllyMutation__
 *
 * To run a mutation, you first call `useDeclareAllyMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeclareAllyMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [declareAllyMutation, { data, loading, error }] = useDeclareAllyMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useDeclareAllyMutation(baseOptions?: Apollo.MutationHookOptions<DeclareAllyMutation, DeclareAllyMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeclareAllyMutation, DeclareAllyMutationVariables>(DeclareAllyDocument, options);
      }
export type DeclareAllyMutationHookResult = ReturnType<typeof useDeclareAllyMutation>;
export type DeclareAllyMutationResult = Apollo.MutationResult<DeclareAllyMutation>;
export type DeclareAllyMutationOptions = Apollo.BaseMutationOptions<DeclareAllyMutation, DeclareAllyMutationVariables>;
export const DeclareWarDocument = gql`
    mutation DeclareWar($input: FactionRelationshipInput!) {
  declareWar(input: $input) {
    id
    status
  }
}
    `;
export type DeclareWarMutationFn = Apollo.MutationFunction<DeclareWarMutation, DeclareWarMutationVariables>;

/**
 * __useDeclareWarMutation__
 *
 * To run a mutation, you first call `useDeclareWarMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeclareWarMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [declareWarMutation, { data, loading, error }] = useDeclareWarMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useDeclareWarMutation(baseOptions?: Apollo.MutationHookOptions<DeclareWarMutation, DeclareWarMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeclareWarMutation, DeclareWarMutationVariables>(DeclareWarDocument, options);
      }
export type DeclareWarMutationHookResult = ReturnType<typeof useDeclareWarMutation>;
export type DeclareWarMutationResult = Apollo.MutationResult<DeclareWarMutation>;
export type DeclareWarMutationOptions = Apollo.BaseMutationOptions<DeclareWarMutation, DeclareWarMutationVariables>;
export const ArchiveFactionDocument = gql`
    mutation ArchiveFaction($input: FactionLifecycleInput!) {
  archiveFaction(input: $input) {
    id
    status
  }
}
    `;
export type ArchiveFactionMutationFn = Apollo.MutationFunction<ArchiveFactionMutation, ArchiveFactionMutationVariables>;

/**
 * __useArchiveFactionMutation__
 *
 * To run a mutation, you first call `useArchiveFactionMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useArchiveFactionMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [archiveFactionMutation, { data, loading, error }] = useArchiveFactionMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useArchiveFactionMutation(baseOptions?: Apollo.MutationHookOptions<ArchiveFactionMutation, ArchiveFactionMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ArchiveFactionMutation, ArchiveFactionMutationVariables>(ArchiveFactionDocument, options);
      }
export type ArchiveFactionMutationHookResult = ReturnType<typeof useArchiveFactionMutation>;
export type ArchiveFactionMutationResult = Apollo.MutationResult<ArchiveFactionMutation>;
export type ArchiveFactionMutationOptions = Apollo.BaseMutationOptions<ArchiveFactionMutation, ArchiveFactionMutationVariables>;
export const CreateGameDocument = gql`
    mutation CreateGame($input: CreateGameInput!) {
  createGame(input: $input) {
    id
    status
  }
}
    `;
export type CreateGameMutationFn = Apollo.MutationFunction<CreateGameMutation, CreateGameMutationVariables>;

/**
 * __useCreateGameMutation__
 *
 * To run a mutation, you first call `useCreateGameMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateGameMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createGameMutation, { data, loading, error }] = useCreateGameMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateGameMutation(baseOptions?: Apollo.MutationHookOptions<CreateGameMutation, CreateGameMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateGameMutation, CreateGameMutationVariables>(CreateGameDocument, options);
      }
export type CreateGameMutationHookResult = ReturnType<typeof useCreateGameMutation>;
export type CreateGameMutationResult = Apollo.MutationResult<CreateGameMutation>;
export type CreateGameMutationOptions = Apollo.BaseMutationOptions<CreateGameMutation, CreateGameMutationVariables>;
export const CreateLocationDocument = gql`
    mutation CreateLocation($input: CreateLocationInput!) {
  createLocation(input: $input) {
    id
    status
  }
}
    `;
export type CreateLocationMutationFn = Apollo.MutationFunction<CreateLocationMutation, CreateLocationMutationVariables>;

/**
 * __useCreateLocationMutation__
 *
 * To run a mutation, you first call `useCreateLocationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateLocationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createLocationMutation, { data, loading, error }] = useCreateLocationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateLocationMutation(baseOptions?: Apollo.MutationHookOptions<CreateLocationMutation, CreateLocationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateLocationMutation, CreateLocationMutationVariables>(CreateLocationDocument, options);
      }
export type CreateLocationMutationHookResult = ReturnType<typeof useCreateLocationMutation>;
export type CreateLocationMutationResult = Apollo.MutationResult<CreateLocationMutation>;
export type CreateLocationMutationOptions = Apollo.BaseMutationOptions<CreateLocationMutation, CreateLocationMutationVariables>;
export const ActivateLocationDocument = gql`
    mutation ActivateLocation($input: LocationLifecycleInput!) {
  activateLocation(input: $input) {
    id
    status
  }
}
    `;
export type ActivateLocationMutationFn = Apollo.MutationFunction<ActivateLocationMutation, ActivateLocationMutationVariables>;

/**
 * __useActivateLocationMutation__
 *
 * To run a mutation, you first call `useActivateLocationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useActivateLocationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [activateLocationMutation, { data, loading, error }] = useActivateLocationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useActivateLocationMutation(baseOptions?: Apollo.MutationHookOptions<ActivateLocationMutation, ActivateLocationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ActivateLocationMutation, ActivateLocationMutationVariables>(ActivateLocationDocument, options);
      }
export type ActivateLocationMutationHookResult = ReturnType<typeof useActivateLocationMutation>;
export type ActivateLocationMutationResult = Apollo.MutationResult<ActivateLocationMutation>;
export type ActivateLocationMutationOptions = Apollo.BaseMutationOptions<ActivateLocationMutation, ActivateLocationMutationVariables>;
export const AddSceneDocument = gql`
    mutation AddScene($input: AddSceneInput!) {
  addScene(input: $input) {
    id
    status
  }
}
    `;
export type AddSceneMutationFn = Apollo.MutationFunction<AddSceneMutation, AddSceneMutationVariables>;

/**
 * __useAddSceneMutation__
 *
 * To run a mutation, you first call `useAddSceneMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddSceneMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addSceneMutation, { data, loading, error }] = useAddSceneMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddSceneMutation(baseOptions?: Apollo.MutationHookOptions<AddSceneMutation, AddSceneMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddSceneMutation, AddSceneMutationVariables>(AddSceneDocument, options);
      }
export type AddSceneMutationHookResult = ReturnType<typeof useAddSceneMutation>;
export type AddSceneMutationResult = Apollo.MutationResult<AddSceneMutation>;
export type AddSceneMutationOptions = Apollo.BaseMutationOptions<AddSceneMutation, AddSceneMutationVariables>;
export const ConnectLocationsDocument = gql`
    mutation ConnectLocations($input: ConnectLocationsInput!) {
  connectLocations(input: $input) {
    id
    status
  }
}
    `;
export type ConnectLocationsMutationFn = Apollo.MutationFunction<ConnectLocationsMutation, ConnectLocationsMutationVariables>;

/**
 * __useConnectLocationsMutation__
 *
 * To run a mutation, you first call `useConnectLocationsMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useConnectLocationsMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [connectLocationsMutation, { data, loading, error }] = useConnectLocationsMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useConnectLocationsMutation(baseOptions?: Apollo.MutationHookOptions<ConnectLocationsMutation, ConnectLocationsMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ConnectLocationsMutation, ConnectLocationsMutationVariables>(ConnectLocationsDocument, options);
      }
export type ConnectLocationsMutationHookResult = ReturnType<typeof useConnectLocationsMutation>;
export type ConnectLocationsMutationResult = Apollo.MutationResult<ConnectLocationsMutation>;
export type ConnectLocationsMutationOptions = Apollo.BaseMutationOptions<ConnectLocationsMutation, ConnectLocationsMutationVariables>;
export const ArchiveLocationDocument = gql`
    mutation ArchiveLocation($input: LocationLifecycleInput!) {
  archiveLocation(input: $input) {
    id
    status
  }
}
    `;
export type ArchiveLocationMutationFn = Apollo.MutationFunction<ArchiveLocationMutation, ArchiveLocationMutationVariables>;

/**
 * __useArchiveLocationMutation__
 *
 * To run a mutation, you first call `useArchiveLocationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useArchiveLocationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [archiveLocationMutation, { data, loading, error }] = useArchiveLocationMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useArchiveLocationMutation(baseOptions?: Apollo.MutationHookOptions<ArchiveLocationMutation, ArchiveLocationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ArchiveLocationMutation, ArchiveLocationMutationVariables>(ArchiveLocationDocument, options);
      }
export type ArchiveLocationMutationHookResult = ReturnType<typeof useArchiveLocationMutation>;
export type ArchiveLocationMutationResult = Apollo.MutationResult<ArchiveLocationMutation>;
export type ArchiveLocationMutationOptions = Apollo.BaseMutationOptions<ArchiveLocationMutation, ArchiveLocationMutationVariables>;
export const CreateMasterBeatDocument = gql`
    mutation CreateMasterBeat($input: CreateMasterBeatInput!) {
  createMasterBeat(input: $input) {
    id
    status
  }
}
    `;
export type CreateMasterBeatMutationFn = Apollo.MutationFunction<CreateMasterBeatMutation, CreateMasterBeatMutationVariables>;

/**
 * __useCreateMasterBeatMutation__
 *
 * To run a mutation, you first call `useCreateMasterBeatMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateMasterBeatMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createMasterBeatMutation, { data, loading, error }] = useCreateMasterBeatMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useCreateMasterBeatMutation(baseOptions?: Apollo.MutationHookOptions<CreateMasterBeatMutation, CreateMasterBeatMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateMasterBeatMutation, CreateMasterBeatMutationVariables>(CreateMasterBeatDocument, options);
      }
export type CreateMasterBeatMutationHookResult = ReturnType<typeof useCreateMasterBeatMutation>;
export type CreateMasterBeatMutationResult = Apollo.MutationResult<CreateMasterBeatMutation>;
export type CreateMasterBeatMutationOptions = Apollo.BaseMutationOptions<CreateMasterBeatMutation, CreateMasterBeatMutationVariables>;
export const UpdateBeatContentDocument = gql`
    mutation UpdateBeatContent($input: UpdateBeatContentInput!) {
  updateBeatContent(input: $input) {
    id
    status
  }
}
    `;
export type UpdateBeatContentMutationFn = Apollo.MutationFunction<UpdateBeatContentMutation, UpdateBeatContentMutationVariables>;

/**
 * __useUpdateBeatContentMutation__
 *
 * To run a mutation, you first call `useUpdateBeatContentMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateBeatContentMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateBeatContentMutation, { data, loading, error }] = useUpdateBeatContentMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useUpdateBeatContentMutation(baseOptions?: Apollo.MutationHookOptions<UpdateBeatContentMutation, UpdateBeatContentMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateBeatContentMutation, UpdateBeatContentMutationVariables>(UpdateBeatContentDocument, options);
      }
export type UpdateBeatContentMutationHookResult = ReturnType<typeof useUpdateBeatContentMutation>;
export type UpdateBeatContentMutationResult = Apollo.MutationResult<UpdateBeatContentMutation>;
export type UpdateBeatContentMutationOptions = Apollo.BaseMutationOptions<UpdateBeatContentMutation, UpdateBeatContentMutationVariables>;
export const AddBeatPrerequisiteDocument = gql`
    mutation AddBeatPrerequisite($input: AddBeatPrerequisiteInput!) {
  addBeatPrerequisite(input: $input) {
    id
    status
  }
}
    `;
export type AddBeatPrerequisiteMutationFn = Apollo.MutationFunction<AddBeatPrerequisiteMutation, AddBeatPrerequisiteMutationVariables>;

/**
 * __useAddBeatPrerequisiteMutation__
 *
 * To run a mutation, you first call `useAddBeatPrerequisiteMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useAddBeatPrerequisiteMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [addBeatPrerequisiteMutation, { data, loading, error }] = useAddBeatPrerequisiteMutation({
 *   variables: {
 *      input: // value for 'input'
 *   },
 * });
 */
export function useAddBeatPrerequisiteMutation(baseOptions?: Apollo.MutationHookOptions<AddBeatPrerequisiteMutation, AddBeatPrerequisiteMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<AddBeatPrerequisiteMutation, AddBeatPrerequisiteMutationVariables>(AddBeatPrerequisiteDocument, options);
      }
export type AddBeatPrerequisiteMutationHookResult = ReturnType<typeof useAddBeatPrerequisiteMutation>;
export type AddBeatPrerequisiteMutationResult = Apollo.MutationResult<AddBeatPrerequisiteMutation>;
export type AddBeatPrerequisiteMutationOptions = Apollo.BaseMutationOptions<AddBeatPrerequisiteMutation, AddBeatPrerequisiteMutationVariables>;
export const CampaignsDocument = gql`
    query Campaigns($gameId: ID!) {
  campaigns(gameId: $gameId) {
    id
    name
    status
    characterCount
    sessionCount
  }
}
    `;

/**
 * __useCampaignsQuery__
 *
 * To run a query within a React component, call `useCampaignsQuery` and pass it any options that fit your needs.
 * When your component renders, `useCampaignsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCampaignsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useCampaignsQuery(baseOptions: Apollo.QueryHookOptions<CampaignsQuery, CampaignsQueryVariables> & ({ variables: CampaignsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CampaignsQuery, CampaignsQueryVariables>(CampaignsDocument, options);
      }
export function useCampaignsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CampaignsQuery, CampaignsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CampaignsQuery, CampaignsQueryVariables>(CampaignsDocument, options);
        }
// @ts-ignore
export function useCampaignsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CampaignsQuery, CampaignsQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignsQuery, CampaignsQueryVariables>;
export function useCampaignsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignsQuery, CampaignsQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignsQuery | undefined, CampaignsQueryVariables>;
export function useCampaignsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignsQuery, CampaignsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CampaignsQuery, CampaignsQueryVariables>(CampaignsDocument, options);
        }
export type CampaignsQueryHookResult = ReturnType<typeof useCampaignsQuery>;
export type CampaignsLazyQueryHookResult = ReturnType<typeof useCampaignsLazyQuery>;
export type CampaignsSuspenseQueryHookResult = ReturnType<typeof useCampaignsSuspenseQuery>;
export type CampaignsQueryResult = Apollo.QueryResult<CampaignsQuery, CampaignsQueryVariables>;
export const CampaignDocument = gql`
    query Campaign($id: ID!, $gameId: ID!) {
  campaign(id: $id, gameId: $gameId) {
    id
    name
    status
    sessionCount
    characters {
      id
      name
      status
      ownerPlayerId
    }
  }
}
    `;

/**
 * __useCampaignQuery__
 *
 * To run a query within a React component, call `useCampaignQuery` and pass it any options that fit your needs.
 * When your component renders, `useCampaignQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCampaignQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useCampaignQuery(baseOptions: Apollo.QueryHookOptions<CampaignQuery, CampaignQueryVariables> & ({ variables: CampaignQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CampaignQuery, CampaignQueryVariables>(CampaignDocument, options);
      }
export function useCampaignLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CampaignQuery, CampaignQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CampaignQuery, CampaignQueryVariables>(CampaignDocument, options);
        }
// @ts-ignore
export function useCampaignSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CampaignQuery, CampaignQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignQuery, CampaignQueryVariables>;
export function useCampaignSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignQuery, CampaignQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignQuery | undefined, CampaignQueryVariables>;
export function useCampaignSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignQuery, CampaignQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CampaignQuery, CampaignQueryVariables>(CampaignDocument, options);
        }
export type CampaignQueryHookResult = ReturnType<typeof useCampaignQuery>;
export type CampaignLazyQueryHookResult = ReturnType<typeof useCampaignLazyQuery>;
export type CampaignSuspenseQueryHookResult = ReturnType<typeof useCampaignSuspenseQuery>;
export type CampaignQueryResult = Apollo.QueryResult<CampaignQuery, CampaignQueryVariables>;
export const CampaignSessionStateDocument = gql`
    query CampaignSessionState($id: ID!, $gameId: ID!) {
  campaignSessionState(id: $id, gameId: $gameId) {
    id
    name
    status
    sessionId
    currentLocation {
      id
      name
      locationType
      connections {
        toLocation {
          id
          name
          locationType
          status
          parentId
          sceneCount
          childCount
        }
        direction
      }
    }
    characters {
      id
      name
      status
      ownerPlayerId
    }
  }
}
    `;

/**
 * __useCampaignSessionStateQuery__
 *
 * To run a query within a React component, call `useCampaignSessionStateQuery` and pass it any options that fit your needs.
 * When your component renders, `useCampaignSessionStateQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCampaignSessionStateQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useCampaignSessionStateQuery(baseOptions: Apollo.QueryHookOptions<CampaignSessionStateQuery, CampaignSessionStateQueryVariables> & ({ variables: CampaignSessionStateQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>(CampaignSessionStateDocument, options);
      }
export function useCampaignSessionStateLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>(CampaignSessionStateDocument, options);
        }
// @ts-ignore
export function useCampaignSessionStateSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>;
export function useCampaignSessionStateSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignSessionStateQuery | undefined, CampaignSessionStateQueryVariables>;
export function useCampaignSessionStateSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>(CampaignSessionStateDocument, options);
        }
export type CampaignSessionStateQueryHookResult = ReturnType<typeof useCampaignSessionStateQuery>;
export type CampaignSessionStateLazyQueryHookResult = ReturnType<typeof useCampaignSessionStateLazyQuery>;
export type CampaignSessionStateSuspenseQueryHookResult = ReturnType<typeof useCampaignSessionStateSuspenseQuery>;
export type CampaignSessionStateQueryResult = Apollo.QueryResult<CampaignSessionStateQuery, CampaignSessionStateQueryVariables>;
export const SessionLauncherDocument = gql`
    query SessionLauncher($campaignId: ID!, $gameId: ID!) {
  sessionLauncher(campaignId: $campaignId, gameId: $gameId) {
    campaignId
    campaignName
    status
    sessionCount
    characters {
      id
      name
      status
      ownerPlayerId
    }
  }
}
    `;

/**
 * __useSessionLauncherQuery__
 *
 * To run a query within a React component, call `useSessionLauncherQuery` and pass it any options that fit your needs.
 * When your component renders, `useSessionLauncherQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSessionLauncherQuery({
 *   variables: {
 *      campaignId: // value for 'campaignId'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useSessionLauncherQuery(baseOptions: Apollo.QueryHookOptions<SessionLauncherQuery, SessionLauncherQueryVariables> & ({ variables: SessionLauncherQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SessionLauncherQuery, SessionLauncherQueryVariables>(SessionLauncherDocument, options);
      }
export function useSessionLauncherLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SessionLauncherQuery, SessionLauncherQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SessionLauncherQuery, SessionLauncherQueryVariables>(SessionLauncherDocument, options);
        }
// @ts-ignore
export function useSessionLauncherSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<SessionLauncherQuery, SessionLauncherQueryVariables>): Apollo.UseSuspenseQueryResult<SessionLauncherQuery, SessionLauncherQueryVariables>;
export function useSessionLauncherSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SessionLauncherQuery, SessionLauncherQueryVariables>): Apollo.UseSuspenseQueryResult<SessionLauncherQuery | undefined, SessionLauncherQueryVariables>;
export function useSessionLauncherSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SessionLauncherQuery, SessionLauncherQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<SessionLauncherQuery, SessionLauncherQueryVariables>(SessionLauncherDocument, options);
        }
export type SessionLauncherQueryHookResult = ReturnType<typeof useSessionLauncherQuery>;
export type SessionLauncherLazyQueryHookResult = ReturnType<typeof useSessionLauncherLazyQuery>;
export type SessionLauncherSuspenseQueryHookResult = ReturnType<typeof useSessionLauncherSuspenseQuery>;
export type SessionLauncherQueryResult = Apollo.QueryResult<SessionLauncherQuery, SessionLauncherQueryVariables>;
export const SessionSummaryDocument = gql`
    query SessionSummary($sessionId: ID!, $gameId: ID!) {
  sessionSummary(sessionId: $sessionId, gameId: $gameId) {
    sessionId
    beatsDiscoveredCount
    entitiesRevealedCount
    finalLocation {
      id
      name
    }
  }
}
    `;

/**
 * __useSessionSummaryQuery__
 *
 * To run a query within a React component, call `useSessionSummaryQuery` and pass it any options that fit your needs.
 * When your component renders, `useSessionSummaryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSessionSummaryQuery({
 *   variables: {
 *      sessionId: // value for 'sessionId'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useSessionSummaryQuery(baseOptions: Apollo.QueryHookOptions<SessionSummaryQuery, SessionSummaryQueryVariables> & ({ variables: SessionSummaryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SessionSummaryQuery, SessionSummaryQueryVariables>(SessionSummaryDocument, options);
      }
export function useSessionSummaryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SessionSummaryQuery, SessionSummaryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SessionSummaryQuery, SessionSummaryQueryVariables>(SessionSummaryDocument, options);
        }
// @ts-ignore
export function useSessionSummarySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<SessionSummaryQuery, SessionSummaryQueryVariables>): Apollo.UseSuspenseQueryResult<SessionSummaryQuery, SessionSummaryQueryVariables>;
export function useSessionSummarySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SessionSummaryQuery, SessionSummaryQueryVariables>): Apollo.UseSuspenseQueryResult<SessionSummaryQuery | undefined, SessionSummaryQueryVariables>;
export function useSessionSummarySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<SessionSummaryQuery, SessionSummaryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<SessionSummaryQuery, SessionSummaryQueryVariables>(SessionSummaryDocument, options);
        }
export type SessionSummaryQueryHookResult = ReturnType<typeof useSessionSummaryQuery>;
export type SessionSummaryLazyQueryHookResult = ReturnType<typeof useSessionSummaryLazyQuery>;
export type SessionSummarySuspenseQueryHookResult = ReturnType<typeof useSessionSummarySuspenseQuery>;
export type SessionSummaryQueryResult = Apollo.QueryResult<SessionSummaryQuery, SessionSummaryQueryVariables>;
export const NpcsDocument = gql`
    query Npcs($gameId: ID!, $excludeArchived: Boolean) {
  npcs(gameId: $gameId, excludeArchived: $excludeArchived) {
    id
    name
    status
    playerVisible
  }
}
    `;

/**
 * __useNpcsQuery__
 *
 * To run a query within a React component, call `useNpcsQuery` and pass it any options that fit your needs.
 * When your component renders, `useNpcsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useNpcsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *      excludeArchived: // value for 'excludeArchived'
 *   },
 * });
 */
export function useNpcsQuery(baseOptions: Apollo.QueryHookOptions<NpcsQuery, NpcsQueryVariables> & ({ variables: NpcsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<NpcsQuery, NpcsQueryVariables>(NpcsDocument, options);
      }
export function useNpcsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<NpcsQuery, NpcsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<NpcsQuery, NpcsQueryVariables>(NpcsDocument, options);
        }
// @ts-ignore
export function useNpcsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<NpcsQuery, NpcsQueryVariables>): Apollo.UseSuspenseQueryResult<NpcsQuery, NpcsQueryVariables>;
export function useNpcsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<NpcsQuery, NpcsQueryVariables>): Apollo.UseSuspenseQueryResult<NpcsQuery | undefined, NpcsQueryVariables>;
export function useNpcsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<NpcsQuery, NpcsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<NpcsQuery, NpcsQueryVariables>(NpcsDocument, options);
        }
export type NpcsQueryHookResult = ReturnType<typeof useNpcsQuery>;
export type NpcsLazyQueryHookResult = ReturnType<typeof useNpcsLazyQuery>;
export type NpcsSuspenseQueryHookResult = ReturnType<typeof useNpcsSuspenseQuery>;
export type NpcsQueryResult = Apollo.QueryResult<NpcsQuery, NpcsQueryVariables>;
export const NpcDocument = gql`
    query Npc($id: ID!, $gameId: ID!) {
  npc(id: $id, gameId: $gameId) {
    id
    name
    status
    playerVisible
    description
    playerDescription
    factionMemberships {
      faction {
        id
        name
      }
      rank
    }
  }
}
    `;

/**
 * __useNpcQuery__
 *
 * To run a query within a React component, call `useNpcQuery` and pass it any options that fit your needs.
 * When your component renders, `useNpcQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useNpcQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useNpcQuery(baseOptions: Apollo.QueryHookOptions<NpcQuery, NpcQueryVariables> & ({ variables: NpcQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<NpcQuery, NpcQueryVariables>(NpcDocument, options);
      }
export function useNpcLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<NpcQuery, NpcQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<NpcQuery, NpcQueryVariables>(NpcDocument, options);
        }
// @ts-ignore
export function useNpcSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<NpcQuery, NpcQueryVariables>): Apollo.UseSuspenseQueryResult<NpcQuery, NpcQueryVariables>;
export function useNpcSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<NpcQuery, NpcQueryVariables>): Apollo.UseSuspenseQueryResult<NpcQuery | undefined, NpcQueryVariables>;
export function useNpcSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<NpcQuery, NpcQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<NpcQuery, NpcQueryVariables>(NpcDocument, options);
        }
export type NpcQueryHookResult = ReturnType<typeof useNpcQuery>;
export type NpcLazyQueryHookResult = ReturnType<typeof useNpcLazyQuery>;
export type NpcSuspenseQueryHookResult = ReturnType<typeof useNpcSuspenseQuery>;
export type NpcQueryResult = Apollo.QueryResult<NpcQuery, NpcQueryVariables>;
export const PlayerCharactersDocument = gql`
    query PlayerCharacters($gameId: ID!) {
  playerCharacters(gameId: $gameId) {
    id
    name
    status
    ownerPlayerId
  }
}
    `;

/**
 * __usePlayerCharactersQuery__
 *
 * To run a query within a React component, call `usePlayerCharactersQuery` and pass it any options that fit your needs.
 * When your component renders, `usePlayerCharactersQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = usePlayerCharactersQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function usePlayerCharactersQuery(baseOptions: Apollo.QueryHookOptions<PlayerCharactersQuery, PlayerCharactersQueryVariables> & ({ variables: PlayerCharactersQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<PlayerCharactersQuery, PlayerCharactersQueryVariables>(PlayerCharactersDocument, options);
      }
export function usePlayerCharactersLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<PlayerCharactersQuery, PlayerCharactersQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<PlayerCharactersQuery, PlayerCharactersQueryVariables>(PlayerCharactersDocument, options);
        }
// @ts-ignore
export function usePlayerCharactersSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<PlayerCharactersQuery, PlayerCharactersQueryVariables>): Apollo.UseSuspenseQueryResult<PlayerCharactersQuery, PlayerCharactersQueryVariables>;
export function usePlayerCharactersSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<PlayerCharactersQuery, PlayerCharactersQueryVariables>): Apollo.UseSuspenseQueryResult<PlayerCharactersQuery | undefined, PlayerCharactersQueryVariables>;
export function usePlayerCharactersSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<PlayerCharactersQuery, PlayerCharactersQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<PlayerCharactersQuery, PlayerCharactersQueryVariables>(PlayerCharactersDocument, options);
        }
export type PlayerCharactersQueryHookResult = ReturnType<typeof usePlayerCharactersQuery>;
export type PlayerCharactersLazyQueryHookResult = ReturnType<typeof usePlayerCharactersLazyQuery>;
export type PlayerCharactersSuspenseQueryHookResult = ReturnType<typeof usePlayerCharactersSuspenseQuery>;
export type PlayerCharactersQueryResult = Apollo.QueryResult<PlayerCharactersQuery, PlayerCharactersQueryVariables>;
export const FactionsDocument = gql`
    query Factions($gameId: ID!) {
  factions(gameId: $gameId) {
    id
    name
    status
    playerVisible
    memberCount
    standingLevelCount
    allies {
      id
      name
    }
    enemies {
      id
      name
    }
  }
}
    `;

/**
 * __useFactionsQuery__
 *
 * To run a query within a React component, call `useFactionsQuery` and pass it any options that fit your needs.
 * When your component renders, `useFactionsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFactionsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useFactionsQuery(baseOptions: Apollo.QueryHookOptions<FactionsQuery, FactionsQueryVariables> & ({ variables: FactionsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FactionsQuery, FactionsQueryVariables>(FactionsDocument, options);
      }
export function useFactionsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FactionsQuery, FactionsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FactionsQuery, FactionsQueryVariables>(FactionsDocument, options);
        }
// @ts-ignore
export function useFactionsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<FactionsQuery, FactionsQueryVariables>): Apollo.UseSuspenseQueryResult<FactionsQuery, FactionsQueryVariables>;
export function useFactionsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FactionsQuery, FactionsQueryVariables>): Apollo.UseSuspenseQueryResult<FactionsQuery | undefined, FactionsQueryVariables>;
export function useFactionsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FactionsQuery, FactionsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FactionsQuery, FactionsQueryVariables>(FactionsDocument, options);
        }
export type FactionsQueryHookResult = ReturnType<typeof useFactionsQuery>;
export type FactionsLazyQueryHookResult = ReturnType<typeof useFactionsLazyQuery>;
export type FactionsSuspenseQueryHookResult = ReturnType<typeof useFactionsSuspenseQuery>;
export type FactionsQueryResult = Apollo.QueryResult<FactionsQuery, FactionsQueryVariables>;
export const FactionDocument = gql`
    query Faction($id: ID!, $gameId: ID!) {
  faction(id: $id, gameId: $gameId) {
    id
    name
    status
    playerVisible
    members {
      npc {
        id
        name
      }
      rank
    }
    standingLevels {
      ordinal
      name
      threshold
    }
    allies {
      id
      name
    }
    enemies {
      id
      name
    }
  }
}
    `;

/**
 * __useFactionQuery__
 *
 * To run a query within a React component, call `useFactionQuery` and pass it any options that fit your needs.
 * When your component renders, `useFactionQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useFactionQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useFactionQuery(baseOptions: Apollo.QueryHookOptions<FactionQuery, FactionQueryVariables> & ({ variables: FactionQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<FactionQuery, FactionQueryVariables>(FactionDocument, options);
      }
export function useFactionLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<FactionQuery, FactionQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<FactionQuery, FactionQueryVariables>(FactionDocument, options);
        }
// @ts-ignore
export function useFactionSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<FactionQuery, FactionQueryVariables>): Apollo.UseSuspenseQueryResult<FactionQuery, FactionQueryVariables>;
export function useFactionSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FactionQuery, FactionQueryVariables>): Apollo.UseSuspenseQueryResult<FactionQuery | undefined, FactionQueryVariables>;
export function useFactionSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<FactionQuery, FactionQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<FactionQuery, FactionQueryVariables>(FactionDocument, options);
        }
export type FactionQueryHookResult = ReturnType<typeof useFactionQuery>;
export type FactionLazyQueryHookResult = ReturnType<typeof useFactionLazyQuery>;
export type FactionSuspenseQueryHookResult = ReturnType<typeof useFactionSuspenseQuery>;
export type FactionQueryResult = Apollo.QueryResult<FactionQuery, FactionQueryVariables>;
export const GamesDocument = gql`
    query Games {
  games {
    id
    name
    status
    campaignCount
    lastActivityAt
  }
}
    `;

/**
 * __useGamesQuery__
 *
 * To run a query within a React component, call `useGamesQuery` and pass it any options that fit your needs.
 * When your component renders, `useGamesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGamesQuery({
 *   variables: {
 *   },
 * });
 */
export function useGamesQuery(baseOptions?: Apollo.QueryHookOptions<GamesQuery, GamesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GamesQuery, GamesQueryVariables>(GamesDocument, options);
      }
export function useGamesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GamesQuery, GamesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GamesQuery, GamesQueryVariables>(GamesDocument, options);
        }
// @ts-ignore
export function useGamesSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GamesQuery, GamesQueryVariables>): Apollo.UseSuspenseQueryResult<GamesQuery, GamesQueryVariables>;
export function useGamesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GamesQuery, GamesQueryVariables>): Apollo.UseSuspenseQueryResult<GamesQuery | undefined, GamesQueryVariables>;
export function useGamesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GamesQuery, GamesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GamesQuery, GamesQueryVariables>(GamesDocument, options);
        }
export type GamesQueryHookResult = ReturnType<typeof useGamesQuery>;
export type GamesLazyQueryHookResult = ReturnType<typeof useGamesLazyQuery>;
export type GamesSuspenseQueryHookResult = ReturnType<typeof useGamesSuspenseQuery>;
export type GamesQueryResult = Apollo.QueryResult<GamesQuery, GamesQueryVariables>;
export const GameDocument = gql`
    query Game($id: ID!) {
  game(id: $id) {
    id
    name
    status
    locationSummary {
      draft
      active
      idle
      archived
    }
    factionSummary {
      draft
      active
      idle
      archived
    }
    npcSummary {
      draft
      active
      idle
      archived
    }
    campaigns {
      id
      name
      status
      characterCount
      sessionCount
    }
  }
}
    `;

/**
 * __useGameQuery__
 *
 * To run a query within a React component, call `useGameQuery` and pass it any options that fit your needs.
 * When your component renders, `useGameQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useGameQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useGameQuery(baseOptions: Apollo.QueryHookOptions<GameQuery, GameQueryVariables> & ({ variables: GameQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<GameQuery, GameQueryVariables>(GameDocument, options);
      }
export function useGameLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<GameQuery, GameQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<GameQuery, GameQueryVariables>(GameDocument, options);
        }
// @ts-ignore
export function useGameSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<GameQuery, GameQueryVariables>): Apollo.UseSuspenseQueryResult<GameQuery, GameQueryVariables>;
export function useGameSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GameQuery, GameQueryVariables>): Apollo.UseSuspenseQueryResult<GameQuery | undefined, GameQueryVariables>;
export function useGameSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<GameQuery, GameQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<GameQuery, GameQueryVariables>(GameDocument, options);
        }
export type GameQueryHookResult = ReturnType<typeof useGameQuery>;
export type GameLazyQueryHookResult = ReturnType<typeof useGameLazyQuery>;
export type GameSuspenseQueryHookResult = ReturnType<typeof useGameSuspenseQuery>;
export type GameQueryResult = Apollo.QueryResult<GameQuery, GameQueryVariables>;
export const LocationsDocument = gql`
    query Locations($gameId: ID!) {
  locations(gameId: $gameId) {
    id
    name
    locationType
    status
    parentId
    sceneCount
    childCount
  }
}
    `;

/**
 * __useLocationsQuery__
 *
 * To run a query within a React component, call `useLocationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useLocationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useLocationsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useLocationsQuery(baseOptions: Apollo.QueryHookOptions<LocationsQuery, LocationsQueryVariables> & ({ variables: LocationsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<LocationsQuery, LocationsQueryVariables>(LocationsDocument, options);
      }
export function useLocationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<LocationsQuery, LocationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<LocationsQuery, LocationsQueryVariables>(LocationsDocument, options);
        }
// @ts-ignore
export function useLocationsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<LocationsQuery, LocationsQueryVariables>): Apollo.UseSuspenseQueryResult<LocationsQuery, LocationsQueryVariables>;
export function useLocationsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<LocationsQuery, LocationsQueryVariables>): Apollo.UseSuspenseQueryResult<LocationsQuery | undefined, LocationsQueryVariables>;
export function useLocationsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<LocationsQuery, LocationsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<LocationsQuery, LocationsQueryVariables>(LocationsDocument, options);
        }
export type LocationsQueryHookResult = ReturnType<typeof useLocationsQuery>;
export type LocationsLazyQueryHookResult = ReturnType<typeof useLocationsLazyQuery>;
export type LocationsSuspenseQueryHookResult = ReturnType<typeof useLocationsSuspenseQuery>;
export type LocationsQueryResult = Apollo.QueryResult<LocationsQuery, LocationsQueryVariables>;
export const LocationDocument = gql`
    query Location($id: ID!, $gameId: ID!) {
  location(id: $id, gameId: $gameId) {
    id
    name
    locationType
    status
    parentLocation {
      id
      name
      locationType
      status
      parentId
      sceneCount
      childCount
    }
    scenes {
      id
      name
    }
    connections {
      toLocation {
        id
        name
        locationType
        status
        parentId
        sceneCount
        childCount
      }
      direction
    }
  }
}
    `;

/**
 * __useLocationQuery__
 *
 * To run a query within a React component, call `useLocationQuery` and pass it any options that fit your needs.
 * When your component renders, `useLocationQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useLocationQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useLocationQuery(baseOptions: Apollo.QueryHookOptions<LocationQuery, LocationQueryVariables> & ({ variables: LocationQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<LocationQuery, LocationQueryVariables>(LocationDocument, options);
      }
export function useLocationLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<LocationQuery, LocationQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<LocationQuery, LocationQueryVariables>(LocationDocument, options);
        }
// @ts-ignore
export function useLocationSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<LocationQuery, LocationQueryVariables>): Apollo.UseSuspenseQueryResult<LocationQuery, LocationQueryVariables>;
export function useLocationSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<LocationQuery, LocationQueryVariables>): Apollo.UseSuspenseQueryResult<LocationQuery | undefined, LocationQueryVariables>;
export function useLocationSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<LocationQuery, LocationQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<LocationQuery, LocationQueryVariables>(LocationDocument, options);
        }
export type LocationQueryHookResult = ReturnType<typeof useLocationQuery>;
export type LocationLazyQueryHookResult = ReturnType<typeof useLocationLazyQuery>;
export type LocationSuspenseQueryHookResult = ReturnType<typeof useLocationSuspenseQuery>;
export type LocationQueryResult = Apollo.QueryResult<LocationQuery, LocationQueryVariables>;
export const ActiveLocationsDocument = gql`
    query ActiveLocations($gameId: ID!) {
  activeLocations(gameId: $gameId) {
    id
    name
    locationType
    status
    parentId
    sceneCount
    childCount
  }
}
    `;

/**
 * __useActiveLocationsQuery__
 *
 * To run a query within a React component, call `useActiveLocationsQuery` and pass it any options that fit your needs.
 * When your component renders, `useActiveLocationsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useActiveLocationsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useActiveLocationsQuery(baseOptions: Apollo.QueryHookOptions<ActiveLocationsQuery, ActiveLocationsQueryVariables> & ({ variables: ActiveLocationsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ActiveLocationsQuery, ActiveLocationsQueryVariables>(ActiveLocationsDocument, options);
      }
export function useActiveLocationsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ActiveLocationsQuery, ActiveLocationsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ActiveLocationsQuery, ActiveLocationsQueryVariables>(ActiveLocationsDocument, options);
        }
// @ts-ignore
export function useActiveLocationsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ActiveLocationsQuery, ActiveLocationsQueryVariables>): Apollo.UseSuspenseQueryResult<ActiveLocationsQuery, ActiveLocationsQueryVariables>;
export function useActiveLocationsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ActiveLocationsQuery, ActiveLocationsQueryVariables>): Apollo.UseSuspenseQueryResult<ActiveLocationsQuery | undefined, ActiveLocationsQueryVariables>;
export function useActiveLocationsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ActiveLocationsQuery, ActiveLocationsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ActiveLocationsQuery, ActiveLocationsQueryVariables>(ActiveLocationsDocument, options);
        }
export type ActiveLocationsQueryHookResult = ReturnType<typeof useActiveLocationsQuery>;
export type ActiveLocationsLazyQueryHookResult = ReturnType<typeof useActiveLocationsLazyQuery>;
export type ActiveLocationsSuspenseQueryHookResult = ReturnType<typeof useActiveLocationsSuspenseQuery>;
export type ActiveLocationsQueryResult = Apollo.QueryResult<ActiveLocationsQuery, ActiveLocationsQueryVariables>;
export const MacGuffinsDocument = gql`
    query MacGuffins($gameId: ID!) {
  macguffins(gameId: $gameId) {
    id
    name
    playerVisible
    possessor {
      ... on NPCRef {
        id
        name
      }
      ... on PlayerCharacterRef {
        id
        name
      }
      ... on LocationRef {
        id
        name
      }
    }
  }
}
    `;

/**
 * __useMacGuffinsQuery__
 *
 * To run a query within a React component, call `useMacGuffinsQuery` and pass it any options that fit your needs.
 * When your component renders, `useMacGuffinsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMacGuffinsQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useMacGuffinsQuery(baseOptions: Apollo.QueryHookOptions<MacGuffinsQuery, MacGuffinsQueryVariables> & ({ variables: MacGuffinsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MacGuffinsQuery, MacGuffinsQueryVariables>(MacGuffinsDocument, options);
      }
export function useMacGuffinsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MacGuffinsQuery, MacGuffinsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MacGuffinsQuery, MacGuffinsQueryVariables>(MacGuffinsDocument, options);
        }
// @ts-ignore
export function useMacGuffinsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<MacGuffinsQuery, MacGuffinsQueryVariables>): Apollo.UseSuspenseQueryResult<MacGuffinsQuery, MacGuffinsQueryVariables>;
export function useMacGuffinsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MacGuffinsQuery, MacGuffinsQueryVariables>): Apollo.UseSuspenseQueryResult<MacGuffinsQuery | undefined, MacGuffinsQueryVariables>;
export function useMacGuffinsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MacGuffinsQuery, MacGuffinsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<MacGuffinsQuery, MacGuffinsQueryVariables>(MacGuffinsDocument, options);
        }
export type MacGuffinsQueryHookResult = ReturnType<typeof useMacGuffinsQuery>;
export type MacGuffinsLazyQueryHookResult = ReturnType<typeof useMacGuffinsLazyQuery>;
export type MacGuffinsSuspenseQueryHookResult = ReturnType<typeof useMacGuffinsSuspenseQuery>;
export type MacGuffinsQueryResult = Apollo.QueryResult<MacGuffinsQuery, MacGuffinsQueryVariables>;
export const MasterNarrativeDocument = gql`
    query MasterNarrative($gameId: ID!) {
  masterNarrative(gameId: $gameId) {
    id
    beats {
      id
      name
      beatType
      scope
      prerequisiteCount
    }
  }
}
    `;

/**
 * __useMasterNarrativeQuery__
 *
 * To run a query within a React component, call `useMasterNarrativeQuery` and pass it any options that fit your needs.
 * When your component renders, `useMasterNarrativeQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useMasterNarrativeQuery({
 *   variables: {
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useMasterNarrativeQuery(baseOptions: Apollo.QueryHookOptions<MasterNarrativeQuery, MasterNarrativeQueryVariables> & ({ variables: MasterNarrativeQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<MasterNarrativeQuery, MasterNarrativeQueryVariables>(MasterNarrativeDocument, options);
      }
export function useMasterNarrativeLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<MasterNarrativeQuery, MasterNarrativeQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<MasterNarrativeQuery, MasterNarrativeQueryVariables>(MasterNarrativeDocument, options);
        }
// @ts-ignore
export function useMasterNarrativeSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<MasterNarrativeQuery, MasterNarrativeQueryVariables>): Apollo.UseSuspenseQueryResult<MasterNarrativeQuery, MasterNarrativeQueryVariables>;
export function useMasterNarrativeSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MasterNarrativeQuery, MasterNarrativeQueryVariables>): Apollo.UseSuspenseQueryResult<MasterNarrativeQuery | undefined, MasterNarrativeQueryVariables>;
export function useMasterNarrativeSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<MasterNarrativeQuery, MasterNarrativeQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<MasterNarrativeQuery, MasterNarrativeQueryVariables>(MasterNarrativeDocument, options);
        }
export type MasterNarrativeQueryHookResult = ReturnType<typeof useMasterNarrativeQuery>;
export type MasterNarrativeLazyQueryHookResult = ReturnType<typeof useMasterNarrativeLazyQuery>;
export type MasterNarrativeSuspenseQueryHookResult = ReturnType<typeof useMasterNarrativeSuspenseQuery>;
export type MasterNarrativeQueryResult = Apollo.QueryResult<MasterNarrativeQuery, MasterNarrativeQueryVariables>;
export const BeatDocument = gql`
    query Beat($id: ID!, $gameId: ID!) {
  beat(id: $id, gameId: $gameId) {
    id
    name
    beatType
    scope
    description
    playerDescription
    prerequisites {
      id
      name
      beatType
      scope
      prerequisiteCount
    }
  }
}
    `;

/**
 * __useBeatQuery__
 *
 * To run a query within a React component, call `useBeatQuery` and pass it any options that fit your needs.
 * When your component renders, `useBeatQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useBeatQuery({
 *   variables: {
 *      id: // value for 'id'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useBeatQuery(baseOptions: Apollo.QueryHookOptions<BeatQuery, BeatQueryVariables> & ({ variables: BeatQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<BeatQuery, BeatQueryVariables>(BeatDocument, options);
      }
export function useBeatLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<BeatQuery, BeatQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<BeatQuery, BeatQueryVariables>(BeatDocument, options);
        }
// @ts-ignore
export function useBeatSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<BeatQuery, BeatQueryVariables>): Apollo.UseSuspenseQueryResult<BeatQuery, BeatQueryVariables>;
export function useBeatSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<BeatQuery, BeatQueryVariables>): Apollo.UseSuspenseQueryResult<BeatQuery | undefined, BeatQueryVariables>;
export function useBeatSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<BeatQuery, BeatQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<BeatQuery, BeatQueryVariables>(BeatDocument, options);
        }
export type BeatQueryHookResult = ReturnType<typeof useBeatQuery>;
export type BeatLazyQueryHookResult = ReturnType<typeof useBeatLazyQuery>;
export type BeatSuspenseQueryHookResult = ReturnType<typeof useBeatSuspenseQuery>;
export type BeatQueryResult = Apollo.QueryResult<BeatQuery, BeatQueryVariables>;
export const AvailableBeatsDocument = gql`
    query AvailableBeats($campaignNarrativeId: ID!, $gameId: ID!) {
  availableBeats(campaignNarrativeId: $campaignNarrativeId, gameId: $gameId) {
    id
    name
    beatType
    scope
    prerequisiteCount
  }
}
    `;

/**
 * __useAvailableBeatsQuery__
 *
 * To run a query within a React component, call `useAvailableBeatsQuery` and pass it any options that fit your needs.
 * When your component renders, `useAvailableBeatsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAvailableBeatsQuery({
 *   variables: {
 *      campaignNarrativeId: // value for 'campaignNarrativeId'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useAvailableBeatsQuery(baseOptions: Apollo.QueryHookOptions<AvailableBeatsQuery, AvailableBeatsQueryVariables> & ({ variables: AvailableBeatsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AvailableBeatsQuery, AvailableBeatsQueryVariables>(AvailableBeatsDocument, options);
      }
export function useAvailableBeatsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AvailableBeatsQuery, AvailableBeatsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AvailableBeatsQuery, AvailableBeatsQueryVariables>(AvailableBeatsDocument, options);
        }
// @ts-ignore
export function useAvailableBeatsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AvailableBeatsQuery, AvailableBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<AvailableBeatsQuery, AvailableBeatsQueryVariables>;
export function useAvailableBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AvailableBeatsQuery, AvailableBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<AvailableBeatsQuery | undefined, AvailableBeatsQueryVariables>;
export function useAvailableBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AvailableBeatsQuery, AvailableBeatsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AvailableBeatsQuery, AvailableBeatsQueryVariables>(AvailableBeatsDocument, options);
        }
export type AvailableBeatsQueryHookResult = ReturnType<typeof useAvailableBeatsQuery>;
export type AvailableBeatsLazyQueryHookResult = ReturnType<typeof useAvailableBeatsLazyQuery>;
export type AvailableBeatsSuspenseQueryHookResult = ReturnType<typeof useAvailableBeatsSuspenseQuery>;
export type AvailableBeatsQueryResult = Apollo.QueryResult<AvailableBeatsQuery, AvailableBeatsQueryVariables>;
export const DiscoveredBeatsDocument = gql`
    query DiscoveredBeats($campaignNarrativeId: ID!, $gameId: ID!) {
  discoveredBeats(campaignNarrativeId: $campaignNarrativeId, gameId: $gameId) {
    id
    name
    beatType
    discoveredInSession
  }
}
    `;

/**
 * __useDiscoveredBeatsQuery__
 *
 * To run a query within a React component, call `useDiscoveredBeatsQuery` and pass it any options that fit your needs.
 * When your component renders, `useDiscoveredBeatsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useDiscoveredBeatsQuery({
 *   variables: {
 *      campaignNarrativeId: // value for 'campaignNarrativeId'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useDiscoveredBeatsQuery(baseOptions: Apollo.QueryHookOptions<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables> & ({ variables: DiscoveredBeatsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>(DiscoveredBeatsDocument, options);
      }
export function useDiscoveredBeatsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>(DiscoveredBeatsDocument, options);
        }
// @ts-ignore
export function useDiscoveredBeatsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>;
export function useDiscoveredBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<DiscoveredBeatsQuery | undefined, DiscoveredBeatsQueryVariables>;
export function useDiscoveredBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>(DiscoveredBeatsDocument, options);
        }
export type DiscoveredBeatsQueryHookResult = ReturnType<typeof useDiscoveredBeatsQuery>;
export type DiscoveredBeatsLazyQueryHookResult = ReturnType<typeof useDiscoveredBeatsLazyQuery>;
export type DiscoveredBeatsSuspenseQueryHookResult = ReturnType<typeof useDiscoveredBeatsSuspenseQuery>;
export type DiscoveredBeatsQueryResult = Apollo.QueryResult<DiscoveredBeatsQuery, DiscoveredBeatsQueryVariables>;
export const CampaignBeatsDocument = gql`
    query CampaignBeats($campaignId: ID!, $gameId: ID!) {
  campaignBeats(campaignId: $campaignId, gameId: $gameId) {
    id
    name
    beatType
    scope
    prerequisiteCount
  }
}
    `;

/**
 * __useCampaignBeatsQuery__
 *
 * To run a query within a React component, call `useCampaignBeatsQuery` and pass it any options that fit your needs.
 * When your component renders, `useCampaignBeatsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCampaignBeatsQuery({
 *   variables: {
 *      campaignId: // value for 'campaignId'
 *      gameId: // value for 'gameId'
 *   },
 * });
 */
export function useCampaignBeatsQuery(baseOptions: Apollo.QueryHookOptions<CampaignBeatsQuery, CampaignBeatsQueryVariables> & ({ variables: CampaignBeatsQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CampaignBeatsQuery, CampaignBeatsQueryVariables>(CampaignBeatsDocument, options);
      }
export function useCampaignBeatsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CampaignBeatsQuery, CampaignBeatsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CampaignBeatsQuery, CampaignBeatsQueryVariables>(CampaignBeatsDocument, options);
        }
// @ts-ignore
export function useCampaignBeatsSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CampaignBeatsQuery, CampaignBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignBeatsQuery, CampaignBeatsQueryVariables>;
export function useCampaignBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignBeatsQuery, CampaignBeatsQueryVariables>): Apollo.UseSuspenseQueryResult<CampaignBeatsQuery | undefined, CampaignBeatsQueryVariables>;
export function useCampaignBeatsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<CampaignBeatsQuery, CampaignBeatsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CampaignBeatsQuery, CampaignBeatsQueryVariables>(CampaignBeatsDocument, options);
        }
export type CampaignBeatsQueryHookResult = ReturnType<typeof useCampaignBeatsQuery>;
export type CampaignBeatsLazyQueryHookResult = ReturnType<typeof useCampaignBeatsLazyQuery>;
export type CampaignBeatsSuspenseQueryHookResult = ReturnType<typeof useCampaignBeatsSuspenseQuery>;
export type CampaignBeatsQueryResult = Apollo.QueryResult<CampaignBeatsQuery, CampaignBeatsQueryVariables>;