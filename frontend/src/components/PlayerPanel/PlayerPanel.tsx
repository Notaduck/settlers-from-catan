import type { PlayerState } from "@/types";
import { TurnPhase, PlayerColor } from "@/types";
import { useGame } from "@/context";
import "./PlayerPanel.css";

// Map PlayerColor enum to CSS color strings
const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

interface PlayerPanelProps {
  players: PlayerState[];
  currentTurn: number;
  turnPhase: TurnPhase;
  dice: number[];
}

const RESOURCE_ICONS: Record<string, string> = {
  wood: "🪵",
  brick: "🧱",
  sheep: "🐑",
  wheat: "🌾",
  ore: "�ite",
};

// Helper to display turn phase
function getTurnPhaseName(phase: TurnPhase): string {
  switch (phase) {
    case TurnPhase.ROLL:
      return "Roll";
    case TurnPhase.TRADE:
      return "Trade";
    case TurnPhase.BUILD:
      return "Build";
    default:
      return "Unknown";
  }
}

export function PlayerPanel({
  players,
  currentTurn,
  turnPhase,
  dice,
}: PlayerPanelProps) {
  const { rollDice, endTurn, currentPlayerId } = useGame();

  const currentPlayer = players[currentTurn];
  const isMyTurn = currentPlayer?.id === currentPlayerId;
  const myPlayer = players.find((p) => p.id === currentPlayerId);

  return (
    <div className="player-panel" data-cy="player-panel">
      {/* Dice display */}
      <div className="dice-section" data-cy="dice-section">
        <h3>Dice</h3>
        <div className="dice-display">
          <div className="die" data-cy="die-1">
            {dice[0] || "?"}
          </div>
          <div className="die" data-cy="die-2">
            {dice[1] || "?"}
          </div>
        </div>
        {isMyTurn && turnPhase === TurnPhase.ROLL && (
          <button
            onClick={rollDice}
            className="btn btn-primary"
            data-cy="roll-dice-btn"
          >
            Roll Dice
          </button>
        )}
      </div>

      {/* Current turn indicator */}
      <div className="turn-section" data-cy="turn-section">
        <h3>Current Turn</h3>
        {currentPlayer && (
          <div className="current-player" data-cy="current-player">
            <span
              className="player-color"
              style={{ backgroundColor: PLAYER_COLORS[currentPlayer.color] }}
            />
            <span data-cy="current-player-name">{currentPlayer.name}</span>
            <span className="turn-phase" data-cy="turn-phase">
              ({getTurnPhaseName(turnPhase)})
            </span>
          </div>
        )}
        {isMyTurn && turnPhase !== TurnPhase.ROLL && (
          <button
            onClick={endTurn}
            className="btn btn-secondary"
            data-cy="end-turn-btn"
          >
            End Turn
          </button>
        )}
      </div>

      {/* My resources */}
      {myPlayer && myPlayer.resources && (
        <div className="resources-section" data-cy="resources-section">
          <h3>Your Resources</h3>
          <div className="resources-grid">
            <div className="resource-item" data-cy="resource-wood">
              <span className="resource-icon">{RESOURCE_ICONS["wood"]}</span>
              <span className="resource-count">{myPlayer.resources.wood}</span>
            </div>
            <div className="resource-item" data-cy="resource-brick">
              <span className="resource-icon">{RESOURCE_ICONS["brick"]}</span>
              <span className="resource-count">{myPlayer.resources.brick}</span>
            </div>
            <div className="resource-item" data-cy="resource-sheep">
              <span className="resource-icon">{RESOURCE_ICONS["sheep"]}</span>
              <span className="resource-count">{myPlayer.resources.sheep}</span>
            </div>
            <div className="resource-item" data-cy="resource-wheat">
              <span className="resource-icon">{RESOURCE_ICONS["wheat"]}</span>
              <span className="resource-count">{myPlayer.resources.wheat}</span>
            </div>
            <div className="resource-item" data-cy="resource-ore">
              <span className="resource-icon">{RESOURCE_ICONS["ore"]}</span>
              <span className="resource-count">{myPlayer.resources.ore}</span>
            </div>
          </div>
        </div>
      )}

      {/* All players */}
      <div className="players-section" data-cy="players-section">
        <h3>Players</h3>
        {players.map((player) => (
          <div
            key={player.id}
            className={`player-row ${
              player.id === currentPlayerId ? "is-me" : ""
            }`}
            data-cy={`player-row-${player.id}`}
          >
            <span
              className="player-color"
              style={{ backgroundColor: PLAYER_COLORS[player.color] }}
            />
            <span className="player-name">{player.name}</span>
            <span className="player-vp" data-cy={`player-vp-${player.id}`}>
              {player.victoryPoints} VP
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
