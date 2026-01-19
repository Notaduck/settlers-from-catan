---
applyTo: '**'
---

# IMPLEMENTATION PLAN - Settlers from Catan

## 1. Interactive Board (CRITICAL)

**Task 1:** Audit clickable vertex/edge rendering and click/tap handlers in `frontend/src/components/Board/Board.tsx`, `Vertex.tsx`, `Edge.tsx` to ensure all required data-cy attributes, hover/disabled states, and valid placement highlighting per `specs/interactive-board.md`.
- Update: Board.tsx, Vertex.tsx, Edge.tsx, Board.css
- Add/extend Playwright E2E: `frontend/tests/interactive-board.spec.ts`

## 2. Setup Phase UI (HIGH)

**Task 2:** Audit setup phase indicators, player turn banner, instructions, highlighting, resource notification, and Playwright selectors in `frontend/src/components/Game/Game.tsx` (and any subcomponents), aligned to all `specs/setup-phase-ui.md` details.
- Update: Game.tsx
- Add/extend Playwright E2E: `frontend/tests/setup-phase.spec.ts`

## 3. Victory Flow (HIGH)

**Task 3:** Complete hidden VP card support in victory point calculation: update `CalculateVictoryPoints` and `CheckVictory` in `backend/internal/game/rules.go` and related logic. Ensure point sources include dev cards.
- Update: `backend/internal/game/rules.go` (+check for TODOs)
- Add unit test: `backend/internal/game/victory_test.go` (new or extend if present)

## 4. Robber Flow (HIGH)

**Task 4:** Validate full backend and frontend flow for discards, move robber, and steal, including edge conditions (no valid victim, discard enforcement, modal UI), cross-checking `specs/robber-flow.md`.
- Update/expand: `backend/internal/game/robber.go`, `backend/internal/handlers/handlers.go`
- Add/extend unit test: `backend/internal/game/robber_test.go`
- Add/extend Playwright E2E: `frontend/tests/robber.spec.ts`

## 5. Trading (MEDIUM)

**Task 5:** Implement backend trading support (player-to-player, bank). Add trade offer/resolve logic and handlers per `specs/trading.md` and proto. Implement trading UI forms/modals and Playwright selectors.
- Add: `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`
- Update: `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto` and `messages.proto` if needed
- Add/extend Playwright E2E: `frontend/tests/trading.spec.ts`

## 6. Development Cards (MEDIUM)

**Task 6:** Implement backend dev card deck management, dev card commands, and frontend UI for buying/playing dev cards, handling each effect per `specs/development-cards.md`.
- Add: `backend/internal/game/devcards.go`, `backend/internal/game/devcards_test.go`
- Update: `backend/internal/handlers/handlers.go`, `proto/catan/v1/types.proto/messages.proto` if needed
- Add/extend Playwright E2E: `frontend/tests/development-cards.spec.ts`

## 7. Longest Road Calculation (MEDIUM)

**Task 7:** Implement/test robust longest road algorithm and triggers. Ensure award/transfer, recalc on path-break, and Playwright coverage match `specs/longest-road.md`.
- Add: `backend/internal/game/longestroad.go`, `backend/internal/game/longestroad_test.go`
- Update: `backend/internal/game/rules.go` if needed
- Add/extend Playwright E2E: `frontend/tests/longest-road.spec.ts`

## 8. Ports & Maritime Trade (LOW)

**Task 8:** Fully implement port schema in proto; generate ports on board, track access, enforce best trade ratio in bank trades per `specs/ports.md`. Add frontend port icons/UI and Playwright selectors.
- Update: `proto/catan/v1/types.proto`, `backend/internal/game/board.go`, `backend/internal/game/ports.go`, `backend/internal/game/ports_test.go`
- Add/extend Playwright E2E: `frontend/tests/ports.spec.ts`

---

**All backend features must have table-driven Go unit tests (backend/internal/game/*_test.go).**
**All Playwright E2E flows must be in dedicated specs (frontend/tests/*.spec.ts), covering all user-facing tasks.**
**Every step should be implemented as a standalone, single-commit change for easy review.**
