import { useState, useMemo } from "react";
import type { ResourceCount } from "@/types";

const RESOURCE_TYPES = ["wood", "brick", "sheep", "wheat", "ore"] as const;
type ResourceType = typeof RESOURCE_TYPES[number];

export interface DiscardModalProps {
  requiredCount: number;
  maxAvailable: ResourceCount;
  onDiscard: (toDiscard: ResourceCount) => void;
  onClose?: () => void;
}

export function DiscardModal({ requiredCount, maxAvailable, onDiscard, onClose }: DiscardModalProps) {
  const [selected, setSelected] = useState<ResourceCount>({ wood: 0, brick: 0, sheep: 0, wheat: 0, ore: 0 });
  const countSelected = useMemo(
    () => RESOURCE_TYPES.reduce((acc, r) => acc + (selected[r] || 0), 0),
    [selected]
  );
  const canSubmit = countSelected === requiredCount && RESOURCE_TYPES.every(r => (selected[r] || 0) <= (maxAvailable[r] || 0));
  const handleInc = (r: ResourceType) => {
    if ((selected[r] || 0) < (maxAvailable[r] || 0) && countSelected < requiredCount) {
      setSelected({ ...selected, [r]: (selected[r] || 0) + 1 });
    }
  };
  const handleDec = (r: ResourceType) => {
    if ((selected[r] || 0) > 0) {
      setSelected({ ...selected, [r]: (selected[r] || 0) - 1 });
    }
  };
  const handleSubmit = () => { if (canSubmit) onDiscard(selected); };
  return (
    <div className="modal discard-modal" data-cy="discard-modal">
      <div className="modal-content">
        <h2>Discard {requiredCount} Cards</h2>
        <div className="resource-list">
          {RESOURCE_TYPES.map((r) => (
            <div key={r} className="resource-item" data-cy={`discard-card-${r}`}>
              <span>{r.charAt(0).toUpperCase() + r.slice(1)}: {(selected[r] || 0)}/{maxAvailable[r] || 0}</span>
              <button onClick={() => handleDec(r)} disabled={!(selected[r] > 0)}>-</button>
              <button onClick={() => handleInc(r)} disabled={countSelected >= requiredCount || (selected[r] || 0) >= (maxAvailable[r] || 0)}>+</button>
            </div>
          ))}
        </div>
        <button data-cy="discard-submit" disabled={!canSubmit} onClick={handleSubmit} className="btn btn-primary">
          Discard
        </button>
        {onClose && <button className="btn btn-secondary" onClick={onClose}>Cancel</button>}
      </div>
    </div>
  );
}
