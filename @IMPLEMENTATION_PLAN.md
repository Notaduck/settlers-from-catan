# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-31 (Iteration 21 / full repo audit; plan updated for missing WS + UI integration)

---

## AUDIT NOTES (Iteration 21)
- E2E audit command `cd frontend && npx playwright test --reporter=list | head -100` timed out after 10s; no fresh results captured.
- Latest E2E status remains 2026-01-27 (Iteration 12): all suites failing, most blocked at lobby/board load.
- Primary root cause: backend WebSocket endpoint + handlers are stubbed (`HandleWebSocket`, `handleClientMessage`, test endpoints). Frontend also expects protobuf JSON via `oneofKind`, but server uses no WS/proto at all.
- Secondary root cause: Game UI missing lobby controls, placement indicators, and integration of existing UI panels (PlayerPanel, trade modals, robber modals, game over overlay).
- Backend game logic for robber/trading/dev cards/ports/longest road/victory exists with unit tests; missing is wiring through handlers + WS broadcasts.

---

## E2E STABILIZATION (Grouped by Spec + Root Cause)

### game-flow.spec.ts
- **Fix E2E: lobby/ready/start flow** — WS not implemented + Game lobby UI missing ready/start controls.
  - Files: `backend/internal/handlers/handlers.go`, `backend/internal/hub/client.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/hooks/useWebSocket.ts`, `frontend/src/components/Game/Game.tsx`
  - Go tests: add `backend/internal/handlers/handlers_test.go` for WS auth + start/ready state persistence.
  - Playwright: `frontend/tests/game-flow.spec.ts`

### interactive-board.spec.ts
- **Fix E2E: placement indicators + board readiness** — placement mode/ready counters not rendered; WS state never arrives.
  - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/placement.ts`
  - Go tests: none (board logic already covered).
  - Playwright: `frontend/tests/interactive-board.spec.ts`

### setup-phase.spec.ts
- **Fix E2E: setup banner/instructions + resource toast** — UI exists but not wired into main Game flow; WS state missing.
  - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/SetupPhasePanel.tsx`, `frontend/src/context/GameContext.tsx`
  - Go tests: none (setup logic already covered).
  - Playwright: `frontend/tests/setup-phase.spec.ts`

### victory.spec.ts
- **Fix E2E: game over broadcast + overlay** — server never emits game over, UI overlay not mounted.
  - Files: `backend/internal/handlers/handlers.go`, `backend/internal/game/rules.go`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/GameOver.tsx`
  - Go tests: add handler-level test for GameOver message emission.
  - Playwright: `frontend/tests/victory.spec.ts`

### robber.spec.ts
- **Fix E2E: discard/move/steal flow** — server handlers + UI modals not wired; test endpoints missing.
  - Files: `backend/internal/handlers/handlers.go`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/Board.tsx`
  - Go tests: add handler test for robber move/steal + discard handling.
  - Playwright: `frontend/tests/robber.spec.ts`

### trading.spec.ts
- **Fix E2E: trade phase + modals** — UI exists but not mounted; server trade handlers not wired; test endpoints missing.
  - Files: `backend/internal/handlers/handlers.go`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`
  - Go tests: add handler test for propose/respond/bank trade message flow.
  - Playwright: `frontend/tests/trading.spec.ts`

### development-cards.spec.ts
- **Fix E2E: dev card buy/play wiring** — WS handlers missing; DEV_MODE endpoints missing; UI panel present but depends on state updates.
  - Files: `backend/internal/handlers/handlers.go`, `frontend/src/components/DevCardPanel.tsx`, `frontend/src/components/Game/Game.tsx`
  - Go tests: add handler test for buy/play dev card messages.
  - Playwright: `frontend/tests/development-cards.spec.ts`

### longest-road.spec.ts
- **Fix E2E: longest road badge + lengths** — PlayerPanel not mounted; WS updates missing.
  - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/PlayerPanel/PlayerPanel.tsx`
  - Go tests: none (longest road already tested).
  - Playwright: `frontend/tests/longest-road.spec.ts`

### ports.spec.ts
- **Fix E2E: port render + trade ratios** — Board ports render only when game board shows; bank trade modal not mounted.
  - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Board/Board.tsx`
  - Go tests: none (ports already tested).
  - Playwright: `frontend/tests/ports.spec.ts`

---

## PRIORITIZED IMPLEMENTATION PLAN (Atomic Tasks)

1) **Implement WebSocket endpoint + protobuf JSON contract**
   - Files: `backend/internal/handlers/handlers.go`, `backend/internal/hub/client.go`, `frontend/src/hooks/useWebSocket.ts`, `frontend/src/context/GameContext.tsx`
   - Scope: Implement `HandleWebSocket` (token auth, load player/game from DB, register client, send initial `ServerMessage` game_state), implement `handleClientMessage` to protojson-decode `ClientMessage`, dispatch to game logic, persist state, broadcast `ServerMessage` updates. Update frontend to use protobuf-ts `toJsonString`/`fromJsonString` instead of raw JSON parse and remove dependency on `oneofKind` JSON shape.
   - Go unit tests: add `backend/internal/handlers/handlers_test.go` for WS auth + message routing + state persistence.
   - Playwright: unblock all suites (primary blocker).
   - Status: **DONE (Iteration 1)** — WS endpoint + message routing/broadcasts implemented; kept `oneofKind` JSON envelope (frontend still expects it).

2) **Add DEV_MODE test endpoints used by Playwright**
   - Files: `backend/internal/handlers/handlers.go`
   - Scope: Implement `/test/grant-resources`, `/test/grant-dev-card`, `/test/force-dice-roll`, `/test/set-game-state` with DB state updates + `game_state` broadcast.
   - Go unit tests: add handler tests for each test endpoint payload and persistence.
   - Playwright: `frontend/tests/robber.spec.ts`, `frontend/tests/trading.spec.ts`, `frontend/tests/development-cards.spec.ts`, `frontend/tests/ports.spec.ts`

3) **Lobby UI + auto-connect + ready/start controls**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/Game.css`, `frontend/src/context/GameContext.tsx`
   - Scope: Show `data-cy="game-loading"` while awaiting state, render lobby list with `data-cy="game-code"`, `ready-btn`, `cancel-ready-btn`, `start-game-btn`, and player readiness. Call WS connect on mount.
   - Go unit tests: none.
   - Playwright: `frontend/tests/game-flow.spec.ts`
   - Status: **DONE (Iteration 1)** — lobby UI rendered, auto-connect added, game-flow E2E passes.

4) **Placement mode indicators and board readiness**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/Game.css`
   - Scope: Add `data-cy="placement-mode"` banner and `data-cy="placement-ready"` counts (valid vertices/edges) to support setup and build mode; wire to placementState.
   - Go unit tests: none.
   - Playwright: `frontend/tests/interactive-board.spec.ts`, `frontend/tests/setup-phase.spec.ts`

5) **Integrate PlayerPanel and turn phase controls**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/PlayerPanel/PlayerPanel.tsx`
   - Scope: Render PlayerPanel during PLAYING, ensure `trade-phase-btn` / `build-phase-btn` toggles exist and call `setTurnPhase`; expose dice + end turn buttons; ensure resource panels render for tests.
   - Go unit tests: none.
   - Playwright: `frontend/tests/game-flow.spec.ts`, `frontend/tests/trading.spec.ts`, `frontend/tests/robber.spec.ts`, `frontend/tests/longest-road.spec.ts`

6) **Robber flow UI wiring**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/Board.tsx`, `backend/internal/handlers/handlers.go`
   - Scope: Show discard modal when required, enable robber move mode on board, show steal modal with candidates, send discard/move/steal messages; broadcast updated state after each step.
   - Go unit tests: add handler test for discard/move/steal flow; core game tests already exist.
   - Playwright: `frontend/tests/robber.spec.ts`

7) **Trading flow wiring (bank + player)**
   - Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`, `backend/internal/handlers/handlers.go`
   - Scope: Mount trade modals, map pendingTrades to incoming offers, wire propose/respond/bank trade messages, broadcast updated state on accept/decline.
   - Go unit tests: add handler tests for propose/respond/bank trade message handling.
   - Playwright: `frontend/tests/trading.spec.ts`, `frontend/tests/ports.spec.ts`

8) **Development cards flow wiring**
   - Files: `backend/internal/handlers/handlers.go`, `frontend/src/components/DevCardPanel.tsx`, `frontend/src/components/Game/Game.tsx`
   - Scope: Wire buy/play dev card messages, send dev_card_bought (private) and game_state updates; ensure play buttons follow timing rules.
   - Go unit tests: add handler tests for buy/play dev card messages.
   - Playwright: `frontend/tests/development-cards.spec.ts`

9) **Victory flow broadcast + overlay**
   - Files: `backend/internal/handlers/handlers.go`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Game/GameOver.tsx`
   - Scope: When state transitions to FINISHED, emit `game_over` payload and show overlay with data-cy selectors; disable further actions.
   - Go unit tests: add handler test verifying `game_over` payload is broadcast on win.
   - Playwright: `frontend/tests/victory.spec.ts`

10) **Longest road display + ports validation**
   - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Board/Board.tsx`
   - Scope: Ensure longest-road holder badge and per-player road length are visible; verify ports are rendered and bank trade ratio reflects ports.
   - Go unit tests: none (already covered).
   - Playwright: `frontend/tests/longest-road.spec.ts`, `frontend/tests/ports.spec.ts`

---

## PLAN VALIDATION & WORKFLOW
- After each atomic fix: run `make test-backend`.
- After any WS/UI change: run `make e2e` (DEV_MODE=true for tests needing helpers).
- After proto changes (if any): run `make generate`, then re-run tests.

#END
