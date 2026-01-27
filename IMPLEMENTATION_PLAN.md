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
   - `ports.spec.ts`, `trading.spec.ts`, `longest-road.spec.ts`, `robber.spec.ts`, `victory.spec.ts`, `setup-phase.spec.ts`, `game-flow.spec.ts`, `development-cards.spec.ts`, `interactive-board.spec.ts`
   - For each, log any failing/flaky test and create atomic fix tasks with file/test locations and root causes.
2. **Document and address root causes for any failures discovered.**

---

## E2E Audit and Stabilization [COMPLETED]
- Full Playwright E2E audit conducted (Jan 27, 2026/Iteration 12).
- All tests failed to run for every spec; failures cluster around backend state initialization, websocket/proto contract, and complete breakdown in game UI interactivity.
- See atomic fix tasks added below for each Playwright spec:

### Atomic Fix Tasks by Spec File (from E2E_STATUS.md audit)

#### development-cards.spec.ts
- [ ] Fix backend/frontend protocol to ensure development cards panel renders and all dev card actions (Monopoly, Knight, Road Building, Year of Plenty) send/receive valid updates. Likely root cause: backend is not serving dev cards state over websocket/proto. Files: `backend/internal/game/devcards.go`, `frontend/src/components/DevCardPanel.tsx`, `proto/catan/v1/*.proto`

#### interactive-board.spec.ts
- [ ] Debug board, settlement, and placement rendering; ensure initial board and interactive handlers receive game state. Files: `frontend/src/components/Board/*`, `frontend/src/context/GameContext.tsx`, protocol socket

#### setup-phase.spec.ts
- [ ] Patch server-side state machine or websocket logic so phase banners and setup instructions initialize and propagate. Test that all setup-phase tiles/placements register. Files: `backend/internal/game/setup.go`, `frontend/tests/setup-phase.spec.ts`, `frontend/src/components/Game/Game.tsx`

#### game-flow.spec.ts
- [ ] Debug initial lobby/game creation, join/ready, and phase transitions between lobby→setup→playing. Files: `frontend/src/context/GameContext.tsx`, `frontend/src/hooks/useWebSocket.ts`, `backend/internal/hub/hub.go`

#### ports.spec.ts
- [ ] Fix port state propagation from backend so all ports render, support maritime trade, and correct bank ratios/settings for each player. Files: `backend/internal/game/ports.go`, `frontend/src/components/Ports.tsx`

#### victory.spec.ts
- [ ] Implement/playtest backend-triggered game-over and VP overlays for client; ensure websocket/game end messages are sent and acted on. Files: `backend/internal/game/victory.go`, `frontend/src/components/VictoryOverlay.tsx`, `proto/catan/v1/*.proto`

#### robber.spec.ts
- [ ] Ensure discard modal, move, and steal interactions fire based on backend/phase state. Patch server event logic and frontend wiring. Files: `backend/internal/game/robber.go`, `frontend/src/components/RobberModal.tsx`, proto

#### trading.spec.ts
- [ ] Debug trading modal events and websocket flows (offer/accept/decline exchanges). Ensure all backend trades propagate to clients and result in correct resource counts. Files: `backend/internal/game/trading.go`, `frontend/src/components/TradingDialog.tsx`, websocket

#### longest-road.spec.ts
- [ ] Patch DFS/bonus logic on backend and ensure all client UI badge/bonus for longest road are wired through protocol, updating in real time. Files: `backend/internal/game/longestroad.go`, `frontend/src/components/LongestRoadBadge.tsx`

---

## Prioritized Feature Audit

### [1] E2E Audit and Stabilization [COMPLETED]
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
2. git commit -m "chore: update implementation plan with atomic E2E fix tasks after full audit"
3. git push
4. EXIT

---
