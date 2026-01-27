# IMPLEMENTATION PLAN - Settlers from Catan

## Ralph Planning Notes (Jan 27, 2026)

### Autonomous Iteration Log (Iteration 12)

- Iteration 12: Full audit and gap analysis across all features and E2E specs.
- Last major E2E audit only reran development-cards; others due for retest next iteration.
- Interactive board is fully implemented and E2E verified.
- Remaining features (setup-phase, victory, robber, trading, dev cards, longest road, ports) need revalidation by E2E or targeted audit.

---

## E2E Stabilization (Critical-Next)

1. **Re-run Playwright E2E suites for all remaining spec files:**
   - `ports.spec.ts`, `trading.spec.ts`, `longest-road.spec.ts`, `robber.spec.ts`, `victory.spec.ts`, `setup-phase.spec.ts`, `game-flow.spec.ts`
   - For each, log any failing/flaky test and create atomic fix tasks with file/test locations and root causes.
2. **Document and address root causes for any failures discovered.**

## Prioritized Feature Audit

### [1] E2E Audit and Stabilization [CRITICAL-NEXT]
- **Files:** All `frontend/tests/*.spec.ts`
- **Action:** Run full audit, categorize failures, and create atomic tasks to resolve.

### [2] Setup Phase UI [HIGH]
- **Files:** `frontend/src/components/Game/Game.tsx`, `frontend/src/context/GameContext.tsx`, `frontend/tests/setup-phase.spec.ts`
- **Action:** Confirm setup-phase banner, placement instructions, and E2E test validity. Address any UI/test gaps.

### [3] Victory Flow [HIGH]
- **Files:** `backend/internal/game/victory_test.go`, `frontend/tests/victory.spec.ts`, `frontend/src/components/Game/Game.tsx`
- **Action:** Ensure Go victory trigger tests and Playwright coverage for game-over events, overlays, and correct VP display.

### [4] Robber Flow [HIGH]
- **Files:** `backend/internal/game/robber.go`, `backend/internal/game/robber_test.go`, `frontend/tests/robber.spec.ts`
- **Action:** Validate discard, move, steal implementation and Playwright coverage. Add missing Go unit tests if found.

### [5] Trading System [MEDIUM]
- **Files:** `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`, `frontend/tests/trading.spec.ts`
- **Action:** Ensure correct implementation of bank/player trades; Playwright coverage for propose/respond/accept/decline. Patch as needed.

### [6] Development Cards [MEDIUM]
- **Files:** `backend/internal/game/devcards.go`, `backend/internal/game/devcards_test.go`, `frontend/tests/development-cards.spec.ts`
- **Action:** Confirm backend logic for dev cards; check Knight, Monopoly, Year of Plenty, Road Building, VP handling. Add/validate E2E.

### [7] Longest Road Calculation [MEDIUM]
- **Files:** `backend/internal/game/longestroad.go`, `backend/internal/game/longestroad_test.go`, `frontend/tests/longest-road.spec.ts`
- **Action:** Ensure DFS and bonus logic is spec-compliant; all events/recalcs tested in Go and E2E.

### [8] Ports (Maritime Trade) [LOW]
- **Files:** `backend/internal/game/ports.go`, `backend/internal/game/ports_test.go`, `frontend/tests/ports.spec.ts`
- **Action:** Confirm board renders 9 ports, UI shows correct ratios, and all trading rules enforced/tested.

---

## General Guidance
- No further implementation for fully complete/verified features (see interactive board logs).
- For each new gap, specify atomic file/test changes in future plan updates. Maintain small, testable commits.

---

## Commit Steps
1. git add -A
2. git commit -m "chore: update implementation plan"
3. git push
4. EXIT

---
