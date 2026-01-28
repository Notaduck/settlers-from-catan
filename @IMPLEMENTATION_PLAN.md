# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-28 (Iteration 15, full E2E failure audit).

Sources reviewed for this iteration:
- All `specs/*` (interactive board, setup, robber, trading, ports, development cards, longest road, victory)
- Backend: `backend/internal/game/*`, `backend/internal/handlers/handlers.go`
- Frontend: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/Game.tsx`
- Playwright specs: `frontend/tests/*.spec.ts`
- Proto: `proto/catan/v1/*.proto`
- E2E artifacts: `E2E_STATUS.md`, `.last-run.json` (50+ failures), previous plan

---

## ITERATION 15 STATUS - MAJOR FUNCTIONAL REGRESSION

- All Playwright E2E test suites are currently failing 100%. No user flows (board, setup, actions, game flow, or endgame) are operational under E2E.
- The prior plan incorrectly claims full coverage/stability—this is now demonstrably obsolete. Audit data from multiple sources confirms global breakage at every UI/backend interaction point.
- **Root causes likely include websocket/proto contract mismatch, server handler faults, and critical UI rendering blockers.**
- The following actionable, atomic tasks are required—ordered by E2E unblock priority and feature dependencies.
- Each item specifies impacted files and required tests. All entries are concrete, commit-sized, and map directly to E2E or spec failures.

---

## PRIORITIZED IMPLEMENTATION PLAN

### 1. E2E CRITICAL STABILIZATION (GLOBAL)

- **[HIGH] Fix Global Websocket/Proto/Server Sync - Board Never Loads, Lobby/Game Fails to Start**
    - Files: `proto/catan/v1/messages.proto`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`
    - Action: Audit and resolve any protocol break, server start error, or websocket handler mismatch preventing initial load/lobby/game creation. Re-run E2E to confirm board UI renders, initial join/start flows proceed. Add/verify test(s) for basic game start in `game-flow.spec.ts`.

- **[HIGH] Fix Interactive Board & Placement Handler - Placement Fails**
    - Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/context/GameContext.tsx`, `backend/internal/game/board.go`, `frontend/tests/interactive-board.spec.ts`
    - Action: Implement SVG vertex/edge rendering, data-cy selectors, and event handlers per spec. Confirm E2E clicks on vertex/edge produce placement and UI updates. Unit test any board-side placement logic if not covered.
    - Test: See all acceptance criteria in `interactive-board.spec.ts`.

- **[HIGH] Fix Setup Phase UI Flow**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/tests/setup-phase.spec.ts`
    - Action: Ensure setup banners, instructions, turn guidance per spec. E2E must pass all initial placement/turn progression tests.
    - Backend: Confirm `state_machine.go` correctly emits/setup phase and advances turns.

- **[HIGH] Restore Phase and Action State Sync Across WS**
    - Files: `frontend/src/context/GameContext.tsx`, `backend/internal/game/state_machine.go`, `proto/catan/v1/messages.proto`
    - Action: All phase and player state transitions must flow correctly between backend and frontend. Unit test backend phase transitions and validate state with E2E.

### 2. BLOCKING FEATURE FLOWS (Per Spec/E2E)

- **Victory Flow - Game Finish and Overlay**
    - Files: `frontend/src/components/Game/Game.tsx`, `backend/internal/game/victory.go`, `backend/internal/game/victory_test.go`, `frontend/tests/victory.spec.ts`
    - Action: Ensure victory detection/overlay is called after eligible turn, final VP revealed. E2E for winner, overlay, new game button, etc.

- **Robber Flow - Discard, Move, Steal (upon rolling 7)**
    - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/Board.tsx`, `backend/internal/game/robber.go`, `backend/internal/game/robber_test.go`, `proto/catan/v1/messages.proto`, `frontend/tests/robber.spec.ts`
    - Action: Add discard modal UI logic, ensure move/steal packets are delivered and resolved in correct order.

- **Trading - Bank and Player Trades**
    - Files: `frontend/src/components/Game/TradeModal.tsx` (or equivalent), `frontend/src/context/GameContext.tsx`, `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`, `frontend/tests/trading.spec.ts`
    - Action: Both player and bank trades per acceptance; unit test for each trade/accept/refusal.

- **Development Cards - Buy/Play (effects, Knight, Monopoly, etc.)**
    - Files: `frontend/src/components/Game/DevCards.tsx` (or equivalent), `frontend/src/context/GameContext.tsx`, `backend/internal/game/devcards.go`, `backend/internal/game/devcards_test.go`, `frontend/tests/development-cards.spec.ts`
    - Action: Buying, playing, all modal/effect logic; unit/E2E test all card effects.

- **Longest Road - DFS and Bonus Logic**
    - Files: `backend/internal/game/longestroad.go`, `backend/internal/game/longestroad_test.go`, `frontend/tests/longest-road.spec.ts`
    - Action: Graph traversal on placement, bonus reassignment, and UI badge.

- **Ports - Maritime Trade Ratios and Board UI**
    - Files: `proto/catan/v1/types.proto`, `backend/internal/game/ports.go`, `backend/internal/game/ports_test.go`, `frontend/src/components/Board/Board.tsx`, `frontend/tests/ports.spec.ts`
    - Action: Implement/coherently sync coastal ratio option, UI port icons. E2E: test trade at 4:1, 3:1, 2:1 ratios.

### 3. PLAYWRIGHT E2E AUDIT: Failing Spec Mapping

For any of the above feature flows, break down by E2E spec (see E2E_STATUS.md for suite-to-feature mapping) and confirm that each atomic acceptance scenario is tracked and finished. Mark re-checked E2E/Go tests per fix.

#### Mapping provided at bottom:
- game-flow.spec.ts: unblock lobby, join, ready, start
- interactive-board.spec.ts: board rendering, placement click, SVG data-cy, etc.
- setup-phase.spec.ts: setup banners, indicator, valid placement
- victory.spec.ts: endgame/overlay, new game button
- robber.spec.ts: discard modal, move, steal logic
- trading.spec.ts: bank, P2P trades
- ports.spec.ts: port icons, bank trade at special rates
- development-cards.spec.ts: buy/play each card, modal, effect
- longest-road.spec.ts: DFS, UI badge, VP sync

---

## ADDITIONAL AUDIT/FIX PROTOCOL

- After every atomic fix, update/confirm both Go unit and Playwright E2E tests as per each spec's required section.
- Validation: `make test-backend`, `make e2e` required after every cycle.
- If new proto fields/messages are added, update `make generate` artifacts and confirm frontend parcel generation.

---

## NEXT ACTIONS

1. Work through atomic tasks above, highest priority first.
2. After each fix, immediately rerun E2E and unit tests to confirm resolution and identify any additional integration issues.
3. Update this plan after each iteration to reflect new failures or fixes discovered by E2E/audit.

---

#END
