# IMPLEMENTATION_PLAN - Settlers from Catan

> Gap analysis against specs/*.md and current codebase. Each task is commit-sized with files + required tests.

## BLOCKERS / PREREQS (Unblocks all E2E)
- [ ] Implement missing HTTP + WebSocket handler layer (currently no `NewHandler`, `HandleWebSocket`, `HandleCreateGame`, `HandleGameRoutes`, or WS message dispatch for roll/build/start/end/ready).
  - Files: `backend/internal/handlers/handlers.go`, `backend/internal/handlers/handlers_test.go`, `backend/internal/hub/*.go`, `backend/internal/db/*.go`
  - Go tests: `backend/internal/handlers/handlers_test.go` (replace TODOs with real state/db/hub assertions)
  - Playwright: existing specs (interactive/setup/etc.) become runnable once handlers exist

## PRIORITY 1: INTERACTIVE BOARD (Complete)
- [x] Vertex/edge rendering, placement highlighting, and `data-cy` attributes present in `frontend/src/components/Board/*`.
  - Playwright: `frontend/tests/interactive-board.spec.ts` already exercises base flow.

## PRIORITY 2: SETUP PHASE UI (Complete)
- [x] Setup banner, turn indicator, placement instruction, and resource toast implemented in `frontend/src/components/Game/Game.tsx`.
  - Playwright: `frontend/tests/setup-phase.spec.ts` covers key behaviors.

## PRIORITY 3: VICTORY FLOW
- [ ] Add authoritative game-over broadcast and final score payload once `CheckVictory` triggers.
  - Files: `backend/internal/game/commands.go`, `backend/internal/game/rules.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/messages.proto`
  - Go tests: expand `backend/internal/game/victory_test.go` for payload scores + winner; add handler coverage in `backend/internal/handlers/handlers_test.go`
  - Playwright: implement real `frontend/tests/victory.spec.ts` flow (no stubs)
- [ ] Frontend consumes `game_over` payload and renders final breakdown from payload (not inferred fields).
  - Files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/GameOver.tsx`, `frontend/src/components/Game/Game.tsx`
  - Playwright: extend `frontend/tests/victory.spec.ts` with winner + scores assertions

## PRIORITY 4: ROBBER FLOW
- [ ] Initialize and advance `RobberPhase` on roll 7 (discard → move → steal), gating turn progression until complete.
  - Files: `backend/internal/game/dice.go`, `backend/internal/game/robber.go`, `backend/internal/game/state_machine.go`
  - Go tests: add flow-level cases in `backend/internal/game/robber_test.go` and `backend/internal/game/dice_test.go`
  - Playwright: convert `frontend/tests/robber.spec.ts` from stubs to real flow
- [ ] Fix client/server wiring for robber actions (discard/move/steal), including victim selection.
  - Files: `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/HexTile.tsx`
  - Go tests: handler tests for discard/move/steal sequencing
  - Playwright: `frontend/tests/robber.spec.ts` verifies discard gating + steal result

## PRIORITY 5: TRADING
- [ ] Implement full trading UI + websocket wiring (bank trade + player trades + incoming modal).
  - Files: `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`
  - Playwright: implement `frontend/tests/trading.spec.ts`
- [ ] Add turn-phase switching (trade ↔ build) and trade expiration at end of turn.
  - Files: `backend/internal/game/commands.go` (new EndTurn/phase switch), `backend/internal/game/trading.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/messages.proto`
  - Go tests: add `EndTurn`/phase transition tests in `backend/internal/game/commands_test.go`; update `backend/internal/game/trading_test.go`
  - Playwright: update `frontend/tests/trading.spec.ts` for phase gating

## PRIORITY 6: DEVELOPMENT CARDS
- [ ] Implement dev card deck, purchase, and all effects (knight/road building/year of plenty/monopoly/VP).
  - Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/game/robber.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/messages.proto`, `proto/catan/v1/types.proto`
  - Go tests: expand `backend/internal/game/devcards_test.go` for all card effects + turn restrictions
  - Playwright: add `frontend/tests/development-cards.spec.ts`
- [ ] Frontend dev card UI: buy button, hand display, play modals, and turn restrictions.
  - Files: `frontend/src/components/Game/*`, `frontend/src/context/GameContext.tsx`
  - Playwright: `frontend/tests/development-cards.spec.ts`

## PRIORITY 7: LONGEST ROAD
- [ ] Recalculate longest road on road/building placement and persist current holder in state.
  - Files: `backend/internal/game/longestroad.go`, `backend/internal/game/commands.go`, `backend/internal/game/state_machine.go`
  - Go tests: expand `backend/internal/game/longestroad_test.go` + integration in `backend/internal/game/commands_test.go`
  - Playwright: add `frontend/tests/longest-road.spec.ts`
- [ ] Display holder + per-player road length in UI.
  - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`
  - Playwright: `frontend/tests/longest-road.spec.ts`

## PRIORITY 8: PORTS (Maritime Trading)
- [ ] Apply port ratios to bank trade and expose best ratios to UI.
  - Files: `backend/internal/game/trading.go`, `backend/internal/game/ports.go`, `backend/internal/handlers/handlers.go`
  - Go tests: extend `backend/internal/game/ports_test.go` for bank trade ratios
  - Playwright: add `frontend/tests/ports.spec.ts`
- [ ] Render ports on board and show available trade ratios in bank trade modal.
  - Files: `frontend/src/components/Board/*`, `frontend/src/components/Game/BankTradeModal.tsx`
  - Playwright: `frontend/tests/ports.spec.ts`
