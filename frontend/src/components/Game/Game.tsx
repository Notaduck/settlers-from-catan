import React, { useState } from "react";

interface GameProps {
  gameCode: string;
  onLeave: () => void;
}

export function Game({ gameCode, onLeave }: GameProps) {
  // TODO: Replace with real context/props/state in final recovery
  const [isReady, setReady] = useState(false);
  const allPlayersReady = true; // placeholder
  const startGame = () => {};

  return (
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
      <button
        onClick={startGame}
        className="btn btn-primary"
        disabled={!allPlayersReady}
        title={!allPlayersReady ? "All players must be ready to start" : "Start the game"}
        data-cy="start-game-btn"
      >
        Start Game
      </button>
    </div>
  );
}
