import type { BoardState, GameState, ResourceCount } from "@/types";
import { GameStatus, TurnPhase } from "@/types";

export interface PlacementState {
  validVertexIds: Set<string>;
  validEdgeIds: Set<string>;
}

const SETTLEMENT_COST = {
  wood: 1,
  brick: 1,
  sheep: 1,
  wheat: 1,
};

const ROAD_COST = {
  wood: 1,
  brick: 1,
};

function isStatus(
  status: GameStatus | string | undefined,
  expected: GameStatus,
  expectedString: string
): boolean {
  return status === expected || status === (expectedString as unknown as GameStatus);
}

function isTurnPhase(
  phase: TurnPhase | string | undefined,
  expected: TurnPhase,
  expectedString: string
): boolean {
  return phase === expected || phase === (expectedString as unknown as TurnPhase);
}

function hasResources(
  resources: ResourceCount | undefined,
  costs: Record<string, number>
): boolean {
  if (!resources) {
    return false;
  }
  return (
    (resources.wood ?? 0) >= (costs.wood ?? 0) &&
    (resources.brick ?? 0) >= (costs.brick ?? 0) &&
    (resources.sheep ?? 0) >= (costs.sheep ?? 0) &&
    (resources.wheat ?? 0) >= (costs.wheat ?? 0) &&
    (resources.ore ?? 0) >= (costs.ore ?? 0)
  );
}

function buildAdjacency(board: BoardState): Map<string, Set<string>> {
  const adjacency = new Map<string, Set<string>>();
  for (const edge of board.edges ?? []) {
    const [v1, v2] = edge.vertices ?? [];
    if (!v1 || !v2) {
      continue;
    }
    if (!adjacency.has(v1)) {
      adjacency.set(v1, new Set());
    }
    if (!adjacency.has(v2)) {
      adjacency.set(v2, new Set());
    }
    adjacency.get(v1)?.add(v2);
    adjacency.get(v2)?.add(v1);
  }
  return adjacency;
}

function violatesDistanceRule(
  vertexId: string,
  vertexById: Map<string, { building?: unknown }>,
  adjacency: Map<string, Set<string>>
): boolean {
  const neighbors = adjacency.get(vertexId);
  if (!neighbors || neighbors.size === 0) {
    return false;
  }
  for (const neighborId of neighbors) {
    const neighbor = vertexById.get(neighborId);
    if (neighbor && neighbor.building) {
      return true;
    }
  }
  return false;
}

function collectPlayerStructures(board: BoardState, playerId: string) {
  const playerRoadVertices = new Set<string>();
  const playerBuildingVertices = new Set<string>();

  for (const vertex of board.vertices ?? []) {
    if (vertex.building?.ownerId === playerId) {
      playerBuildingVertices.add(vertex.id);
    }
  }

  for (const edge of board.edges ?? []) {
    if (edge.road?.ownerId === playerId) {
      for (const vertexId of edge.vertices ?? []) {
        if (vertexId) {
          playerRoadVertices.add(vertexId);
        }
      }
    }
  }

  return { playerRoadVertices, playerBuildingVertices };
}

function collectUnroadedBuildingVertices(
  board: BoardState,
  playerId: string
): Set<string> {
  const verticesWithRoad = new Set<string>();
  for (const edge of board.edges ?? []) {
    if (edge.road?.ownerId !== playerId) {
      continue;
    }
    for (const vertexId of edge.vertices ?? []) {
      if (vertexId) {
        verticesWithRoad.add(vertexId);
      }
    }
  }

  const unroaded = new Set<string>();
  for (const vertex of board.vertices ?? []) {
    if (vertex.building?.ownerId !== playerId) {
      continue;
    }
    if (!verticesWithRoad.has(vertex.id)) {
      unroaded.add(vertex.id);
    }
  }

  return unroaded;
}

function edgeConnectsToVertices(
  edgeVertices: string[] | undefined,
  candidates: Set<string>
): boolean {
  if (!edgeVertices) {
    return false;
  }
  return edgeVertices.some((vertexId) => candidates.has(vertexId));
}

export function getPlacementState(
  gameState: GameState | null,
  currentPlayerId: string | null
): PlacementState {
  const emptyState: PlacementState = {
    validVertexIds: new Set(),
    validEdgeIds: new Set(),
  };

  if (!gameState || !currentPlayerId || !gameState.board) {
    return emptyState;
  }

  const currentPlayer = gameState.players[gameState.currentTurn ?? 0];
  if (!currentPlayer || currentPlayer.id !== currentPlayerId) {
    return emptyState;
  }

  const board = gameState.board;
  const adjacency = buildAdjacency(board);
  const vertexById = new Map(
    board.vertices.map((vertex) => [vertex.id, vertex])
  );

  const { playerRoadVertices, playerBuildingVertices } =
    collectPlayerStructures(board, currentPlayerId);

  const isSetup = isStatus(
    gameState.status,
    GameStatus.SETUP,
    "GAME_STATUS_SETUP"
  );
  const isBuildPhase =
    isStatus(
      gameState.status,
      GameStatus.PLAYING,
      "GAME_STATUS_PLAYING"
    ) &&
    isTurnPhase(
      gameState.turnPhase,
      TurnPhase.BUILD,
      "TURN_PHASE_BUILD"
    );

  if (!isSetup && !isBuildPhase) {
    return emptyState;
  }

  const canPlaceSettlement =
    isSetup ||
    hasResources(currentPlayer.resources, SETTLEMENT_COST);
  const canPlaceRoad =
    isSetup || hasResources(currentPlayer.resources, ROAD_COST);

  if (isSetup && (gameState.setupPhase?.placementsInTurn ?? 0) === 0) {
    if (!canPlaceSettlement) {
      return emptyState;
    }
    for (const vertex of board.vertices ?? []) {
      if (vertex.building) {
        continue;
      }
      if (violatesDistanceRule(vertex.id, vertexById, adjacency)) {
        continue;
      }
      emptyState.validVertexIds.add(vertex.id);
    }
    return emptyState;
  }

  if (isSetup && (gameState.setupPhase?.placementsInTurn ?? 0) === 1) {
    if (!canPlaceRoad) {
      return emptyState;
    }
    const setupRoadVertices = collectUnroadedBuildingVertices(
      board,
      currentPlayerId
    );
    const roadAnchorVertices =
      setupRoadVertices.size > 0 ? setupRoadVertices : playerBuildingVertices;
    for (const edge of board.edges ?? []) {
      if (edge.road) {
        continue;
      }
      if (edgeConnectsToVertices(edge.vertices, roadAnchorVertices)) {
        emptyState.validEdgeIds.add(edge.id);
      }
    }
    return emptyState;
  }

  if (isBuildPhase) {
    if (canPlaceSettlement) {
      for (const vertex of board.vertices ?? []) {
        if (vertex.building) {
          continue;
        }
        if (violatesDistanceRule(vertex.id, vertexById, adjacency)) {
          continue;
        }
        if (!playerRoadVertices.has(vertex.id)) {
          continue;
        }
        emptyState.validVertexIds.add(vertex.id);
      }
    }

    if (canPlaceRoad) {
      for (const edge of board.edges ?? []) {
        if (edge.road) {
          continue;
        }
        if (
          edgeConnectsToVertices(edge.vertices, playerBuildingVertices) ||
          edgeConnectsToVertices(edge.vertices, playerRoadVertices)
        ) {
          emptyState.validEdgeIds.add(edge.id);
        }
      }
    }
  }

  return emptyState;
}
