# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-22. Sources: specs/*, current backend/frontend code, E2E_STATUS.md, frontend/test-results/. Quick E2E audit command timed out after 30s (see E2E Stabilization).

## PRIORITY 1: INTERACTIVE BOARD (DONE)
- ✅ Vertex/edge rendering, click handlers, placement highlights, and data-cy attributes are implemented in `frontend/src/components/Board/Board.tsx`.
- ✅ Placement validity is computed in `frontend/src/components/Board/placement.ts`.
- E2E status is unknown (see E2E Stabilization section).

## PRIORITY 2: SETUP PHASE UI (DONE)
- ✅ Setup banner, turn indicator, placement instruction, and resource toast implemented in `frontend/src/components/Game/Game.tsx` + `frontend/src/context/GameContext.tsx`.
- E2E status is unknown (see E2E Stabilization section).

## PRIORITY 3: VICTORY FLOW (DONE)
- ✅ Victory detection + GameOver payload implemented in `backend/internal/game/rules.go` and broadcast in `backend/internal/handlers/handlers.go`.
- ✅ Game over overlay + score breakdown implemented in `frontend/src/components/Game/GameOver.tsx`.
- E2E status is unknown (see E2E Stabilization section).

## PRIORITY 4: ROBBER FLOW (DONE)
- ✅ Discard/move/steal commands implemented in `backend/internal/game/robber.go` + handlers.
- ✅ Robber UI in `frontend/src/components/Game/DiscardModal.tsx`, `StealModal.tsx`, and board robber click targets.
- E2E status is unknown (see E2E Stabilization section).

## PRIORITY 5: TURN PHASE SWITCHING + BUILD CONTROLS (MISSING)
1) Add explicit turn-phase switching (TRADE → BUILD + optional back) ✅
- Implemented: `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
- Tests: `backend/internal/game/state_machine_test.go`, `backend/internal/handlers/handlers_test.go`, `frontend/tests/trading.spec.ts`
- E2E (trading.spec.ts) timed out at 120s; partial run shows timeouts in bank trade + trade accept/offer modal flows (details in E2E_STATUS.md).

2) Add build-mode selection for settlement/road/city (UI + placement)
- Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/placement.ts`, `frontend/src/components/Board/Board.tsx`
- Go tests: (none) backend already validates structure types
- Playwright: extend `frontend/tests/game-flow.spec.ts` (build city and road in BUILD phase)

3) Add city placement affordances + valid city targets
- Files: `frontend/src/components/Board/placement.ts`, `frontend/src/components/Game/Game.tsx`
- Go tests: (none)
- Playwright: extend `frontend/tests/game-flow.spec.ts` and/or `frontend/tests/interactive-board.spec.ts`

## PRIORITY 6: TRADING COMPLETION (PARTIAL)
1) Expire pending trades at end of turn (and/or mark CANCELLED)
- Files: `backend/internal/game/state_machine.go`, `backend/internal/game/trading.go`
- Go tests: extend `backend/internal/game/trading_test.go`
- Playwright: extend `frontend/tests/trading.spec.ts`

2) Enforce “one active trade per proposer”
- Files: `backend/internal/game/trading.go`, `backend/internal/handlers/handlers.go`
- Go tests: extend `backend/internal/game/trading_test.go`
- Playwright: extend `frontend/tests/trading.spec.ts`

3) Add counter-offer flow in UI
- Files: `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/context/GameContext.tsx`
- Playwright: extend `frontend/tests/trading.spec.ts`

## PRIORITY 7: DEVELOPMENT CARDS COMPLETION (PARTIAL)
1) Enforce “one dev card per turn” + active-player-only plays
- Files: `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`, `frontend/src/context/GameContext.tsx`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: extend `frontend/tests/development-cards.spec.ts`

2) Fix “just bought this turn” tracking for multiple cards of same type
- Files: `backend/internal/game/devcards.go`, `proto/catan/v1/types.proto`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: update `frontend/tests/development-cards.spec.ts` for mixed old+new cards

3) Road Building placement mode + UI guidance
- Files: `backend/internal/game/devcards.go`, `frontend/src/components/Board/placement.ts`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: extend `frontend/tests/development-cards.spec.ts`

4) Redact dev-card info from non-owners in GameState broadcasts
- Files: `backend/internal/handlers/handlers.go`, `backend/internal/hub/hub.go` (per-client broadcast), `frontend/src/context/GameContext.tsx`
- Go tests: add redaction coverage in `backend/internal/handlers/handlers_test.go`
- Playwright: optional (verify opponents cannot see card types)

## PRIORITY 8: LONGEST ROAD UI + LIVE DISPLAY (PARTIAL)
1) Show longest-road holder + per-player road lengths
- Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`
- Go tests: (none)
- Playwright: update `frontend/tests/longest-road.spec.ts`

## PRIORITY 9: PORTS DETERMINISTIC PLACEMENT (IMPROVEMENT)
1) Replace random port placement with fixed coastal vertex pairs
- Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`
- Go tests: `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go`
- Playwright: re-run `frontend/tests/ports.spec.ts` for stability

## ADDITIONAL SPEC: 3D BOARD VISUALIZATION (DONE)
- ✅ R3F-based 3D board, vertices, edges, and ports implemented in `frontend/src/components/Board/Board.tsx`.
- ✅ 2D fallback in `frontend/src/components/Board/Board.tsx` + `frontend/src/components/Board/HexTile.tsx`.
- E2E coverage still pending via `frontend/tests/interactive-board.spec.ts`.

---

## E2E STABILIZATION (MANDATORY)
Source: `E2E_STATUS.md` (last audit January 22, 2026). `frontend/test-results/` contains dev-cards failures. Quick audit command (`cd frontend && npx playwright test --reporter=list 2>&1 | head -100`) timed out after 30s in this environment.

### Failing Spec Group: `development-cards.spec.ts` (10/12 failing)
Likely root causes: nondeterministic dev-card draws, resource updates not awaited, and UI timing waits.
- Fix E2E: Year of Plenty modal selection (deterministic dev-card grant or seed deck)
- Fix E2E: Monopoly modal selection (deterministic dev-card grant or seed deck)
- Fix E2E: Knight flow + Largest Army check (ensure dev-card grant and turn phase are stable)
- Fix E2E: “card types render with play buttons” (ensure UI updates after dev-card purchases)
- Fix E2E: “dev cards panel count updates” (wait for WS updates via helpers)
- Fix E2E: “buy dev card with resources” (verify dev-mode grant endpoints + update waits)

### Unverified Spec Files (run audit, then file fixes if failing)
- `interactive-board.spec.ts`
- `setup-phase.spec.ts`
- `robber.spec.ts`
- `trading.spec.ts`
- `longest-road.spec.ts`
- `ports.spec.ts`
- `victory.spec.ts`

Audit tasks (each is a small commit):
- Run single-spec Playwright for each unverified file and capture failures in `E2E_STATUS.md`.
- For each failure, add a targeted fix task in this plan with files + tests.
- Investigate why the quick audit command hangs >30s (Playwright webServer, worker count, or backend lifecycle) and document outcome in `E2E_STATUS.md`.
