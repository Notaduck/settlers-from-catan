# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-26.
Sources reviewed: `specs/*`, `backend/internal/game/*.go`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/**/*`, `frontend/tests/*.spec.ts`, `proto/catan/v1/*.proto`, `E2E_STATUS.md`, `frontend/test-results/`.
Quick E2E audit command (`cd frontend && npx playwright test --reporter=list 2>&1 | head -100`) timed out after 10s.

Current snapshot:
- Backend rules for robber/trading/dev cards/longest road/ports/victory appear implemented and unit-tested; remaining gaps are mostly UI/selector mismatches and a couple of backend rule gaps (trade expiry, one-trade-per-proposer, dev-card-per-turn).
- Frontend is missing build controls for BUILD phase (road/settlement/city selection), longest-road UI, and a few data-cy alignment fixes.
- E2E failures align with trade UI gating (dice=7 randomness), trade status enum mismatch, missing build buttons, and robber move/steal flow not surfacing the steal modal.

---

## PRIORITY 1: INTERACTIVE BOARD (1 flaky test)
- Stabilize setup placement highlights by ensuring UI/test waits for non-empty `.vertex--valid` before assertion; files: `frontend/src/components/Board/placement.ts`, `frontend/src/components/Board/Board.tsx`, `frontend/tests/interactive-board.spec.ts`; Go tests: none; Playwright: update `interactive-board.spec.ts` to wait for valid vertices count > 0 and/or add a deterministic “placement ready” marker.

## PRIORITY 2: SETUP PHASE UI (passing)
- No changes planned unless regressions appear; files: none; Go tests: none; Playwright: none.

## PRIORITY 3: VICTORY FLOW + BUILD CONTROLS (blocks victory/longest-road E2E)
- Add BUILD-phase controls (`data-cy='build-road-btn'`, `build-settlement-btn`, `build-city-btn'`) and a selected build-mode state so board clicks only apply to the chosen structure; files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/Game.css`, `frontend/src/components/Board/placement.ts`; Go tests: none; Playwright: update `frontend/tests/helpers.ts`, `victory.spec.ts`, `longest-road.spec.ts` to use build buttons consistently.
- Add city upgrade placement validation + highlighting in build mode (only your settlements, not already cities); files: `frontend/src/components/Board/placement.ts`; Go tests: none; Playwright: add/extend victory flow step for city build in `frontend/tests/victory.spec.ts`.
- Align resource selectors used by victory spec by adding `data-cy='player-{resource}'` aliases (keep existing `resource-{resource}`); files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`; Go tests: none; Playwright: `frontend/tests/victory.spec.ts` should pass without selector edits.

## PRIORITY 4: ROBBER FLOW (steal modal not visible)
- Fix MoveRobber payload shape and steal triggering: ensure client sends only `{q,r}` HexCoord and that steal requests don’t fail proto parsing; reset steal modal state when a new robber phase starts; files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`; Go tests: add cases in `backend/internal/game/robber_test.go` only if server-side behavior changes; Playwright: re-run `frontend/tests/robber.spec.ts` and stabilize any waits around robber-hex clicks.

## PRIORITY 5: TRADING (phase gating + enum mismatch + expiry)
- Normalize trade status handling in UI (string enum vs numeric) so incoming trade modal appears; files: `frontend/src/components/Game/Game.tsx`; Go tests: none; Playwright: `frontend/tests/trading.spec.ts` (recipient modal should appear).
- Make trade tests deterministic by forcing non-7 dice rolls before expecting TRADE phase; files: `frontend/tests/trading.spec.ts` (use `forceDiceRoll` helper), possibly `frontend/tests/helpers.ts` to add a `rollDiceToTradePhase` helper; Go tests: none.
- Expire pending trades at end of turn; files: `backend/internal/game/state_machine.go`; Go tests: extend `backend/internal/game/trading_test.go` to verify pending trades removed after `EndTurn`; Playwright: add a trade-expiry assertion to `frontend/tests/trading.spec.ts`.
- Enforce “one active trade per proposer”; files: `backend/internal/game/trading.go`; Go tests: add table case in `backend/internal/game/trading_test.go`; Playwright: add test for duplicate propose blocked in `frontend/tests/trading.spec.ts`.

## PRIORITY 6: DEVELOPMENT CARDS (per-turn rule + road building placement)
- Enforce “one dev card per turn” (add per-player turn marker); files: `proto/catan/v1/types.proto` (new field), `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`, `frontend/src/context/GameContext.tsx`; Go tests: extend `backend/internal/game/devcards_test.go` for one-per-turn; Playwright: add test to `frontend/tests/development-cards.spec.ts`.
- Road Building placement mode: surface remaining free roads and allow edge highlights without resources when `road_building_roads_remaining > 0`; files: `frontend/src/components/Board/placement.ts`, `frontend/src/components/Game/Game.tsx`; Go tests: none; Playwright: add a road-building placement test to `frontend/tests/development-cards.spec.ts`.
- Show “You drew: {Card}” toast on `devCardBought` (client-only, not broadcast); files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`; Go tests: none; Playwright: add a simple assertion in `frontend/tests/development-cards.spec.ts`.

## PRIORITY 7: LONGEST ROAD (UI missing)
- Show longest-road holder + per-player road length with data-cy attributes; reuse existing client-side road-length calculation; files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx` (pass board); Go tests: none; Playwright: update `frontend/tests/longest-road.spec.ts` to assert `[data-cy='longest-road-holder']` and `road-length-{playerId}`.

## PRIORITY 8: PORTS (determinism + enum normalization)
- Make port placement deterministic at standard coastal positions; files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`; Go tests: update `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go` to assert deterministic layout; Playwright: re-run `frontend/tests/ports.spec.ts`.
- Normalize port enums in BankTradeModal (string vs numeric) so ratios resolve; files: `frontend/src/components/Game/BankTradeModal.tsx`; Go tests: none; Playwright: confirm ratio display in `frontend/tests/ports.spec.ts`.

---

## E2E STABILIZATION (MANDATORY)

Failing spec groups from `E2E_STATUS.md` (2026-01-22):
- `interactive-board.spec.ts` (1 flaky): likely placement-state not ready; fix task = add explicit wait for non-zero valid vertices or add a placement-ready marker in UI.
- `longest-road.spec.ts` (6 failing): missing build buttons + road-length UI; fix task = add build controls + road-length display + update tests to use buttons.
- `ports.spec.ts` (8 failing): trade UI absent when roll hits 7 + port enums not normalized; fix task = force dice to non-7 in tests + normalize enums + deterministic ports.
- `robber.spec.ts` (2 failing): steal modal never appears; fix task = align moveRobber payload and reset steal modal state on new robber phase.
- `trading.spec.ts` (6 failing + 1 flaky): trade modal gating on enum mismatch and random dice; fix task = normalize TradeStatus + force dice roll in tests + add expiry/duplicate-trade rules.
- `victory.spec.ts` (1 failing): selector mismatch (`player-ore`) + missing build-city control; fix task = add data-cy alias + build-city button/flow.
- `development-cards.spec.ts` (not rerun in recent audit): re-run after dev-card-per-turn and road-building placement updates.

---

## Reasoning & Assumptions
- Trade UI failures are most likely from non-deterministic dice rolls hitting 7, which skips TRADE phase; the plan chooses deterministic dice in tests instead of changing core rules.
- Robber “parse error” likely stems from client payload shape; safest fix is to ensure HexCoord is `{q,r}` and avoid sending undefined/extra fields, while also resetting modal state to avoid stale UI.
- Build controls are required because E2E helpers expect explicit build buttons and current UI allows implicit build-on-click without structure selection.
