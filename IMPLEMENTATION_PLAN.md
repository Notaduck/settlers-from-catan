# Implementation Plan - Settlers from Catan

> Tasks are ordered by priority and dependency. Each task is commit-sized and lists files/tests to touch.

## Priority 1: Interactive Board (CRITICAL)

- [x] Vertex rendering with ownership visuals and data-cy attributes
  - Files: frontend/src/components/Board/Board.tsx, frontend/src/components/Board/Vertex.tsx, frontend/src/components/Board/Board.css
  - Go tests: none
  - Playwright: frontend/tests/interactive-board.spec.ts (existing)

- [x] Edge rendering with ownership visuals and data-cy attributes
  - Files: frontend/src/components/Board/Board.tsx, frontend/src/components/Board/Edge.tsx, frontend/src/components/Board/Board.css
  - Go tests: none
  - Playwright: frontend/tests/interactive-board.spec.ts (existing)

- [x] Initialize setup phase state when the game starts (ensure setup_phase exists and turn order advances)
  - Files: backend/internal/handlers/handlers.go (startGame uses protobuf update), backend/internal/game/state_machine.go (helper if needed)
  - Go tests: backend/internal/handlers/handlers_test.go (startGame sets setup_phase), backend/internal/game/state_machine_test.go (if new helper)
  - Playwright: none

- [x] Add placement mode indicator for current player (settlement vs road vs build) with `data-cy="placement-mode"`
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/Game/Game.css, frontend/src/context/GameContext.tsx (derived placement mode)
  - Go tests: none
  - Playwright: frontend/tests/interactive-board.spec.ts (assert placement mode)

- [x] Compute valid placements and highlight/click valid vertices and edges (setup + normal play)
  - Files: frontend/src/components/Board/Board.tsx, frontend/src/components/Board/Vertex.tsx, frontend/src/components/Board/Edge.tsx, frontend/src/components/Board/Board.css, frontend/src/context/GameContext.tsx (derived placement state), frontend/src/components/Board/placement.ts (new helper)
  - Go tests: none
  - Playwright: frontend/tests/interactive-board.spec.ts (click settlement, click road, invalid placement blocked, highlight class)

- [x] (Foundational) Normalize board graph to standard 54 vertices / 72 edges
  - Files: backend/internal/game/board.go, backend/internal/game/board_test.go
  - Go tests: backend/internal/game/board_test.go (update expectations)
  - Playwright: none

---

## Priority 2: Setup Phase UI

- [x] Setup phase banner and turn indicator (round + current player)
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/Game/Game.css
  - Go tests: none
  - Playwright: frontend/tests/setup-phase.spec.ts (new, uses data-cy setup-phase-banner, setup-turn-indicator)

- [x] Setup instructions and placement counts (settlement 1/2, road 1/2)
  - Files: frontend/src/components/Game/Game.tsx
  - Go tests: none
  - Playwright: frontend/tests/setup-phase.spec.ts (uses data-cy setup-instruction)

- [ ] Resource grant notification after Round 2 settlement
  - Files: frontend/src/context/GameContext.tsx (resource diffing), frontend/src/components/Game/Game.tsx (toast)
  - Go tests: none
  - Playwright: frontend/tests/setup-phase.spec.ts (resource toast visible)

- [ ] Setup completion transition to PLAYING (dice available, setup banner removed)
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/PlayerPanel/PlayerPanel.tsx
  - Go tests: none
  - Playwright: frontend/tests/setup-phase.spec.ts (dice button available after setup)

---

## Priority 3: Victory Flow

- [ ] Add victory evaluation helper and call it after point-gaining actions
  - Files: backend/internal/game/rules.go (CheckVictory helper), backend/internal/game/commands.go (post-build checks), backend/internal/handlers/handlers.go (game over broadcast)
  - Go tests: backend/internal/game/victory_test.go (new)
  - Playwright: none

- [ ] Broadcast GameOver payload and lock game state on victory
  - Files: backend/internal/handlers/handlers.go
  - Go tests: backend/internal/game/victory_test.go, backend/internal/handlers/handlers_test.go (game over broadcast)
  - Playwright: frontend/tests/victory.spec.ts (new)

- [ ] Game over UI with winner and final scores
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/Game/GameOver.tsx (new), frontend/src/context/GameContext.tsx
  - Go tests: none
  - Playwright: frontend/tests/victory.spec.ts

---

## Priority 4: Robber Flow

- [ ] Proto: add discard_cards message + robber phase state (pending discards / current step)
  - Files: proto/catan/v1/messages.proto, proto/catan/v1/types.proto
  - Go tests: none
  - Playwright: none

- [ ] Backend robber commands: discard, move robber, steal
  - Files: backend/internal/game/robber.go (new), backend/internal/game/dice.go, backend/internal/game/state_machine.go (if needed)
  - Go tests: backend/internal/game/robber_test.go (new)
  - Playwright: none

- [ ] WebSocket handlers for discard/move/steal + state transitions
  - Files: backend/internal/handlers/handlers.go
  - Go tests: backend/internal/game/robber_test.go
  - Playwright: frontend/tests/robber.spec.ts (new)

- [ ] Robber UI: discard modal, robber move, steal modal
  - Files: frontend/src/components/Game/DiscardModal.tsx (new), frontend/src/components/Game/StealModal.tsx (new), frontend/src/components/Board/HexTile.tsx, frontend/src/context/GameContext.tsx
  - Go tests: none
  - Playwright: frontend/tests/robber.spec.ts

---

## Priority 5: Trading

- [ ] Proto: add pending trades to GameState and bank trade message
  - Files: proto/catan/v1/types.proto, proto/catan/v1/messages.proto
  - Go tests: none
  - Playwright: none

- [ ] Backend trading logic (propose/respond/bank/expire)
  - Files: backend/internal/game/trading.go (new), backend/internal/handlers/handlers.go
  - Go tests: backend/internal/game/trading_test.go (new)
  - Playwright: frontend/tests/trading.spec.ts (new)

- [ ] Trade UI (trade/build toggle, bank trade, propose trade, incoming trade)
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/Game/BankTradeModal.tsx (new), frontend/src/components/Game/ProposeTradeModal.tsx (new), frontend/src/components/Game/IncomingTradeModal.tsx (new), frontend/src/context/GameContext.tsx
  - Go tests: none
  - Playwright: frontend/tests/trading.spec.ts

---

## Priority 6: Development Cards

- [ ] Proto: add buy_dev_card message and dev card hand/deck state
  - Files: proto/catan/v1/messages.proto, proto/catan/v1/types.proto
  - Go tests: none
  - Playwright: none

- [ ] Backend dev card deck, buy, and play effects
  - Files: backend/internal/game/devcards.go (new), backend/internal/handlers/handlers.go
  - Go tests: backend/internal/game/devcards_test.go (new)
  - Playwright: frontend/tests/development-cards.spec.ts (new)

- [ ] Dev card UI (hand display, buy/play controls, card modals)
  - Files: frontend/src/components/PlayerPanel/DevCards.tsx (new), frontend/src/components/Game/Game.tsx, frontend/src/context/GameContext.tsx
  - Go tests: none
  - Playwright: frontend/tests/development-cards.spec.ts

---

## Priority 7: Longest Road

- [ ] Longest road algorithm and tracking (award/transfer)
  - Files: backend/internal/game/longestroad.go (new), backend/internal/handlers/handlers.go
  - Go tests: backend/internal/game/longestroad_test.go (new)
  - Playwright: frontend/tests/longest-road.spec.ts (new)

- [ ] Longest road UI display
  - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/PlayerPanel/PlayerPanel.tsx
  - Go tests: none
  - Playwright: frontend/tests/longest-road.spec.ts

---

## Priority 8: Ports

- [ ] Proto: add Port definitions to board state
  - Files: proto/catan/v1/types.proto
  - Go tests: none
  - Playwright: none

- [ ] Backend port generation and access tracking
  - Files: backend/internal/game/board.go, backend/internal/game/commands.go
  - Go tests: backend/internal/game/ports_test.go (new)
  - Playwright: frontend/tests/ports.spec.ts (new)

- [ ] Apply port ratios to bank trade
  - Files: backend/internal/game/trading.go
  - Go tests: backend/internal/game/trading_test.go
  - Playwright: frontend/tests/ports.spec.ts

- [ ] Port UI (render ports + trade ratios)
  - Files: frontend/src/components/Board/Port.tsx (new), frontend/src/components/Board/Board.tsx, frontend/src/components/Game/BankTradeModal.tsx
  - Go tests: none
  - Playwright: frontend/tests/ports.spec.ts

---

## Discovered Issues / Risks

- Board graph currently produces more than 54 vertices / 72 edges (see TODOs in backend/internal/game/board_test.go); affects placement correctness and longest road.
- WebSocket handlers mix JSON maps and protobuf state (endTurn/playerReady use map); new GameState fields (pending trades, robber state) can be dropped unless handlers use protobuf updates.

## Validation Notes

- `make test-backend` passed.
- `make typecheck` passed.
- `make lint` reports existing warnings:
  - `frontend/src/components/Game/Game.tsx` has `react-hooks/exhaustive-deps` warnings.
  - `proto/buf.yaml` warns about deprecated DEFAULT category.
- `make e2e` not run (servers not started per instructions).
