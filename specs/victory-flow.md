# Victory Flow - Win Detection & Game Over

**Priority**: HIGH - Required to end the game properly

## Overview

First player to reach 10+ victory points on their turn wins. The game must detect this and display the winner.

## Victory Point Sources

| Source             | Points |
| ------------------ | ------ |
| Settlement         | 1      |
| City               | 2      |
| Longest Road       | 2      |
| Largest Army       | 2      |
| Victory Point Card | 1 each |

## Acceptance Criteria

### Victory Detection

- [ ] Check for victory after each point-gaining action
- [ ] Actions that gain points: build settlement, build city, play VP card, gain Longest Road, gain Largest Army
- [ ] Victory requires 10+ VP on YOUR turn
- [ ] Hidden VP cards count (revealed on win)

### Victory Check Timing

- [ ] After `PlaceSettlement` (setup doesn't count - need 10)
- [ ] After `PlaceCity`
- [ ] After `PlayDevCard` (VP card or Knight → Largest Army)
- [ ] After road placement (may gain Longest Road)
- [ ] NOT during other players' turns

### Game End Flow

- [ ] When player reaches 10+ VP, game status → FINISHED
- [ ] Broadcast `GameOver` message with winner and final scores
- [ ] No more actions allowed after game over
- [ ] Final VP revealed for all players (including hidden VP cards)

### Game Over UI

- [ ] Full-screen victory overlay
- [ ] Show winner name and final VP count
- [ ] Show breakdown: settlements, cities, Longest Road, Largest Army, VP cards
- [ ] Show all players' final scores
- [ ] "New Game" button to return to lobby/create new game

### Score Breakdown Display

- [ ] Per player: public VP (visible buildings/bonuses)
- [ ] Per player: hidden VP (revealed at end)
- [ ] Per player: total VP
- [ ] Highlight winner row

### Data Attributes (for Playwright)

- [ ] `data-cy="game-over-overlay"`
- [ ] `data-cy="winner-name"`
- [ ] `data-cy="winner-vp"`
- [ ] `data-cy="final-score-{playerId}"`
- [ ] `data-cy="new-game-btn"`

## Required Go Unit Tests

File: `backend/internal/game/victory_test.go` (new)

```go
// Test: 10 VP triggers victory
// Test: 9 VP does not trigger victory
// Test: Victory only checked on active player's turn
// Test: CheckVictory counts settlements (1 each)
// Test: CheckVictory counts cities (2 each)
// Test: CheckVictory includes Longest Road bonus
// Test: CheckVictory includes Largest Army bonus
// Test: CheckVictory includes hidden VP cards
// Test: Game status changes to FINISHED on victory
// Test: GameOver payload includes winner and all scores
```

## Required Playwright E2E Tests

File: `frontend/tests/victory.spec.ts`

```typescript
// Test: Game over screen shows when player wins
// Test: Winner name and VP displayed correctly
// Test: All players' final scores visible
// Test: Hidden VP cards revealed in final scores
// Test: New game button returns to create game flow
```

## Implementation Notes

- `CalculateVictoryPoints(player)` already exists in rules.go
- `CheckVictory(gameState, playerId)` returns bool, exists but not called
- Add victory check to: `PlaceSettlement`, `PlaceCity`, road placement (after Longest Road check), `PlayDevCard`
- `GameOverPayload` exists in proto with `winner_id` and `scores`
- Consider: reveal all hands (VP cards) in GameOver payload
