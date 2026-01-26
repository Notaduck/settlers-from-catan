# E2E Test Status

Last full audit: January 22, 2026 - Iteration 2
Latest partial audit: January 26, 2026 - Iteration 2 (interactive-board)

## Latest Partial Audit (Iteration 2)

| Spec File                 | Passed | Failed | Flaky | Total | Status |
| ------------------------- | ------ | ------ | ----- | ----- | ------ |
| interactive-board.spec.ts | 7      | 0      | 0     | 7     | ✅ PASSING |

## Previous Partial Audit (Iteration 3)

| Spec File                 | Passed | Failed | Flaky | Total | Status |
| ------------------------- | ------ | ------ | ----- | ----- | ------ |
| interactive-board.spec.ts | 6      | 0      | 1     | 7     | ⚠️ FLAKY (placement highlight) |
| longest-road.spec.ts      | 1      | 6      | 0     | 7     | ❌ FAILING |
| ports.spec.ts             | 1      | 8      | 0     | 9     | ❌ FAILING |
| robber.spec.ts            | 5      | 2      | 0     | 7     | ❌ FAILING |
| setup-phase.spec.ts       | 2      | 0      | 0     | 2     | ✅ PASSING |
| trading.spec.ts           | 5      | 6      | 1     | 12    | ❌ FAILING (1 flaky) |
| victory.spec.ts           | 4      | 1      | 0     | 5     | ❌ FAILING |
| **TOTAL (Iteration 3)**   | **24** | **23** | **2** | **49** | ⚠️ MIXED |

## Previously Reported (Not Rerun in Iteration 3)

| Spec File                 | Passed | Failed | Total | Status | Notes |
| ------------------------- | ------ | ------ | ----- | ------ | ----- |
| game-flow.spec.ts         | 4      | 0      | 4     | ✅ PASSING | Last run Iteration 2 |
| development-cards.spec.ts | 2      | 10     | 12    | ⚠️ PARTIAL | Last run Iteration 2 (may be outdated) |

## Key Findings (Iteration 3)

**✅ WORKING**
- setup-phase.spec.ts passes 2/2.
- interactive-board.spec.ts mostly stable (1 flaky on valid placement highlighting).

**❌ FAILURES CLUSTER A - Trade UI availability**
- ports.spec.ts and trading.spec.ts frequently fail to find or click `[data-cy='bank-trade-btn']` or open bank trade modal.
- trading.spec.ts also fails trade offer modal appearance and trade phase button expectations.

**❌ FAILURES CLUSTER B - Longest Road UI/flow**
- longest-road.spec.ts fails on missing road length elements (`[data-cy='road-length-{playerId}']`) and inability to click `[data-cy='build-road-btn']`.
- Some setups show resource counts not matching expected preconditions.

**❌ FAILURES CLUSTER C - Robber steal UI**
- robber.spec.ts fails on steal modal visibility (`[data-cy='steal-modal']` not found) for adjacent-player and steal-transfer tests.
- Backend logs show `Failed to parse move robber request` during these failures.

**❌ FAILURES CLUSTER D - Victory flow setup**
- victory.spec.ts fails in “Game shows victory screen when player reaches 10 VP” due to missing `[data-cy='player-ore']` resource readout.

## Notable Flaky Tests (Iteration 3)

- interactive-board.spec.ts: `highlights valid placements during setup` (initial run failed, retry passed).
- trading.spec.ts: `Cannot propose trade without resources` (timeout opening propose trade modal).

## Detailed Failures (Iteration 3)

### interactive-board.spec.ts (6/7, 1 flaky)
- ⚠️ `highlights valid placements during setup` (valid vertex count 0 on first run)

### longest-road.spec.ts (1/7)
- ❌ `5 connected roads awards Longest Road bonus` (resource preconditions mismatch)
- ❌ `Longer road takes Longest Road from another player` (cannot click build-road button)
- ❌ `Tie does not transfer Longest Road (current holder keeps)` (cannot click build-road button)
- ❌ `Longest Road badge shows correct player name` (cannot click build-road button)
- ❌ `Road length displayed for all players` (missing `road-length-{playerId}` data-cy)
- ❌ `Longest Road updates in real-time for all players` (cannot click build-road button)

### ports.spec.ts (1/9)
- ❌ `Bank trade shows 4:1 by default` (bank trade button/modal not available)
- ❌ `Cannot bank trade without enough resources` (bank trade button/modal not available)
- ❌ `Player with 3:1 port sees 3:1 trade ratio` (bank trade button/modal not available)
- ❌ `Bank trade executes successfully with valid resources` (bank trade button/modal not available)
- ❌ `Port access validation - settlement on port vertex grants access` (bank trade button/modal not available)
- ❌ `2:1 specific port provides best ratio for that resource only` (bank trade button/modal not available)
- ❌ `Multiple players can have different port access` (bank trade button/modal not available)
- ❌ `Trade ratio updates when selecting different resources` (bank trade button/modal not available)

### robber.spec.ts (5/7)
- ❌ `Steal UI shows adjacent players` (steal modal not visible)
- ❌ `Stealing transfers a resource` (steal modal not visible)

### trading.spec.ts (5/12, 1 flaky)
- ❌ `Bank trade 4:1 works` (bank trade button/modal not available)
- ❌ `Trade recipient sees offer modal` (trade offer modal not visible)
- ❌ `Accepting trade transfers resources` (trade offer modal not visible)
- ❌ `Declining trade notifies proposer` (resource preconditions mismatch)
- ❌ `Cannot trade outside trade phase` (trade buttons not present/disabled check fails)
- ❌ `Bank trade button shows during trade phase only` (trade button not present/disabled check fails)
- ⚠️ `Cannot propose trade without resources` (flaky: propose trade button timeout)

### victory.spec.ts (4/5)
- ❌ `Game shows victory screen when player reaches 10 VP` (missing `[data-cy='player-ore']` resource readout)

## Notes

- Test artifacts (screenshots, traces) are in `frontend/test-results/`.
- Run single spec: `cd frontend && npx playwright test <spec>.spec.ts --reporter=list`
- Backend runs on port 8080 via Playwright webServer config.
- Update Jan 26, 2026: `interactive-board.spec.ts` rerun and passing 7/7 (no flake observed).
