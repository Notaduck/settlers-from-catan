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
} from "@/types";
import { GameStatus, StructureType, TurnPhase } from "@/types";
import { useWebSocket } from "@/hooks/useWebSocket";
import { getPlacementState, type PlacementState } from "@/components/Board/placement";

// Game context state
interface GameContextState {
  gameState: GameState | null;
  currentPlayerId: string | null;
  isConnected: boolean;
  error: string | null;
}

type PlacementMode = "settlement" | "road" | "build";

// Actions for the reducer
type GameAction =
  | { type: "SET_GAME_STATE"; payload: GameState }
  | { type: "SET_PLAYER_ID"; payload: string }
  | { type: "PLAYER_JOINED"; payload: PlayerState }
  | { type: "PLAYER_LEFT"; payload: string }
  | { type: "SET_CONNECTED"; payload: boolean }
  | { type: "SET_ERROR"; payload: string | null }
  | { type: "RESET" };

const initialState: GameContextState = {
  gameState: null,
  currentPlayerId: null,
  isConnected: false,
  error: null,
};

function gameReducer(
  state: GameContextState,
  action: GameAction
): GameContextState {
  switch (action.type) {
    case "SET_GAME_STATE":
      return { ...state, gameState: action.payload };
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
  placementMode: PlacementMode | null;
  placementState: PlacementState;
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

  const placementMode = getPlacementMode(state.gameState, state.currentPlayerId);
  const placementState = useMemo(
    () => getPlacementState(state.gameState, state.currentPlayerId),
    [state.gameState, state.currentPlayerId]
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
    placementMode,
    placementState,
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
  return status === expected || status === (expectedString as GameStatus);
}

function isTurnPhase(
  phase: TurnPhase | string | undefined,
  expected: TurnPhase,
  expectedString: string
): boolean {
  return phase === expected || phase === (expectedString as TurnPhase);
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
