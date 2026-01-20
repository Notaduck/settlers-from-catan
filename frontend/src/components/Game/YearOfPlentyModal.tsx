import { useState } from "react";
import { Resource } from "@/types";

const RESOURCES = [
  { key: Resource.WOOD, label: "Wood", dataCy: "wood" },
  { key: Resource.BRICK, label: "Brick", dataCy: "brick" },
  { key: Resource.SHEEP, label: "Sheep", dataCy: "sheep" },
  { key: Resource.WHEAT, label: "Wheat", dataCy: "wheat" },
  { key: Resource.ORE, label: "Ore", dataCy: "ore" },
];

export interface YearOfPlentyModalProps {
  open: boolean;
  onClose: () => void;
  onSubmit: (resources: Resource[]) => void;
}

export function YearOfPlentyModal({ open, onClose, onSubmit }: YearOfPlentyModalProps) {
  const [selected, setSelected] = useState<Resource[]>([]);

  const handleSelectResource = (resource: Resource) => {
    if (selected.length < 2) {
      setSelected([...selected, resource]);
    }
  };

  const handleRemove = (index: number) => {
    setSelected(selected.filter((_, i) => i !== index));
  };

  const handleSubmit = () => {
    if (selected.length === 2) {
      onSubmit(selected);
      setSelected([]);
      onClose();
    }
  };

  const handleCancel = () => {
    setSelected([]);
    onClose();
  };

  const canSubmit = selected.length === 2;

  if (!open) return null;

  return (
    <div className="modal year-of-plenty-modal" data-cy="year-of-plenty-modal">
      <div className="modal-content">
        <h2>Year of Plenty</h2>
        <p>Select 2 resources from the bank</p>
        
        <div className="selected-resources">
          <h3>Selected ({selected.length}/2)</h3>
          <div className="selected-list">
            {selected.map((resource, index) => {
              const resourceInfo = RESOURCES.find(r => r.key === resource);
              return (
                <div key={index} className="selected-item">
                  <span>{resourceInfo?.label || "Unknown"}</span>
                  <button onClick={() => handleRemove(index)} className="btn btn-small">Remove</button>
                </div>
              );
            })}
            {selected.length === 0 && <div className="empty-hint">No resources selected yet</div>}
          </div>
        </div>

        <div className="resource-selector">
          <h3>Available Resources</h3>
          <div className="resource-list">
            {RESOURCES.map(resource => (
              <button
                key={resource.key}
                className="resource-btn"
                data-cy={`year-of-plenty-select-${resource.dataCy}`}
                onClick={() => handleSelectResource(resource.key)}
                disabled={selected.length >= 2}
              >
                {resource.label}
              </button>
            ))}
          </div>
        </div>

        <div className="modal-actions">
          <button
            className="btn btn-primary"
            data-cy="year-of-plenty-submit"
            onClick={handleSubmit}
            disabled={!canSubmit}
          >
            Take Resources
          </button>
          <button className="btn btn-secondary" onClick={handleCancel}>
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
