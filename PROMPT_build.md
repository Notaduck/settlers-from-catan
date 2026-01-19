# Ralph Build Mode - Settlers from Catan

You are implementing Settlers from Catan with Go backend (unit tested) and React frontend (e2e tested).

**CRITICAL: You are running in an autonomous loop. NEVER ask questions. Make decisions and continue. If uncertain, make the best choice, document it in IMPLEMENTATION_PLAN.md, and proceed.**

## 0. Orient

0a. Study `specs/*` to learn the application specifications.

0b. Study `@IMPLEMENTATION_PLAN.md` to understand what needs to be done.

0c. Study `backend/internal/game/*` to understand existing game logic.

0d. For reference, frontend source code is in `frontend/src/*`.

## 1. Implement

Your task is to implement functionality per the specifications.

1. Follow `@IMPLEMENTATION_PLAN.md` and choose the **most important incomplete item**.
2. Before making changes, **search the codebase** (don't assume not implemented).
3. Implement the feature with appropriate tests:
   - **Backend changes**: Add Go unit tests in `*_test.go` files
   - **Frontend changes**: Add/update Playwright tests in `frontend/tests/*.spec.ts`
4. Use `data-cy` attributes for Playwright selectors.

## 2. Validate

After implementing, run validation:

```bash
# Backend unit tests (ALWAYS run)
make test-backend

# TypeScript typecheck (ALWAYS run)
make typecheck

# Linting (ALWAYS run)
make lint

# Verify the projects build
make build

# E2E tests - SKIP if servers not running. Note in plan and continue.
# Do NOT start servers yourself. E2E will be validated separately.
```

If backend tests or typecheck fail, fix them. If e2e fails due to servers not running, note it and continue.

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
