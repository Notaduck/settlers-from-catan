# Ports - Maritime Trade

**Priority**: LOW - Enhancement to trading, not required for basic play

## Overview

Ports allow better trade ratios with the bank:

- **3:1 Generic Port**: Trade 3 of any resource for 1 of any
- **2:1 Specific Port**: Trade 2 of one specific resource for 1 of any

Standard Catan board has 9 ports around the coast.

## Acceptance Criteria

### Proto Definition

- [ ] Add `Port` message to types.proto
- [ ] Port has: location (edge or vertex pair), type (generic/specific), resource (if specific)
- [ ] Add `repeated Port ports` to BoardState

### Board Generation

- [ ] Generate 9 ports at standard coastal positions
- [ ] 4 generic (3:1) ports
- [ ] 5 specific (2:1) ports: 1 each for wood, brick, wheat, sheep, ore
- [ ] Ports placed on coastal edges/vertices

### Port Rendering

- [ ] Show port icons on board at correct positions
- [ ] Generic ports show "3:1" label
- [ ] Specific ports show resource icon + "2:1"

### Port Access

- [ ] Player has port access if they have settlement/city on port vertex
- [ ] Track which ports each player has access to

### Port Trading

- [ ] Bank trade modal shows available ratios
- [ ] Default 4:1 always available
- [ ] If player has 3:1 port, show 3:1 option
- [ ] If player has 2:1 port for resource X, show 2:1 for X
- [ ] Use best available ratio automatically

### Data Attributes (for Playwright)

- [ ] `data-cy="port-{index}"` on each port
- [ ] `data-cy="trade-ratio-{resource}"` showing best ratio

## Required Go Unit Tests

File: `backend/internal/game/ports_test.go` (new)

```go
// Test: Board generates 9 ports
// Test: Port distribution is 4 generic + 5 specific
// Test: Player gains port access on settlement placement
// Test: GetBestTradeRatio returns 4 by default
// Test: GetBestTradeRatio returns 3 with generic port
// Test: GetBestTradeRatio returns 2 with specific port for that resource
// Test: PortTrade validates correct resource count
```

## Required Playwright E2E Tests

File: `frontend/tests/ports.spec.ts`

```typescript
// Test: Ports render on board
// Test: Bank trade shows 4:1 by default
// Test: Player with 3:1 port sees 3:1 option
// Test: Player with 2:1 wheat port can trade 2 wheat
```

## Implementation Notes

- Port positions are fixed in standard Catan
- Coastal vertices determined by hex adjacency (vertices with <3 adjacent hexes)
- Consider: ports as edge-based (between two coastal vertices)
- Store player port access in PlayerState for quick lookup
