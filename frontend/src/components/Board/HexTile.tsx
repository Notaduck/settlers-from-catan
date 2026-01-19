import type { Hex } from "@/types";
import { TileResource } from "@/types";

interface HexTileProps {
  hex: Hex;
  x: number;
  y: number;
  size: number;
  hasRobber: boolean;
}

// Color mapping for tile resources
const RESOURCE_COLORS: Record<TileResource, string> = {
  [TileResource.UNSPECIFIED]: "#808080",
  [TileResource.WOOD]: "#228B22",
  [TileResource.BRICK]: "#CD5C5C",
  [TileResource.SHEEP]: "#90EE90",
  [TileResource.WHEAT]: "#FFD700",
  [TileResource.ORE]: "#708090",
  [TileResource.DESERT]: "#DEB887",
};

// Human-readable resource names
const RESOURCE_NAMES: Record<TileResource, string> = {
  [TileResource.UNSPECIFIED]: "",
  [TileResource.WOOD]: "Wood",
  [TileResource.BRICK]: "Brick",
  [TileResource.SHEEP]: "Sheep",
  [TileResource.WHEAT]: "Wheat",
  [TileResource.ORE]: "Ore",
  [TileResource.DESERT]: "Desert",
};

// Generate hexagon points for pointy-top orientation
function getHexPoints(size: number): string {
  const points: string[] = [];
  for (let i = 0; i < 6; i++) {
    const angle = (Math.PI / 3) * i - Math.PI / 6;
    const px = size * Math.cos(angle);
    const py = size * Math.sin(angle);
    points.push(`${px},${py}`);
  }
  return points.join(" ");
}

export function HexTile({ hex, x, y, size, hasRobber }: HexTileProps) {
  const color = RESOURCE_COLORS[hex.resource];
  const resourceName = RESOURCE_NAMES[hex.resource];
  const points = getHexPoints(size * 0.95);

  // Determine if this number is "good" (6 or 8)
  const isHighProbability = hex.number === 6 || hex.number === 8;
  const isDesert = hex.resource === TileResource.DESERT;

  // Create a hex ID from coordinates
  const hexId = hex.coord ? `hex-${hex.coord.q}-${hex.coord.r}` : "hex-unknown";

  return (
    <g transform={`translate(${x}, ${y})`} className="hex-tile" data-cy={hexId}>
      {/* Hex background */}
      <polygon
        points={points}
        fill={color}
        stroke="#333"
        strokeWidth="2"
        className="hex-polygon"
      />

      {/* Resource name label at top */}
      <text
        x="0"
        y={-size * 0.55}
        textAnchor="middle"
        fontSize={size * 0.18}
        fontWeight="bold"
        fill="#fff"
        stroke="#333"
        strokeWidth="0.5"
        paintOrder="stroke"
      >
        {resourceName}
      </text>

      {/* Number token (not for desert) */}
      {hex.number > 0 && (
        <>
          <circle
            cx="0"
            cy={size * 0.1}
            r={size * 0.3}
            fill="#f5f5dc"
            stroke="#333"
            strokeWidth="1"
          />
          <text
            x="0"
            y={size * 0.1}
            textAnchor="middle"
            dominantBaseline="central"
            fontSize={size * 0.28}
            fontWeight={isHighProbability ? "bold" : "normal"}
            fill={isHighProbability ? "#c41e3a" : "#333"}
          >
            {hex.number}
          </text>
          {/* Probability dots */}
          <text
            x="0"
            y={size * 0.28}
            textAnchor="middle"
            fontSize={size * 0.1}
            fill="#666"
          >
            {"•".repeat(6 - Math.abs(7 - hex.number))}
          </text>
        </>
      )}

      {/* Robber */}
      {hasRobber && (
        <g className="robber">
          <ellipse
            cx="0"
            cy={isDesert ? -size * 0.05 : size * 0.1}
            rx={size * 0.15}
            ry={size * 0.25}
            fill="#222"
          />
          <circle
            cx="0"
            cy={isDesert ? -size * 0.3 : -size * 0.15}
            r={size * 0.12}
            fill="#222"
          />
        </g>
      )}
    </g>
  );
}
