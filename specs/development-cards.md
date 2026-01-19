# Development Cards

**Priority**: MEDIUM - Adds strategic depth, enables Largest Army

## Overview

Development cards are purchased for (1 ore, 1 wheat, 1 sheep) and provide one-time effects or victory points.

## Card Types

| Card           | Count | Effect                                           |
| -------------- | ----- | ------------------------------------------------ |
| Knight         | 14    | Move robber + steal. Counts toward Largest Army. |
| Victory Point  | 5     | +1 VP (hidden until game end or 10 VP)           |
| Road Building  | 2     | Place 2 roads for free                           |
| Year of Plenty | 2     | Take any 2 resources from bank                   |
| Monopoly       | 2     | Name a resource, all players give you theirs     |

Total: 25 cards

## Acceptance Criteria

### Deck Management

- [ ] Initialize shuffled deck of 25 cards at game start
- [ ] Track remaining cards in deck
- [ ] Track each player's hand of dev cards
- [ ] Distinguish "just bought this turn" (can't play) vs "playable"

### Buying Cards

- [ ] "Buy Development Card" button in build phase
- [ ] Costs: 1 ore, 1 wheat, 1 sheep
- [ ] Card drawn from deck, added to player hand
- [ ] Card NOT playable until next turn
- [ ] Show "You drew: Knight" notification

### Playing Cards - General

- [ ] Can play 1 dev card per turn (except VP)
- [ ] Can play before or after rolling
- [ ] Knight can be played to avoid rolling 7 consequences
- [ ] VP cards revealed automatically at game end

### Knight Effect

- [ ] Move robber to new hex
- [ ] Steal from adjacent player (same as rolling 7)
- [ ] Increment player's knight count
- [ ] Check for Largest Army after play

### Road Building Effect

- [ ] Enter "place 2 roads" mode
- [ ] Roads placed for free (no resource cost)
- [ ] Must place both roads (if able)
- [ ] Normal road placement rules apply

### Year of Plenty Effect

- [ ] Modal to select 2 resources
- [ ] Resources taken from bank
- [ ] Can pick same resource twice

### Monopoly Effect

- [ ] Modal to select 1 resource type
- [ ] All other players lose ALL of that resource
- [ ] Active player gains all collected resources
- [ ] Show count stolen: "You collected 7 wheat!"

### Victory Point Cards

- [ ] Hidden from other players until revealed
- [ ] Count toward victory but don't show in public VP
- [ ] Auto-reveal when player reaches 10+ VP
- [ ] Titles: "Library", "Market", "Chapel", "University", "Palace"

### Data Attributes (for Playwright)

- [ ] `data-cy="buy-dev-card-btn"`
- [ ] `data-cy="dev-card-{type}"` in hand display
- [ ] `data-cy="play-dev-card-btn"` on each playable card
- [ ] `data-cy="monopoly-select-{resource}"`
- [ ] `data-cy="year-of-plenty-select-{resource}"`

## Required Go Unit Tests

File: `backend/internal/game/devcards_test.go` (new)

```go
// Test: InitDevCardDeck creates 25 shuffled cards
// Test: BuyDevCard deducts resources and adds to hand
// Test: BuyDevCard fails if insufficient resources
// Test: BuyDevCard fails if deck empty
// Test: Newly bought card not playable this turn
// Test: PlayKnight moves robber and increments knight count
// Test: PlayKnight triggers Largest Army check
// Test: PlayRoadBuilding allows 2 free road placements
// Test: PlayYearOfPlenty adds 2 resources to player
// Test: PlayMonopoly collects all of resource from other players
// Test: VictoryPointCard adds hidden VP
// Test: Cannot play more than 1 dev card per turn
```

## Required Playwright E2E Tests

File: `frontend/tests/development-cards.spec.ts`

```typescript
// Test: Can buy development card with correct resources
// Test: Cannot buy without resources
// Test: Playing knight moves robber
// Test: Monopoly collects resources from all players
// Test: Year of plenty grants 2 resources
// Test: Road building allows 2 free roads
// Test: VP cards count toward victory
```

## Implementation Notes

- DevCard deck: `[]DevCardType` shuffled at game start
- Player hand: `map[DevCardType]int` for counts
- Track `playedDevCardThisTurn bool` per player
- Track `boughtDevCardsThisTurn []DevCardType` to prevent same-turn play
- Knight count for Largest Army: `player.knightsPlayed int`
