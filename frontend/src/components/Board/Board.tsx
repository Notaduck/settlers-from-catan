import type {
  BoardState,
  Edge as EdgeState,
  HexCoord,
  PlayerState,
  Vertex,
} from "@/types";
import { PlayerColor } from "@/types";
import { HexTile } from "./HexTile";
import { Edge as EdgeSegment } from "./Edge";
import { Vertex as VertexMarker } from "./Vertex";
import "./Board.css";

interface BoardProps {
  board: BoardState;
  players: PlayerState[];
  validVertexIds?: Set<string>;
  validEdgeIds?: Set<string>;
  onBuildSettlement?: (vertexId: string) => void;
  onBuildRoad?: (edgeId: string) => void;
  isRobberMoveMode?: boolean;
  onSelectRobberHex?: (hex: { coord?: { q: number; r: number } }) => void;
}

// Hex size in pixels
const HEX_SIZE = 60;
const VERTEX_MATCH_TOLERANCE = 0.02;

const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

const VERTEX_OFFSETS = [
  { direction: "N", dq: -1 / 3, dr: 2 / 3 },
  { direction: "NE", dq: 1 / 3, dr: 1 / 3 },
  { direction: "SE", dq: 2 / 3, dr: -1 / 3 },
  { direction: "S", dq: 1 / 3, dr: -2 / 3 },
  { direction: "SW", dq: -1 / 3, dr: -1 / 3 },
  { direction: "NW", dq: -2 / 3, dr: 1 / 3 },
];

function axialToPixel(
  q: number,
  r: number,
  size: number
): { x: number; y: number } {
  const x = size * (Math.sqrt(3) * q + (Math.sqrt(3) / 2) * r);
  const y = size * ((3 / 2) * r);
  return { x, y };
}

// Convert axial coordinates to pixel position (pointy-top orientation)
function hexToPixel(coord: HexCoord, size: number): { x: number; y: number } {
  const q = coord.q ?? 0;
  const r = coord.r ?? 0;
  return axialToPixel(q, r, size);
}

// Check if a coord is valid (has defined q and r)
function isValidCoord(coord: HexCoord | undefined): coord is HexCoord {
  return (
    coord !== undefined &&
    typeof coord.q === "number" &&
    typeof coord.r === "number"
  );
}

function parseVertexId(id: string): { q: number; r: number } | null {
  const parts = id.split(",");
  if (parts.length !== 2) {
    return null;
  }
  const q = Number(parts[0]);
  const r = Number(parts[1]);
  if (Number.isNaN(q) || Number.isNaN(r)) {
    return null;
  }
  return { q, r };
}

function getVertexDataCy(
  vertex: Vertex,
  coord: { q: number; r: number }
): string {
  const matches: Array<{ q: number; r: number; direction: string }> = [];

  for (const hex of vertex.adjacentHexes ?? []) {
    if (typeof hex.q !== "number" || typeof hex.r !== "number") {
      continue;
    }
    for (const offset of VERTEX_OFFSETS) {
      const expectedQ = hex.q + offset.dq;
      const expectedR = hex.r + offset.dr;
      if (
        Math.abs(coord.q - expectedQ) < VERTEX_MATCH_TOLERANCE &&
        Math.abs(coord.r - expectedR) < VERTEX_MATCH_TOLERANCE
      ) {
        matches.push({
          q: hex.q,
          r: hex.r,
          direction: offset.direction,
        });
      }
    }
  }

  if (matches.length === 0) {
    return `vertex-${coord.q}-${coord.r}`;
  }

  matches.sort(
    (a, b) => a.q - b.q || a.r - b.r || a.direction.localeCompare(b.direction)
  );

  const chosen = matches[0];
  return `vertex-${chosen.q}-${chosen.r}-${chosen.direction}`;
}

function normalizeCoordValue(value: number): number {
  return Object.is(value, -0) ? 0 : value;
}

function formatEdgeCoord(value: number): string {
  return normalizeCoordValue(value).toFixed(1);
}

function getEdgeDataCy(
  edge: EdgeState,
  v1Coord: { q: number; r: number },
  v2Coord: { q: number; r: number }
): string {
  const ordered = [v1Coord, v2Coord].sort(
    (a, b) => a.q - b.q || a.r - b.r
  );
  const [first, second] = ordered;
  if (!first || !second) {
    return `edge-${edge.id}`;
  }
  return `edge-${formatEdgeCoord(first.q)}-${formatEdgeCoord(
    first.r
  )}-${formatEdgeCoord(second.q)}-${formatEdgeCoord(second.r)}`;
}

export function Board({
  board,
  players,
  validVertexIds,
  validEdgeIds,
  onBuildSettlement,
  onBuildRoad,
  isRobberMoveMode,
  onSelectRobberHex,
}: BoardProps) {
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
  const vertexPositions = board.vertices
    .map((vertex) => {
      const coord = parseVertexId(vertex.id);
      if (!coord) {
        return null;
      }
      const pos = axialToPixel(coord.q, coord.r, HEX_SIZE);
      return { pos, coord, vertex };
    })
    .filter(
      (
        item
      ): item is { pos: { x: number; y: number }; coord: { q: number; r: number }; vertex: Vertex } =>
        item !== null
    );
  const vertexById = new Map(
    vertexPositions.map((item) => [
      item.vertex.id,
      { pos: item.pos, coord: item.coord },
    ])
  );
  const edgePositions = board.edges
    .map((edge) => {
      const [v1Id, v2Id] = edge.vertices ?? [];
      if (!v1Id || !v2Id) {
        return null;
      }
      const v1 = vertexById.get(v1Id);
      const v2 = vertexById.get(v2Id);
      if (!v1 || !v2) {
        return null;
      }
      return { edge, v1, v2 };
    })
    .filter(
      (
        item
      ): item is {
        edge: EdgeState;
        v1: { pos: { x: number; y: number }; coord: { q: number; r: number } };
        v2: { pos: { x: number; y: number }; coord: { q: number; r: number } };
      } => item !== null
    );
  const allPositions = positions.concat(vertexPositions.map((item) => item.pos));
  const minX = Math.min(...allPositions.map((p) => p.x));
  const maxX = Math.max(...allPositions.map((p) => p.x));
  const minY = Math.min(...allPositions.map((p) => p.y));
  const maxY = Math.max(...allPositions.map((p) => p.y));

  const padding = HEX_SIZE * 1.5;
  const width = maxX - minX + padding * 2;
  const height = maxY - minY + padding * 2;
  const offsetX = -minX + padding;
  const offsetY = -minY + padding;

  const robberHex = board.robberHex;
  const playerColors = new Map(
    players.map((player) => [
      player.id,
      PLAYER_COLORS[player.color] ?? "#808080",
    ])
  );

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
              const isSelectable =
                isRobberMoveMode && (!isRobber); // Cannot re-select current robber position
              return (
                <HexTile
                  key={`${coord.q},${coord.r}`}
                  hex={hex}
                  x={pos.x}
                  y={pos.y}
                  size={HEX_SIZE}
                  hasRobber={isRobber || false}
                  isRobberMoveSelectable={!!isSelectable}
                  onSelectRobberHex={isSelectable ? onSelectRobberHex : undefined}
                />
              );
            })}
          {edgePositions.map(({ edge, v1, v2 }) => {
            const roadOwnerId = edge.road?.ownerId;
            const ownerColor = roadOwnerId
              ? playerColors.get(roadOwnerId)
              : undefined;
            const isValid = Boolean(validEdgeIds?.has(edge.id));
            return (
              <EdgeSegment
                key={edge.id}
                edge={edge}
                x1={v1.pos.x}
                y1={v1.pos.y}
                x2={v2.pos.x}
                y2={v2.pos.y}
                ownerColor={ownerColor}
                dataCy={getEdgeDataCy(edge, v1.coord, v2.coord)}
                isValid={isValid}
                onClick={
                  isValid && onBuildRoad ? () => onBuildRoad(edge.id) : undefined
                }
              />
            );
          })}
          {vertexPositions.map(({ pos, coord, vertex }) => {
            const ownerColor = vertex.building
              ? playerColors.get(vertex.building.ownerId)
              : undefined;
            const isValid = Boolean(validVertexIds?.has(vertex.id));
            return (
              <VertexMarker
                key={vertex.id}
                vertex={vertex}
                x={pos.x}
                y={pos.y}
                ownerColor={ownerColor}
                dataCy={getVertexDataCy(vertex, coord)}
                isValid={isValid}
                onClick={
                  isValid && onBuildSettlement
                    ? () => onBuildSettlement(vertex.id)
                    : undefined
                }
              />
            );
          })}
        </g>
      </svg>
    </div>
  );
}
