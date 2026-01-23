import type { BoardState, PlayerState } from "@/types";
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
  board?: BoardState | null;
  currentTurn: number;
  turnPhase: TurnPhase;
  dice: number[];
  gameStatus?: string;
  isGameOver?: boolean;
  longestRoadPlayerId?: string | null;
}

const RESOURCE_ICONS: Record<string, string> = {
  wood: "🪵",
  brick: "🧱",
  sheep: "🐑",
  wheat: "🌾",
  ore: "�ite",
};

function getLongestRoadLengths(
  board: BoardState | null | undefined,
  players: PlayerState[]
): Record<string, number> {
  const lengths: Record<string, number> = {};
  if (!board) {
    for (const player of players) {
      lengths[player.id] = 0;
    }
    return lengths;
  }

  const vertexToEdges = new Map<string, string[]>();
  const edgeVertices = new Map<string, [string, string]>();
  const edges = board.edges ?? [];

  for (const edge of edges) {
    const [v1, v2] = edge.vertices ?? [];
    if (!v1 || !v2) {
      continue;
    }
    edgeVertices.set(edge.id, [v1, v2]);
    if (!vertexToEdges.has(v1)) {
      vertexToEdges.set(v1, []);
    }
    if (!vertexToEdges.has(v2)) {
      vertexToEdges.set(v2, []);
    }
    vertexToEdges.get(v1)?.push(edge.id);
    vertexToEdges.get(v2)?.push(edge.id);
  }

  const blockedVerticesByPlayer = new Map<string, Set<string>>();
  for (const player of players) {
    blockedVerticesByPlayer.set(player.id, new Set());
  }

  for (const vertex of board.vertices ?? []) {
    const ownerId = vertex.building?.ownerId;
    if (!ownerId) {
      continue;
    }
    for (const player of players) {
      if (player.id !== ownerId) {
        blockedVerticesByPlayer.get(player.id)?.add(vertex.id);
      }
    }
  }

  for (const player of players) {
    const playerEdges = edges.filter(
      (edge) => edge.road?.ownerId === player.id
    );
    if (playerEdges.length === 0) {
      lengths[player.id] = 0;
      continue;
    }

    const playerEdgeIds = new Set(playerEdges.map((edge) => edge.id));
    const blockedVertices = blockedVerticesByPlayer.get(player.id) ?? new Set();

    const dfs = (edgeId: string, visited: Set<string>): number => {
      visited.add(edgeId);
      let maxLength = 1;
      const vertices = edgeVertices.get(edgeId) ?? [];
      for (const vertexId of vertices) {
        if (blockedVertices.has(vertexId)) {
          continue;
        }
        const connectedEdges = vertexToEdges.get(vertexId) ?? [];
        for (const nextEdgeId of connectedEdges) {
          if (nextEdgeId === edgeId) {
            continue;
          }
          if (!playerEdgeIds.has(nextEdgeId)) {
            continue;
          }
          if (visited.has(nextEdgeId)) {
            continue;
          }
          const length = 1 + dfs(nextEdgeId, visited);
          if (length > maxLength) {
            maxLength = length;
          }
        }
      }
      visited.delete(edgeId);
      return maxLength;
    };

    let longest = 0;
    for (const edge of playerEdges) {
      const length = dfs(edge.id, new Set());
      if (length > longest) {
        longest = length;
      }
    }
    lengths[player.id] = longest;
  }

  return lengths;
}

function isTurnPhase(
  phase: TurnPhase | string | undefined,
  expected: TurnPhase,
  expectedString: string
): boolean {
  return phase === expected || phase === (expectedString as unknown as TurnPhase);
}

// Helper to display turn phase
function getTurnPhaseName(phase: TurnPhase | string | undefined): string {
  if (isTurnPhase(phase, TurnPhase.ROLL, "TURN_PHASE_ROLL")) {
    return "Roll";
  }
  if (isTurnPhase(phase, TurnPhase.TRADE, "TURN_PHASE_TRADE")) {
    return "Trade";
  }
  if (isTurnPhase(phase, TurnPhase.BUILD, "TURN_PHASE_BUILD")) {
    return "Build";
  }
  return "Unknown";
}

export function PlayerPanel({
  players,
  currentTurn,
  turnPhase,
  dice,
  gameStatus,
  isGameOver = false,
}: PlayerPanelProps) {
  const { rollDice, endTurn, currentPlayerId } = useGame();

  const currentPlayer = players[currentTurn];
  const isMyTurn = currentPlayer?.id === currentPlayerId;
  const isRollPhase = isTurnPhase(turnPhase, TurnPhase.ROLL, "TURN_PHASE_ROLL");
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
        {dice.length > 0 && dice[0] && dice[1] && (
          <div className="dice-result" data-cy="dice-result">
            Total: {dice[0] + dice[1]}
          </div>
        )}
         {isMyTurn && isRollPhase && !isGameOver && (gameStatus === undefined || gameStatus === "PLAYING" || gameStatus === "GAME_STATUS_PLAYING") && (
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
        {isMyTurn && !isRollPhase && !isGameOver && (
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
