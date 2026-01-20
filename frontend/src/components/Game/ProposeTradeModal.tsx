import { useState, useMemo } from "react";
import type { ResourceCount } from "@/types";

const RESOURCE_TYPES = ["wood", "brick", "sheep", "wheat", "ore"] as const;
type ResourceType = typeof RESOURCE_TYPES[number];

const RESOURCE_LABELS: Record<ResourceType, string> = {
  wood: "Wood",
  brick: "Brick",
  sheep: "Sheep",
  wheat: "Wheat",
  ore: "Ore",
};

export interface ProposeTradeModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (offer: ResourceCount, request: ResourceCount, toPlayerId?: string | null) => void;
  players: { id: string; name: string }[];
  myResources: ResourceCount;
}

export function ProposeTradeModal({
  open,
  onClose,
  onSubmit,
  players,
  myResources,
}: ProposeTradeModalProps) {
  const [offering, setOffering] = useState<ResourceCount>({
    wood: 0,
    brick: 0,
    sheep: 0,
    wheat: 0,
    ore: 0,
  });
  const [requesting, setRequesting] = useState<ResourceCount>({
    wood: 0,
    brick: 0,
    sheep: 0,
    wheat: 0,
    ore: 0,
  });
  const [targetPlayerId, setTargetPlayerId] = useState<string | null>(null);

  const offeringTotal = useMemo(
    () => RESOURCE_TYPES.reduce((acc, r) => acc + (offering[r] || 0), 0),
    [offering]
  );

  const requestingTotal = useMemo(
    () => RESOURCE_TYPES.reduce((acc, r) => acc + (requesting[r] || 0), 0),
    [requesting]
  );

  const canSubmit = useMemo(() => {
    // Must offer at least 1 resource
    if (offeringTotal === 0) return false;
    // Must request at least 1 resource
    if (requestingTotal === 0) return false;
    // Must have resources to offer
    return RESOURCE_TYPES.every((r) => (offering[r] || 0) <= (myResources[r] || 0));
  }, [offeringTotal, requestingTotal, offering, myResources]);

  const handleIncOffer = (r: ResourceType) => {
    if ((offering[r] || 0) < (myResources[r] || 0)) {
      setOffering({ ...offering, [r]: (offering[r] || 0) + 1 });
    }
  };

  const handleDecOffer = (r: ResourceType) => {
    if ((offering[r] || 0) > 0) {
      setOffering({ ...offering, [r]: (offering[r] || 0) - 1 });
    }
  };

  const handleIncRequest = (r: ResourceType) => {
    setRequesting({ ...requesting, [r]: (requesting[r] || 0) + 1 });
  };

  const handleDecRequest = (r: ResourceType) => {
    if ((requesting[r] || 0) > 0) {
      setRequesting({ ...requesting, [r]: (requesting[r] || 0) - 1 });
    }
  };

  const handleSubmit = () => {
    if (canSubmit) {
      onSubmit(offering, requesting, targetPlayerId);
      // Reset state
      setOffering({ wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 });
      setRequesting({ wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 });
      setTargetPlayerId(null);
      onClose();
    }
  };

  const handleCancel = () => {
    // Reset state
    setOffering({ wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 });
    setRequesting({ wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 });
    setTargetPlayerId(null);
    onClose();
  };

  if (!open) return null;

  return (
    <div className="modal-overlay" data-cy="propose-trade-modal">
      <div className="modal-content propose-trade-modal">
        <h3>Propose Trade</h3>

        <div className="trade-target-section">
          <label htmlFor="trade-target">Trade with:</label>
          <select
            id="trade-target"
            value={targetPlayerId ?? "all"}
            onChange={(e) => setTargetPlayerId(e.target.value === "all" ? null : e.target.value)}
            data-cy="trade-target-select"
          >
            <option value="all">All Players</option>
            {players.map((player) => (
              <option key={player.id} value={player.id}>
                {player.name}
              </option>
            ))}
          </select>
        </div>

        <div className="trade-sections">
          <div className="trade-section">
            <h4>I Offer</h4>
            <div className="resource-list">
              {RESOURCE_TYPES.map((r) => (
                <div key={r} className="resource-item" data-cy={`trade-offer-${r}`}>
                  <span className="resource-name">{RESOURCE_LABELS[r]}</span>
                  <div className="resource-controls">
                    <button
                      onClick={() => handleDecOffer(r)}
                      disabled={!(offering[r] > 0)}
                      aria-label={`Decrease ${RESOURCE_LABELS[r]}`}
                    >
                      -
                    </button>
                    <span className="resource-count">
                      {offering[r] || 0}/{myResources[r] || 0}
                    </span>
                    <button
                      onClick={() => handleIncOffer(r)}
                      disabled={(offering[r] || 0) >= (myResources[r] || 0)}
                      aria-label={`Increase ${RESOURCE_LABELS[r]}`}
                    >
                      +
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="trade-section">
            <h4>I Request</h4>
            <div className="resource-list">
              {RESOURCE_TYPES.map((r) => (
                <div key={r} className="resource-item" data-cy={`trade-request-${r}`}>
                  <span className="resource-name">{RESOURCE_LABELS[r]}</span>
                  <div className="resource-controls">
                    <button
                      onClick={() => handleDecRequest(r)}
                      disabled={!(requesting[r] > 0)}
                      aria-label={`Decrease ${RESOURCE_LABELS[r]}`}
                    >
                      -
                    </button>
                    <span className="resource-count">{requesting[r] || 0}</span>
                    <button
                      onClick={() => handleIncRequest(r)}
                      aria-label={`Increase ${RESOURCE_LABELS[r]}`}
                    >
                      +
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="trade-summary">
          <p>
            Offering: <strong>{offeringTotal}</strong> resources
          </p>
          <p>
            Requesting: <strong>{requestingTotal}</strong> resources
          </p>
        </div>

        <div className="modal-actions">
          <button
            data-cy="propose-trade-submit-btn"
            onClick={handleSubmit}
            disabled={!canSubmit}
            className="btn btn-primary"
          >
            Propose Trade
          </button>
          <button
            data-cy="propose-trade-cancel-btn"
            onClick={handleCancel}
            className="btn btn-secondary"
          >
            Cancel
          </button>
        </div>

        {!canSubmit && offeringTotal > 0 && requestingTotal > 0 && (
          <p className="error-message" data-cy="propose-trade-error">
            You don't have enough resources to make this offer
          </p>
        )}
        {!canSubmit && offeringTotal === 0 && (
          <p className="info-message">Select resources to offer</p>
        )}
        {!canSubmit && requestingTotal === 0 && offeringTotal > 0 && (
          <p className="info-message">Select resources to request</p>
        )}
      </div>
    </div>
  );
}
