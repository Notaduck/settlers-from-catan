import { useMemo } from "react";

const RESOURCE_TYPES = ["wood", "brick", "sheep", "wheat", "ore"] as const;
type ResourceType = typeof RESOURCE_TYPES[number];

const RESOURCE_LABELS: Record<ResourceType, string> = {
  wood: "Wood",
  brick: "Brick",
  sheep: "Sheep",
  wheat: "Wheat",
  ore: "Ore",
};

export interface IncomingTradeModalProps {
  open: boolean;
  onAccept: () => void;
  onDecline: () => void;
  fromPlayer: string;
  offer: Record<string, number>;
  request: Record<string, number>;
}

export function IncomingTradeModal({
  open,
  onAccept,
  onDecline,
  fromPlayer,
  offer,
  request,
}: IncomingTradeModalProps) {
  const offerTotal = useMemo(
    () => RESOURCE_TYPES.reduce((acc, r) => acc + (offer[r] || 0), 0),
    [offer]
  );

  const requestTotal = useMemo(
    () => RESOURCE_TYPES.reduce((acc, r) => acc + (request[r] || 0), 0),
    [request]
  );

  if (!open) return null;

  return (
    <div className="modal-overlay" data-cy="incoming-trade-modal">
      <div className="modal-content incoming-trade-modal">
        <h3>Trade Offer from {fromPlayer}</h3>
        <p className="trade-description">
          {fromPlayer} wants to trade with you
        </p>

        <div className="trade-sections">
          <div className="trade-section">
            <h4>They Offer</h4>
            <div className="resource-list">
              {RESOURCE_TYPES.map((r) => {
                const count = offer[r] || 0;
                if (count === 0) return null;
                return (
                  <div
                    key={r}
                    className="resource-item"
                    data-cy={`incoming-offer-${r}`}
                  >
                    <span className="resource-name">{RESOURCE_LABELS[r]}</span>
                    <span className="resource-count">{count}</span>
                  </div>
                );
              })}
            </div>
            {offerTotal === 0 && (
              <p className="empty-message">No resources offered</p>
            )}
          </div>

          <div className="trade-arrow">
            <span>⇄</span>
          </div>

          <div className="trade-section">
            <h4>They Request</h4>
            <div className="resource-list">
              {RESOURCE_TYPES.map((r) => {
                const count = request[r] || 0;
                if (count === 0) return null;
                return (
                  <div
                    key={r}
                    className="resource-item"
                    data-cy={`incoming-request-${r}`}
                  >
                    <span className="resource-name">{RESOURCE_LABELS[r]}</span>
                    <span className="resource-count">{count}</span>
                  </div>
                );
              })}
            </div>
            {requestTotal === 0 && (
              <p className="empty-message">No resources requested</p>
            )}
          </div>
        </div>

        <div className="trade-summary">
          <p>
            You give: <strong>{requestTotal}</strong> resources
          </p>
          <p>
            You receive: <strong>{offerTotal}</strong> resources
          </p>
        </div>

        <div className="modal-actions">
          <button
            data-cy="accept-trade-btn"
            onClick={onAccept}
            className="btn btn-primary"
          >
            Accept Trade
          </button>
          <button
            data-cy="decline-trade-btn"
            onClick={onDecline}
            className="btn btn-secondary"
          >
            Decline
          </button>
        </div>

        <p className="help-text">
          Accepting this trade will exchange resources immediately.
        </p>
      </div>
    </div>
  );
}
