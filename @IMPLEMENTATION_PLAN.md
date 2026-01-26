# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-26. Iteration N+1.
Sources reviewed: `specs/*`, backend (game logic, handlers), frontend (context, Board, Game UI, Playwright specs), proto (API, types), E2E_STATUS, latest test artifacts.

---

## PRIORITY 1: INTERACTIVE BOARD (E2E flaky)
- **Status:** ✅ Completed (2026-01-26)
- **Result:** Added `data-cy="placement-ready"` with deterministic counts in `frontend/src/components/Game/Game.tsx` and updated Playwright waits in `frontend/tests/interactive-board.spec.ts`. Placement highlight E2E now waits for valid vertices/edges to be ready.
- **Go tests:** none; already covered
- **Playwright tests:** updated `interactive-board.spec.ts` for highlight readiness waits

## PRIORITY 2: SETUP PHASE UI (passing)
- No new tasks unless regressions found. All criteria, selectors, and E2E pass as of latest audit.

## PRIORITY 3: VICTORY FLOW (passing)
- No new tasks unless regressions found. 
- Most recent failures were for missing selectors; those have been addressed (`player-{resource}` aliases in `PlayerPanel`). Ensure test maintains tolerance for resource counts after second setup.

## PRIORITY 4: ROBBER FLOW (E2E failing)
- **Status:** ✅ Completed (2026-01-26)
- **Result:** Sanitized MoveRobber payload to `{q, r}` only and reset steal modal state when robber phase starts; files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`.
- **Go tests:** none; no server-side changes
- **Playwright tests:** `frontend/tests/robber.spec.ts` (7/7 passing)

## PRIORITY 5: TRADING (trade UI gating, enum mismatch, expiry)
- **Task 1:** Normalize Frontend/Backend `TradeStatus` handling. Make frontend treat enum as number (not string) for trade/port UI and modal selectors; files: `frontend/src/components/Game/Game.tsx` and related modal components. ✅ Completed (2026-01-26)
  - Added `isTradeStatus` helper and used `TradeStatus.PENDING` with string fallback `"TRADE_STATUS_PENDING"` to handle protojson enum strings.
  - E2E run: `npx playwright test trading.spec.ts --reporter=list` (2026-01-26) → 4 passed, 8 failed. Failures include missing `[data-cy="bank-trade-btn"]`/trade modal visibility and resource count mismatch in multi-trade test. See `frontend/test-results/` artifacts.
- **Task 2:** Expire trades at end of turn; update trading logic in `backend/internal/game/state_machine.go` and cover with `trading_test.go`. 
- **Task 3:** Enforce one-active-trade-per-proposer rule in `backend/internal/game/trading.go` and test in `trading_test.go`.
- **Task 4:** Make Playwright trade tests deterministic—force non-7 rolls so TRADE phase is guaranteed; add helper in `frontend/tests/helpers.ts` and update `frontend/tests/trading.spec.ts` (and `ports.spec.ts` as needed).
- **Go tests:** All new/changed rules in `trading_test.go`
- **Playwright tests:** Stabilize `trading.spec.ts`, `ports.spec.ts`

## PRIORITY 6: DEVELOPMENT CARDS (per-turn rule/road placement)
- **Task 1:** Enforce one-dev-card-per-turn: add marker field in Proto (`types.proto`), update logic in backend (`devcards.go`/`state_machine.go`) and reflect in client logic; test in `devcards_test.go`.
- **Task 2:** Road Building card: Placement mode logic and UI (free roads, edge highlight without resource check) in `frontend/src/components/Board/placement.ts`, `Game.tsx`. Playwright E2E: test 2-for-1 placement in `development-cards.spec.ts`.
- **Task 3:** When dev card bought: add explicit toast in `Game.tsx`/`GameContext.tsx`.
- **Go tests:** Extend `devcards_test.go`
- **Playwright tests:** Add assertions to `development-cards.spec.ts`

## PRIORITY 7: LONGEST ROAD (flow/race issues)
- **Task:** Finalize deterministic E2E helpers for building minimum-length connected roads; relax test assertions in `longest-road.spec.ts` to tolerate length >= 5, and debug/correct road graph traversal for awarding/losing bonus. Files: `frontend/tests/longest-road.spec.ts`, backend logic in `longestroad.go` as needed for traversal, and UI display in `PlayerPanel`.
- **Go tests:** Extend/add edge cases to `longestroad_test.go`
- **Playwright tests:** Validate `longest-road.spec.ts`

## PRIORITY 8: PORTS (E2E failing)
- **Task 1:** Make port placement completely deterministic; ensure coastal positions standard; logic in `backend/internal/game/ports.go`, backed by asserts in `ports_test.go`/`board_test.go`.
- **Task 2:** Normalize port/resource/enum mapping in `frontend/src/components/Game/BankTradeModal.tsx` and all consumers; ensure correct display of ratios for each player based on port access.
- **Go tests:** Update/extend `ports_test.go`, `board_test.go`
- **Playwright tests:** Stabilize/verify `ports.spec.ts`

---

## E2E STABILIZATION (MANDATORY)
Group and enumerate failed E2E cases by spec:

- **interactive-board.spec.ts (1 flaky):** Fix highlight/select stable waiting or marker in UI+test (`Board.tsx`, `placement.ts`, `interactive-board.spec.ts`).
- **longest-road.spec.ts (multiple):** Update test helpers and backend traversal to ensure robust/consistent Longest Road calculation. Fix assertion to accept `>= 5` as valid, debug bonus logic for real-time update.
- **ports.spec.ts (all failing):** Fix port/trade state, normalize enums/ratios, align E2E test with backend deterministic port layout. File: `BankTradeModal.tsx`, `ports.go`, test helpers.
- **robber.spec.ts (steal modal):** Align client MoveRobber message structure; ensure modal reset per phase. File: `GameContext.tsx`, `Game.tsx`, selectors in Playwright.
- **trading.spec.ts (trade modal):** Force TRADE phase w/ deterministic dice; align test/UI on status enum and phase/click preconditions; ensure modal selectors are present.
- **victory.spec.ts:** Monitored after previous fix; assert passing unless new failures.
- **development-cards.spec.ts:** Re-run after per-turn/placement updates—should pass after changes above.

---

## Reasoning & Assumptions
- Backend logic covers core rules but deterministic state, enum/string mismatches, and placement flow still break end-to-end automation.
- Interactive board and placement highlights unblock all setup, building, and road/port actions—thus highest priority.
- Robber, trade, and port failures are mainly data shape, enum, and test determinism, not rules bugs.
- All fixes keep build small/unit-testable with referenced files/tests listed above.

---

#END
