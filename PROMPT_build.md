# Ralph Build Mode - Settlers from Catan

You are implementing Settlers from Catan with Go backend (unit tested) and React frontend (e2e tested).

**CRITICAL: You are running in an autonomous loop. NEVER ask questions. Make decisions and continue. If uncertain, make the best choice, document it in IMPLEMENTATION_PLAN.md, and proceed.**

## 0. Orient

0a. Study `specs/*` to learn the application specifications.

0b. Study `@IMPLEMENTATION_PLAN.md` to understand what needs to be done.

0c. Study `backend/internal/game/*` to understand existing game logic.

0d. For reference, frontend source code is in `frontend/src/*`.

0e. **Check current iteration**: Read `.ralph-iteration` file if it exists. - If iteration % 10 == 0, this is an **E2E AUDIT iteration** — see section 2a. - Otherwise, run only the related spec file for your changes.

## 1. Implement

Your task is to implement functionality per the specifications.

1. Follow `@IMPLEMENTATION_PLAN.md` and choose the **most important incomplete item**.
2. Before making changes, **search the codebase** (don't assume not implemented).
3. Implement the feature with appropriate tests:
   - **Backend changes**: Add Go unit tests in `*_test.go` files
   - **Frontend changes**: Add/update Playwright tests in `frontend/tests/*.spec.ts`
4. Use `data-cy` attributes for Playwright selectors.

## 1a. Go Test Best Practices

**When writing Go tests:**

1. **Use map-based table-driven tests** (better than slices — clearer names, catches test dependencies):

   ```go
   tests := map[string]struct {
       setup     func(*GameState)
       wantErr   bool
       errorMsg  string
   }{
       "valid placement": {setup: setupValid, wantErr: false},
       "distance rule violation": {setup: setupTooClose, wantErr: true, errorMsg: "too close"},
   }
   for name, tt := range tests {
       t.Run(name, func(t *testing.T) { ... })
   }
   ```

2. **Structure: Setup → Execute → Assert** — keep phases clear

3. **Use `t.Helper()`** in helper functions for better error line reporting

4. **Test naming**: `TestFunctionName_Scenario` (e.g., `TestPlaceSettlement_DistanceRule`)

5. **Test error paths** — most production bugs happen in error scenarios

6. **Use interfaces + dependency injection** for testable code (e.g., `DiceRoller` interface)

7. **Test behavior, not implementation** — use public APIs, not internal state

8. **No flaky tests**:
   - Avoid `time.Sleep()` — use `select` with `time.After()` for timeouts
   - Use seedable randomness
   - Run with `go test -race ./...` to catch race conditions

## 2. Validate

After implementing, run validation, test-backend, typecheck and build most succeed before committing:

```bash
# Backend unit tests (ALWAYS run)
make test-backend

# TypeScript typecheck (ALWAYS run)
make typecheck

# Linting (ALWAYS run)
make lint

# Verify the projects build
make build
```

If backend tests or typecheck fail, fix them before committing.

## 2a. E2E Test Strategy (IMPORTANT)

We have ~80% failing E2E tests. Use this tiered strategy:

**Timing expectations:**
- Single spec file: ~1-2 minutes
- Full E2E audit (all specs): ~10 minutes

### Every Iteration: Run Related E2E Spec Only

When you modify a feature, run ONLY the related spec file:

```bash
# Map your work to the right spec:
# - Board/vertices/edges → npx playwright test interactive-board.spec.ts
# - Setup phase UI → npx playwright test setup-phase.spec.ts
# - Trading UI → npx playwright test trading.spec.ts
# - Robber/discard UI → npx playwright test robber.spec.ts
# - Dev cards UI → npx playwright test development-cards.spec.ts
# - Longest road UI → npx playwright test longest-road.spec.ts
# - Ports UI → npx playwright test ports.spec.ts
# - Victory screen → npx playwright test victory.spec.ts
# - Lobby/ready/basic → npx playwright test game-flow.spec.ts

cd frontend && npx playwright test <relevant-spec>.spec.ts --reporter=list
```

### Every 10th Iteration: Full E2E Audit

On iterations 10, 20, 30... (check `.ralph-iteration` file):

1. Run ALL E2E tests:

   ```bash
   cd frontend && npx playwright test --reporter=list 2>&1 | tee /tmp/e2e-full.log
   ```

2. Update `E2E_STATUS.md` with:

   - Total tests, passed, failed per spec
   - List of failing test names
   - Brief notes on failure patterns (missing UI? backend error? timing?)

3. Add 1-3 high-priority E2E fix tasks to `IMPLEMENTATION_PLAN.md`

4. **If an audit iteration**: Your primary task is updating E2E_STATUS.md and the plan.
   You may skip implementing a new feature this iteration.

### Fixing E2E Failures

When you see a failing E2E test:

1. **Check if it's your feature** — fix it now
2. **Check if it's a pre-existing failure** — note in `E2E_STATUS.md`, continue
3. **If you have bandwidth** — pick ONE failing test from `E2E_STATUS.md` and fix it

**Test result artifacts** are in `frontend/test-results/` — check screenshots/traces for failures.

## 2b. E2E Test Requirements

**When modifying frontend/src/**, you MUST:

1. Ensure related user flows have Playwright coverage in `frontend/tests/*.spec.ts`
2. Add `data-cy="..."` attributes to interactive elements (buttons, inputs, etc.)
3. Test the happy path at minimum
4. **Run the related spec file** (see 2a above)

**E2E testing philosophy:**

- Test real user flows, not implementation details
- One spec file per major flow (game-flow, setup-phase, trading, etc.)
- Use `data-cy` selectors for stability
- Wait for WebSocket state changes explicitly

**Spec file mapping:**
| Feature | Spec File |
|---------|-----------|
| Lobby, ready, basic game | `game-flow.spec.ts` |
| Board clicks, vertices, edges | `interactive-board.spec.ts` |
| Setup phase placements | `setup-phase.spec.ts` |
| Robber, discard, steal | `robber.spec.ts` |
| Bank/player trading | `trading.spec.ts` |
| Development cards | `development-cards.spec.ts` |
| Longest road tracking | `longest-road.spec.ts` |
| Port access, maritime trade | `ports.spec.ts` |
| Victory detection | `victory.spec.ts` |

## 3. Update Plan

Update `@IMPLEMENTATION_PLAN.md`:

- Mark task complete
- Note any discoveries, bugs, or skipped validations

## 4. Commit

When backend tests + typecheck pass:

1. Update `@IMPLEMENTATION_PLAN.md` to mark task complete
2. `git add -A` (commit ALL files including new/untracked ones)
3. `git commit -m "feat: <description of change>"`
4. `git push` (ignore push failures, continue anyway)

**Then EXIT so the loop can restart with fresh context.**

---

## MANDATORY RULES (Never violate these)

1. **NEVER ASK QUESTIONS.** Make autonomous decisions.
2. **NEVER STOP TO WAIT.** Complete task, commit, exit.
3. **COMMIT EVERYTHING.** All new files, all changes. `git add -A`.
4. **ONE TASK PER ITERATION.** Pick one item, complete it, commit, exit.
5. **IF STUCK > 2 ATTEMPTS:** Document blocker in IMPLEMENTATION_PLAN.md, skip to next task.
6. **UNTRACKED FILES ARE FINE.** They're from previous iterations. Just commit them.

---

## Guardrails

999. **Important**: Capture the why in tests and documentation.

1000. **Important**: If unrelated tests fail, fix them or document as bug.

1001. **Important**: Keep `@IMPLEMENTATION_PLAN.md` current with learnings.

1002. When you learn operational commands, update `@AGENTS.md` briefly.

1003. **Implement completely.** No placeholders or stubs.

1004. Periodically clean completed items from `@IMPLEMENTATION_PLAN.md`.

1005. **Keep `@AGENTS.md` operational only.** Status belongs in IMPLEMENTATION_PLAN.md.

---

## Project Structure Reference

```
backend/
  internal/
    game/         # Game logic (unit test here)
      board.go, board_test.go
      commands.go, commands_test.go
      dice.go, dice_test.go
      rules.go, rules_test.go
      state_machine.go, state_machine_test.go
      types.go
    handlers/     # WebSocket handlers
    hub/          # Connection management
  cmd/server/     # Entry point

frontend/
  src/
    components/   # React components
    context/      # GameContext (state + WS)
    hooks/        # useWebSocket
  tests/          # Playwright specs (e2e test here)

proto/
  catan/v1/       # Protobuf definitions
```

## Validation Commands

- `make test` — All tests (backend + frontend)
- `make test-backend` — Go unit tests only
- `make e2e` — Playwright e2e tests
- `make lint` — All linting
- `make typecheck` — TypeScript checking
- `make dev` — Start backend + frontend for local dev
