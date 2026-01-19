# Longest Road Calculation

**Priority**: MEDIUM - Required for accurate victory point tracking

## Overview

"Longest Road" (2 VP bonus) goes to the first player with a continuous road of 5+ segments. The road must be unbroken - opponent settlements/cities break the chain.

## Acceptance Criteria

### Road Length Algorithm

- [ ] Calculate longest continuous road for each player
- [ ] Road segments must be connected (share a vertex)
- [ ] Opponent buildings break the chain
- [ ] Own buildings do NOT break the chain
- [ ] Handle branches: find the maximum path, not total roads

### Longest Road Tracking

- [ ] Award "Longest Road" to first player with 5+ road length
- [ ] Transfer if another player exceeds current holder
- [ ] Tie goes to current holder (no transfer on tie)
- [ ] If holder's road broken (by opponent building), recalculate

### Recalculation Triggers

- [ ] After any road placement
- [ ] After any settlement/city placement (may break roads)

### UI Display

- [ ] Show "Longest Road: {Player} (X roads)" in game info
- [ ] Show road length per player in player panel
- [ ] Highlight longest road on board (optional)

### Data Attributes (for Playwright)

- [ ] `data-cy="longest-road-holder"` showing current holder
- [ ] `data-cy="road-length-{playerId}"` per player

## Required Go Unit Tests

File: `backend/internal/game/longestroad_test.go` (new)

```go
// Test: Single road segment = length 1
// Test: Two connected roads = length 2
// Test: Branching roads returns max branch length
// Test: Opponent settlement breaks road chain
// Test: Own settlement does NOT break chain
// Test: Circular road counts correctly
// Test: 5+ roads awards Longest Road bonus
// Test: Longer road takes Longest Road from holder
// Test: Tie does not transfer Longest Road
// Test: Breaking opponent's road may transfer bonus

// Algorithm test cases:
// Test: Linear road of 6 = length 6
// Test: Y-shaped road (3+3+3) = length 6 (longest path)
// Test: Figure-8 road calculates correctly
```

## Algorithm

```
LongestRoad(player) -> int:
    maxLength = 0
    for each road segment owned by player:
        length = DFS(road, visited={}, player)
        maxLength = max(maxLength, length)
    return maxLength

DFS(currentEdge, visited, player) -> int:
    visited.add(currentEdge)
    maxPath = 1  // Current segment

    for each adjacentEdge connected to currentEdge:
        if adjacentEdge in visited:
            continue
        if adjacentEdge not owned by player:
            continue
        // Check if connection vertex is blocked by opponent
        connectingVertex = getSharedVertex(currentEdge, adjacentEdge)
        if hasOpponentBuilding(connectingVertex, player):
            continue  // Road broken

        path = 1 + DFS(adjacentEdge, visited, player)
        maxPath = max(maxPath, path)

    visited.remove(currentEdge)  // Backtrack for other branches
    return maxPath
```

## Required Playwright E2E Tests

File: `frontend/tests/longest-road.spec.ts`

```typescript
// Test: 5 connected roads shows Longest Road badge
// Test: Longer road takes Longest Road from another player
// Test: Broken road loses Longest Road
```

## Implementation Notes

- Use DFS with backtracking to find longest path
- Cache result per player, invalidate on road/building placement
- Edge adjacency: edges share a vertex
- Vertex blocking: check `vertex.building` owner != player
- Consider: store road graph separately for faster traversal
