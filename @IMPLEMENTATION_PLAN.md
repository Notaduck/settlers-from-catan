# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-23. Sources: `specs/*`, backend/frontend code scan, `E2E_STATUS.md`, `frontend/test-results/`.
Quick E2E audit command (`cd frontend && npx playwright test --reporter=list 2>&1 | head -100`) timed out after 120s.

Current snapshot:
- Backend rules appear feature-complete for core specs (robber/trading/dev cards/longest road/ports/victory) with unit tests in place.
- Primary gaps are frontend build controls, UI selectors, enum normalization, and a small set of backend wiring gaps (trade expiry).
- E2E failures align with missing UI affordances and mismatched payload/selector expectations.

---

## PRIORITY 1: INTERACTIVE BOARD (1 FLAKY TEST)
- Stabilize setup placement highlights (valid vertices intermittently absent).
  - Files: `frontend/src/components/Board/placement.ts`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/Game.tsx`
  - Go tests: none
  - Playwright: stabilize `frontend/tests/interactive-board.spec.ts` (wait for `data-cy="placement-mode"` + non-zero `.vertex--valid` count before assertion).

## PRIORITY 2: SETUP PHASE UI (PASSING)
- No changes planned (E2E passes 2/2).

## PRIORITY 3: VICTORY FLOW (UI SELECTOR + BUILD CONTROLS)
- Add/align resource selector used in victory test (`player-ore` vs `resource-ore`) and expose build-city control.
  - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/placement.ts`
  - Go tests: none
  - Playwright: update `frontend/tests/victory.spec.ts` to use correct selector and build flow (`build-city-btn`).

## PRIORITY 4: ROBBER FLOW (MOVE PAYLOAD + MODAL RESET)
- Fix MoveRobber client payload to match proto (remove `s` coord); reset Steal modal when new robber phase starts.
  - Files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
  - Go tests: extend `backend/internal/game/robber_test.go` only if behavior changes (otherwise none)
  - Playwright: rerun `frontend/tests/robber.spec.ts` (steal modal should appear).

## PRIORITY 5: TRADING (ENUM NORMALIZATION + EXPIRY)
- Normalize `TradeStatus` handling (string vs numeric enum) so incoming trade modal appears.
  - Files: `frontend/src/components/Game/Game.tsx`
  - Go tests: none
  - Playwright: update/re-run `frontend/tests/trading.spec.ts` (recipient sees offer modal).

- Expire pending trades on end turn (spec requirement).
  - Files: `backend/internal/game/state_machine.go`
  - Go tests: extend `backend/internal/game/trading_test.go` (verify expiry after `EndTurn`)
  - Playwright: add to `frontend/tests/trading.spec.ts` (expired trades dismissed).

- Enforce one active trade per proposer (spec requirement).
  - Files: `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`
  - Go tests: add table case for duplicate pending trade
  - Playwright: add to `frontend/tests/trading.spec.ts` (second propose blocked).

## PRIORITY 6: DEVELOPMENT CARDS (RULES + ROAD BUILDING PLACEMENT)
- Enforce "one dev card per turn" and allow Knight to trigger robber flow.
  - Files: `proto/catan/v1/types.proto` (new per-player turn marker), `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
  - Go tests: extend `backend/internal/game/devcards_test.go` (one-per-turn + knight->robber phase)
  - Playwright: extend `frontend/tests/development-cards.spec.ts`.

- Road Building placement should highlight edges even without resources (use `road_building_roads_remaining`).
  - Files: `frontend/src/components/Board/placement.ts`, `frontend/src/components/Game/Game.tsx`
  - Go tests: none
  - Playwright: extend `frontend/tests/development-cards.spec.ts` (road-building placement mode).

## PRIORITY 7: LONGEST ROAD (UI MISSING)
- Add longest road holder + per-player road length UI (`data-cy="longest-road-holder"`, `data-cy="road-length-{playerId}"`).
  - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`
  - Go tests: none
  - Playwright: update `frontend/tests/longest-road.spec.ts`.

- Add build controls for road/settlement/city (required by E2E helpers).
  - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/Game.css`, `frontend/src/components/Board/placement.ts`
  - Go tests: none
  - Playwright: update `frontend/tests/longest-road.spec.ts`, `frontend/tests/victory.spec.ts`, `frontend/tests/trading.spec.ts` (use build buttons).

## PRIORITY 8: PORTS (DETERMINISM + UI NORMALIZATION)
- Replace random port placement with fixed standard coastal positions (deterministic).
  - Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`
  - Go tests: update `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go`
  - Playwright: rerun `frontend/tests/ports.spec.ts` for stability.

- Normalize port enum handling in BankTradeModal (string vs numeric).
  - Files: `frontend/src/components/Game/BankTradeModal.tsx`
  - Go tests: none
  - Playwright: update/validate `frontend/tests/ports.spec.ts`.

---

## E2E STABILIZATION (MANDATORY)
Source: `E2E_STATUS.md` (last audit 2026-01-22). Quick audit timed out after 120s.

### Failing Spec Group: `interactive-board.spec.ts` (1 flaky)
- Likely cause: placement state not ready when assertion runs.
- Fix E2E: wait for `data-cy="placement-mode"` + `.vertex--valid` count > 0 before asserting highlight.

### Failing Spec Group: `longest-road.spec.ts` (6 failing)
- Likely cause: missing `build-road-btn` and missing road-length UI.
- Fix E2E: add build controls + `data-cy="road-length-{playerId}"` and `data-cy="longest-road-holder"`.

### Failing Spec Group: `ports.spec.ts` (8 failing)
- Likely cause: trade UI not visible (phase mismatch) + non-deterministic ports.
- Fix E2E: ensure trade phase after forced dice roll; use deterministic port placement; normalize port enums in UI.

### Failing Spec Group: `robber.spec.ts` (2 failing)
- Likely cause: `moveRobber` payload contains extra `s`, parse error -> no steal modal.
- Fix E2E: send `{q,r}` only; reset steal modal state on new robber phase.

### Failing Spec Group: `trading.spec.ts` (6 failing + 1 flaky)
- Likely cause: incoming trade modal gated on numeric enum; trade UI hidden when not in trade phase.
- Fix E2E: normalize trade status handling; add phase waits or toggle to TRADE; ensure bank-trade/propose-trade buttons present.

### Failing Spec Group: `victory.spec.ts` (1 failing)
- Likely cause: selector mismatch (`player-ore` vs `resource-ore`) + missing build-city button.
- Fix E2E: align selectors and expose build controls.

### Unverified/Outdated
- `development-cards.spec.ts` last full run 2026-01-22 (status may be stale). Re-run after dev-card rule updates.

