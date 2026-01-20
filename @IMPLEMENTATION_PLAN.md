# IMPLEMENTATION_PLAN - Settlers from Catan

Prioritized gap analysis against `specs/*.md` and current codebase. Each task is commit-sized and lists target files + required tests.

## PRIORITY 1: INTERACTIVE BOARD
- No gaps found versus `specs/interactive-board.md`. Existing vertex/edge rendering, click handlers, and `data-cy` attributes already present in `frontend/src/components/Board/Board.tsx`, `frontend/src/components/Board/Vertex.tsx`, and `frontend/src/components/Board/Edge.tsx`. Playwright coverage: `frontend/tests/interactive-board.spec.ts`.

## PRIORITY 2: SETUP PHASE UI
- No gaps found versus `specs/setup-phase-ui.md`. Setup banner/instructions/resource toast already in `frontend/src/components/Game/Game.tsx` + `frontend/src/context/GameContext.tsx`. Playwright coverage: `frontend/tests/setup-phase.spec.ts`.

## PRIORITY 3: VICTORY FLOW
- Ensure victory detection triggers on dev card effects and bonus transfers (Largest Army / Longest Road), and set `GameStatus` to FINISHED + broadcast `gameOver`. Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/handlers/handlers.go`, `backend/internal/game/rules.go`. Go tests: `backend/internal/game/victory_test.go`, `backend/internal/game/devcards_test.go`. Playwright: extend `frontend/tests/victory.spec.ts` to confirm real victory trigger (not just test endpoint).

## PRIORITY 4: ROBBER FLOW
- No functional gaps found versus `specs/robber-flow.md`, but update E2E to use `forceDiceRoll` so tests exercise discard/move/steal flow. Files: `frontend/tests/robber.spec.ts`, `frontend/tests/helpers.ts`. Go tests: already in `backend/internal/game/robber_test.go`.

## PRIORITY 5: TRADING + TURN PHASES
- Implement `SetTurnPhase` to switch `TRADE ↔ BUILD`, enforce build actions only in BUILD phase, and broadcast `turnChanged`. Files: `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`, `backend/internal/game/commands.go`, `proto/catan/v1/messages.proto` (if message wiring needs update), `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`. Go tests: `backend/internal/game/state_machine_test.go`, `backend/internal/handlers/handlers_test.go`. Playwright: update `frontend/tests/trading.spec.ts` and `frontend/tests/game-flow.spec.ts`.
- Add build action controls and placement mode selection for `settlement | road | city` in BUILD phase. Files: `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Board/placement.ts`, `frontend/src/components/Board/Board.tsx`. Playwright: update `frontend/tests/longest-road.spec.ts`, `frontend/tests/victory.spec.ts`, `frontend/tests/ports.spec.ts` to use `data-cy="build-settlement-btn|build-road-btn|build-city-btn"`.
- Expire pending trades on end turn and enforce “one active trade per proposer.” Files: `backend/internal/game/trading.go`, `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`. Go tests: `backend/internal/game/trading_test.go`, `backend/internal/game/state_machine_test.go`. Playwright: extend `frontend/tests/trading.spec.ts` with expiry + reject duplicate offer.
- Add counter-offer UX for incoming trades (reuse propose modal). Files: `frontend/src/components/Game/IncomingTradeModal.tsx`, `frontend/src/components/Game/ProposeTradeModal.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`. Playwright: extend `frontend/tests/trading.spec.ts` for counter-flow.

## PRIORITY 6: DEVELOPMENT CARDS
- Enforce “one dev card per turn,” active-player-only, and turn/phase validation; add per-player tracking in state if needed. Files: `backend/internal/game/devcards.go`, `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto` (if adding tracking), `proto/catan/v1/messages.proto` + `make generate`. Go tests: `backend/internal/game/devcards_test.go`. Playwright: extend `frontend/tests/development-cards.spec.ts` with “second play same turn fails”.
- Implement Road Building free-placement mode (two roads, no resource cost) with UI guidance and server validation. Files: `backend/internal/game/devcards.go`, `backend/internal/game/commands.go`, `backend/internal/game/state_machine.go`, `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`, `frontend/src/components/Game/Game.tsx`, `frontend/src/components/Board/placement.ts`. Go tests: `backend/internal/game/devcards_test.go`, `backend/internal/game/commands_test.go`. Playwright: extend `frontend/tests/development-cards.spec.ts` with full Road Building flow.
- Hide dev card details from non-owners in websocket payloads (redact `dev_cards`, `dev_card_count`, `dev_cards_purchased_turn` for other players). Files: `backend/internal/handlers/handlers.go`, `frontend/src/context/GameContext.tsx`. Go tests: `backend/internal/handlers/handlers_test.go`. Playwright: add check in `frontend/tests/development-cards.spec.ts` that opponents cannot see card types.

## PRIORITY 7: LONGEST ROAD
- Recalculate `longest_road_player_id` after road/settlement/city placement; apply +2 VP bonus transfers, including breaks by opponent buildings. Files: `backend/internal/game/commands.go`, `backend/internal/game/longestroad.go`, `backend/internal/game/rules.go`, `backend/internal/game/state_machine.go`. Go tests: `backend/internal/game/longestroad_test.go`, `backend/internal/game/commands_test.go`. Playwright: ensure `frontend/tests/longest-road.spec.ts` passes.
- Add UI for longest road holder + per-player road length (`data-cy="longest-road-holder"` and `data-cy="road-length-{playerId}"`). Files: `frontend/src/components/PlayerPanel/PlayerPanel.tsx`, `frontend/src/components/Game/Game.tsx`. Playwright: `frontend/tests/longest-road.spec.ts`.

## PRIORITY 8: PORTS
- Make port placement deterministic and align with standard coastal positions (avoid `rand.Perm`), then ensure ports are generated in stable order. Files: `backend/internal/game/ports.go`, `backend/internal/game/board.go`. Go tests: `backend/internal/game/ports_test.go`, `backend/internal/game/board_test.go`. Playwright: update `frontend/tests/ports.spec.ts` if selectors/order assumptions change.

## NOTES / ASSUMPTIONS
- Ports currently use randomized coastal selection; plan assumes switching to fixed vertex-pair list for the standard board.
- Dev card visibility is currently broadcast to all clients; plan assumes redaction on outbound state snapshots rather than changing client storage.
