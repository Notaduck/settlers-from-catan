import type { BoardState, HexCoord } from "@/types";
import { HexTile } from "./HexTile";
import "./Board.css";

interface BoardProps {
  board: BoardState;
}

// Hex size in pixels
const HEX_SIZE = 60;

// Convert axial coordinates to pixel position (pointy-top orientation)
function hexToPixel(coord: HexCoord, size: number): { x: number; y: number } {
  const q = coord.q ?? 0;
  const r = coord.r ?? 0;
  const x = size * (Math.sqrt(3) * q + (Math.sqrt(3) / 2) * r);
  const y = size * ((3 / 2) * r);
  return { x, y };
}

// Check if a coord is valid (has defined q and r)
function isValidCoord(coord: HexCoord | undefined): coord is HexCoord {
  return (
    coord !== undefined &&
    typeof coord.q === "number" &&
    typeof coord.r === "number"
  );
}

export function Board({ board }: BoardProps) {
  // Filter hexes with valid coords
  const validHexes = board.hexes.filter((hex) => isValidCoord(hex.coord));

  if (validHexes.length === 0) {
    return (
      <div className="board-container" data-cy="board-loading">
        <p>Loading board...</p>
      </div>
    );
  }

  // Calculate board dimensions
  const positions = validHexes.map((hex) => hexToPixel(hex.coord!, HEX_SIZE));
  const minX = Math.min(...positions.map((p) => p.x));
  const maxX = Math.max(...positions.map((p) => p.x));
  const minY = Math.min(...positions.map((p) => p.y));
  const maxY = Math.max(...positions.map((p) => p.y));

  const padding = HEX_SIZE * 1.5;
  const width = maxX - minX + padding * 2;
  const height = maxY - minY + padding * 2;
  const offsetX = -minX + padding;
  const offsetY = -minY + padding;

  const robberHex = board.robberHex;

  return (
    <div className="board-container" data-cy="board">
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="board-svg"
        preserveAspectRatio="xMidYMid meet"
        data-cy="board-svg"
      >
        <g transform={`translate(${offsetX}, ${offsetY})`}>
          {validHexes.map((hex) => {
            const coord = hex.coord!;
            const pos = hexToPixel(coord, HEX_SIZE);
            const isRobber =
              robberHex && coord.q === robberHex.q && coord.r === robberHex.r;
            return (
              <HexTile
                key={`${coord.q},${coord.r}`}
                hex={hex}
                x={pos.x}
                y={pos.y}
                size={HEX_SIZE}
                hasRobber={isRobber || false}
              />
            );
          })}
        </g>
      </svg>
    </div>
  );
}
