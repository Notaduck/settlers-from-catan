import { useMemo } from "react";
import type { PlayerState } from "@/types";
import { DevCardType } from "@/types";

const DEV_CARD_LABELS: Record<number, string> = {
  [DevCardType.KNIGHT]: "Knight",
  [DevCardType.ROAD_BUILDING]: "Road Building",
  [DevCardType.YEAR_OF_PLENTY]: "Year of Plenty",
  [DevCardType.MONOPOLY]: "Monopoly",
  [DevCardType.VICTORY_POINT]: "Victory Point",
};

const DEV_CARD_KEYS: Record<number, string> = {
  [DevCardType.KNIGHT]: "knight",
  [DevCardType.ROAD_BUILDING]: "road-building",
  [DevCardType.YEAR_OF_PLENTY]: "year-of-plenty",
  [DevCardType.MONOPOLY]: "monopoly",
  [DevCardType.VICTORY_POINT]: "victory-point",
};

export interface DevelopmentCardsPanelProps {
  currentPlayer: PlayerState | null;
  canBuy: boolean;
  canPlay: boolean;
  onBuyCard: () => void;
  onPlayCard: (cardType: DevCardType) => void;
  turnCounter: number;
}

export function DevelopmentCardsPanel({
  currentPlayer,
  canBuy,
  canPlay,
  onBuyCard,
  onPlayCard,
  turnCounter,
}: DevelopmentCardsPanelProps) {
  const devCards = useMemo(() => {
    if (!currentPlayer?.devCards) return [];
    
    return Object.entries(currentPlayer.devCards)
      .map(([typeStr, count]) => {
        const type = parseInt(typeStr, 10);
        return {
          type,
          count: count || 0,
          label: DEV_CARD_LABELS[type] || "Unknown",
          key: DEV_CARD_KEYS[type] || "unknown",
        };
      })
      .filter(card => card.count > 0);
  }, [currentPlayer]);

  const totalCards = currentPlayer?.devCardCount || 0;

  return (
    <div className="dev-cards-panel" data-cy="dev-cards-panel">
      <div className="dev-cards-header">
        <h3>Development Cards ({totalCards})</h3>
        <button
          className="btn btn-small"
          data-cy="buy-dev-card-btn"
          onClick={onBuyCard}
          disabled={!canBuy}
          title={!canBuy ? "Not enough resources or not your turn" : "Buy Development Card (1 Ore, 1 Wheat, 1 Sheep)"}
        >
          Buy Card
        </button>
      </div>
      
      {devCards.length === 0 ? (
        <div className="dev-cards-empty">No cards in hand</div>
      ) : (
        <div className="dev-cards-list">
          {devCards.map(card => (
            <div
              key={card.type}
              className="dev-card-item"
              data-cy={`dev-card-${card.key}`}
            >
              <div className="dev-card-info">
                <span className="dev-card-name">{card.label}</span>
                <span className="dev-card-count">×{card.count}</span>
              </div>
               {card.type !== DevCardType.VICTORY_POINT && (
                 <button
                   className="btn btn-small"
                   data-cy={`play-dev-card-btn-${card.key}`}
                   onClick={() => onPlayCard(card.type as DevCardType)}
                   disabled={!canPlay || (currentPlayer?.devCardsPurchasedTurn?.[card.type] === turnCounter)}
                   title={
                     !canPlay
                       ? "Cannot play dev cards right now"
                       : currentPlayer?.devCardsPurchasedTurn?.[card.type] === turnCounter
                       ? "You may not play a development card in the turn you bought it."
                       : `Play ${card.label}`
                   }
                 >
                   Play
                 </button>
               )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
