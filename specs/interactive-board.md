# Interactive Board - Vertex & Edge Rendering + Click Handlers

**Priority**: CRITICAL - Unblocks all placement features (setup, building, roads)

## Overview

The board currently renders hex tiles but does NOT render clickable vertices (for settlements/cities) or edges (for roads). Without this, players cannot place anything.

## Acceptance Criteria

### Vertex Rendering

- [ ] Render vertices as SVG circles at hex corners
- [ ] Each vertex has a unique coordinate (derived from adjacent hex coords)
- [ ] Empty vertices show as small dots (or invisible until hover)
- [ ] Vertices with settlements show settlement icon
- [ ] Vertices with cities show city icon
- [ ] Vertices use player color when occupied

### Edge Rendering

- [ ] Render edges as SVG lines/rectangles between adjacent vertices
- [ ] Each edge has a unique coordinate pair
- [ ] Empty edges are invisible (or subtle on hover)
- [ ] Edges with roads show road icon/line in player color

### Click Handlers

- [ ] Vertices are clickable when placement is valid
- [ ] Clicking a valid vertex sends `build_structure` message with vertex coord
- [ ] Edges are clickable when road placement is valid
- [ ] Clicking a valid edge sends `build_structure` message with edge coord
- [ ] Invalid positions are not clickable (or show disabled state)

### Placement Mode

- [ ] UI indicates when player should place (settlement vs road)
- [ ] Valid placement positions are highlighted
- [ ] During setup phase, highlight valid initial settlement spots
- [ ] During normal play, highlight spots connected to player's roads

### Data Attributes (for Playwright)

- [ ] `data-cy="vertex-{q}-{r}-{direction}"` on each vertex
- [ ] `data-cy="edge-{q1}-{r1}-{q2}-{r2}"` on each edge
- [ ] `data-cy="placement-mode"` indicator showing current mode

## Required Go Unit Tests

Backend vertex/edge logic is already tested. No new backend tests needed for this spec.

## Required Playwright E2E Tests

File: `frontend/tests/interactive-board.spec.ts`

```typescript
// Test: Vertices render on board
// - Start game, verify vertex elements exist with data-cy attributes

// Test: Clicking vertex during setup places settlement
// - Start 2-player game, reach setup phase
// - Click valid vertex, verify settlement appears

// Test: Clicking edge during setup places road
// - After placing settlement, click adjacent edge
// - Verify road appears in player color

// Test: Invalid vertices not clickable
// - Try clicking vertex adjacent to existing settlement
// - Verify no placement occurs (distance rule)

// Test: Placement highlighting shows valid spots
// - During setup, verify valid vertices have highlight class
```

## Implementation Notes

- Vertex coordinates: Use direction from hex center (N, NE, SE, S, SW, NW)
- Edge coordinates: Use the two vertices it connects
- SVG layering: Hexes at bottom, edges middle, vertices top
- Consider using CSS transitions for hover states
