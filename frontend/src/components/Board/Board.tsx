import type {
  BoardState,
  Edge as EdgeState,
  Hex,
  HexCoord,
  PlayerState,
  Port as PortState,
  Vertex,
} from "@/types";
import { BuildingType, PlayerColor, PortType, TileResource } from "@/types";
import { Canvas, useFrame } from "@react-three/fiber";
import {
  Html,
  Instance,
  Instances,
  OrbitControls,
  useCursor,
} from "@react-three/drei";
import * as THREE from "three";
import { useMemo, useRef, useState } from "react";
import { HexTile } from "./HexTile";
import { Edge as EdgeSegment } from "./Edge";
import { Vertex as VertexMarker } from "./Vertex";
import { Port } from "./Port";
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

const HEX_SIZE = 1.2;
const HEX_SIZE_2D = 60;
const TILE_RADIUS = HEX_SIZE * 0.92;
const EDGE_Y = 0.08;
const VERTEX_Y = 0.1;
const VERTEX_MATCH_TOLERANCE = 0.02;

const PLAYER_COLORS: Record<PlayerColor, string> = {
  [PlayerColor.UNSPECIFIED]: "#808080",
  [PlayerColor.RED]: "#e74c3c",
  [PlayerColor.BLUE]: "#3498db",
  [PlayerColor.GREEN]: "#2ecc71",
  [PlayerColor.ORANGE]: "#e67e22",
};

const RESOURCE_CONFIG: Record<
  TileResource,
  { color: string; height: number; label: string }
> = {
  [TileResource.UNSPECIFIED]: {
    color: "#808080",
    height: 0.2,
    label: "",
  },
  [TileResource.WOOD]: {
    color: "#2f8f3b",
    height: 0.7,
    label: "Wood",
  },
  [TileResource.BRICK]: {
    color: "#b0543f",
    height: 0.5,
    label: "Brick",
  },
  [TileResource.SHEEP]: {
    color: "#9fd99a",
    height: 0.28,
    label: "Sheep",
  },
  [TileResource.WHEAT]: {
    color: "#f2c14e",
    height: 0.32,
    label: "Wheat",
  },
  [TileResource.ORE]: {
    color: "#7f8c8d",
    height: 0.85,
    label: "Ore",
  },
  [TileResource.DESERT]: {
    color: "#caa472",
    height: 0.16,
    label: "Desert",
  },
};

const PORT_RESOURCE_LABELS: Record<number, string> = {
  0: "",
  1: "W",
  2: "B",
  3: "S",
  4: "Wh",
  5: "O",
};

const TILE_RESOURCE_ALIASES: Record<string, TileResource> = {
  TILE_RESOURCE_UNSPECIFIED: TileResource.UNSPECIFIED,
  TILE_RESOURCE_WOOD: TileResource.WOOD,
  TILE_RESOURCE_BRICK: TileResource.BRICK,
  TILE_RESOURCE_SHEEP: TileResource.SHEEP,
  TILE_RESOURCE_WHEAT: TileResource.WHEAT,
  TILE_RESOURCE_ORE: TileResource.ORE,
  TILE_RESOURCE_DESERT: TileResource.DESERT,
};

const RESOURCE_ALIASES: Record<string, number> = {
  RESOURCE_UNSPECIFIED: 0,
  RESOURCE_WOOD: 1,
  RESOURCE_BRICK: 2,
  RESOURCE_SHEEP: 3,
  RESOURCE_WHEAT: 4,
  RESOURCE_ORE: 5,
};

const PORT_TYPE_ALIASES: Record<string, PortType> = {
  PORT_TYPE_UNSPECIFIED: PortType.UNSPECIFIED,
  PORT_TYPE_GENERIC: PortType.GENERIC,
  PORT_TYPE_SPECIFIC: PortType.SPECIFIC,
};

const BUILDING_TYPE_ALIASES: Record<string, BuildingType> = {
  BUILDING_TYPE_UNSPECIFIED: BuildingType.UNSPECIFIED,
  BUILDING_TYPE_SETTLEMENT: BuildingType.SETTLEMENT,
  BUILDING_TYPE_CITY: BuildingType.CITY,
};

const PLAYER_COLOR_ALIASES: Record<string, PlayerColor> = {
  PLAYER_COLOR_UNSPECIFIED: PlayerColor.UNSPECIFIED,
  PLAYER_COLOR_RED: PlayerColor.RED,
  PLAYER_COLOR_BLUE: PlayerColor.BLUE,
  PLAYER_COLOR_GREEN: PlayerColor.GREEN,
  PLAYER_COLOR_ORANGE: PlayerColor.ORANGE,
};

function normalizeTileResource(
  resource: TileResource | string | undefined
): TileResource {
  if (typeof resource === "number") {
    return resource;
  }
  if (!resource) {
    return TileResource.UNSPECIFIED;
  }
  return TILE_RESOURCE_ALIASES[resource] ?? TileResource.UNSPECIFIED;
}

function normalizeResource(resource: number | string | undefined): number {
  if (typeof resource === "number") {
    return resource;
  }
  if (!resource) {
    return 0;
  }
  return RESOURCE_ALIASES[resource] ?? 0;
}

function normalizePortType(type: PortType | string | undefined): PortType {
  if (typeof type === "number") {
    return type;
  }
  if (!type) {
    return PortType.UNSPECIFIED;
  }
  return PORT_TYPE_ALIASES[type] ?? PortType.UNSPECIFIED;
}

function normalizeBuildingType(
  type: BuildingType | string | undefined
): BuildingType {
  if (typeof type === "number") {
    return type;
  }
  if (!type) {
    return BuildingType.UNSPECIFIED;
  }
  return BUILDING_TYPE_ALIASES[type] ?? BuildingType.UNSPECIFIED;
}

function normalizePlayerColor(
  color: PlayerColor | string | undefined
): PlayerColor {
  if (typeof color === "number") {
    return color;
  }
  if (!color) {
    return PlayerColor.UNSPECIFIED;
  }
  return PLAYER_COLOR_ALIASES[color] ?? PlayerColor.UNSPECIFIED;
}

const VERTEX_OFFSETS = [
  { direction: "N", dq: -1 / 3, dr: 2 / 3 },
  { direction: "NE", dq: 1 / 3, dr: 1 / 3 },
  { direction: "SE", dq: 2 / 3, dr: -1 / 3 },
  { direction: "S", dq: 1 / 3, dr: -2 / 3 },
  { direction: "SW", dq: -1 / 3, dr: -1 / 3 },
  { direction: "NW", dq: -2 / 3, dr: 1 / 3 },
];

function axialToWorld(
  q: number,
  r: number,
  size: number
): { x: number; z: number } {
  const x = size * (Math.sqrt(3) * q + (Math.sqrt(3) / 2) * r);
  const z = size * ((3 / 2) * r);
  return { x, z };
}

function axialToPixel2D(
  q: number,
  r: number,
  size: number
): { x: number; y: number } {
  const x = size * (Math.sqrt(3) * q + (Math.sqrt(3) / 2) * r);
  const y = size * ((3 / 2) * r);
  return { x, y };
}

function hexToWorld(coord: HexCoord, size: number): { x: number; z: number } {
  const q = coord.q ?? 0;
  const r = coord.r ?? 0;
  return axialToWorld(q, r, size);
}

function hexToPixel2D(
  coord: HexCoord,
  size: number
): { x: number; y: number } {
  const q = coord.q ?? 0;
  const r = coord.r ?? 0;
  return axialToPixel2D(q, r, size);
}

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

function lightenColor(color: string, amount: number): string {
  const base = new THREE.Color(color);
  const white = new THREE.Color("#ffffff");
  base.lerp(white, amount);
  return `#${base.getHexString()}`;
}

type HexPosition = {
  hex: Hex;
  position: { x: number; z: number };
};

type VertexPosition = {
  vertex: Vertex;
  coord: { q: number; r: number };
  position: { x: number; z: number };
};

type EdgePosition = {
  edge: EdgeState;
  v1: VertexPosition;
  v2: VertexPosition;
};

type PortPosition = {
  port: PortState;
  index: number;
  position: { x: number; z: number };
};

type BoardLayout = {
  hexes: HexPosition[];
  vertices: VertexPosition[];
  edges: EdgePosition[];
  ports: PortPosition[];
  center: { x: number; z: number };
};

function buildLayout(board: BoardState): BoardLayout {
  const validHexes = board.hexes.filter((hex) => isValidCoord(hex.coord));
  const hexes = validHexes.map((hex) => ({
    hex,
    position: hexToWorld(hex.coord!, HEX_SIZE),
  }));

  const vertices = board.vertices
    .map((vertex) => {
      const coord = parseVertexId(vertex.id);
      if (!coord) {
        return null;
      }
      const position = axialToWorld(coord.q, coord.r, HEX_SIZE);
      return { vertex, coord, position };
    })
    .filter((item): item is VertexPosition => item !== null);

  const vertexById = new Map(
    vertices.map((item) => [item.vertex.id, item])
  );

  const edges = board.edges
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
    .filter((item): item is EdgePosition => item !== null);

  const ports = (board.ports ?? [])
    .map((port, index) => {
      const [v1Id, v2Id] = port.location ?? [];
      if (!v1Id || !v2Id) {
        return null;
      }
      const v1 = vertexById.get(v1Id);
      const v2 = vertexById.get(v2Id);
      if (!v1 || !v2) {
        return null;
      }
      return {
        port,
        index,
        position: {
          x: (v1.position.x + v2.position.x) / 2,
          z: (v1.position.z + v2.position.z) / 2,
        },
      };
    })
    .filter((item): item is PortPosition => item !== null);

  const allPositions = [
    ...hexes.map((item) => item.position),
    ...vertices.map((item) => item.position),
  ];
  if (allPositions.length === 0) {
    return {
      hexes,
      vertices,
      edges,
      ports,
      center: { x: 0, z: 0 },
    };
  }
  const minX = Math.min(...allPositions.map((p) => p.x));
  const maxX = Math.max(...allPositions.map((p) => p.x));
  const minZ = Math.min(...allPositions.map((p) => p.z));
  const maxZ = Math.max(...allPositions.map((p) => p.z));

  return {
    hexes,
    vertices,
    edges,
    ports,
    center: {
      x: (minX + maxX) / 2,
      z: (minZ + maxZ) / 2,
    },
  };
}

function isWebGLAvailable(): boolean {
  if (typeof document === "undefined") {
    return false;
  }
  try {
    const canvas = document.createElement("canvas");
    return !!(
      (canvas.getContext("webgl") || canvas.getContext("experimental-webgl")) &&
      window.WebGLRenderingContext
    );
  } catch {
    return false;
  }
}

function shouldUse2DBoard(): boolean {
  if (typeof navigator === "undefined") {
    return true;
  }
  const isAutomation = navigator.webdriver === true;
  const isHeadless = /Headless|Playwright/i.test(navigator.userAgent);
  return isAutomation || isHeadless || !isWebGLAvailable();
}

function Robber({ height }: { height: number }) {
  return (
    <group position={[0, height, 0]}>
      <mesh castShadow>
        <cylinderGeometry args={[0.12, 0.18, 0.4, 12]} />
        <meshStandardMaterial color="#1c1c1c" />
      </mesh>
      <mesh position={[0, 0.3, 0]} castShadow>
        <sphereGeometry args={[0.14, 16, 16]} />
        <meshStandardMaterial color="#111111" />
      </mesh>
    </group>
  );
}

function HexTileInstances({
  hexes,
  robberHex,
  isRobberMoveMode,
  onSelectRobberHex,
}: {
  hexes: HexPosition[];
  robberHex?: HexCoord;
  isRobberMoveMode?: boolean;
  onSelectRobberHex?: (hex: { coord?: { q: number; r: number } }) => void;
}) {
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  useCursor(Boolean(hoveredId));

  const geometry = useMemo(
    () => new THREE.CylinderGeometry(TILE_RADIUS, TILE_RADIUS, 1, 6),
    []
  );
  const material = useMemo(
    () =>
      new THREE.MeshStandardMaterial({
        roughness: 0.65,
        metalness: 0.08,
        vertexColors: true,
      }),
    []
  );

  return (
    <Instances geometry={geometry} material={material} castShadow receiveShadow>
      {hexes.map((item) => {
        const resource = normalizeTileResource(item.hex.resource);
        const config = RESOURCE_CONFIG[resource];
        const coord = item.hex.coord;
        const id = coord ? `${coord.q},${coord.r}` : item.hex.resource.toString();
        const isRobber =
          robberHex && coord && coord.q === robberHex.q && coord.r === robberHex.r;
        const isSelectable = Boolean(isRobberMoveMode && !isRobber);
        const color =
          hoveredId === id || isSelectable
            ? lightenColor(config.color, 0.18)
            : config.color;
        const height = config.height;

        return (
          <Instance
            key={id}
            position={[item.position.x, height / 2, item.position.z]}
            scale={[1, height, 1]}
            color={color}
            onPointerOver={() => setHoveredId(id)}
            onPointerOut={() => setHoveredId(null)}
            onPointerDown={
              isSelectable ? () => onSelectRobberHex?.(item.hex) : undefined
            }
          />
        );
      })}
    </Instances>
  );
}

function HexOverlay({
  item,
  isRobberMoveMode,
  robberHex,
  onSelectRobberHex,
}: {
  item: HexPosition;
  isRobberMoveMode?: boolean;
  robberHex?: HexCoord;
  onSelectRobberHex?: (hex: Hex) => void;
}) {
  const resource = normalizeTileResource(item.hex.resource);
  const config = RESOURCE_CONFIG[resource];
  const coord = item.hex.coord;
  const isRobber =
    robberHex && coord && coord.q === robberHex.q && coord.r === robberHex.r;
  const isSelectable = Boolean(isRobberMoveMode && !isRobber);
  const label = config.label;
  const number = item.hex.number ?? 0;
  const isHighProbability = number === 6 || number === 8;
  const dataCy = isSelectable
    ? `robber-hex-${coord?.q}-${coord?.r}`
    : `hex-${coord?.q}-${coord?.r}`;
  const labelHeight = config.height + 0.12;

  return (
    <group position={[item.position.x, 0, item.position.z]}>
      {isRobber && <Robber height={config.height + 0.1} />}
      <Html
        className="tile-label"
        position={[0, labelHeight, 0]}
        center
        transform
        distanceFactor={8}
        style={{ pointerEvents: "none" }}
      >
        <div className="tile-label__name">{label}</div>
        {number > 0 && (
          <div
            className={`tile-label__token${
              isHighProbability ? " tile-label__token--hot" : ""
            }`}
          >
            {number}
          </div>
        )}
      </Html>
      <Html
        position={[0, labelHeight, 0]}
        center
        transform
        distanceFactor={8}
      >
        <button
          type="button"
          data-cy={dataCy}
          className={`board-hit board-hit--hex${
            isSelectable ? " board-hit--active" : ""
          }`}
          onClick={isSelectable ? () => onSelectRobberHex?.(item.hex) : undefined}
          style={{ pointerEvents: isSelectable ? "auto" : "none" }}
          aria-hidden="true"
        />
      </Html>
    </group>
  );
}

function VertexMarker3D({
  vertex,
  position,
  ownerColor,
  dataCy,
  isValid,
  onClick,
}: {
  vertex: Vertex;
  position: { x: number; z: number };
  ownerColor?: string;
  dataCy: string;
  isValid: boolean;
  onClick?: () => void;
}) {
  const [hovered, setHovered] = useState(false);
  const pulseRef = useRef<THREE.Mesh>(null);
  useCursor(hovered || isValid);

  useFrame((state) => {
    if (!pulseRef.current) {
      return;
    }
    if (hovered || isValid) {
      const pulse = 1 + Math.sin(state.clock.elapsedTime * 4) * 0.08;
      pulseRef.current.scale.set(pulse, pulse, pulse);
    } else {
      pulseRef.current.scale.set(1, 1, 1);
    }
  });

  const baseColor = ownerColor ?? (isValid ? "#f7d774" : "#dfe5f2");
  const displayColor = hovered
    ? lightenColor(baseColor, 0.2)
    : baseColor;
  const building = vertex.building;
  const buildingType = normalizeBuildingType(building?.type);
  const isOccupied = Boolean(building);

  return (
    <group position={[position.x, VERTEX_Y, position.z]}>
      <mesh
        ref={pulseRef}
        castShadow
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
        onPointerDown={isValid ? onClick : undefined}
      >
        <sphereGeometry args={[0.12, 18, 18]} />
        <meshStandardMaterial color={displayColor} />
      </mesh>
      {buildingType === BuildingType.SETTLEMENT && (
        <group position={[0, 0.18, 0]}>
          <mesh castShadow>
            <boxGeometry args={[0.26, 0.18, 0.26]} />
            <meshStandardMaterial color={ownerColor ?? "#6b6b6b"} />
          </mesh>
          <mesh position={[0, 0.16, 0]} castShadow>
            <coneGeometry args={[0.2, 0.2, 4]} />
            <meshStandardMaterial color={ownerColor ?? "#6b6b6b"} />
          </mesh>
        </group>
      )}
      {buildingType === BuildingType.CITY && (
        <group position={[0, 0.24, 0]}>
          <mesh castShadow>
            <boxGeometry args={[0.32, 0.2, 0.32]} />
            <meshStandardMaterial color={ownerColor ?? "#6b6b6b"} />
          </mesh>
          <mesh position={[0, 0.18, 0]} castShadow>
            <boxGeometry args={[0.18, 0.2, 0.18]} />
            <meshStandardMaterial color={ownerColor ?? "#6b6b6b"} />
          </mesh>
        </group>
      )}
      <Html position={[0, 0.2, 0]} center transform distanceFactor={8}>
        <button
          type="button"
          data-cy={dataCy}
          className={`board-hit board-hit--vertex ${isOccupied ? "vertex--occupied" : "vertex--empty"}${
            isValid ? " vertex--valid board-hit--active" : ""
          }`}
          onClick={isValid ? onClick : undefined}
          style={{ pointerEvents: isValid ? "auto" : "none" }}
          aria-hidden="true"
        />
      </Html>
    </group>
  );
}

function EdgeSegment3D({
  edge,
  v1,
  v2,
  ownerColor,
  dataCy,
  isValid,
  onClick,
}: {
  edge: EdgeState;
  v1: VertexPosition;
  v2: VertexPosition;
  ownerColor?: string;
  dataCy: string;
  isValid: boolean;
  onClick?: () => void;
}) {
  const [hovered, setHovered] = useState(false);
  useCursor(hovered || isValid);

  const hasRoad = Boolean(edge.road);
  const start = new THREE.Vector3(v1.position.x, 0, v1.position.z);
  const end = new THREE.Vector3(v2.position.x, 0, v2.position.z);
  const direction = new THREE.Vector3().subVectors(end, start);
  const length = direction.length();
  direction.normalize();

  const midpoint = new THREE.Vector3()
    .addVectors(start, end)
    .multiplyScalar(0.5);

  const quaternion = new THREE.Quaternion();
  quaternion.setFromUnitVectors(new THREE.Vector3(0, 1, 0), direction);

  const thickness = hasRoad ? 0.09 : 0.04;
  const baseColor = hasRoad
    ? ownerColor ?? "#d6d6d6"
    : isValid || hovered
      ? "#f7d774"
      : "#d6d6d6";
  const opacity = hasRoad ? 1 : isValid || hovered ? 0.75 : 0.25;

  const geometry = useMemo(
    () => new THREE.CylinderGeometry(thickness, thickness, length, 10),
    [length, thickness]
  );

  return (
    <group
      position={[midpoint.x, EDGE_Y, midpoint.z]}
      quaternion={quaternion}
    >
      <mesh
        geometry={geometry}
        castShadow
        onPointerOver={() => setHovered(true)}
        onPointerOut={() => setHovered(false)}
        onPointerDown={isValid ? onClick : undefined}
      >
        <meshStandardMaterial
          color={baseColor}
          transparent={!hasRoad}
          opacity={opacity}
        />
      </mesh>
      <Html position={[0, 0, 0]} center transform distanceFactor={8}>
        <button
          type="button"
          data-cy={dataCy}
          className={`board-hit board-hit--edge ${hasRoad ? "edge--occupied" : "edge--empty"}${
            isValid ? " edge--valid board-hit--active" : ""
          }`}
          onClick={isValid ? onClick : undefined}
          style={{ pointerEvents: isValid ? "auto" : "none" }}
          aria-hidden="true"
        />
      </Html>
    </group>
  );
}

function PortMarker3D({ port, position, index }: PortPosition) {
  const portType = normalizePortType(port.type);
  const isGeneric = portType === PortType.GENERIC;
  const label = isGeneric ? "3:1" : "2:1";
  const resourceValue = normalizeResource(port.resource);
  const resourceLabel =
    !isGeneric && resourceValue ? PORT_RESOURCE_LABELS[resourceValue] : "";

  return (
    <Html position={[position.x, 0.2, position.z]} center transform>
      <div
        className={`port-marker${
          isGeneric ? " port-marker--generic" : " port-marker--specific"
        }`}
        data-cy={`port-${index}`}
      >
        <div className="port-marker__label">{label}</div>
        {resourceLabel && (
          <div className="port-marker__resource">{resourceLabel}</div>
        )}
      </div>
    </Html>
  );
}

function BoardScene({
  layout,
  board,
  players,
  validVertexIds,
  validEdgeIds,
  onBuildSettlement,
  onBuildRoad,
  isRobberMoveMode,
  onSelectRobberHex,
}: {
  layout: BoardLayout;
  board: BoardState;
  players: PlayerState[];
  validVertexIds?: Set<string>;
  validEdgeIds?: Set<string>;
  onBuildSettlement?: (vertexId: string) => void;
  onBuildRoad?: (edgeId: string) => void;
  isRobberMoveMode?: boolean;
  onSelectRobberHex?: (hex: { coord?: { q: number; r: number } }) => void;
}) {
  const playerColors = useMemo(
    () =>
      new Map(
        players.map((player) => [
          player.id,
          PLAYER_COLORS[normalizePlayerColor(player.color)] ?? "#808080",
        ])
      ),
    [players]
  );

  return (
    <>
      <OrbitControls
        enablePan={false}
        maxPolarAngle={1.2}
        minPolarAngle={0.6}
        minDistance={6}
        maxDistance={18}
      />
      <group position={[-layout.center.x, 0, -layout.center.z]}>
        <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.02, 0]} receiveShadow>
          <planeGeometry args={[50, 50]} />
          <shadowMaterial opacity={0.35} />
        </mesh>
        <HexTileInstances
          hexes={layout.hexes}
          robberHex={board.robberHex}
          isRobberMoveMode={isRobberMoveMode}
          onSelectRobberHex={onSelectRobberHex}
        />
        {layout.hexes.map((item) => (
          <HexOverlay
            key={`overlay-${item.hex.coord?.q}-${item.hex.coord?.r}`}
            item={item}
            robberHex={board.robberHex}
            isRobberMoveMode={isRobberMoveMode}
            onSelectRobberHex={onSelectRobberHex}
          />
        ))}
        {layout.edges.map(({ edge, v1, v2 }) => {
          const roadOwnerId = edge.road?.ownerId;
          const ownerColor = roadOwnerId
            ? playerColors.get(roadOwnerId)
            : undefined;
          const isValid = Boolean(validEdgeIds?.has(edge.id));
          return (
            <EdgeSegment3D
              key={edge.id}
              edge={edge}
              v1={v1}
              v2={v2}
              ownerColor={ownerColor}
              dataCy={getEdgeDataCy(edge, v1.coord, v2.coord)}
              isValid={isValid}
              onClick={
                isValid && onBuildRoad ? () => onBuildRoad(edge.id) : undefined
              }
            />
          );
        })}
        {layout.vertices.map(({ vertex, coord, position }) => {
          const ownerColor = vertex.building
            ? playerColors.get(vertex.building.ownerId)
            : undefined;
          const isValid = Boolean(validVertexIds?.has(vertex.id));
          return (
            <VertexMarker3D
              key={vertex.id}
              vertex={vertex}
              position={position}
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
        {layout.ports.map((port) => (
          <PortMarker3D key={`port-${port.index}`} {...port} />
        ))}
      </group>
    </>
  );
}

function Board({
  board,
  players,
  validVertexIds,
  validEdgeIds,
  onBuildSettlement,
  onBuildRoad,
  isRobberMoveMode,
  onSelectRobberHex,
}: BoardProps) {
  const use2DBoard = shouldUse2DBoard();
  const validHexes = board.hexes.filter((hex) => isValidCoord(hex.coord));
  const layout = useMemo(() => buildLayout(board), [board]);

  if (use2DBoard) {
    return (
      <Board2D
        board={board}
        players={players}
        validVertexIds={validVertexIds}
        validEdgeIds={validEdgeIds}
        onBuildSettlement={onBuildSettlement}
        onBuildRoad={onBuildRoad}
        isRobberMoveMode={isRobberMoveMode}
        onSelectRobberHex={onSelectRobberHex}
      />
    );
  }

  if (validHexes.length === 0) {
    return (
      <div className="board-container" data-cy="board-loading">
        <p>Loading board...</p>
      </div>
    );
  }

  return (
    <div className="board-container" data-cy="game-board-container">
      <div className="board-canvas-wrapper" data-cy="board-canvas">
        <Canvas
          className="board-canvas"
          data-cy="board"
          shadows
          dpr={[1, 2]}
          camera={{ position: [8, 10, 8], fov: 45, near: 0.1, far: 100 }}
          gl={{ antialias: true, alpha: true }}
        >
          <ambientLight intensity={0.55} />
          <directionalLight
            position={[6, 12, 6]}
            intensity={0.95}
            castShadow
            shadow-mapSize-width={1024}
            shadow-mapSize-height={1024}
          />
          <BoardScene
            layout={layout}
            board={board}
            players={players}
            validVertexIds={validVertexIds}
            validEdgeIds={validEdgeIds}
            onBuildSettlement={onBuildSettlement}
            onBuildRoad={onBuildRoad}
            isRobberMoveMode={isRobberMoveMode}
            onSelectRobberHex={onSelectRobberHex}
          />
        </Canvas>
      </div>
    </div>
  );
}

function Board2D({
  board,
  players,
  validVertexIds,
  validEdgeIds,
  onBuildSettlement,
  onBuildRoad,
  isRobberMoveMode,
  onSelectRobberHex,
}: BoardProps) {
  const validHexes = board.hexes.filter((hex) => isValidCoord(hex.coord));

  if (validHexes.length === 0) {
    return (
      <div className="board-container" data-cy="board-loading">
        <p>Loading board...</p>
      </div>
    );
  }

  const positions = validHexes.map((hex) =>
    hexToPixel2D(hex.coord!, HEX_SIZE_2D)
  );
  const vertexPositions = board.vertices
    .map((vertex) => {
      const coord = parseVertexId(vertex.id);
      if (!coord) {
        return null;
      }
      const pos = axialToPixel2D(coord.q, coord.r, HEX_SIZE_2D);
      return { pos, coord, vertex };
    })
    .filter(
      (
        item
      ): item is {
        pos: { x: number; y: number };
        coord: { q: number; r: number };
        vertex: Vertex;
      } => item !== null
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

  const allPositions = positions.concat(
    vertexPositions.map((item) => item.pos)
  );
  const minX = Math.min(...allPositions.map((p) => p.x));
  const maxX = Math.max(...allPositions.map((p) => p.x));
  const minY = Math.min(...allPositions.map((p) => p.y));
  const maxY = Math.max(...allPositions.map((p) => p.y));

  const padding = HEX_SIZE_2D * 1.5;
  const width = maxX - minX + padding * 2;
  const height = maxY - minY + padding * 2;
  const offsetX = -minX + padding;
  const offsetY = -minY + padding;

  const robberHex = board.robberHex;
  const playerColors = new Map(
    players.map((player) => [
      player.id,
      PLAYER_COLORS[normalizePlayerColor(player.color)] ?? "#808080",
    ])
  );

  return (
    <div className="board-container" data-cy="game-board-container">
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="board-svg"
        preserveAspectRatio="xMidYMid meet"
        data-cy="board"
      >
        <g transform={`translate(${offsetX}, ${offsetY})`}>
          {validHexes.map((hex) => {
            const coord = hex.coord!;
            const pos = hexToPixel2D(coord, HEX_SIZE_2D);
            const isRobber =
              robberHex && coord.q === robberHex.q && coord.r === robberHex.r;
            const isSelectable = isRobberMoveMode && !isRobber;
            return (
              <HexTile
                key={`${coord.q},${coord.r}`}
                hex={hex}
                x={pos.x}
                y={pos.y}
                size={HEX_SIZE_2D}
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
                  isValid && onBuildRoad
                    ? () => onBuildRoad(edge.id)
                    : undefined
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
          {board.ports?.map((port, index) => {
            const [v1Id, v2Id] = port.location ?? [];
            if (!v1Id || !v2Id) {
              return null;
            }
            const v1 = vertexById.get(v1Id);
            const v2 = vertexById.get(v2Id);
            if (!v1 || !v2) {
              return null;
            }
            const midX = (v1.pos.x + v2.pos.x) / 2;
            const midY = (v1.pos.y + v2.pos.y) / 2;
            return (
              <Port
                key={`port-${index}`}
                port={port}
                x={midX}
                y={midY}
                index={index}
              />
            );
          })}
        </g>
      </svg>
    </div>
  );
}

export default Board;
export { Board };
