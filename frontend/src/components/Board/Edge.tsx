import type { Edge } from "@/types";

interface EdgeProps {
  edge: Edge;
  x1: number;
  y1: number;
  x2: number;
  y2: number;
  ownerColor?: string;
  dataCy?: string;
}

export function Edge({ edge, x1, y1, x2, y2, ownerColor, dataCy }: EdgeProps) {
  const hasRoad = Boolean(edge.road);
  const strokeColor = ownerColor ?? "#d6d6d6";

  return (
    <g
      className={`edge ${hasRoad ? "edge--occupied" : "edge--empty"}`}
      data-cy={dataCy}
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
