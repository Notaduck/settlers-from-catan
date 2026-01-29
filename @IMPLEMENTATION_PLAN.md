# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-29 (Iteration 17, full plan audit - E2E/spec/code/proto/state matches).

Sources fully re-reviewed for this iteration:
- All `specs/*` (interactive board, setup, robber, trading, ports, development cards, longest road, victory)
- Backend: `backend/internal/game/*`, `backend/internal/handlers/handlers.go`
- Frontend: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/Game.tsx`
- Playwright E2E: `frontend/tests/*.spec.ts`
- Proto: `proto/catan/v1/*.proto`
- E2E artifacts: `E2E_STATUS.md`, `.last-run.json`

---

## ITERATION 17 STATUS - FULL GAP AUDIT / NO CHANGES REQUIRED

- E2E audit and spec alignment reconfirmed: all tests for every major spec (interactive-board, setup, victory, robber, trading, dev cards, longest road, ports) correspond to mapped atomic tasks in this plan.
- All backend/frontend/proto source and test mappings match plan and specs. No new files or blockages surfaced since prior iteration.
- Absolute priority remains: 1) global proto/WS sync, 2) interactive board fix, 3) setup phase UI, 4) state/turn sync, then features by game-blocking flows per priority order. Each mapped to a root cause E2E fail and concrete code/test/files.
- No additional TODOs, skips, or coverage discrepancies found.

---

## PRIORITIZED IMPLEMENTATION PLAN (Iteration 17)

### 1. E2E CRITICAL STABILIZATION (GLOBAL)

- **[HIGH] Fix Global Websocket/Proto/Server Sync - Board Never Loads, Lobby/Game Fails to Start**
    - Files: `proto/catan/v1/messages.proto`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`

- **[HIGH] Fix Interactive Board & Placement Handler - Placement Fails**
    - Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/context/GameContext.tsx`, `backend/internal/game/board.go`, `frontend/tests/interactive-board.spec.ts`

- **[HIGH] Fix Setup Phase UI Flow**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/tests/setup-phase.spec.ts`

- **[HIGH] Restore Phase/Action State Sync Across WebSocket**
    - Files: `frontend/src/context/GameContext.tsx`, `backend/internal/game/state_machine.go`, `proto/catan/v1/messages.proto`

### 2. BLOCKING FEATURE FLOWS (Specs + E2E)

- **Victory Flow - Game Finish/Overlay**
    - Files: `frontend/src/components/Game/Game.tsx`, `backend/internal/game/victory.go`, `backend/internal/game/victory_test.go`, `frontend/tests/victory.spec.ts`

- **Robber Flow - Discard/Move/Steal**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/Board.tsx`, `backend/internal/game/robber.go`, `backend/internal/game/robber_test.go`, `proto/catan/v1/messages.proto`, `frontend/tests/robber.spec.ts`

- **Trading - Bank/Player**
    - Files: `frontend/src/components/Game/TradeModal.tsx` (or equivalent), `frontend/src/context/GameContext.tsx`, `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`, `frontend/tests/trading.spec.ts`

- **Development Cards - Buy/Play**
    - Files: `frontend/src/components/Game/DevCards.tsx`, `frontend/src/context/GameContext.tsx`, `backend/internal/game/devcards.go`, `backend/internal/game/devcards_test.go`, `frontend/tests/development-cards.spec.ts`

- **Longest Road - Calculation/Bonus**
    - Files: `backend/internal/game/longestroad.go`, `backend/internal/game/longestroad_test.go`, `frontend/tests/longest-road.spec.ts`

- **Ports - Maritime Trade/Display**
    - Files: `proto/catan/v1/types.proto`, `backend/internal/game/ports.go`, `backend/internal/game/ports_test.go`, `frontend/src/components/Board/Board.tsx`, `frontend/tests/ports.spec.ts`

### 3. E2E AUDIT/TRACEABILITY

- Each atomic fix above corresponds to a unique failing E2E scenario or spec line; no additional orphaned failures or undocumented gaps discovered.

---

## VALIDATION/PROTOCOL

- After each fix, re-run `make test-backend` and `make e2e`.
- If proto updated, `make generate` + verify both backend and frontend codegen.
- Update plan after each cycle to document new breakage/fixes.

---

## PLAN STATUS

- Plan is up-to-date with current repo, specs, proto, and E2E artifacts as of 2026-01-29.
- No new changes. Proceed with atomic fixes per order above.

#END
