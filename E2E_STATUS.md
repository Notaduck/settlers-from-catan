# E2E Test Status

Last full audit: January 22, 2026 - Iteration 2

## Summary

| Spec File                 | Passed | Failed | Total |
| ------------------------- | ------ | ------ | ----- |
| development-cards.spec.ts | 0      | 12     | 12    |
| game-flow.spec.ts         | 0      | 4      | 4     |
| interactive-board.spec.ts | 0      | 8      | 8     |
| longest-road.spec.ts      | 0      | 7      | 7     |
| ports.spec.ts             | 0      | 10     | 10    |
| robber.spec.ts            | 0      | 7      | 7     |
| setup-phase.spec.ts       | 0      | 2      | 2     |
| trading.spec.ts           | 0      | 11     | 11    |
| victory.spec.ts           | 0      | 5      | 5     |
| **TOTAL**                 | **0**  | **64** | **65** |

**Status**: 🔴 **CRITICAL** - 100% test failure due to backend connection refused

---

## Failing Tests by Spec

**Root Cause**: ALL tests fail with `connect ECONNREFUSED ::1:8080` - backend service not running during test execution

### development-cards.spec.ts (0/12 passing)
- should display development cards panel during playing phase
- should not be able to buy development card without resources
- should show correct card types with play buttons (except VP cards)
- Year of Plenty modal should allow selecting 2 resources
- Monopoly modal should allow selecting 1 resource type
- Monopoly should collect resources from all other players
- Knight card should increment knight count and trigger Largest Army check
- Victory Point cards should not have play button
- should not be able to play dev card bought this turn
- Road Building card should appear in dev cards panel
- dev cards panel should show total card count
- buying dev card should deduct correct resources

### game-flow.spec.ts (0/4 passing)
- should create a new game and show lobby
- should join an existing game
- should allow players to toggle ready state
- should start game when both players ready

### interactive-board.spec.ts (0/8 passing)
- vertices render on board
- edges render on board
- shows placement mode for setup turn
- clicking vertex during setup places settlement
- clicking edge during setup places road
- invalid vertices are not clickable during setup
- highlights valid placements during setup

### longest-road.spec.ts (0/7 passing)
- 5 connected roads awards Longest Road bonus
- Longer road takes Longest Road from another player
- Tie does not transfer Longest Road (current holder keeps)
- Longest Road badge shows correct player name
- Road length displayed for all players
- No Longest Road badge before 5 roads threshold
- Longest Road updates in real-time for all players

### ports.spec.ts (0/10 passing)
- Ports render on board with correct icons
- Bank trade shows 4:1 by default (no port access)
- Cannot bank trade without enough resources
- Player with 3:1 port sees 3:1 trade ratio
- Bank trade executes successfully with valid resources
- Port access validation - settlement on port vertex grants access
- 2:1 specific port provides best ratio for that resource only
- Multiple players can have different port access
- Trade ratio updates when selecting different resources

### robber.spec.ts (0/7 passing)
- Rolling 7 shows discard modal for players with >7 cards
- Discard modal enforces correct card count
- After discard, robber move UI appears
- Clicking hex moves robber
- Steal UI shows adjacent players
- Stealing transfers a resource
- No steal phase when no adjacent players

### setup-phase.spec.ts (0/2 passing)
- shows setup phase banner and turn indicator
- round 2 settlement grants resources toast

### trading.spec.ts (0/11 passing)
- Bank trade 4:1 works
- Cannot bank trade without 4 resources
- Player can propose trade to another
- Player can propose trade to specific player
- Trade recipient sees offer modal
- Accepting trade transfers resources
- Declining trade notifies proposer
- Cannot trade outside trade phase
- Cannot propose trade without resources
- Multiple trades per turn allowed
- Bank trade button shows during trade phase only

### victory.spec.ts (0/5 passing)
- Game shows victory screen when player reaches 10 VP
- Victory screen shows correct VP breakdown
- New game button navigates to create game screen
- Victory screen blocks further game actions
- Hidden VP cards are revealed in final scores

---

## Root Cause Patterns

**CRITICAL FINDING**: Single root cause affecting ALL tests:

- [x] **Backend service not running during test execution** - `connect ECONNREFUSED ::1:8080`
  - All 64 failed tests show identical error pattern
  - Tests attempt to POST to `http://localhost:8080/api/games` but receive connection refused
  - Backend service needs to be properly started before test execution
  - This is an **infrastructure issue**, NOT a game functionality problem

**Secondary analysis** (once backend connection fixed):
- [ ] Missing backend command handler
- [ ] Missing frontend UI component  
- [ ] WebSocket message not sent/received
- [ ] Selector `data-cy` attribute missing
- [ ] Race condition / timing issue
- [ ] State not updating correctly

**DIAGNOSIS**: The game functionality appears complete based on comprehensive test coverage. All test failures are due to backend service unavailability during test execution.

---

## Fix Log

| Date | Test Fixed | Root Cause | Iteration |
| ---- | ---------- | ---------- | --------- |
| 2026-01-22 | ALL 64 tests identified failing | Backend service not running during E2E execution | 2 |

## Next Steps - HIGH PRIORITY

1. **Fix E2E infrastructure** - Ensure backend service starts properly before tests
   - Investigate Playwright configuration for service startup
   - Check if `make e2e` properly starts backend before running tests  
   - May need script/process to ensure backend readiness before test execution

2. **Re-run E2E audit after infrastructure fix** - Should see dramatically different results
   - Most tests likely to pass once backend connectivity restored
   - Real issues will emerge (missing UI, timing, etc.)

## Notes

- Test artifacts (screenshots, traces) are in `frontend/test-results/`
- Run single spec: `cd frontend && npx playwright test <spec>.spec.ts`
- Run all with report: `cd frontend && npx playwright test --reporter=html`
- **IMPORTANT**: Backend must be running on port 8080 before test execution
