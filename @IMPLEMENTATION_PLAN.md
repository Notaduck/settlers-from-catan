# IMPLEMENTATION_PLAN - Settlers from Catan

> Gap analysis against `specs/*.md` + current codebase. Each task is commit-sized and lists files + required tests.

## BLOCKERS / FOUNDATIONAL (Unblocks all E2E + gameplay)
- [x] Implement HTTP/WS handler layer (missing `NewHandler`, `HandleWebSocket`, `HandleCreateGame`, `HandleGameRoutes`) with DB persistence and hub broadcast wiring. Files: `backend/internal/handlers/handlers.go`, `backend/internal/handlers/handlers_test.go`, `backend/internal/hub/client.go`, `backend/cmd/server/main.go`; Go tests: `backend/internal/handlers/handlers_test.go` (real DB + WS hub assertions); Playwright: `frontend/tests/game-flow.spec.ts`.
- [x] Add WS routing for core messages (ready/start/roll/build/end-turn) and broadcast authoritative `game_state` for every state change. Files: `backend/internal/handlers/handlers.go`, `backend/internal/game/state_machine.go`; Go tests: `backend/internal/game/state_machine_test.go`.

## PRIORITY 1: INTERACTIVE BOARD
- [x] Vertex/edge rendering, placement highlighting, and `data-cy` attributes implemented on the board UI. Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Edge.tsx`, `frontend/src/components/Board/Vertex.tsx`; Go tests: none (backend vertex/edge logic already covered); Playwright: `frontend/tests/interactive-board.spec.ts` (blocked by WS handlers above).

## PRIORITY 2: SETUP PHASE UI
- [x] Setup banner, turn indicator, placement instruction, and resource toast implemented. Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`; Go tests: none (state machine tests already cover setup); Playwright: `frontend/tests/setup-phase.spec.ts` (blocked by WS handlers above).

## PRIORITY 3: VICTORY FLOW
- [ ] Add authoritative victory detection + `game_over` broadcast with score breakdown, and prevent actions after FINISHED. Files: `backend/internal/game/commands.go`, `backend/internal/game/rules.go`, `backend/internal/game/longestroad.go`, `backend/internal/handlers/handlers.go`; Go tests: extend `backend/internal/game/victory_test.go`; Playwright: `frontend/tests/victory.spec.ts`.
- [ ] Frontend consumes `game_over` payload and renders breakdown from payload (not inferred fields) and disables actions. Files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/GameOver.tsx`, `frontend/src/components/Game/Game.tsx`; Go tests: none; Playwright: `frontend/tests/victory.spec.ts`.

## PRIORITY 4: ROBBER FLOW
- [ ] Initialize and progress RobberPhase on roll 7 (discard -> move -> steal), gating turn progression until completion. Files: `backend/internal/game/dice.go`, `backend/internal/game/robber.go`, `backend/internal/game/state_machine.go`; Go tests: `backend/internal/game/robber_test.go`, `backend/internal/game/dice_test.go`; Playwright: `frontend/tests/robber.spec.ts`.
- [ ] Wire discard/move/steal handlers and UI, including robber-hex click support and steal selection flow without invalid MoveRobber calls. Files: `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/HexTile.tsx`; Go tests: handler coverage in `backend/internal/handlers/handlers_test.go`; Playwright: `frontend/tests/robber.spec.ts`.

## PRIORITY 5: TRADING
- [ ] Implement explicit trade/build phase switching, end-turn logic, and trade expiry at end of turn. Files: `backend/internal/game/commands.go`, `backend/internal/game/trading.go`, `backend/internal/handlers/handlers.go`; Go tests: `backend/internal/game/commands_test.go`, `backend/internal/game/trading_test.go`; Playwright: `frontend/tests/trading.spec.ts`.
- [ ] Build trading UI (bank + player proposals + incoming modal) and websocket wiring; show controls only in TRADE phase. Files: `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`; Go tests: none; Playwright: `frontend/tests/trading.spec.ts`.

## PRIORITY 6: DEVELOPMENT CARDS
- [ ] Implement deck management, buying, per-card effects, and per-turn play rules (incl. Largest Army). Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/game/robber.go`, `backend/internal/game/rules.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto`, `proto/catan/v1/messages.proto`; Go tests: extend `backend/internal/game/devcards_test.go`; Playwright: add `frontend/tests/development-cards.spec.ts`.
- [ ] Frontend dev-card UI (buy button, hand display, play modals). Files: `frontend/src/components/Game/*`, `frontend/src/context/GameContext.tsx`; Go tests: none; Playwright: `frontend/tests/development-cards.spec.ts`.

## PRIORITY 7: LONGEST ROAD
- [ ] Recalculate and persist longest-road holder on road/building placement; apply tie rules and update victory checks. Files: `backend/internal/game/longestroad.go`, `backend/internal/game/commands.go`, `backend/internal/game/rules.go`; Go tests: `backend/internal/game/longestroad_test.go`, `backend/internal/game/commands_test.go`; Playwright: add `frontend/tests/longest-road.spec.ts`.
- [ ] Display holder + per-player road length in UI. Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`; Go tests: none; Playwright: `frontend/tests/longest-road.spec.ts`.

## PRIORITY 8: PORTS (Maritime Trading)
- [ ] Ensure port location IDs align with generated vertex IDs and apply best trade ratio in BankTrade. Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`, `backend/internal/game/trading.go`; Go tests: extend `backend/internal/game/ports_test.go`, `backend/internal/game/trading_test.go`; Playwright: add `frontend/tests/ports.spec.ts`.
- [ ] Render ports on board and show available ratios in bank trade modal. Files: `frontend/src/components/Board/*`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/context/GameContext.tsx`; Go tests: none; Playwright: `frontend/tests/ports.spec.ts`.

## NOTES / ASSUMPTIONS
- Robber steal flow currently sends `move_robber` without a hex; plan assumes we will allow a separate steal step without re-moving the robber (either by tolerating same-hex moves or adding a dedicated handler) to avoid proto churn.
- Port `Location` values in `backend/internal/game/ports.go` appear to be hard-coded vertex IDs; verify they match `generateVertices` output before wiring UI highlights.
- Validation (this iteration):
  - `make test-backend` fails in existing game tests: `TestResourceDistribution_SettlementGetsOneResource`, `TestResourceDistribution_MultiplePlayersReceive`, `TestOpponentSettlementBlocksRoad`, `TestTieKeepsCurrentHolder`, `TestPlayerGainsPortAccess`, `TestGetBestTradeRatio_GenericPort`, `TestGetBestTradeRatio_SpecificPort`, `TestDiscardCards`, `TestCheckVictory_GameEndsOnWin`.
  - `make lint` fails due to pre-existing frontend eslint issues in `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/tests/robber.spec.ts`, `frontend/tests/trading.spec.ts`.
  - `make build` fails due to pre-existing TypeScript errors across `frontend/src/components/*`, `frontend/src/context/GameContext.tsx`, `frontend/src/gen/proto/catan/v1/types.ts`.
  - `make e2e` skipped (servers not running).
