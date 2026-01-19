// AUTO-GENERATED FILE - DO NOT EDIT
// Generated from asyncapi.yaml

export type Resource = "wood" | "brick" | "sheep" | "wheat" | "ore";
export const ResourceValues = ["wood", "brick", "sheep", "wheat", "ore"] as const;

export type TileResource = "wood" | "brick" | "sheep" | "wheat" | "ore" | "desert";
export const TileResourceValues = ["wood", "brick", "sheep", "wheat", "ore", "desert"] as const;

export type BuildingType = "settlement" | "city";
export const BuildingTypeValues = ["settlement", "city"] as const;

export type StructureType = "settlement" | "city" | "road";
export const StructureTypeValues = ["settlement", "city", "road"] as const;

export type GameStatus = "waiting" | "setup" | "playing" | "finished";
export const GameStatusValues = ["waiting", "setup", "playing", "finished"] as const;

export type TurnPhase = "roll" | "trade" | "build";
export const TurnPhaseValues = ["roll", "trade", "build"] as const;

export type PlayerColor = "red" | "blue" | "green" | "orange";
export const PlayerColorValues = ["red", "blue", "green", "orange"] as const;

export type DevCardType = "knight" | "road_building" | "year_of_plenty" | "monopoly" | "victory_point";
export const DevCardTypeValues = ["knight", "road_building", "year_of_plenty", "monopoly", "victory_point"] as const;

export interface HexCoord {
  q: number;
  r: number;
};

export interface Hex {
  coord: HexCoord;
  resource: TileResource;
  number: number;
};

export interface Building {
  type: BuildingType;
  ownerId: string;
};

export interface Road {
  ownerId: string;
};

export interface Vertex {
  id: string;
  adjacentHexes: HexCoord[];
  building?: Building;
};

export interface Edge {
  id: string;
  vertices: string[];
  road?: Road;
};

export interface ResourceCount {
  wood?: number;
  brick?: number;
  sheep?: number;
  wheat?: number;
  ore?: number;
};

export interface PlayerState {
  id: string;
  name: string;
  color: PlayerColor;
  resources: ResourceCount;
  devCardCount?: number;
  knightsPlayed?: number;
  victoryPoints: number;
  connected?: boolean;
};

export interface BoardState {
  hexes: Hex[];
  vertices: Vertex[];
  edges: Edge[];
  robberHex: HexCoord;
};

export interface GameState {
  id: string;
  code: string;
  board: BoardState;
  players: PlayerState[];
  currentTurn: number;
  turnPhase: TurnPhase;
  dice: number[];
  status: GameStatus;
  longestRoadPlayerId?: string;
  largestArmyPlayerId?: string;
  setupPhaseRound?: number;
};

export interface TradeOffer {
  id: string;
  proposerId: string;
  targetId?: string;
  offering: ResourceCount;
  requesting: ResourceCount;
  status: "pending" | "accepted" | "rejected" | "cancelled";
};

export interface JoinGameMessage {
  type: "JOIN_GAME";
  gameId: string;
};

export interface StartGameMessage {
  type: "START_GAME";
};

export interface RollDiceMessage {
  type: "ROLL_DICE";
};

export interface BuildStructureMessage {
  type: "BUILD";
  structureType: StructureType;
  location: string;
};

export interface ProposeTradeMessage {
  type: "PROPOSE_TRADE";
  targetId?: string;
  offering: ResourceCount;
  requesting: ResourceCount;
};

export interface RespondTradeMessage {
  type: "RESPOND_TRADE";
  tradeId: string;
  accept: boolean;
};

export interface MoveRobberMessage {
  type: "MOVE_ROBBER";
  hex: HexCoord;
  victimId?: string;
};

export interface EndTurnMessage {
  type: "END_TURN";
};

export interface PlayDevCardMessage {
  type: "PLAY_DEV_CARD";
  cardType: DevCardType;
  targetResource?: Resource;
  resources?: Resource[];
};

export interface GameStateMessage {
  type: "GAME_STATE";
  payload: GameState;
};

export interface PlayerJoinedMessage {
  type: "PLAYER_JOINED";
  payload: {
    player: PlayerState;
  };
};

export interface PlayerLeftMessage {
  type: "PLAYER_LEFT";
  payload: {
    playerId: string;
  };
};

export interface DiceRolledMessage {
  type: "DICE_ROLLED";
  payload: {
    playerId: string;
    values: number[];
    resourcesDistributed?: {
  playerId?: string;
  resources?: ResourceCount;
}[];
  };
};

export interface BuildingPlacedMessage {
  type: "BUILDING_PLACED";
  payload: {
    playerId: string;
    buildingType: BuildingType;
    vertexId: string;
  };
};

export interface RoadPlacedMessage {
  type: "ROAD_PLACED";
  payload: {
    playerId: string;
    edgeId: string;
  };
};

export interface TradeProposedMessage {
  type: "TRADE_PROPOSED";
  payload: TradeOffer;
};

export interface TradeResolvedMessage {
  type: "TRADE_RESOLVED";
  payload: {
    tradeId: string;
    accepted: boolean;
    acceptedBy?: string;
  };
};

export interface RobberMovedMessage {
  type: "ROBBER_MOVED";
  payload: {
    playerId: string;
    hex: HexCoord;
    victimId?: string;
    stolenResource?: Resource;
  };
};

export interface TurnChangedMessage {
  type: "TURN_CHANGED";
  payload: {
    activePlayerId: string;
    phase: TurnPhase;
  };
};

export interface GameStartedMessage {
  type: "GAME_STARTED";
  payload: GameState;
};

export interface GameOverMessage {
  type: "GAME_OVER";
  payload: {
    winnerId: string;
    scores: {
  playerId: string;
  points: number;
}[];
  };
};

export interface ErrorMessage {
  type: "ERROR";
  payload: {
    code?: string;
    message: string;
  };
};

export interface CreateGameRequest {
  playerName: string;
};

export interface CreateGameResponse {
  gameId: string;
  code: string;
  sessionToken: string;
  playerId: string;
};

export interface JoinGameRequest {
  playerName: string;
};

export interface JoinGameResponse {
  gameId: string;
  sessionToken: string;
  playerId: string;
  players: {
  id?: string;
  name?: string;
  color?: PlayerColor;
}[];
};

export interface GameInfoResponse {
  code: string;
  status: GameStatus;
  playerCount: number;
  players: {
  name?: string;
  color?: PlayerColor;
}[];
};

// Message type unions
export type ClientMessage =
  | JoinGameMessage
  | StartGameMessage
  | RollDiceMessage
  | BuildStructureMessage
  | ProposeTradeMessage
  | RespondTradeMessage
  | MoveRobberMessage
  | EndTurnMessage
  | PlayDevCardMessage;

export type ServerMessage =
  | GameStateMessage
  | PlayerJoinedMessage
  | PlayerLeftMessage
  | DiceRolledMessage
  | BuildingPlacedMessage
  | RoadPlacedMessage
  | TradeProposedMessage
  | TradeResolvedMessage
  | RobberMovedMessage
  | TurnChangedMessage
  | GameStartedMessage
  | GameOverMessage
  | ErrorMessage;

// Message type constants
export const MessageTypes = {
  // Client -> Server
  JOIN_GAME: 'JOIN_GAME',
  START_GAME: 'START_GAME',
  ROLL_DICE: 'ROLL_DICE',
  BUILD: 'BUILD',
  PROPOSE_TRADE: 'PROPOSE_TRADE',
  RESPOND_TRADE: 'RESPOND_TRADE',
  MOVE_ROBBER: 'MOVE_ROBBER',
  END_TURN: 'END_TURN',
  PLAY_DEV_CARD: 'PLAY_DEV_CARD',
  // Server -> Client
  GAME_STATE: 'GAME_STATE',
  PLAYER_JOINED: 'PLAYER_JOINED',
  PLAYER_LEFT: 'PLAYER_LEFT',
  DICE_ROLLED: 'DICE_ROLLED',
  BUILDING_PLACED: 'BUILDING_PLACED',
  ROAD_PLACED: 'ROAD_PLACED',
  TRADE_PROPOSED: 'TRADE_PROPOSED',
  TRADE_RESOLVED: 'TRADE_RESOLVED',
  ROBBER_MOVED: 'ROBBER_MOVED',
  TURN_CHANGED: 'TURN_CHANGED',
  GAME_STARTED: 'GAME_STARTED',
  GAME_OVER: 'GAME_OVER',
  ERROR: 'ERROR',
} as const;
