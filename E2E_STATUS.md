# E2E Test Status

Last full audit: January 27, 2026 - Iteration 12

## Full Audit Results (Iteration 12)

| Spec File                 | Passed | Failed | Flaky | Total | Status |
| ------------------------- | ------ | ------ | ----- | ----- | ------ |
| development-cards.spec.ts | 0      | 24     | 0     | 24    | ❌ FAIL (all tests failed to run to completion)
| interactive-board.spec.ts | 0      | 14     | 0     | 14    | ❌ FAIL (all tests failed)
| setup-phase.spec.ts       | 0      | ~7     | 0     | ~7    | ❌ FAIL (all tests failed)
| game-flow.spec.ts         | 0      | 8      | 0     | 8     | ❌ FAIL (lobby/game start E2E failure)
| ports.spec.ts             | 0      | 5      | 0     | 5     | ❌ FAIL (port rendering/trade fails)
| victory.spec.ts           | 0      | ~5     | 0     | ~5    | ❌ FAIL (victory detection/game over)
| robber.spec.ts            | 0      | ~6     | 0     | ~6    | ❌ FAIL (robber flow discard/move/steal fails)
| trading.spec.ts           | 0      | ~6     | 0     | ~6    | ❌ FAIL (all trade UI/acceptance fails)
| longest-road.spec.ts      | 0      | 8      | 0     | 8     | ❌ FAIL (all road bonus UI/logic fails)

**Total run:** 83 / **Passed:** 0 / **Failed:** 83

## Failing Test Details

**development-cards.spec.ts**
- All tests failed (panel not rendered, button not active, no UI updating for buys/plays)
- Monopoly, Knight, Road Building, Year of Plenty all fail at modal/activation step

**game-flow.spec.ts**
- All tests fail at or before game lobby, create/join buttons, or player ready

**interactive-board.spec.ts**
- All tests fail: board/interactivity, setup placements not activating, vertex/edge selection broken

**setup-phase.spec.ts**
- All tests fail: phase banners/instructions never shown, setup never completes

**victory.spec.ts**
- All tests fail: victory condition/game-over overlay never triggers

**robber.spec.ts**
- All tests fail: discard modal, robber move, steal target selection not rendered

**trading.spec.ts**
- All tests fail: bank trade modal, offer, accept/decline, and proposal modals not working

**ports.spec.ts**
- All tests fail: ports do not appear, port-based trade unavailable

**longest-road.spec.ts**
- All tests fail: road badge, DFS, bonus logic never updates/shows for any player

## Failure Pattern Notes
- Failures are global and seem to affect all interactivity, game phase transitions, and UI rendering.
- Likely root cause is backend server not responding, major websocket/proto break, or build artifact mismatch.
- No tests even make it past initial UI or lobby for any suite.

## Action Items
- See IMPLEMENTATION_PLAN.md for atomic fix tasks generated from these failures.
- For each spec file, create at least one atomic task corresponding to root cause or major failure for that flow.
- Hypothesis: Backend protocol or state flow is failing. Immediate next step: Backend/frontend integration validation, or protocol hotfix, must be performed.

