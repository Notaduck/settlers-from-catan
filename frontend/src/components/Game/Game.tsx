import React from "react";
import { useGame } from "@/context/GameContext";
import Board from "@/components/Board/Board";
import { SetupPhasePanel } from "./SetupPhasePanel";
import { DevCardPanel } from "../DevCardPanel";

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

import { StructureType, GameStatus, PlayerColor } from "@/gen/proto/catan/v1/types";

const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

export function Game({ gameCode: _gameCode, onLeave: _onLeave }: GameProps) {
  // Silence unused prop lint:
  void _gameCode;
  void _onLeave;
  // Always call useGame ONCE at top
  const {
    gameState,
    placementState,
    placementMode,
    build,
    currentPlayerId,
    setReady,
    startGame,
    isConnected,
  } = useGame();

  console.log(
    "[Game render] gameState:",
    gameState && {
      status: gameState.status,
      players: gameState.players,
      turn: gameState.currentTurn,
      setupPhase: gameState.setupPhase,
    }
  );

  if (!gameState) {
    return (
      <div className="game-loading" data-cy="game-loading">
        {isConnected ? "Syncing game state..." : "Connecting to game..."}
      </div>
    );
  }

  // Show E2E-visible waiting room if status is WAITING
  if (gameState.status === GameStatus.WAITING) {
    const players = gameState.players ?? [];
    const me = players.find((p) => p.id === currentPlayerId);
    const isReady = Boolean(me?.isReady);
    const isHost = Boolean(me?.isHost);
    const allReady = players.length >= 2 && players.every((p) => p.isReady);

    return (
      <div className="game-waiting" data-cy="game-waiting">
        <h2>Lobby</h2>
        <p>Share this code with friends to join the game.</p>
        <div className="game-code" data-cy="game-code">
          Game Code: <strong>{gameState.code}</strong>
        </div>

        <div className="players-list">
          <h3>Players</h3>
          {players.map((player) => (
            <div
              key={player.id}
              className={`player-item ${player.isReady ? "ready" : ""}`}
            >
              <span
                className="player-color"
                style={{
                  backgroundColor:
                    PLAYER_COLORS[(player.color as PlayerColor) ?? PlayerColor.UNSPECIFIED],
                }}
              />
              <span className="player-name">
                {player.name}
                {player.isHost && <span className="host-badge">HOST</span>}
              </span>
              <span
                className={`ready-status ${player.isReady ? "ready" : "not-ready"}`}
              >
                {player.isReady ? "✓ Ready" : "Not Ready"}
              </span>
            </div>
          ))}
        </div>

        <div className="lobby-actions">
          {isReady ? (
            <button
              className="btn btn-secondary"
              data-cy="cancel-ready-btn"
              onClick={() => setReady(false)}
            >
              Cancel Ready
            </button>
          ) : (
            <button
              className="btn btn-primary btn-ready"
              data-cy="ready-btn"
              onClick={() => setReady(true)}
            >
              Ready
            </button>
          )}
          {isHost && (
            <button
              className="btn btn-primary"
              data-cy="start-game-btn"
              disabled={!allReady}
              onClick={() => startGame()}
            >
              Start Game
            </button>
          )}
        </div>
        {!allReady && (
          <div className="waiting-hint">Waiting for all players to be ready...</div>
        )}
      </div>
    );
  }

  // Setup phase UI
  if (gameState?.status === GameStatus.SETUP) {
    // BoardState is optional in type but required here
    const board = gameState.board!;
    // Placement handlers
    const onBuildSettlement =
      placementMode === "settlement"
        ? (vertexId: string) => build(StructureType.SETTLEMENT, vertexId)
        : undefined;
    const onBuildRoad =
      placementMode === "road"
        ? (edgeId: string) => build(StructureType.ROAD, edgeId)
        : undefined;
    return (
      <div className="game-board-container">
        <div data-cy="game-phase">SETUP</div>
        <SetupPhasePanel />
        <Board
          board={board}
          players={gameState.players}
          validVertexIds={placementState?.validVertexIds}
          validEdgeIds={placementState?.validEdgeIds}
          onBuildSettlement={onBuildSettlement}
          onBuildRoad={onBuildRoad}
        />
      </div>
    );
  }

  // Main gameplay
  if (
    gameState?.status === GameStatus.PLAYING &&
    gameState.players?.length
  ) {
    const board = gameState.board!;
    const onBuildSettlement =
      placementMode === "build"
        ? (vertexId: string) => build(StructureType.SETTLEMENT, vertexId)
        : undefined;
    const onBuildRoad =
      placementMode === "build"
        ? (edgeId: string) => build(StructureType.ROAD, edgeId)
        : undefined;
    return (
      <div className="game-board-container">
        <Board
          board={board}
          players={gameState.players}
          validVertexIds={placementState?.validVertexIds}
          validEdgeIds={placementState?.validEdgeIds}
          onBuildSettlement={onBuildSettlement}
          onBuildRoad={onBuildRoad}
        />
        <div className="main-game-ui">
          <div data-cy="game-phase">PLAYING</div>
          <DevCardPanel />
        </div>
      </div>
    );
  }

  // Fallback state
  const sessionToken = sessionStorage.getItem("sessionToken");
  return (
    <div className="lobby-actions">
      <div data-cy="main-game-placeholder">
        {sessionToken == null
          ? "ERROR: No session token found in sessionStorage"
          : `Game in progress or not started. Status: ${
              typeof gameState?.status !== "undefined"
                ? String(gameState.status)
                : "undefined"
            }`}
      </div>
    </div>
  );
}
