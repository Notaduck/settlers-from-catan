# agent.md — Settlers from Catan (Go backend + React frontend)

You are working in a monorepo:

- `backend/` — Go server + game engine
- `frontend/` — React (Vite) UI + Playwright e2e
- `proto/` — buf-managed protobuf schema
- `Makefile` — canonical dev/test commands

Primary goals:

1. Keep the rules/game engine correct and deterministic.
2. Everything is unit tested (Go + TS where applicable).
3. Critical user flows are end-to-end tested with Playwright.
4. Never break protobuf generation or websocket contract.

---

## Ground rules

### Don’t edit generated files

- Go generated: `backend/gen/proto/**`
- TS generated: `frontend/src/gen/proto/**`
  If proto changes are needed: update `.proto` files under `proto/` and run `make generate`.

### Prefer deterministic behavior

- Any randomness (board gen, dice) must be seedable and testable.
- Unit tests must not depend on timing, sleep, or real network where avoidable.

### Keep code changes small and test-first

- If implementing a rule: add failing tests first (Go).
- If implementing a UI flow: update/add Playwright spec(s). test with make e2e

---

## Commands (canonical)

From repo root:

### Install + generate + build

- `make install`
- `make generate`
- `make build`

### Dev

- `make dev` (backend + frontend)
- Backend: `make dev-backend` (port :8080)
- Frontend: `make dev-frontend` (port :3000)

### Tests

- `make test` (all unit tests)
- `make test-backend`
- `make test-frontend` (currently may be minimal)
- `make e2e` (starts backend/frontend and runs Playwright)

### Lint

- `make lint`
- `make lint-proto`
- `make typecheck`

### Housekeeping

- `make clean`
- `make db-reset`

When you change anything:

- Run `make test` at minimum.
- If UI or websocket behavior changes, also run `make e2e`.

---

## Architecture expectations

### Backend (Go)

Likely relevant areas:

- Game logic: `backend/internal/game/*`
- Websocket hub: `backend/internal/hub/*`
- HTTP handlers: `backend/internal/handlers/*`
- Server entry: `backend/cmd/server/main.go`

Backend responsibilities:

- Own the source-of-truth game state and rules.
- Validate all client commands (never trust the UI).
- Broadcast authoritative updates over websocket.
- Keep state transitions explicit and testable.

Testing guidelines (Go):

- Use table-driven tests.
- Cover:
  - Board generation invariants (no invalid coords, correct tile distribution).
  - Turn/state machine transitions (lobby -> started -> turns).
  - Command validation (reject invalid moves).
  - Deterministic playthrough tests where possible.
- Avoid flaky tests: no real sockets unless explicitly testing handlers/hub; prefer in-memory constructs.

### Frontend (React + TS)

Likely relevant areas:

- `frontend/src/context/GameContext.tsx` (client state + websocket)
- `frontend/src/hooks/useWebSocket.ts`
- UI components:
  - `frontend/src/components/Lobby/*`
  - `frontend/src/components/Game/*`
  - `frontend/src/components/Board/*`

Frontend responsibilities:

- Render server-provided state.
- Send user intent as commands.
- Never compute authoritative game logic locally (only UI convenience).

Testing guidelines (Frontend):

- Prefer Playwright for “real user flows”.
- If adding component-level tests later, keep them pure and mock websocket.

---

## Playwright E2E strategy

Location:

- Specs: `frontend/tests/*.spec.ts`
- Config: `frontend/playwright.config.ts`

Run:

- `make e2e` (requires backend/frontend running)
- `cd frontend && npm test`
- `cd frontend && npm run test:headed` (see browser)

E2E principles:

- Test the real flow:
  1. Create game
  2. Join game
  3. Toggle ready
  4. Start game
  5. Board renders correctly and state updates arrive
- Avoid brittle selectors:
  - Prefer `data-cy="..."` attributes in UI.
- Ensure tests are isolated:
  - Each spec should create its own game/lobby.
  - Backend should support easy reset (in-memory store) or unique game IDs.

If tests are flaky:

- Add explicit waits for websocket state changes.
- Use Playwright's built-in auto-waiting and retryability.

---

## Protobuf workflow

- Proto schema lives under: `proto/catan/v1/*`
- Use buf:
  - `make generate` runs `buf generate`
  - `make lint-proto` lints proto files

Rules:

- Any contract change between backend and frontend must go through proto.
- After changing `.proto`:
  - Run `make generate`
  - Update server/client usage accordingly
  - Update unit tests + Playwright tests

---

## Definition of done (DoD)

A change is “done” when:

- `make test` passes locally.
- `make e2e` passes if UI/websocket behavior changed.
- No generated files were hand-edited.
- New behavior has tests:
  - Go rule/state changes => Go unit tests.
  - User-visible flows => Playwright coverage (or updated existing spec).
- The change is deterministic and does not introduce flakiness.

---

## When in doubt

Prefer:

- Correctness > feature speed.
- Explicit state machines and validation.
- Deterministic seeds for randomness.
- Small PR-sized changes with tests.

If something is unclear:

- Inspect existing tests first:
  - `backend/internal/game/*_test.go`
  - `frontend/tests/game-flow.spec.ts`
    and extend patterns already used in the repo.
