import React from "react";
import { GameState, PlayerState } from "@/types";

interface GameOverProps {
  gameState: GameState;
  onNewGame: () => void;
}

export function GameOver({ gameState, onNewGame }: GameOverProps) {
  if (!gameState || (gameState.status !== "GAME_STATUS_FINISHED" && gameState.status !== 4)) {
    return null;
  }
  // Find the winner and compute totals
  const players = gameState.players || [];
  // Winner id may come from GameState.winnerId or derive from max victory points
  // @ts-expect-error (winnerId possibly not in types generated)
  const winnerId: string = (gameState.winnerId as string | undefined) || (players.reduce((cur, x) => (x.victoryPoints || 0) > (cur.victoryPoints || 0) ? x : cur, players[0])?.id);

  const getBreakdown = (player: PlayerState) => {
    return [
      { label: "Settlements", value: player.settlementCount ?? player.settlements ?? 0 },
      { label: "Cities", value: player.cityCount ?? player.cities ?? 0 },
      { label: "Longest Road", value: player.hasLongestRoad ? 2 : 0 },
      { label: "Largest Army", value: player.hasLargestArmy ? 2 : 0 },
      { label: "VP Cards", value: player.vpDevCard ?? player.victoryPointCards ?? 0 },
    ];
  };

  const getTotalVp = (player: PlayerState) => {
    let vp = 0;
    const breakdown = getBreakdown(player);
    vp = breakdown.reduce((sum, { value }) => sum + (value || 0), 0);
    return vp;
  };

  // Assume each player row has a player.id uniquely present
  return (
    <div className="game-over-overlay" data-cy="game-over-overlay">
      <div className="game-over-modal">
        <h2 data-cy="winner-name">Winner: {players.find(p => p.id === winnerId)?.name ?? "?"}</h2>
        <div className="winner-vp" data-cy="winner-vp">
          VP: {players.find(p => p.id === winnerId)?.victoryPoints ?? getTotalVp(players.find(p => p.id === winnerId) ?? players[0])}
        </div>
        <h3>Score Breakdown</h3>
        <table className="score-breakdown">
          <thead>
            <tr><th>Player</th><th>Settlements</th><th>Cities</th><th>Longest Road</th><th>Largest Army</th><th>VP Cards</th><th>Total</th></tr>
          </thead>
          <tbody>
            {players.map(player => {
              const breakdown = getBreakdown(player);
              const total = getTotalVp(player);
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
