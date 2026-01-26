import {
  createContext,
  useContext,
  useReducer,
  useCallback,
  useMemo,
  type ReactNode,
} from "react";
import type {
  GameState,
  PlayerState,
  ServerMessage,
  ClientMessage,
  ResourceCount,
  GameOverPayload,
} from "@/types";
import { GameStatus, StructureType, TurnPhase, DevCardType, Resource } from "@/types";
import { useWebSocket } from "@/hooks/useWebSocket";
import { getPlacementState, type PlacementState } from "@/components/Board/placement";

// Game context state
interface GameContextState {
  gameState: GameState | null;
  gameOver: GameOverPayload | null;
  currentPlayerId: string | null;
  isConnected: boolean;
  error: string | null;
  resourceGain: ResourceGain | null;
}

type PlacementMode = "settlement" | "road" | "build";

interface ResourceGain {
  playerId: string;
  resources: ResourceCount;
}

// Actions for the reducer
type GameAction =
  | { type: "SET_GAME_STATE"; payload: GameState }
  | { type: "SET_GAME_OVER"; payload: GameOverPayload | null }
  | { type: "SET_PLAYER_ID"; payload: string }
  | { type: "PLAYER_JOINED"; payload: PlayerState }
  | { type: "PLAYER_LEFT"; payload: string }
  | { type: "SET_CONNECTED"; payload: boolean }
  | { type: "SET_ERROR"; payload: string | null }
  | { type: "CLEAR_RESOURCE_GAIN" }
  | { type: "RESET" };

const initialState: GameContextState = {
  gameState: null,
  gameOver: null,
  currentPlayerId: null,
  isConnected: false,
  error: null,
  resourceGain: null,
};

function gameReducer(
  state: GameContextState,
  action: GameAction
): GameContextState {
  switch (action.type) {
    case "SET_GAME_STATE": {
      const resourceGain = getSetupResourceGain(
        state.gameState,
        action.payload,
        state.currentPlayerId
      );
      const isFinished = isStatus(
        action.payload.status,
        GameStatus.FINISHED,
        "GAME_STATUS_FINISHED"
      );
      return {
        ...state,
        gameState: action.payload,
        gameOver: isFinished ? state.gameOver : null,
        resourceGain: resourceGain ?? state.resourceGain,
      };
    }
    case "SET_GAME_OVER":
      return { ...state, gameOver: action.payload };
    case "SET_PLAYER_ID":
      return { ...state, currentPlayerId: action.payload };
    case "PLAYER_JOINED":
      if (!state.gameState) return state;
      return {
        ...state,
        gameState: {
          ...state.gameState,
          players: [...state.gameState.players, action.payload],
        },
      };
    case "PLAYER_LEFT":
      if (!state.gameState) return state;
      return {
        ...state,
        gameState: {
          ...state.gameState,
          players: state.gameState.players.filter(
            (p) => p.id !== action.payload
          ),
        },
      };
    case "SET_CONNECTED":
      return { ...state, isConnected: action.payload };
    case "SET_ERROR":
      return { ...state, error: action.payload };
    case "CLEAR_RESOURCE_GAIN":
      return { ...state, resourceGain: null };
    case "RESET":
      return initialState;
    default:
      return state;
  }
}

// Context value type
interface GameContextValue extends GameContextState {
  connect: () => void;
  disconnect: () => void;
  sendMessage: (message: ClientMessage) => void;
  rollDice: () => void;
  build: (structureType: StructureType, location: string) => void;
  endTurn: () => void;
  startGame: () => void;
  setReady: (ready: boolean) => void;
  clearResourceGain: () => void;
  placementMode: PlacementMode | null;
  placementState: PlacementState;
  // --- Robber UI ---
  isRobberDiscardRequired: boolean;
  robberDiscardAmount: number;
  robberDiscardMax: ResourceCount | null;
  isRobberMoveRequired: boolean;
  isRobberStealRequired: boolean;
  sendRobberDiscard: (toDiscard: ResourceCount) => void;
  sendRobberMove: (hex: { q: number; r: number }, victimId?: string) => void;
  sendRobberSteal: (victimId: string) => void;
  robberStealCandidates: { id: string; name: string; avatarUrl?: string }[];
  // --- Dev Cards ---
  buyDevCard: () => void;
  playDevCard: (cardType: DevCardType, targetResource?: Resource, resources?: Resource[]) => void;
  // --- Trading ---
  proposeTrade: (offering: ResourceCount, requesting: ResourceCount, targetPlayerId?: string | null) => void;
  respondTrade: (tradeId: string, accept: boolean) => void;
  bankTrade: (offering: ResourceCount, resourceRequested: Resource) => void;
  setTurnPhase: (phase: TurnPhase) => void;
}


const GameContext = createContext<GameContextValue | null>(null);

interface GameProviderProps {
  children: ReactNode;
  playerId: string | null;
}

export function GameProvider({ children, playerId }: GameProviderProps) {
  const [state, dispatch] = useReducer(gameReducer, {
    ...initialState,
    currentPlayerId: playerId,
  });

  const handleMessage = useCallback((data: ServerMessage) => {
    const msg = data.message;
    if (msg.oneofKind === undefined) return;

    switch (msg.oneofKind) {
      case "gameState":
        if (msg.gameState?.state) {
          dispatch({ type: "SET_GAME_STATE", payload: msg.gameState.state });
        }
        break;
      case "gameStarted":
        if (msg.gameStarted?.state) {
          dispatch({ type: "SET_GAME_STATE", payload: msg.gameStarted.state });
        }
        break;
      case "playerJoined":
        if (msg.playerJoined?.player) {
          dispatch({ type: "PLAYER_JOINED", payload: msg.playerJoined.player });
        }
        break;
      case "playerLeft":
        if (msg.playerLeft?.playerId) {
          dispatch({ type: "PLAYER_LEFT", payload: msg.playerLeft.playerId });
        }
        break;
      case "error":
        if (msg.error?.message) {
          dispatch({ type: "SET_ERROR", payload: msg.error.message });
        }
        break;
      case "gameOver":
        dispatch({ type: "SET_GAME_OVER", payload: msg.gameOver ?? null });
        break;
      // Handle other message types as needed
    }
  }, []);

  const { isConnected, error, connect, disconnect, sendMessage } = useWebSocket(
    {
      onMessage: handleMessage,
      onConnect: () => dispatch({ type: "SET_CONNECTED", payload: true }),
      onDisconnect: () => dispatch({ type: "SET_CONNECTED", payload: false }),
    }
  );

  const rollDice = useCallback(() => {
    sendMessage({
      message: { oneofKind: "rollDice", rollDice: {} },
    } as ClientMessage);
  }, [sendMessage]);

  const build = useCallback(
    (structureType: StructureType, location: string) => {
      sendMessage({
        message: {
          oneofKind: "buildStructure",
          buildStructure: { structureType, location },
        },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const endTurn = useCallback(() => {
    sendMessage({
      message: { oneofKind: "endTurn", endTurn: {} },
    } as ClientMessage);
  }, [sendMessage]);

  const setTurnPhase = useCallback((phase: TurnPhase) => {
    sendMessage({
      message: { oneofKind: "setTurnPhase", setTurnPhase: { phase } },
    } as ClientMessage);
  }, [sendMessage]);

  const startGame = useCallback(() => {
    sendMessage({
      message: { oneofKind: "startGame", startGame: {} },
    } as ClientMessage);
  }, [sendMessage]);

  const setReady = useCallback(
    (ready: boolean) => {
      sendMessage({
        message: { oneofKind: "playerReady", playerReady: { ready } },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const clearResourceGain = useCallback(() => {
    dispatch({ type: "CLEAR_RESOURCE_GAIN" });
  }, []);

  const placementMode = getPlacementMode(state.gameState, state.currentPlayerId);
  const placementState = useMemo(
    () => getPlacementState(state.gameState, state.currentPlayerId),
    [state.gameState, state.currentPlayerId]
  );

  // --- Robber Phase Derivation ---
  const robberPhase = state.gameState?.robberPhase;
  const currentPlayerId = state.currentPlayerId ?? "";

  // Discard
  const isRobberDiscardRequired = !!(
    robberPhase && robberPhase.discardPending?.includes(currentPlayerId)
  );
  const robberDiscardAmount = robberPhase?.discardRequired?.[currentPlayerId] ?? 0;
  const robberDiscardMax = useMemo(() => {
    if (!state.gameState?.players || !currentPlayerId) return null;
    return state.gameState.players.find(p => p.id === currentPlayerId)?.resources ?? null;
  }, [state.gameState, currentPlayerId]);

  // Move
  const isRobberMoveRequired = !!(
    robberPhase && robberPhase.movePendingPlayerId === currentPlayerId
  );

  // Steal
  const isRobberStealRequired = !!(
    robberPhase && robberPhase.stealPendingPlayerId === currentPlayerId
  );

  // Candidates for StealModal
  const robberStealCandidates = useMemo(() => {
    // Find hex where robber just moved to (usually board.robberHex)
    const board = state.gameState?.board;
    if (!board || !board.robberHex) return [];
    // Victim: any player (ID not currentPlayerId) with a building on adjacent vertex
    const adjacentPlayers = new Set<string>();
    (board.vertices ?? []).forEach((v) => {
      if (
        v.building &&
        v.building.ownerId !== currentPlayerId &&
        v.adjacentHexes?.some(
          h => h.q === board.robberHex!.q && h.r === board.robberHex!.r
        )
      ) {
        adjacentPlayers.add(v.building.ownerId);
      }
    });
    // Map to player objects
    return (state.gameState?.players ?? [])
      .filter(p => adjacentPlayers.has(p.id))
      .map(p => ({ id: p.id, name: p.name }));
  }, [state.gameState?.board, state.gameState?.players, currentPlayerId]);

  // Handlers
  const sendRobberDiscard = useCallback((toDiscard: ResourceCount) => {
    sendMessage({
      message: {
        oneofKind: "discardCards",
        discardCards: { resources: toDiscard },
      },
    } as ClientMessage);
  }, [sendMessage]);

  const sendRobberMove = useCallback((hex: { q: number; r: number }, victimId?: string) => {
    sendMessage({
      message: {
        oneofKind: "moveRobber",
        moveRobber: {
          hex: { q: hex.q, r: hex.r },
          victimId,
        },
      },
    } as ClientMessage);
  }, [sendMessage]);

  const sendRobberSteal = useCallback((victimId: string) => {
    // For the UI: call moveRobber with undefined hex (should be ignored on handler), but set victimId.
    sendMessage({
      message: {
        oneofKind: "moveRobber",
        moveRobber: { victimId },
      },
    } as ClientMessage);
  }, [sendMessage]);

  const buyDevCard = useCallback(() => {
    sendMessage({
      message: {
        oneofKind: "buyDevCard",
        buyDevCard: {},
      },
    } as ClientMessage);
  }, [sendMessage]);

  const playDevCard = useCallback(
    (cardType: DevCardType, targetResource?: Resource, resources?: Resource[]) => {
      sendMessage({
        message: {
          oneofKind: "playDevCard",
          playDevCard: {
            cardType,
            targetResource,
            resources: resources ?? [],
          },
        },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const proposeTrade = useCallback(
    (offering: ResourceCount, requesting: ResourceCount, targetPlayerId?: string | null) => {
      sendMessage({
        message: {
          oneofKind: "proposeTrade",
          proposeTrade: {
            targetId: targetPlayerId ?? undefined,
            offering,
            requesting,
          },
        },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const respondTrade = useCallback(
    (tradeId: string, accept: boolean) => {
      sendMessage({
        message: {
          oneofKind: "respondTrade",
          respondTrade: {
            tradeId,
            accept,
          },
        },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const bankTrade = useCallback(
    (offering: ResourceCount, resourceRequested: Resource) => {
      sendMessage({
        message: {
          oneofKind: "bankTrade",
          bankTrade: {
            offering,
            resourceRequested,
          },
        },
      } as ClientMessage);
    },
    [sendMessage]
  );

  const value: GameContextValue = {
    ...state,
    isConnected,
    error: state.error || error,
    connect,
    disconnect,
    sendMessage,
    rollDice,
    build,
    endTurn,
    startGame,
    setReady,
    resourceGain: state.resourceGain,
    clearResourceGain,
    placementMode,
    placementState,
    isRobberDiscardRequired,
    robberDiscardAmount,
    robberDiscardMax,
    isRobberMoveRequired,
    isRobberStealRequired,
    sendRobberDiscard,
    sendRobberMove,
    sendRobberSteal,
    robberStealCandidates,
    buyDevCard,
    playDevCard,
    proposeTrade,
    respondTrade,
    bankTrade,
    setTurnPhase,
  };

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}

export function useGame() {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error("useGame must be used within a GameProvider");
  }
  return context;
}

function isStatus(
  status: GameStatus | string | undefined,
  expected: GameStatus,
  expectedString: string
): boolean {
  return status === expected || status === (expectedString as unknown as GameStatus);
}

function isTurnPhase(
  phase: TurnPhase | string | undefined,
  expected: TurnPhase,
  expectedString: string
): boolean {
  return phase === expected || phase === (expectedString as unknown as TurnPhase);
}

function getSetupResourceGain(
  previousState: GameState | null,
  nextState: GameState,
  currentPlayerId: string | null
): ResourceGain | null {
  if (!previousState || !currentPlayerId) {
    return null;
  }

  if (
    !isStatus(nextState.status, GameStatus.SETUP, "GAME_STATUS_SETUP") ||
    nextState.setupPhase?.round !== 2
  ) {
    return null;
  }

  const previousPlayer = previousState.players.find(
    (player) => player.id === currentPlayerId
  );
  const nextPlayer = nextState.players.find(
    (player) => player.id === currentPlayerId
  );

  if (!previousPlayer || !nextPlayer) {
    return null;
  }

  const delta = getPositiveResourceDelta(
    previousPlayer.resources,
    nextPlayer.resources
  );
  if (!delta) {
    return null;
  }

  return { playerId: currentPlayerId, resources: delta };
}

function getPositiveResourceDelta(
  previousResources: ResourceCount | undefined,
  nextResources: ResourceCount | undefined
): ResourceCount | null {
  if (!previousResources || !nextResources) {
    return null;
  }

  const delta: ResourceCount = { wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 };
  let hasGain = false;
  const keys = ["wood", "brick", "sheep", "wheat", "ore"] as const;

  for (const key of keys) {
    const diff = (nextResources[key] ?? 0) - (previousResources[key] ?? 0);
    if (diff > 0) {
      delta[key] = diff;
      hasGain = true;
    }
  }

  return hasGain ? delta : null;
}

function getPlacementMode(
  gameState: GameState | null,
  currentPlayerId: string | null
): PlacementMode | null {
  if (!gameState || !currentPlayerId) {
    return null;
  }

  const currentPlayer = gameState.players[gameState.currentTurn ?? 0];
  if (!currentPlayer || currentPlayer.id !== currentPlayerId) {
    return null;
  }

  if (isStatus(gameState.status, GameStatus.SETUP, "GAME_STATUS_SETUP")) {
    const placementsInTurn = gameState.setupPhase?.placementsInTurn ?? 0;
    return placementsInTurn === 0 ? "settlement" : "road";
  }

  if (
    isStatus(gameState.status, GameStatus.PLAYING, "GAME_STATUS_PLAYING") &&
    isTurnPhase(
      gameState.turnPhase,
      TurnPhase.BUILD,
      "TURN_PHASE_BUILD"
    )
  ) {
    return "build";
  }

  return null;
}
