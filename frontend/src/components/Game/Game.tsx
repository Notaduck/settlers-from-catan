import { useEffect, useMemo } from "react";
import { useGame } from "@/context";
import { Board } from "@/components/Board";
import { PlayerPanel } from "@/components/PlayerPanel";
import { GameStatus, PlayerColor } from "@/types";
import "./Game.css";

// Map PlayerColor enum to CSS color strings
const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

export function Game({ gameCode, onLeave }: GameProps) {
  const {
    connect,
    disconnect,
    isConnected,
    gameState,
    error,
    startGame,
    setReady,
    currentPlayerId,
    placementMode,
  } = useGame();

  useEffect(() => {
    connect();
    return () => disconnect();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount/unmount

  const players = gameState?.players ?? [];

  // Find current player's state
  const currentPlayer = useMemo(
    () => players.find((p) => p.id === currentPlayerId),
    [players, currentPlayerId]
  );

  // Check if all players are ready
  const allPlayersReady = useMemo(
    () => players.length >= 2 && players.every((p) => p.isReady),
    [players]
  );

  // Check if current player is host/ready
  const isHost = currentPlayer?.isHost ?? false;
  const isReady = currentPlayer?.isReady ?? false;
  // Handle both enum value and JSON string representation
  const isWaiting =
    gameState?.status === GameStatus.WAITING ||
    gameState?.status === ("GAME_STATUS_WAITING" as unknown as GameStatus);

  const placementModeLabel = useMemo(() => {
    switch (placementMode) {
      case "settlement":
        return "Place Settlement";
      case "road":
        return "Place Road";
      case "build":
        return "Build";
      default:
        return null;
    }
  }, [placementMode]);

  if (error) {
    return (
      <div className="game-error" data-cy="game-error">
        <h2>Connection Error</h2>
        <p>{error}</p>
        <button
          onClick={onLeave}
          className="btn btn-secondary"
          data-cy="back-to-lobby-btn"
        >
          Back to Lobby
        </button>
      </div>
    );
  }

  if (!isConnected) {
    return (
      <div className="game-loading" data-cy="game-loading">
        <p>Connecting to game...</p>
      </div>
    );
  }

  if (!gameState) {
    return (
      <div className="game-loading" data-cy="game-loading">
        <p>Loading game state...</p>
      </div>
    );
  }

  return (
    <div className="game" data-cy="game">
      <div className="game-header">
        <div className="game-code" data-cy="game-code">
          Game Code: <strong>{gameCode}</strong>
        </div>
        <button
          onClick={onLeave}
          className="btn btn-small"
          data-cy="leave-game-btn"
        >
          Leave Game
        </button>
      </div>

      {isWaiting ? (
        <div className="game-waiting" data-cy="game-waiting">
          <h2>Game Lobby</h2>
          <p>Share the game code with your friends to join!</p>
          <div className="players-list" data-cy="players-list">
            <h3>Players ({gameState.players.length}/4)</h3>
            {gameState.players.map((player) => (
              <div
                key={player.id}
                className={`player-item ${player.isReady ? "ready" : ""}`}
                data-cy={`player-item-${player.id}`}
              >
                <span
                  className="player-color"
                  style={{ backgroundColor: PLAYER_COLORS[player.color] }}
                />
                <span className="player-name">
                  {player.name}
                  {player.isHost && (
                    <span className="host-badge" data-cy="host-badge">
                      HOST
                    </span>
                  )}
                </span>
                <span
                  className={`ready-status ${
                    player.isReady ? "ready" : "not-ready"
                  }`}
                  data-cy={`player-ready-status-${player.id}`}
                >
                  {player.isReady ? "✓ Ready" : "Not Ready"}
                </span>
              </div>
            ))}
          </div>

          <div className="lobby-actions">
            {!isReady ? (
              <button
                onClick={() => setReady(true)}
                className="btn btn-ready"
                data-cy="ready-btn"
              >
                I'm Ready
              </button>
            ) : (
              <button
                onClick={() => setReady(false)}
                className="btn btn-secondary"
                data-cy="cancel-ready-btn"
              >
                Cancel Ready
              </button>
            )}

            {isHost && (
              <button
                onClick={startGame}
                className="btn btn-primary"
                disabled={!allPlayersReady}
                title={
                  !allPlayersReady
                    ? "All players must be ready to start"
                    : "Start the game"
                }
                data-cy="start-game-btn"
              >
                Start Game
              </button>
            )}
          </div>

          {gameState.players.length < 2 && (
            <p className="waiting-hint">Waiting for at least 2 players...</p>
          )}
          {gameState.players.length >= 2 && !allPlayersReady && (
            <p className="waiting-hint">
              Waiting for all players to be ready...
            </p>
          )}
        </div>
      ) : (
        <div className="game-board-container" data-cy="game-board-container">
          {placementModeLabel && (
            <div className="placement-mode" data-cy="placement-mode">
              {placementModeLabel}
            </div>
          )}
          <div className="game-board-content">
            {gameState.board && (
              <Board board={gameState.board} players={gameState.players} />
            )}
            <PlayerPanel
              players={gameState.players}
              currentTurn={gameState.currentTurn}
              turnPhase={gameState.turnPhase}
              dice={gameState.dice}
            />
          </div>
        </div>
      )}
    </div>
  );
}
