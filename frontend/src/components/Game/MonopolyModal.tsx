import { useState } from "react";
import { Resource } from "@/types";

const RESOURCES = [
  { key: Resource.WOOD, label: "Wood", dataCy: "wood" },
  { key: Resource.BRICK, label: "Brick", dataCy: "brick" },
  { key: Resource.SHEEP, label: "Sheep", dataCy: "sheep" },
  { key: Resource.WHEAT, label: "Wheat", dataCy: "wheat" },
  { key: Resource.ORE, label: "Ore", dataCy: "ore" },
];

export interface MonopolyModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (resource: Resource) => void;
}

export function MonopolyModal({ open, onClose, onSubmit }: MonopolyModalProps) {
  const [selected, setSelected] = useState<Resource | null>(null);

  const handleSubmit = () => {
  if (selected !== null) {
    console.log('[MonopolyModal] handleSubmit called', selected);
    onSubmit(selected);
    setSelected(null);
    onClose();
  }
};

  const handleCancel = () => {
  console.log('[MonopolyModal] cancelled by user');
  setSelected(null);
  onClose();
};

  if (!open) return null;

  return (
    <div className="modal monopoly-modal" data-cy="monopoly-modal">
      <div className="modal-content">
        <h2>Monopoly</h2>
        <p>Select a resource type to collect from all other players</p>
        
        <div className="resource-selector">
          {RESOURCES.map(resource => (
            <button
              key={resource.key}
              className={`resource-btn ${selected === resource.key ? 'selected' : ''}`}
              data-cy={`monopoly-select-${resource.dataCy}`}
              onClick={() => setSelected(resource.key)}
            >
              {resource.label}
            </button>
          ))}
        </div>

        <div className="modal-actions">
          <button
            className="btn btn-primary"
            data-cy="monopoly-submit"
            onClick={handleSubmit}
            disabled={selected === null}
          >
            Collect Resource
          </button>
          <button className="btn btn-secondary" onClick={handleCancel}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
