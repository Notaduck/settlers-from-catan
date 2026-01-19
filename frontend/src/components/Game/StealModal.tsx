import React from "react";

export interface StealTargetInfo {
  id: string;
  name: string;
  avatarUrl?: string; // Optional for now
}

export interface StealModalProps {
  candidates: StealTargetInfo[];
  onSteal: (id: string) => void;
  onCancel?: () => void;
}

export function StealModal({ candidates, onSteal, onCancel }: StealModalProps) {
  return (
    <div className="modal steal-modal" data-cy="steal-modal">
      <div className="modal-content">
        <h2>Choose Player to Steal From</h2>
        <div className="steal-candidates">
          {candidates.map((c) => (
            <button
              key={c.id}
              data-cy={`steal-player-${c.id}`}
              className="steal-player-btn"
              onClick={() => onSteal(c.id)}
            >
              {/* Avatar if available */}
              {c.avatarUrl && <img src={c.avatarUrl} alt={c.name} className="avatar" />} {c.name}
            </button>
          ))}
        </div>
        {onCancel && <button className="btn btn-secondary" onClick={onCancel}>Cancel</button>}
      </div>
    </div>
  );
}
