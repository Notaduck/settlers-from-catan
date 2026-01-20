import type { Edge as EdgeType } from "@/types";

interface EdgeProps {
  edge: EdgeType;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  ownerColor?: string;
  dataCy?: string;
  isValid?: boolean;
  onClick?: () => void;
}

export function Edge({
  edge,
  x1,
  y1,
  x2,
  y2,
  ownerColor,
  dataCy,
  isValid,
  onClick,
}: EdgeProps) {
  const hasRoad = Boolean(edge.road);
  const strokeColor = ownerColor ?? "#d6d6d6";

  return (
    <g
      className={`edge ${hasRoad ? "edge--occupied" : "edge--empty"} ${
        isValid ? "edge--valid" : ""
      }`}
      data-cy={dataCy}
      onClick={onClick}
    >
      <line
        x1={x1}
        y1={y1}
        x2={x2}
        y2={y2}
        className={hasRoad ? "edge-road" : "edge-line"}
        stroke={hasRoad ? strokeColor : undefined}
      />
    </g>
  );
}
