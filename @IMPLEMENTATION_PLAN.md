# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-28 (Iteration 16, plan audit - post comprehensive E2E/code/spec/proto review).

Sources re-reviewed for this iteration:
- All `specs/*` (interactive board, setup, robber, trading, ports, development cards, longest road, victory)
- Backend: `backend/internal/game/*`, `backend/internal/handlers/handlers.go`
- Frontend: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/Game.tsx`
- Playwright E2E: `frontend/tests/*.spec.ts`
- Proto: `proto/catan/v1/*.proto`
- E2E artifacts: `E2E_STATUS.md`, `.last-run.json` (recent global failures)

---

## ITERATION 16 STATUS - FULL GAP SYNTHESIS / PLAN UPDATE

- All Playwright E2E test suites remain non-functional at time of this plan (as per audit, 0/83 passing, matching all atomic requirements/specs).
- This plan has been reconfirmed against all acceptance criteria in `specs/*` and each failing test suite. No missing spec or coverage gaps were found beyond what is already listed below.
- For each required flow, E2E atomic tasks and Go/TS tests have been mapped. Every spec line and E2E failure is traceable to at least one concrete atomic fix task.
- Backend, frontend (UI and context), proto, and test coverage all present; required file and test mappings are up-to-date per actual repo state.
- Below is the actionable prioritized task list, grouped first by global protocol/unblockers, then by feature flows, then by E2E coverage.

---

## PRIORITIZED IMPLEMENTATION PLAN (Iteration 16)

### 1. E2E CRITICAL STABILIZATION (GLOBAL)

- **[HIGH] Fix Global Websocket/Proto/Server Sync - Board Never Loads, Lobby/Game Fails to Start**
    - Files: `proto/catan/v1/messages.proto`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`
    - Action: Audit and resolve any protocol break, server start error, or websocket handler mismatch preventing initial load/lobby/game creation. Re-run E2E to confirm board UI renders, initial join/start flows proceed. Add/verify test(s) for basic game start in `game-flow.spec.ts`.

- **[HIGH] Fix Interactive Board & Placement Handler - Placement Fails**
    - Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/context/GameContext.tsx`, `backend/internal/game/board.go`, `frontend/tests/interactive-board.spec.ts`
    - Action: Implement SVG vertex/edge rendering, data-cy selectors, and event handlers per `interactive-board.md`. Confirm E2E clicks produce board state change and UI update.
    - Tests: See `interactive-board.spec.ts` acceptance criteria.

- **[HIGH] Fix Setup Phase UI Flow**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/tests/setup-phase.spec.ts`
    - Action: Show banners, turn UI, placement instructions as per `setup-phase-ui.md`. Backend must emit/setup phase and advance turns. E2E must pass all setup flow/placement tests.

- **[HIGH] Restore Phase/Action State Sync Across WebSocket**
    - Files: `frontend/src/context/GameContext.tsx`, `backend/internal/game/state_machine.go`, `proto/catan/v1/messages.proto`
    - Action: All phase/player transitions must transmit accurately, reflected in frontend state. Unit/E2E test turn and setup advances.

### 2. BLOCKING FEATURE FLOWS (Specs + E2E)

- **Victory Flow - Game Finish/Overlay**
    - Files: `frontend/src/components/Game/Game.tsx`, `backend/internal/game/victory.go`, `backend/internal/game/victory_test.go`, `frontend/tests/victory.spec.ts`
    - Action: Victory detection and overlay; E2E for winning/ending/new game state.

- **Robber Flow - Discard/Move/Steal**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/Board.tsx`, `backend/internal/game/robber.go`, `backend/internal/game/robber_test.go`, `proto/catan/v1/messages.proto`, `frontend/tests/robber.spec.ts`
    - Action: UI/modal logic, packet flow, test all discard/move/steal scenarios per `robber-flow.md` and E2E.

- **Trading - Bank/Player**
    - Files: `frontend/src/components/Game/TradeModal.tsx` (or equivalent), `frontend/src/context/GameContext.tsx`, `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`, `frontend/tests/trading.spec.ts`
    - Action: Both trade UI types, modal, packet, backend transfer/validation. Full Go/TS/E2E for every accept/decline/counter case.

- **Development Cards - Buy/Play**
    - Files: `frontend/src/components/Game/DevCards.tsx`, `frontend/src/context/GameContext.tsx`, `backend/internal/game/devcards.go`, `backend/internal/game/devcards_test.go`, `frontend/tests/development-cards.spec.ts`
    - Action: All dev card effects, modal handling, Largest Army, etc. E2E: verify modal and board/game effect for each.

- **Longest Road - Calculation/Bonus**
    - Files: `backend/internal/game/longestroad.go`, `backend/internal/game/longestroad_test.go`, `frontend/tests/longest-road.spec.ts`
    - Action: DFS for award, UI indicator. E2E: all changing/bonus transfer scenarios.

- **Ports - Maritime Trade/Display**
    - Files: `proto/catan/v1/types.proto`, `backend/internal/game/ports.go`, `backend/internal/game/ports_test.go`, `frontend/src/components/Board/Board.tsx`, `frontend/tests/ports.spec.ts`
    - Action: Board port rendering, trade ratios by player access, E2E for all ratios on board and modal.

### 3. E2E AUDIT/TRACEABILITY

- Each atomic fix above corresponds to a unique failing E2E scenario or spec line; no additional orphaned failures or undocumented gaps discovered.
- Mapping: `game-flow.spec.ts`, `interactive-board.spec.ts`, `setup-phase.spec.ts`, `victory.spec.ts`, `robber.spec.ts`, `trading.spec.ts`, `development-cards.spec.ts`, `longest-road.spec.ts`, `ports.spec.ts`.

---

## VALIDATION/PROTOCOL
- After each fix, re-run `make test-backend` and `make e2e`.
- If proto updated, `make generate` + verify both backend and frontend codegen.
- Update plan after each cycle to document new breakage/fixes.

---

## PLAN STATUS

This plan is current, complete, and maps every required feature/E2E/test to concrete work. No new gaps as of this audit.

#END
