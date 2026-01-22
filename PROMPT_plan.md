# Ralph Planning Mode - Settlers from Catan

You are analyzing a Settlers from Catan implementation to create a prioritized implementation plan.

**CRITICAL: You are running in an autonomous loop. NEVER ask questions. Make decisions and continue. If uncertain, make the best choice and document your reasoning in IMPLEMENTATION_PLAN.md.**

## 0. Orient

0a. Study `specs/*` to learn what needs to be built:

- `specs/interactive-board.md` - CRITICAL: vertex/edge click handlers
- `specs/setup-phase-ui.md` - Setup flow UI
- `specs/robber-flow.md` - Move robber, steal, discard
- `specs/trading.md` - Player and bank trading
- `specs/ports.md` - Maritime trade ratios
- `specs/development-cards.md` - Dev card deck and effects
- `specs/longest-road.md` - Graph algorithm for road length
- `specs/victory-flow.md` - Win detection and game over

0b. Study `@IMPLEMENTATION_PLAN.md` (if present) to understand existing plan state.

0c. Study existing backend code:

- `backend/internal/game/*.go` - Game logic
- `backend/internal/handlers/handlers.go` - WebSocket handlers

0d. Study existing frontend code:

- `frontend/src/context/GameContext.tsx` - Client state, WS messages
- `frontend/src/components/Board/Board.tsx` - Board rendering
- `frontend/src/components/Game/Game.tsx` - Game UI
- `frontend/tests/*.spec.ts` - Existing Playwright tests

0e. Study `proto/catan/v1/*.proto` to understand the API contract.

0f. **Check E2E test status** (IMPORTANT):

- Review `E2E_STATUS.md` if it exists for known failing tests
- Check `frontend/test-results/` for recent failure artifacts
- Run quick E2E audit if needed: `cd frontend && npx playwright test --reporter=list 2>&1 | head -100`

## 1. Gap Analysis & Plan Creation

Compare existing source code against `specs/*`. Create/update `@IMPLEMENTATION_PLAN.md` as a prioritized bullet point list.

Consider:

- TODO comments, minimal implementations, placeholders
- Skipped or flaky tests
- Inconsistent patterns between specs and code
- Dependencies between features (interactive-board unblocks everything)
- **Failing E2E tests** — group related failures as tasks

## 1a. E2E Audit Tasks

When updating the plan, include an **E2E Stabilization** section:

1. Group failing E2E tests by spec file
2. Identify root causes (missing backend logic? missing UI? broken selector?)
3. Create atomic fix tasks: "Fix E2E: [test name] - [likely cause]"

**Current E2E spec files:**

- `game-flow.spec.ts` — Lobby, ready, basic game start
- `interactive-board.spec.ts` — Board vertex/edge clicks
- `setup-phase.spec.ts` — Setup placements
- `robber.spec.ts` — Robber move, discard, steal
- `trading.spec.ts` — Bank and player trading
- `development-cards.spec.ts` — Dev card buying/playing
- `longest-road.spec.ts` — Road length tracking
- `ports.spec.ts` — Maritime trade ratios
- `victory.spec.ts` — Win detection

**Priority Order:**

1. Interactive Board (unblocks all placement)
2. Setup Phase UI (needed for playable game)
3. Victory Flow (needed to end game)
4. Robber Flow (core mechanic)
5. Trading (strategic depth)
6. Development Cards (strategic depth)
7. Longest Road algorithm (VP accuracy)
8. Ports (enhancement)

## 2. Commit and Exit

After updating IMPLEMENTATION_PLAN.md:

1. `git add -A`
2. `git commit -m "chore: update implementation plan"`
3. `git push` (ignore failures)
4. **EXIT immediately**

---

## MANDATORY RULES (Never violate these)

1. **NEVER ASK QUESTIONS.** Make autonomous decisions.
2. **NEVER STOP TO WAIT.** Complete analysis, update plan, commit, exit.
3. **PLAN ONLY.** Do NOT implement any code.
4. **ONE ITERATION = ONE PLAN UPDATE.** Then exit for fresh context.
5. **DO NOT ASSUME MISSING.** Search code before adding tasks.

---

## Task Format

Each task should specify:

- What file(s) to modify
- What Go unit tests to add (backend)
- What Playwright e2e tests to add (frontend)

Keep tasks small: 1 task = 1 commit-sized change.

## Backpressure Commands (for reference)

These will be run during BUILD mode:

- `make test-backend` - Go unit tests
- `make e2e` - Playwright end-to-end tests
- `make lint` - Linting
- `make typecheck` - TypeScript type checking

## ULTIMATE GOAL

We want to achieve a fully playable Settlers from Catan game where:

- Every Go function in `backend/internal/game/*` has unit tests
- Every user-visible flow has Playwright e2e coverage
- The game follows standard Catan rules

If a required spec is missing, create it at `specs/FILENAME.md`. Document all planned work in `@IMPLEMENTATION_PLAN.md`.
