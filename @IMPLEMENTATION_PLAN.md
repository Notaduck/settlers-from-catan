# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-02-01 (Iteration 22 / quick E2E audit + spec/code sync)

---

## AUDIT NOTES (Iteration 22)
- Reviewed specs, backend game/handlers, frontend context/board/game, and proto contracts.
- Quick E2E audit run: `cd frontend && npx playwright test --reporter=list 2>&1 | head -100`. The pipe cut the run early; logs show WS initially closed then reconnected and game state reached SETUP with setupPhase updates.
- `frontend/test-results/` now only contains one failure artifact: `development-cards-...-panel-during-playing-phase`. Snapshot shows the game stuck in **Setup Phase - Round 2** with `Place Road (2/2)` after `completeSetupPhase`, so setup completion -> PLAYING transition is still flaky or the UI/test sync is missing a wait.
- UI gaps confirmed by code scan: missing `data-cy="placement-ready"` element; no build/trade phase buttons; no build action buttons (`build-road-btn`, etc.); trade modals not mounted; `/test/grant-dev-card` endpoint returns 501 (unimplemented).

---

## E2E STABILIZATION (Grouped by Spec + Root Cause)

### interactive-board.spec.ts
- **Root cause:** `data-cy="placement-ready"` missing (tests poll valid vertex/edge counts).
- **Fix:** Add placement-ready element with `data-valid-vertices` / `data-valid-edges` derived from `placementState`.

### setup-phase.spec.ts
- **Root cause:** Setup completion to PLAYING appears flaky (dev-card test snapshot stuck in Round 2).
- **Fix:** Verify `PlaceSetupRoad` advances setup to PLAYING; ensure client gets the final state; add explicit wait for `data-cy="game-phase"` -> PLAYING after `completeSetupPhase`.

### development-cards.spec.ts
- **Root cause:** Dev cards panel unreachable because game remains in SETUP; DEV_MODE endpoint missing for `grant-dev-card`.
- **Fix:** Implement `/test/grant-dev-card`; unblock PLAYING transition after setup.

### trading.spec.ts
- **Root cause:** Missing trade/build phase buttons and trade modals wiring.
- **Fix:** Add phase toggles, mount BankTrade/Propose/Incoming modals, wire to GameContext actions.

### longest-road.spec.ts
- **Root cause:** Missing build action buttons (`build-road-btn`) and build phase toggle.
- **Fix:** Add build controls; ensure road placement highlights drive valid edges.

### robber.spec.ts
- **Root cause:** Blocked by setup completion; robber UI wiring should be rechecked after PLAYING stability.
- **Fix:** Ensure discard/move/steal UI gating matches `robberPhase` and board exposes `robber-hex-*` selectors.

### ports.spec.ts
- **Root cause:** Ports render but bank trade modal not mounted.
- **Fix:** Mount BankTrade modal and surface port ratio UI; re-run E2E.

### victory.spec.ts
- **Root cause:** No current failure artifact; re-verify after setup/build flow stabilizes.

### game-flow.spec.ts
- **Root cause:** No current failure artifact; re-verify after setup/build flow stabilizes.

---

## PRIORITIZED IMPLEMENTATION PLAN (Atomic Tasks)

1) **Add placement-ready indicator + counts**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/SetupPhasePanel.tsx`, `frontend/src/components/Game/Game.css`
   - Scope: Render `data-cy="placement-ready"` with `data-valid-vertices` / `data-valid-edges` counts from `placementState` for setup/build.
   - Go unit tests: none.
   - Playwright: `frontend/tests/interactive-board.spec.ts`, `frontend/tests/setup-phase.spec.ts`

2) **Make setup completion deterministic and wait for PLAYING**
   - Files: `backend/internal/game/state_machine.go` (if setup advance bug found), `frontend/tests/helpers.ts`, `frontend/src/components/Game/Game.tsx`
   - Scope: Confirm final setup road advances to PLAYING; add helper wait for `data-cy="game-phase"` -> PLAYING after `completeSetupPhase`.
   - Go unit tests: add regression test if setup advance bug found.
   - Playwright: `frontend/tests/development-cards.spec.ts`, `frontend/tests/trading.spec.ts`, `frontend/tests/robber.spec.ts`, `frontend/tests/longest-road.spec.ts`, `frontend/tests/ports.spec.ts`

3) **Implement DEV_MODE `/test/grant-dev-card`**
   - Files: `backend/internal/handlers/handlers.go`, `backend/internal/handlers/handlers_test.go`
   - Scope: Add handler to grant a specific dev card, persist state, broadcast updated game state.
   - Go unit tests: add handler test cases for valid/invalid card types and missing player.
   - Playwright: `frontend/tests/development-cards.spec.ts`

4) **Add build/trade phase controls + build action buttons**
   - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/placement.ts`
   - Scope: Add `data-cy="trade-phase-btn"` / `data-cy="build-phase-btn"` toggles; add `build-road-btn`, `build-settlement-btn`, `build-city-btn` to set placement intent; update placementState to respect selected build type.
   - Go unit tests: none.
   - Playwright: `frontend/tests/trading.spec.ts`, `frontend/tests/longest-road.spec.ts`

5) **Mount trading modals and wire GameContext actions**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`
   - Scope: Open/close modals from buttons; call `bankTrade`, `proposeTrade`, `respondTrade`; map `pendingTrades` to incoming trade UI.
   - Go unit tests: none.
   - Playwright: `frontend/tests/trading.spec.ts`, `frontend/tests/ports.spec.ts`

6) **Robber flow UI verification after PLAYING stability**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/Board.tsx`
   - Scope: Validate discard modal count, robber move hex targets, and steal modal candidates; add any missing data-cy or gating.
   - Go unit tests: none (core logic already tested).
   - Playwright: `frontend/tests/robber.spec.ts`

7) **Dev cards UX wiring after setup fix**
   - Files: `frontend/src/components/DevCardPanel.tsx`, `frontend/src/components/Game/Game.tsx`
   - Scope: Ensure dev card panel renders in PLAYING; validate play buttons and modals in E2E.
   - Go unit tests: none (core logic already tested).
   - Playwright: `frontend/tests/development-cards.spec.ts`

8) **Ports + bank ratio display verification**
   - Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Port.tsx`, `frontend/src/components/Game/BankTradeModal.tsx`
   - Scope: Confirm port rendering and ratio display matches port access.
   - Go unit tests: none (ports already tested).
   - Playwright: `frontend/tests/ports.spec.ts`

9) **Victory + game-over overlay recheck**
   - Files: `frontend/src/components/Game/GameOver.tsx`, `backend/internal/handlers/handlers.go`
   - Scope: Ensure `game_over` payload broadcasts on win and overlay selectors match E2E.
   - Go unit tests: add handler test if win broadcast regression found.
   - Playwright: `frontend/tests/victory.spec.ts`

---

## PLAN VALIDATION & WORKFLOW
- After each atomic fix: run `make test-backend`.
- After any WS/UI change: run `make e2e` (DEV_MODE=true for tests needing helpers).
- After proto changes (if any): run `make generate`, then re-run tests.

#END
