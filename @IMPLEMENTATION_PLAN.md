# IMPLEMENTATION_PLAN - Settlers from Catan

> Gap analysis against `specs/*.md` + current codebase. Each task is commit-sized and lists files + required tests.

## BLOCKERS / FOUNDATIONAL (Completed)
- [x] HTTP/WS handler layer foundational wiring (see prior commits)

## PRIORITY 1: INTERACTIVE BOARD
- [x] Vertex/edge rendering, placement highlighting, and `data-cy` attributes implemented on the board UI. Files: `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Edge.tsx`, `frontend/src/components/Board/Vertex.tsx`; Go tests: none (backend vertex/edge logic already covered); Playwright: `frontend/tests/interactive-board.spec.ts` (blocked by WS handlers above).

## PRIORITY 2: SETUP PHASE UI
- [x] Setup banner, turn indicator, placement instruction, and resource toast implemented. Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`; Go tests: none (state machine tests already cover setup); Playwright: `frontend/tests/setup-phase.spec.ts`.

## PRIORITY 3: VICTORY FLOW
- [x] Add authoritative victory detection + `game_over` broadcast with score breakdown, and prevent actions after FINISHED. Files: `backend/internal/game/rules.go`, `backend/internal/handlers/handlers.go`; Go tests: extend `backend/internal/game/victory_test.go`; Playwright: `frontend/tests/victory.spec.ts`.
- [x] Frontend consumes `game_over` payload and renders breakdown from payload (not inferred fields) and disables actions. Files: `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/GameOver.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/PlayerPanel/PlayerPanel.tsx`; Go tests: none; Playwright: `frontend/tests/victory.spec.ts`.

## PRIORITY 4: ROBBER FLOW
- [x] Initialize and progress RobberPhase on roll 7 (discard -> move -> steal), gating turn progression until completion. Files: `backend/internal/game/dice.go`, `backend/internal/game/robber.go`, `backend/internal/game/state_machine.go`; Go tests: `backend/internal/game/robber_test.go`, `backend/internal/game/dice_test.go`; Playwright: `frontend/tests/robber.spec.ts`.
- [x] Wire discard/move/steal handlers and UI, including robber-hex click support and steal selection flow without invalid MoveRobber calls. Files: `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Game/DiscardModal.tsx`, `frontend/src/components/Game/StealModal.tsx`, `frontend/src/components/Board/HexTile.tsx`; Go tests: handler coverage in `backend/internal/handlers/handlers_test.go`; Playwright: `frontend/tests/robber.spec.ts`.

## PRIORITY 5: TRADING
- [ ] Implement phase switching (TRADE <-> BUILD), end-turn logic, and trade expiry/cleanup at end of turn. Files: `backend/internal/game/commands.go`, `backend/internal/game/trading.go`, `backend/internal/handlers/handlers.go`; Go tests: `backend/internal/game/commands_test.go`, `backend/internal/game/trading_test.go`; Playwright: `frontend/tests/trading.spec.ts` (if missing, create).
- [ ] Backend: Complete all `// TODO` and incomplete logic (see `backend/internal/game/trading.go`); add port support for trade ratios <4 on bank trades (spec: ports). Go tests: add/extend `backend/internal/game/trading_test.go`.
- [ ] Frontend: Fully implement BankTradeModal, ProposeTradeModal, IncomingTradeModal; wire modal actions to websocket client sends (GameContext), not stub code. Files: `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`; Playwright: `frontend/tests/trading.spec.ts` (create if missing).

## PRIORITY 6: DEVELOPMENT CARDS
- [ ] Implement dev card deck management, drawing logic, per-card effect logic for all spec cards (Knight, Road Building, Monopoly, Year of Plenty, Victory Point), and per-turn play rules (one per turn except VP). Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/game/robber.go`, `backend/internal/game/rules.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto`, `proto/catan/v1/messages.proto`; Go tests: `backend/internal/game/devcards_test.go`.
- [ ] Complete all visible `// TODO` cases and placeholder logic in `devcards.go`. Go tests: extend `backend/internal/game/devcards_test.go` for each card/effect.
- [ ] Frontend: Build dev card hand UI (buy button, card list, play buttons/options), wire to websocket messages for buying and playing cards. Files: `frontend/src/components/Game/*`, `frontend/src/context/GameContext.tsx`; Playwright: `frontend/tests/development-cards.spec.ts` (create this test for full story coverage).

## PRIORITY 7: LONGEST ROAD
- [ ] Validate longest road logic for all edge cases, recalculate and persist road holder on road/building placements, enforce tie rules and update victory checks. Files: `backend/internal/game/longestroad.go`, `backend/internal/game/commands.go`, `backend/internal/game/rules.go`; Go tests: `backend/internal/game/longestroad_test.go`, `backend/internal/game/commands_test.go`.
- [ ] Frontend: Display current holder, per-player road length, and optional highlight on board. Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`; Playwright: `frontend/tests/longest-road.spec.ts` (create this test if not present).

## PRIORITY 8: PORTS (Maritime Trading)
- [ ] Ensure port location IDs are consistent between backend generation and frontend rendering (vertex and edge IDs), and backend returns best available ratio in bank trading. Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`, `backend/internal/game/trading.go`; Go tests: extend `backend/internal/game/ports_test.go`, `backend/internal/game/trading_test.go`.
- [ ] Frontend: Render ports on board, show available trade ratios in BankTradeModal, and highlight accessible ports for active player. Files: `frontend/src/components/Board/*`, `frontend/src/components/Game/BankTradeModal.tsx`, `frontend/src/context/GameContext.tsx`; Playwright: `frontend/tests/ports.spec.ts` (create this test to cover UI/logic).

## PLAYWRIGHT E2E REQUIRED FILES
- [ ] `frontend/tests/development-cards.spec.ts` (cover all dev card actions/effects)
- [ ] `frontend/tests/longest-road.spec.ts` (cover longest road claim/loss/victory)
- [ ] `frontend/tests/ports.spec.ts` (cover port rendering, access, trade ratio)

## VALIDATION AND FLAKY TESTS
- [ ] Resolve failing or skipped backend unit tests (listed in prior plan)
- [ ] Resolve frontend lint, type errors; ensure all new files pass `make lint` and `make typecheck`
- [ ] Ensure all new features covered by at least one atomic Playwright e2e spec

## NOTES / ASSUMPTIONS
- All frontend trading modals and dev card UI must become fully functional (no TODO/stub modals left)
- Backend TODOs identified above must be completed for final rules/spec conformance
- All new E2E coverage must actually exist: create the three missing Playwright specs as top priority within each feature area
