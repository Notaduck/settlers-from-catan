# IMPLEMENTATION_PLAN - Settlers from Catan

> Gap analysis against specs/*.md and current codebase. Each task is commit-sized and lists files + required tests.

## BLOCKERS / PREREQS (Unblocks all E2E + gameplay)
- [ ] Build the missing HTTP/WS handler layer (currently no `NewHandler`, `HandleWebSocket`, `HandleCreateGame`, `HandleGameRoutes`, and no WS dispatch for core actions).
  - Files: `backend/internal/handlers/handlers.go`, `backend/internal/handlers/handlers_test.go`, `backend/internal/hub/client.go`
  - Go tests: `backend/internal/handlers/handlers_test.go` (replace TODO stubs with real DB + hub assertions)
  - Playwright: `frontend/tests/game-flow.spec.ts` (create/join/ready/start should pass once handlers exist)
- [ ] Add WS routing for core gameplay messages (start, ready, roll, build, end turn) and persist authoritative state.
  - Files: `backend/internal/handlers/handlers.go`, `backend/internal/game/commands.go`, `backend/internal/game/state_machine.go`
  - Go tests: expand `backend/internal/game/commands_test.go` for end-turn transitions; add handler coverage in `backend/internal/handlers/handlers_test.go`
  - Playwright: `frontend/tests/game-flow.spec.ts`, `frontend/tests/interactive-board.spec.ts`, `frontend/tests/setup-phase.spec.ts`

## PRIORITY 1: INTERACTIVE BOARD (Complete)
- [x] Vertex/edge rendering, placement highlighting, and `data-cy` attributes implemented in `frontend/src/components/Board/*`.
  - Playwright: `frontend/tests/interactive-board.spec.ts` already exercises base flow once handlers exist.

## PRIORITY 2: SETUP PHASE UI (Complete)
- [x] Setup banner, turn indicator, placement instruction, and resource toast implemented in `frontend/src/components/Game/Game.tsx`.
  - Playwright: `frontend/tests/setup-phase.spec.ts` covers key behaviors once handlers exist.

## PRIORITY 3: VICTORY FLOW
- [ ] Authoritative game-over detection + broadcast with score payload (triggered after VP-gaining actions on active turn).
  - Files: `backend/internal/game/commands.go`, `backend/internal/game/rules.go`, `backend/internal/handlers/handlers.go`
  - Go tests: expand `backend/internal/game/victory_test.go` for payload scores + winner; add handler assertions in `backend/internal/handlers/handlers_test.go`
  - Playwright: implement real `frontend/tests/victory.spec.ts` (no stubbed waits)
- [ ] Frontend consumes `game_over` payload and renders breakdown from payload (not inferred fields).
  - Files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/GameOver.tsx`, `frontend/src/components/Game/Game.tsx`
  - Playwright: extend `frontend/tests/victory.spec.ts` with winner + score row assertions

## PRIORITY 4: ROBBER FLOW
- [ ] Initialize and advance `RobberPhase` on roll 7 (discard → move → steal), gating turn progression until complete.
  - Files: `backend/internal/game/dice.go`, `backend/internal/game/robber.go`, `backend/internal/game/state_machine.go`
  - Go tests: add flow-level cases in `backend/internal/game/robber_test.go` and `backend/internal/game/dice_test.go`
  - Playwright: replace stubs in `frontend/tests/robber.spec.ts` with real flow
- [ ] Wire discard/move/steal handlers and messages, including victim selection and “no adjacent players” skip.
  - Files: `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/HexTile.tsx`
  - Go tests: handler tests for discard/move/steal sequencing
  - Playwright: `frontend/tests/robber.spec.ts` verifies discard gating + steal result

## PRIORITY 5: TRADING
- [ ] Add explicit trade ↔ build phase switching + end-turn logic (trade offers expire at turn end).
  - Files: `backend/internal/game/commands.go`, `backend/internal/game/trading.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/messages.proto`
  - Go tests: add phase transition tests in `backend/internal/game/commands_test.go`; update `backend/internal/game/trading_test.go`
  - Playwright: `frontend/tests/trading.spec.ts` for phase gating + end-turn expiry
- [ ] Implement trading UI + websocket wiring (bank trade + propose/respond + incoming modal).
  - Files: `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`
  - Playwright: `frontend/tests/trading.spec.ts`

## PRIORITY 6: DEVELOPMENT CARDS
- [ ] Implement deck, purchase, and per-card effects; track “bought this turn” vs playable cards (requires proto fields for hand).
  - Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/game/robber.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto`, `proto/catan/v1/messages.proto`
  - Go tests: expand `backend/internal/game/devcards_test.go` for deck sizing, draw, and each effect + turn restriction
  - Playwright: add `frontend/tests/development-cards.spec.ts`
- [ ] Frontend dev-card UI: buy button, hand display, and per-card play modals.
  - Files: `frontend/src/components/Game/*`, `frontend/src/context/GameContext.tsx`
  - Playwright: `frontend/tests/development-cards.spec.ts`

## PRIORITY 7: LONGEST ROAD
- [ ] Recalculate and persist longest-road holder on road/building placement; apply tie rules.
  - Files: `backend/internal/game/longestroad.go`, `backend/internal/game/commands.go`, `backend/internal/game/state_machine.go`
  - Go tests: expand `backend/internal/game/longestroad_test.go`; add integration cases in `backend/internal/game/commands_test.go`
  - Playwright: add `frontend/tests/longest-road.spec.ts`
- [ ] Display holder + per-player road length in UI.
  - Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`
  - Playwright: `frontend/tests/longest-road.spec.ts`

## PRIORITY 8: PORTS (Maritime Trading)
- [ ] Apply port ratios to bank trade (use `GetBestTradeRatio`) and expose best ratio to UI.
  - Files: `backend/internal/game/trading.go`, `backend/internal/game/ports.go`, `backend/internal/handlers/handlers.go`
  - Go tests: extend `backend/internal/game/ports_test.go` for bank trade ratios
  - Playwright: add `frontend/tests/ports.spec.ts`
- [ ] Render ports on board and show ratio options in bank trade modal.
  - Files: `frontend/src/components/Board/*`, `frontend/src/components/Game/BankTradeModal.tsx`
  - Playwright: `frontend/tests/ports.spec.ts`

## NOTES / UNCERTAINTIES (Documented per instructions)
- Port `Location` values are hard-coded strings; verify they match generated vertex IDs before wiring UI highlights. If mismatched, adjust `standardPortEdges` to align with `generateVertices` output.
