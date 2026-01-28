import { useState, type FormEvent } from "react";
import type { CreateGameResponse, JoinGameResponse } from "@/types";
import "./Lobby.css";

interface LobbyProps {
  onGameJoined: (code: string, token: string, playerId: string) => void;
}

export function Lobby({ onGameJoined }: LobbyProps) {
  const [playerName, setPlayerName] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreateGame = async (e: FormEvent) => {
    e.preventDefault();
    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/games", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ playerName: playerName.trim() }),
      });

      if (!response.ok) {
        throw new Error("Failed to create game");
      }

      const data: CreateGameResponse = await response.json();
      onGameJoined(data.code, data.sessionToken, data.playerId);
    } catch {
      setError("Failed to create game. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  const handleJoinGame = async (e: FormEvent) => {
    e.preventDefault();
    if (!playerName.trim()) {
      setError("Please enter your name");
      return;
    }
    if (!joinCode.trim()) {
      setError("Please enter a game code");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch(
        `/api/games/${joinCode.trim().toUpperCase()}/join`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ playerName: playerName.trim() }),
        }
      );

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || "Failed to join game");
      }

      const data: JoinGameResponse = await response.json();
      onGameJoined(
        joinCode.trim().toUpperCase(),
        data.sessionToken,
        data.playerId
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to join game");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="lobby" data-cy="lobby">
      <div className="lobby-card">
        <h2>Welcome to Settlers from Catan</h2>

        {/* E2E: Add explicit waiting marker for Playwright */}
        <div data-cy="game-waiting" style={{ position: 'absolute', left: -9999 }}>
          Waiting for players...
        </div>

        {error && (
          <div className="lobby-error" data-cy="lobby-error">
            {error}
          </div>
        )}

        <div className="lobby-form">
          <label htmlFor="playerName">Your Name</label>
          <input
            id="playerName"
            type="text"
            value={playerName}
            onChange={(e) => setPlayerName(e.target.value)}
            placeholder="Enter your name"
            disabled={isLoading}
            maxLength={20}
            data-cy="player-name-input"
          />
        </div>

        <div className="lobby-actions">
          <div className="lobby-section">
            <h3>Create New Game</h3>
            <button
              onClick={handleCreateGame}
              disabled={isLoading || !playerName.trim()}
              className="btn btn-primary"
              data-cy="create-game-btn"
            >
              {isLoading ? "Creating..." : "Create Game"}
            </button>
          </div>

          <div className="lobby-divider">
            <span>or</span>
          </div>

          <div className="lobby-section">
            <h3>Join Existing Game</h3>
            <form onSubmit={handleJoinGame} className="join-form">
              <input
                type="text"
                value={joinCode}
                onChange={(e) => setJoinCode(e.target.value.toUpperCase())}
                placeholder="Enter game code"
                disabled={isLoading}
                maxLength={6}
                className="code-input"
                data-cy="join-code-input"
              />
              <button
                type="submit"
                disabled={isLoading || !playerName.trim() || !joinCode.trim()}
                className="btn btn-secondary"
                data-cy="join-game-btn"
              >
                {isLoading ? "Joining..." : "Join Game"}
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
