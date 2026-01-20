import { useState } from 'react';
import type { BoardState, Resource, ResourceCount } from "@/types";
import { PortType } from "@/types";

interface BankTradeModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (offering: Resource, offeringAmount: number, requested: Resource) => void;
  resources: ResourceCount;
  board?: BoardState;
  playerId?: string;
}

const RESOURCE_LABELS: Record<number, string> = {
  1: "Wood",
  2: "Brick",
  3: "Sheep",
  4: "Wheat",
  5: "Ore",
};

const RESOURCE_VALUES: Resource[] = [1, 2, 3, 4, 5]; // WOOD, BRICK, SHEEP, WHEAT, ORE

// Calculate best trade ratio for a player for a specific resource
function getBestTradeRatio(playerId: string | undefined, resource: Resource, board: BoardState | undefined): number {
  if (!playerId || !board || !board.ports) {
    return 4; // Default bank rate
  }

  let best = 4;
  for (const port of board.ports) {
    // Check if player has access to this port (has building on port vertex)
    let hasAccess = false;
    for (const vertexId of port.location) {
      const vertex = board.vertices.find(v => v.id === vertexId);
      if (vertex?.building?.ownerId === playerId) {
        hasAccess = true;
        break;
      }
    }

    if (hasAccess) {
      if (port.type === PortType.GENERIC) {
        best = Math.min(best, 3);
      } else if (port.type === PortType.SPECIFIC && port.resource === resource) {
        best = Math.min(best, 2);
      }
    }
  }

  return best;
}

export function BankTradeModal({ open, onClose, onSubmit, resources, board, playerId }: BankTradeModalProps) {
  const [offering, setOffering] = useState<Resource>(1); // WOOD
  const [requested, setRequested] = useState<Resource>(2); // BRICK

  if (!open) return null;

  const tradeRatio = getBestTradeRatio(playerId, offering, board);
  
  // Get resource amounts as Record
  const resourceAmounts: Record<number, number> = {
    1: resources.wood ?? 0,
    2: resources.brick ?? 0,
    3: resources.sheep ?? 0,
    4: resources.wheat ?? 0,
    5: resources.ore ?? 0,
  };

  const canTrade = resourceAmounts[offering] >= tradeRatio;

  const handleSubmit = () => {
    if (canTrade) {
      onSubmit(offering, tradeRatio, requested);
      onClose();
    }
  };

  return (
    <div className="modal-overlay" data-cy="bank-trade-modal">
      <div className="modal-content">
        <h3>Bank Trade</h3>
        
        <div className="trade-ratio-info" data-cy="trade-ratio-display">
          <p>Current ratio: <strong data-cy={`trade-ratio-${offering}`}>{tradeRatio}:1</strong></p>
          <p className="text-sm">
            {tradeRatio === 4 && "Default bank rate"}
            {tradeRatio === 3 && "Using 3:1 generic port"}
            {tradeRatio === 2 && `Using 2:1 ${RESOURCE_LABELS[offering]} port`}
          </p>
        </div>

        <div className="trade-selection">
          <div className="trade-section">
            <label>Give {tradeRatio}:</label>
            <select
              value={offering}
              onChange={(e) => setOffering(Number(e.target.value) as Resource)}
              data-cy="bank-trade-offering-select"
            >
              {RESOURCE_VALUES.map(res => (
                <option key={res} value={res}>
                  {RESOURCE_LABELS[res]} (have: {resourceAmounts[res]})
                </option>
              ))}
            </select>
          </div>

          <div className="trade-arrow">→</div>

          <div className="trade-section">
            <label>Receive 1:</label>
            <select
              value={requested}
              onChange={(e) => setRequested(Number(e.target.value) as Resource)}
              data-cy="bank-trade-requesting-select"
            >
              {RESOURCE_VALUES.map(res => (
                <option key={res} value={res}>
                  {RESOURCE_LABELS[res]}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="modal-actions">
          <button
            data-cy="bank-trade-submit-btn"
            onClick={handleSubmit}
            disabled={!canTrade}
            className="btn btn-primary"
          >
            Trade
          </button>
          <button
            data-cy="bank-trade-cancel-btn"
            onClick={onClose}
            className="btn btn-secondary"
          >
            Cancel
          </button>
        </div>

        {!canTrade && (
          <p className="error-message" data-cy="bank-trade-error">
            Not enough {RESOURCE_LABELS[offering]} (need {tradeRatio}, have {resourceAmounts[offering]})
          </p>
        )}
      </div>
    </div>
  );
}
