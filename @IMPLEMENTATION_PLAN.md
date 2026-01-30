# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-30 (Iteration 20 / plan audited, NO CHANGES REQUIRED. All specs, proto contracts, code, and E2E results fully align with the current atomic plan.)

--

## AUDIT STATUS: PLAN FULLY ALIGNED, NO CHANGES

- All `specs/*` requirements reviewed for interactive board, setup, robber, trading, development cards, ports, longest road, victory
- All backend Go/game/handlers, frontend context/board/game, and modal code audited
- All proto/catan/v1/* schemas checked
- All E2E test/spec failures correspond directly to granular atomic tasks already in plan
- No missing features, no untracked or orphan test failures, no additional TODOs or refactors needed at this time

--

## PRIORITIZED IMPLEMENTATION PLAN (Iteration 20 / unchanged)

### 1. E2E CRITICAL STABILIZATION (GLOBAL)
- **[HIGH] Fix Global WebSocket/Proto/Server Sync - Board Never Loads, Lobby/Game Fails to Start**
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
- All failing E2E mapped to atomic plan tasks above. No new orphan failures/sample cases found. Full alignment: code/specs/proto/tests.

--

## PLAN VALIDATION & WORKFLOW
- After every fix: run `make test-backend`, `make e2e`. After proto changes: `make generate` and verify both backend/frontend build.
- Update plan each iteration post-fix, document any new gap or regression.

--

Current state: Plan unchanged, fully aligned. Proceed per atomic tasks, cycle/commit after every change.

#END
