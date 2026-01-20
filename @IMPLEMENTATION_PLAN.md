# IMPLEMENTATION_PLAN - Settlers from Catan

Prioritized gap analysis against `specs/*.md` and current codebase. Each task is commit-sized and lists target files + required tests.

## PRIORITY 1: INTERACTIVE BOARD
- ✅ **COMPLETE**: Interactive board fully implemented. Vertex/edge rendering, click handlers, and `data-cy` attributes already present in `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Vertex.tsx`, and `frontend/src/components/Board/Edge.tsx`. Playwright coverage: `frontend/tests/interactive-board.spec.ts`.

## PRIORITY 2: SETUP PHASE UI
- ✅ **COMPLETE**: Setup phase UI fully implemented. Setup banner/instructions/resource toast already in `frontend/src/components/Game/Game.tsx` + `frontend/src/context/GameContext.tsx`. Playwright coverage: `frontend/tests/setup-phase.spec.ts`.

## PRIORITY 3: VICTORY FLOW
- ✅ **COMPLETE**: Victory detection implemented. `PlayDevCard` sets `GameStatus` to FINISHED when victory is detected and handler broadcasts `gameOver` after dev card plays. VP cards keep scoring count when revealed. Added `backend/internal/game/devcards_victory_test.go`. Playwright `frontend/tests/victory.spec.ts` covers victory triggers.

## PRIORITY 4: ROBBER FLOW
- ✅ **COMPLETE**: Robber flow fully implemented. Updated E2E to use `forceDiceRoll` and fully exercise discard/move/steal flow. Files: `frontend/tests/robber.spec.ts`, `frontend/tests/helpers.ts`. Go tests covered in `backend/internal/game/robber_test.go`.

## PRIORITY 5: TURN PHASES & TRADE/BUILD SWITCHING
**STATUS**: 🔶 **PARTIALLY IMPLEMENTED** - Core gap identified
- **MISSING**: `handleSetTurnPhase` handler in `backend/internal/handlers/handlers.go` to process `SetTurnPhaseMessage`
- **MISSING**: Trade/Build toggle buttons in `frontend/src/components/Game/Game.tsx` to switch between `TURN_PHASE_TRADE` ↔ `TURN_PHASE_BUILD`
- **MISSING**: Build action controls (settlement/road/city buttons) for BUILD phase placement mode selection
- **IMPLEMENTED**: Phase validation (trading only in TRADE phase, building only in BUILD phase), automatic ROLL→TRADE transition after dice

### Tasks:
- Add `handleSetTurnPhase` handler in `backend/internal/handlers/handlers.go` and `backend/internal/game/state_machine.go`
- Add Trade/Build phase toggle buttons in `frontend/src/components/Game/Game.tsx` with `data-cy="trade-phase-btn"` and `data-cy="build-phase-btn"`
- Add build action controls (`data-cy="build-settlement-btn|build-road-btn|build-city-btn"`) and placement mode selection
- Go tests: `backend/internal/game/state_machine_test.go`, `backend/internal/handlers/handlers_test.go`
- Playwright: update `frontend/tests/trading.spec.ts` and `frontend/tests/game-flow.spec.ts` for manual phase switching

## PRIORITY 6: TRADING COMPLETION
**STATUS**: 🔶 **PARTIALLY IMPLEMENTED** - Backend trading exists, some UI gaps
- **IMPLEMENTED**: Backend trading logic in `backend/internal/game/trading.go`, bank trade, propose/respond handlers
- **IMPLEMENTED**: Trading modals in frontend (`BankTradeModal`, `ProposeTradeModal`, `IncomingTradeModal`)
- **MISSING**: Trade expiry on end turn - trades should expire when turn ends
- **MISSING**: "One active trade per proposer" enforcement
- **MISSING**: Counter-offer UX (reuse propose modal for incoming trades)

### Tasks:
- Expire pending trades on end turn in `backend/internal/game/state_machine.go` and `backend/internal/game/trading.go`
- Enforce "one active trade per proposer" validation in `backend/internal/handlers/handlers.go`
- Add counter-offer button to `frontend/src/components/Game/IncomingTradeModal.tsx`
- Go tests: extend `backend/internal/game/trading_test.go` with expiry + duplicate offer tests
- Playwright: extend `frontend/tests/trading.spec.ts` with expiry and counter-offer flows

## PRIORITY 7: DEVELOPMENT CARDS COMPLETION
**STATUS**: 🔶 **PARTIALLY IMPLEMENTED** - Core logic exists, some restrictions missing
- **IMPLEMENTED**: Dev card deck, buy/play logic in `backend/internal/game/devcards.go`, frontend dev card panel
- **MISSING**: "One dev card per turn" enforcement (can currently play multiple)
- **MISSING**: Active-player-only validation for dev card plays
- **MISSING**: Road Building free-placement mode with UI guidance
- **MISSING**: Dev card visibility restrictions (opponents can see card types in websocket payload)

### Tasks:
- Add per-player "played dev card this turn" tracking in `backend/internal/game/state_machine.go`
- Implement Road Building special placement mode in `backend/internal/game/devcards.go` and frontend placement system
- Redact `dev_cards`, `dev_card_count`, `dev_cards_purchased_turn` from non-owners in `backend/internal/handlers/handlers.go`
- Go tests: extend `backend/internal/game/devcards_test.go` with one-per-turn and Road Building tests
- Playwright: extend `frontend/tests/development-cards.spec.ts` with "second play fails" and Road Building flow

## PRIORITY 8: LONGEST ROAD ALGORITHM
**STATUS**: 🔶 **PARTIALLY IMPLEMENTED** - Algorithm exists but not integrated with placement
- **IMPLEMENTED**: Longest road algorithm in `backend/internal/game/longestroad.go`
- **MISSING**: Auto-recalculation after road/settlement/city placement
- **MISSING**: +2 VP bonus transfers when longest road holder changes
- **MISSING**: UI display for longest road holder and per-player road lengths

### Tasks:
- Trigger longest road recalculation in `backend/internal/game/commands.go` after placements
- Apply +2 VP bonus transfers in `backend/internal/game/state_machine.go`
- Add longest road UI display in `frontend/src/components/PlayerPanel/PlayerPanel.tsx` with `data-cy="longest-road-holder"`
- Go tests: extend `backend/internal/game/longestroad_test.go` with placement integration tests
- Playwright: ensure `frontend/tests/longest-road.spec.ts` passes with real recalculation

## PRIORITY 9: PORTS DETERMINISTIC PLACEMENT  
**STATUS**: ✅ **IMPLEMENTED** - Ports work, but randomized placement may cause test flake
- **IMPROVEMENT NEEDED**: Make port placement deterministic using fixed coastal vertex-pair list instead of `rand.Perm`
- Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`
- Go tests: `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go`
- Playwright: verify `frontend/tests/ports.spec.ts` remains stable with fixed placement

## NOTES / ASSUMPTIONS
- Interactive board and setup phase UI are fully complete and spec-compliant
- Victory flow and robber mechanics are fully implemented
- Turn phase switching is the critical missing piece that blocks proper game flow
- Trading backend is solid but needs frontend UX improvements
- Development cards need enforcement rules and Road Building special mode
- Longest road needs integration with placement actions for live VP updates