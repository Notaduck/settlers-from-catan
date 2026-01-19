import type { Vertex } from "@/types";
import { BuildingType } from "@/types";

interface VertexProps {
  vertex: Vertex;
  x: number;
  y: number;
  ownerColor?: string;
  dataCy?: string;
  isValid?: boolean;
  onClick?: () => void;
}

export function Vertex({
  vertex,
  x,
  y,
  ownerColor,
  dataCy,
  isValid,
  onClick,
}: VertexProps) {
  const building = vertex.building;
  const hasBuilding = Boolean(building);
  const fillColor = ownerColor ?? "#6b6b6b";

  return (
    <g
      transform={`translate(${x}, ${y})`}
      className={`vertex ${hasBuilding ? "vertex--occupied" : "vertex--empty"} ${
        isValid ? "vertex--valid" : ""
      }`}
      data-cy={dataCy}
      onClick={onClick}
    >
      {!hasBuilding && <circle r={3.2} className="vertex-dot" />}
      {building?.type === BuildingType.SETTLEMENT && (
        <polygon
          points="0,-8 7,-2 7,8 -7,8 -7,-2"
          fill={fillColor}
          stroke="#222"
          strokeWidth="1"
          className="vertex-settlement"
        />
      )}
      {building?.type === BuildingType.CITY && (
        <g className="vertex-city" stroke="#222" strokeWidth="1">
          <rect x={-8} y={-6} width={16} height={12} fill={fillColor} />
          <rect x={-4} y={-12} width={8} height={6} fill={fillColor} />
        </g>
      )}
    </g>
  );
}
