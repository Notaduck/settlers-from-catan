import type { Port as PortType, Resource } from "@/types";

interface PortProps {
  port: PortType;
  x: number; // Midpoint x between two vertices
  y: number; // Midpoint y between two vertices
  index: number;
}

const RESOURCE_LABELS: Record<Resource, string> = {
  0: "", // RESOURCE_UNSPECIFIED
  1: "W", // RESOURCE_WOOD
  2: "B", // RESOURCE_BRICK
  3: "S", // RESOURCE_SHEEP
  4: "Wh", // RESOURCE_WHEAT
  5: "O", // RESOURCE_ORE
};

export function Port({ port, x, y, index }: PortProps) {
  const isGeneric = port.type === 1; // PORT_TYPE_GENERIC
  const label = isGeneric ? "3:1" : "2:1";
  const resourceLabel = !isGeneric && port.resource ? RESOURCE_LABELS[port.resource] : "";

  return (
    <g
      transform={`translate(${x}, ${y})`}
      data-cy={`port-${index}`}
      className="port"
    >
      {/* Port background circle */}
      <circle
        cx="0"
        cy="0"
        r="12"
        fill={isGeneric ? "#8B4513" : "#4682B4"}
        stroke="#333"
        strokeWidth="1.5"
        opacity="0.9"
      />
      
      {/* Trade ratio label */}
      <text
        x="0"
        y={resourceLabel ? "-2" : "1"}
        textAnchor="middle"
        dominantBaseline="middle"
        fontSize="8"
        fontWeight="bold"
        fill="#fff"
      >
        {label}
      </text>
      
      {/* Resource label for specific ports */}
      {resourceLabel && (
        <text
          x="0"
          y="6"
          textAnchor="middle"
          dominantBaseline="middle"
          fontSize="7"
          fontWeight="bold"
          fill="#FFD700"
        >
          {resourceLabel}
        </text>
      )}
    </g>
  );
}
