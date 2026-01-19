# Setup Phase UI

**Priority**: HIGH - Required for game to be playable after start

## Overview

Backend handles setup phase logic (snake draft, free placements, round transitions). Frontend needs UI to guide players through placing their initial 2 settlements and 2 roads.

## Acceptance Criteria

### Setup Phase Indicators

- [ ] Display "Setup Phase - Round 1" or "Round 2" banner
- [ ] Show whose turn it is to place
- [ ] Show what to place next: "Place Settlement" or "Place Road"
- [ ] Show placement count (e.g., "Settlement 1/2")

### Turn Flow

- [ ] Round 1: Players place in order (P1 → P2 → P3 → P4)
- [ ] Round 2: Reverse order (P4 → P3 → P2 → P1)
- [ ] Each turn: Settlement first, then adjacent road
- [ ] Turn auto-advances after road placed

### Placement Guidance

- [ ] Highlight ALL valid vertices for first settlement (distance rule only)
- [ ] After settlement placed, highlight only adjacent edges for road
- [ ] After road placed, turn passes to next player
- [ ] Round 2 second settlement grants starting resources (shown in UI)

### Resource Grant Display

- [ ] After Round 2 settlement, show resources gained
- [ ] Toast/notification: "You received: 1 Wood, 1 Brick, 1 Wheat"

### Data Attributes (for Playwright)

- [ ] `data-cy="setup-phase-banner"` with round number
- [ ] `data-cy="setup-instruction"` showing "Place Settlement" or "Place Road"
- [ ] `data-cy="setup-turn-indicator"` showing current player

## Required Go Unit Tests

Backend setup logic already tested in `state_machine_test.go`. No new backend tests.

## Required Playwright E2E Tests

File: `frontend/tests/setup-phase.spec.ts`

```typescript
// Test: Setup phase banner shows after game start
// - Start 2-player game with all ready
// - Verify setup phase banner visible with "Round 1"

// Test: First player can place settlement
// - Verify instruction shows "Place Settlement"
// - Click valid vertex, settlement appears
// - Instruction changes to "Place Road"

// Test: After road, turn passes to next player
// - Player 1 places settlement + road
// - Verify turn indicator shows Player 2

// Test: Round 2 grants resources for second settlement
// - Complete Round 1 placements
// - Player places Round 2 settlement adjacent to resource hexes
// - Verify resource notification appears

// Test: Setup completes and game enters PLAYING state
// - All players complete setup placements
// - Verify game phase changes, dice roll becomes available
```

## Implementation Notes

- Setup state comes from `gameState.setupPhase` in websocket payload
- Use `setupPhase.round` (1 or 2) and `setupPhase.currentPlayerIndex`
- Track `setupPhase.settlementsPlaced` and `setupPhase.roadsPlaced` arrays
- GameContext should expose `isMySetupTurn` computed property
