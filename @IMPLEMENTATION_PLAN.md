# IMPLEMENTATION_PLAN - Settlers from Catan (Ralph Planning Mode)

Last updated: 2026-01-22. Sources: specs/*, current backend/frontend code, E2E_STATUS.md, frontend/test-results/.

## PRIORITY 1: INTERACTIVE BOARD (DONE)
- ✅ Vertex + edge rendering, click handlers, placement highlights, and data-cy attributes are implemented in `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Vertex.tsx`, and `frontend/src/components/Board/Edge.tsx`.
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
1) Add explicit turn-phase switching (TRADE ↔ BUILD)
- Files: `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
- Go tests: `backend/internal/game/state_machine_test.go` (phase transition validation), `backend/internal/handlers/handlers_test.go` (set_turn_phase message)
- Playwright: update `frontend/tests/trading.spec.ts` + `frontend/tests/game-flow.spec.ts`

2) Add build-mode selection for settlement/road/city
- Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/placement.ts`, `frontend/src/components/Board/Board.tsx`
- Go tests: (none) backend already validates structure types
- Playwright: extend `frontend/tests/game-flow.spec.ts` (build city and road in BUILD phase)

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
1) Enforce “one dev card per turn” and active-player-only plays
- Files: `proto/catan/v1/types.proto`, `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`, `frontend/src/context/GameContext.tsx`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: extend `frontend/tests/development-cards.spec.ts`

2) Fix “just bought this turn” tracking for multiple cards of same type
- Files: `proto/catan/v1/types.proto`, `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: update `frontend/tests/development-cards.spec.ts` for mixed old+new cards

3) Road Building placement mode + UI guidance
- Files: `backend/internal/game/devcards.go`, `frontend/src/components/Board/placement.ts`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`
- Go tests: extend `backend/internal/game/devcards_test.go`
- Playwright: extend `frontend/tests/development-cards.spec.ts`

4) Redact dev-card info from non-owners
- Files: `backend/internal/handlers/handlers.go`, `backend/internal/hub/hub.go` (per-client broadcast), `frontend/src/context/GameContext.tsx`
- Go tests: add redaction coverage in `backend/internal/handlers/handlers_test.go`
- Playwright: optional (verify opponents cannot see card types)

## PRIORITY 8: LONGEST ROAD UI + LIVE DISPLAY (PARTIAL)
1) Show longest-road holder + per-player road lengths
- Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Board/placement.ts` or new helper (client-side DFS)
- Go tests: (none)
- Playwright: update `frontend/tests/longest-road.spec.ts`

## PRIORITY 9: PORTS DETERMINISTIC PLACEMENT (IMPROVEMENT)
1) Replace random port placement with fixed coastal vertex pairs
- Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`
- Go tests: `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go`
- Playwright: re-run `frontend/tests/ports.spec.ts` for stability

---

## E2E STABILIZATION (MANDATORY)
Source: `E2E_STATUS.md` (last audit January 22, 2026). Only dev-cards spec is known failing; other specs are untested.

### Failing Spec Group: `development-cards.spec.ts`
Likely root causes: nondeterministic dev-card draws, resource updates not awaited, and UI timing waits.
- Fix E2E: Year of Plenty modal selection (use deterministic dev-card grant or seed deck)
- Fix E2E: Monopoly modal selection (use deterministic dev-card grant or seed deck)
- Fix E2E: Knight flow + Largest Army check (ensure dev-card grant and turn phase are stable)
- Fix E2E: “card types render with play buttons” (ensure UI updates after dev-card purchases)
- Fix E2E: “dev cards panel count updates” (wait for WS updates via helpers)

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
