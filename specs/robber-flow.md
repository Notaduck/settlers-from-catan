# Robber Flow - Move, Steal, Discard

**Priority**: HIGH - Core game mechanic triggered every ~6 rolls

## Overview

When a 7 is rolled:

1. Players with >7 cards must discard half
2. Active player moves the robber to a new hex
3. Active player steals 1 random card from a player adjacent to that hex

Backend has partial implementation (robber activation, discard calculation, production blocking). Missing: move handler, steal logic, discard message, and all UI.

## Acceptance Criteria

### Discard Phase (>7 cards)

- [ ] When 7 rolled, players with >7 cards see discard modal
- [ ] Modal shows: "Discard X cards" (half, rounded down)
- [ ] Player selects which cards to discard
- [ ] Submit button disabled until correct count selected
- [ ] All discards must complete before robber moves
- [ ] Other players see "Waiting for X to discard..."

### Move Robber

- [ ] After discards, active player sees "Move the Robber"
- [ ] All hexes except current robber location are clickable
- [ ] Clicking hex moves robber there
- [ ] Desert IS a valid robber destination
- [ ] Robber icon moves to new hex on board

### Steal Phase

- [ ] After robber placed, if players adjacent, show steal choice
- [ ] List players with settlements/cities on that hex
- [ ] Player icons show (but not their cards - hidden info)
- [ ] Clicking player steals 1 random resource from them
- [ ] Stolen resource added to active player, shown in notification
- [ ] If no adjacent players, skip steal phase

### Backend Implementation Needed

- [ ] `handleMoveRobber` handler in handlers.go
- [ ] `MoveRobber(gameState, playerId, hexCoord)` command
- [ ] `StealFromPlayer(gameState, thiefId, victimId)` - random card
- [ ] `DiscardCards(gameState, playerId, cards)` command
- [ ] Add `discard_cards` message type to proto

### Data Attributes (for Playwright)

- [ ] `data-cy="discard-modal"` for discard UI
- [ ] `data-cy="discard-card-{resource}"` for each card type
- [ ] `data-cy="discard-submit"` submit button
- [ ] `data-cy="robber-hex-{q}-{r}"` clickable hexes
- [ ] `data-cy="steal-player-{id}"` steal target buttons

## Required Go Unit Tests

File: `backend/internal/game/robber_test.go` (new)

```go
// Test: MoveRobber updates robber position
// Test: MoveRobber fails if same hex
// Test: MoveRobber fails if not active player
// Test: StealFromPlayer transfers random resource
// Test: StealFromPlayer fails if victim has no cards
// Test: StealFromPlayer fails if victim not adjacent to robber
// Test: DiscardCards removes correct resources
// Test: DiscardCards fails if wrong count
// Test: DiscardCards fails if player doesn't have cards
```

## Required Playwright E2E Tests

File: `frontend/tests/robber.spec.ts`

```typescript
// Test: Rolling 7 shows discard modal for players with >7 cards
// Test: Discard modal enforces correct card count
// Test: After discard, robber move UI appears
// Test: Clicking hex moves robber
// Test: Steal UI shows adjacent players
// Test: Stealing transfers a resource
```

## Implementation Notes

- Robber hex stored in `gameState.board.robberHex`
- Use `GetPlayersToDiscard(gameState)` to find who must discard
- Steal is random: pick random index from victim's resource array
- Consider: block turn progression until robber flow complete
