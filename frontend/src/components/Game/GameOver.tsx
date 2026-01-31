import { BuildingType, GameOverPayload, GameState, PlayerState } from "@/types";
import "./GameOver.css";

interface GameOverProps {
  gameState: GameState;
  gameOver: GameOverPayload | null;
  onNewGame: () => void;
}

export function GameOver({ gameState, gameOver, onNewGame }: GameOverProps) {
  if (!gameState || (!gameOver && (gameState.status as unknown as string) !== "GAME_STATUS_FINISHED" && gameState.status !== 4)) {
    return null;
  }
  const scoreMap = new Map(
    (gameOver?.scores ?? []).map((score) => [score.playerId, score.points])
  );
  const players = gameState.players || [];
  const fallbackPlayer = players[0];
  const winnerId =
    gameOver?.winnerId ||
    (fallbackPlayer
      ? players.reduce(
          (cur, x) =>
            (x.victoryPoints || 0) > (cur.victoryPoints || 0) ? x : cur,
          fallbackPlayer
        )?.id
      : undefined);

  const getBuildingCounts = (playerId: string) => {
    let settlements = 0;
    let cities = 0;
    for (const vertex of gameState.board?.vertices ?? []) {
      if (vertex.building?.ownerId !== playerId) {
        continue;
      }
      const buildingTypeValue = vertex.building?.type as unknown;
      const buildingTypeName =
        typeof buildingTypeValue === "string"
          ? buildingTypeValue
          : BuildingType[buildingTypeValue as BuildingType];
      if (
        buildingTypeValue === BuildingType.SETTLEMENT ||
        buildingTypeName === "BUILDING_TYPE_SETTLEMENT" ||
        buildingTypeName === "SETTLEMENT"
      ) {
        settlements += 1;
      }
      if (
        buildingTypeValue === BuildingType.CITY ||
        buildingTypeName === "BUILDING_TYPE_CITY" ||
        buildingTypeName === "CITY"
      ) {
        cities += 1;
      }
    }
    return { settlements, cities };
  };

  const getBreakdown = (player: PlayerState) => {
    const { settlements, cities } = getBuildingCounts(player.id);
    return [
      { label: "Settlements", value: settlements },
      { label: "Cities", value: cities },
      {
        label: "Longest Road",
        value: gameState.longestRoadPlayerId === player.id ? 2 : 0,
      },
      {
        label: "Largest Army",
        value: gameState.largestArmyPlayerId === player.id ? 2 : 0,
      },
      { label: "VP Cards", value: player.victoryPointCards ?? 0 },
    ];
  };

  const getTotalVp = (player: PlayerState) => {
    const breakdown = getBreakdown(player);
    return breakdown.reduce((sum, { value }) => sum + (value || 0), 0);
  };

  // Assume each player row has a player.id uniquely present
  return (
    <div className="game-over-overlay" data-cy="game-over-overlay">
      <div className="game-over-modal">
        <h2 data-cy="winner-name">Winner: {players.find(p => p.id === winnerId)?.name ?? "?"}</h2>
        <div className="winner-vp" data-cy="winner-vp">
          VP: {scoreMap.get(winnerId ?? "") ?? (fallbackPlayer ? getTotalVp(players.find(p => p.id === winnerId) ?? fallbackPlayer) : 0)}
        </div>
        <h3>Score Breakdown</h3>
        <table className="score-breakdown">
          <thead>
            <tr><th>Player</th><th>Settlements</th><th>Cities</th><th>Longest Road</th><th>Largest Army</th><th>VP Cards</th><th>Total</th></tr>
          </thead>
          <tbody>
            {players.map(player => {
              const breakdown = getBreakdown(player);
              const total = scoreMap.get(player.id) ?? getTotalVp(player);
              return (
                <tr key={player.id} data-cy={`final-score-${player.id}`}
                  className={player.id === winnerId ? "winner-row" : ""}>
                  <td>{player.name}</td>
                  {breakdown.map(b => <td key={b.label}>{b.value}</td>)}
                  <td>{total}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <button className="btn btn-primary" data-cy="new-game-btn" onClick={onNewGame}>
          New Game
        </button>
      </div>
    </div>
  );
}
