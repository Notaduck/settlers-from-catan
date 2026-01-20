# IMPLEMENTATION PLAN - Settlers from Catan

> Prioritized, commit-sized tasks derived from gap analysis and code/spec review. Every planned change lists affected files and required test updates.

----

## STATUS SUMMARY (2026-01-20)

**All 8 priority spec features are BACKEND COMPLETE:**
- ✅ Interactive Board
- ✅ Setup Phase UI
- ✅ Victory Flow
- ✅ Robber Flow
- ✅ Trading System
- ✅ Development Cards
- ✅ Longest Road
- ✅ Ports

**Backend health:**
- ✅ All Go unit tests pass (make test-backend)
- ✅ All builds pass (make build)
- ✅ TypeScript passes (make typecheck)
- ✅ Lint passes with 2 acceptable warnings (make lint)

**E2E Test Coverage:**
- ✅ Interactive board (full implementation)
- ✅ Setup phase (full implementation)
- ⚠️ Victory flow (placeholder only)
- ⚠️ Robber flow (placeholder only)
- ⚠️ Trading (placeholder only)
- ⚠️ Development cards (placeholder only)
- ❌ Longest road (not created)
- ❌ Ports (not created)

**Frontend UI Gaps:**
- ⚠️ ProposeTradeModal (stub/TODO)
- ⚠️ IncomingTradeModal (stub/TODO)
- ⚠️ Knight card → robber move integration (missing)
- ⚠️ Road Building → free placement mode (missing)

----

## REMAINING WORK

### PRIORITY 1: E2E TEST IMPLEMENTATION

All backend logic is complete. Focus now shifts to comprehensive E2E test coverage to ensure all user flows work end-to-end.

#### Task 1.1: Implement Victory Flow E2E Tests
**Spec:** `specs/victory-flow.md`
**File:** `frontend/tests/victory.spec.ts`

**Current state:** Placeholder tests only (30 lines)

**Required tests:**
- [ ] Game over screen shows when player wins
- [ ] Winner name and VP displayed correctly
- [ ] All players' final scores visible
- [ ] Hidden VP cards revealed in final scores
- [ ] New game button works (returns to lobby/create)

**Implementation notes:**
- May require test helper to simulate/accelerate game to victory condition
- Consider backend test endpoint to set game state to near-victory
- Verify GameOver overlay blocks further actions

**Files to modify:**
- `frontend/tests/victory.spec.ts` (expand from placeholder)

**Validation:**
- `make e2e` passes victory tests

---

#### Task 1.2: Implement Robber Flow E2E Tests
**Spec:** `specs/robber-flow.md`
**File:** `frontend/tests/robber.spec.ts`

**Current state:** Placeholder tests with window.__test stubs (69 lines)

**Required tests:**
- [ ] Rolling 7 shows discard modal for players with >7 cards
- [ ] Discard modal enforces correct card count
- [ ] Discard submit works and closes modal
- [ ] After discard, robber move UI appears
- [ ] Clicking hex moves robber
- [ ] Steal UI shows adjacent players
- [ ] Stealing transfers a resource

**Implementation notes:**
- Replace `window.__test` stubs with real game flow
- Use test helpers to grant >7 cards to players
- Verify discard count calculation (half, rounded down)
- Verify robber placement restrictions (cannot place on same hex)

**Files to modify:**
- `frontend/tests/robber.spec.ts` (replace stubs with real flow)

**Validation:**
- `make e2e` passes robber tests

---

#### Task 1.3: Implement Trading E2E Tests
**Spec:** `specs/trading.md`
**File:** `frontend/tests/trading.spec.ts`

**Current state:** Placeholder tests only (40 lines)

**Required tests:**
- [ ] Bank trade 4:1 works
- [ ] Cannot bank trade without 4 resources
- [ ] Player can propose trade to another
- [ ] Trade recipient sees offer modal
- [ ] Accepting trade transfers resources
- [ ] Declining trade notifies proposer
- [ ] Cannot trade outside trade phase

**Implementation notes:**
- Test both bank and player-to-player trading
- Verify resource counts update correctly
- Test trade expiration at turn end
- ProposeTradeModal and IncomingTradeModal must be fully implemented first (see Priority 2)

**Files to modify:**
- `frontend/tests/trading.spec.ts` (expand from placeholder)

**Validation:**
- `make e2e` passes trading tests

---

#### Task 1.4: Implement Development Cards E2E Tests
**Spec:** `specs/development-cards.md`
**File:** `frontend/tests/development-cards.spec.ts`

**Current state:** Placeholder tests only (118 lines)

**Required tests:**
- [ ] Can buy development card with correct resources
- [ ] Cannot buy without resources
- [ ] Playing knight moves robber (requires Task 2.1)
- [ ] Monopoly collects resources from all players
- [ ] Year of plenty grants 2 resources
- [ ] Road building allows 2 free roads (requires Task 2.2)
- [ ] VP cards count toward victory
- [ ] Cannot play more than 1 dev card per turn

**Implementation notes:**
- Test all 5 dev card types
- Verify deck initialization (25 cards, correct distribution)
- Test "cannot play card bought this turn" rule
- Requires Knight and Road Building UI integration (see Priority 2)

**Files to modify:**
- `frontend/tests/development-cards.spec.ts` (expand from placeholder)

**Validation:**
- `make e2e` passes dev card tests

---

#### Task 1.5: Create Longest Road E2E Tests
**Spec:** `specs/longest-road.md`
**File:** `frontend/tests/longest-road.spec.ts` (NEW)

**Current state:** Does not exist

**Required tests:**
- [ ] 5 connected roads shows Longest Road badge
- [ ] Longer road takes Longest Road from another player
- [ ] Broken road loses Longest Road
- [ ] Longest Road badge shows in UI with player name
- [ ] Road length per player displayed correctly

**Implementation notes:**
- Backend logic complete (DFS algorithm, recalc triggers)
- UI shows badges (per existing plan)
- Test tie-breaking rule (current holder keeps on tie)

**Files to create:**
- `frontend/tests/longest-road.spec.ts` (new)

**Validation:**
- `make e2e` passes longest road tests

---

#### Task 1.6: Create Ports E2E Tests
**Spec:** `specs/ports.md`
**File:** `frontend/tests/ports.spec.ts` (NEW)

**Current state:** Does not exist

**Required tests:**
- [ ] Ports render on board with correct icons
- [ ] Bank trade shows 4:1 by default
- [ ] Player with 3:1 port sees 3:1 option in trade modal
- [ ] Player with 2:1 wheat port can trade 2 wheat for 1 other
- [ ] Port access granted when settlement placed on port vertex

**Implementation notes:**
- Backend logic complete (port generation, trade ratios)
- Frontend rendering complete (Port.tsx, BankTradeModal.tsx)
- Test getBestTradeRatio calculation in UI

**Files to create:**
- `frontend/tests/ports.spec.ts` (new)

**Validation:**
- `make e2e` passes ports tests

---

### PRIORITY 2: FRONTEND UI COMPLETION

Complete stub/TODO UI components to enable full E2E testing.

#### Task 2.1: Knight Card → Robber Move Integration
**Spec:** `specs/development-cards.md` (Knight effect)
**Blocker for:** Task 1.4 (Dev cards E2E)

**Current state:** Knight card plays via `playDevCard` but does not trigger robber move UI

**Required implementation:**
- [ ] After Knight card played, trigger robber move mode
- [ ] Reuse existing robber move UI (hex selection)
- [ ] After robber placed, show steal UI (if applicable)
- [ ] Knight count increments (backend already handles)
- [ ] Largest Army check triggers (backend already handles)

**Files to modify:**
- `frontend/src/components/Game/Game.tsx` (add Knight → robber mode logic)
- `frontend/src/context/GameContext.tsx` (track Knight-triggered robber state)

**Backend:** Already complete (`PlayDevCard` for Knight increments `knightsPlayed`, `RecalculateLargestArmy`)

**Validation:**
- Unit test: playing Knight increments knight count
- E2E test: Knight → robber → steal flow works
- `make test-backend` passes
- `make e2e` passes

---

#### Task 2.2: Road Building → Free Placement Mode
**Spec:** `specs/development-cards.md` (Road Building effect)
**Blocker for:** Task 1.4 (Dev cards E2E)

**Current state:** Road Building card plays via `playDevCard` but does not enable free road placement

**Required implementation:**
- [ ] After Road Building played, enter "place 2 roads" mode
- [ ] Highlight valid road placements (same as normal)
- [ ] Roads placed without resource cost
- [ ] Track roads placed count (2 max)
- [ ] Exit mode after 2 roads placed
- [ ] Normal road placement rules still apply (adjacency, etc.)

**Files to modify:**
- `frontend/src/components/Game/Game.tsx` (add Road Building placement mode)
- `frontend/src/context/GameContext.tsx` (track free road placement state)
- Backend handler may need `free_placement` flag or separate message type

**Backend considerations:**
- `PlayDevCard` for Road Building validates card but doesn't grant resources
- May need separate `PlaceRoadFree` message or flag in `BuildStructure`
- Review `backend/internal/handlers/handlers.go` for free placement support

**Validation:**
- E2E test: Road Building → place 2 roads → resources not deducted
- `make test-backend` passes
- `make e2e` passes

---

#### Task 2.3: Complete ProposeTradeModal UI
**Spec:** `specs/trading.md`
**Blocker for:** Task 1.3 (Trading E2E)

**Current state:** Stub implementation with TODO comment

**Required implementation:**
- [ ] Resource offer selection UI (increment/decrement buttons per resource)
- [ ] Resource request selection UI (increment/decrement buttons per resource)
- [ ] Player target selection (dropdown or buttons: specific player or "All")
- [ ] Validation: cannot offer resources player doesn't have
- [ ] Submit button calls `proposeTrade` from GameContext
- [ ] Cancel button closes modal

**Files to modify:**
- `frontend/src/components/Game/ProposeTradeModal.tsx` (full implementation)

**Data attributes:**
- `data-cy="propose-trade-modal"`
- `data-cy="trade-offer-{resource}"`
- `data-cy="trade-request-{resource}"`
- `data-cy="propose-trade-submit-btn"`
- `data-cy="propose-trade-cancel-btn"`

**Backend:** Already complete (`handleProposeTrade`, `ProposeTrade` command)

**Validation:**
- E2E test: propose trade flow works
- `make typecheck` passes
- `make lint` passes
- `make e2e` passes

---

#### Task 2.4: Complete IncomingTradeModal UI
**Spec:** `specs/trading.md`
**Blocker for:** Task 1.3 (Trading E2E)

**Current state:** Stub implementation with TODO comment

**Required implementation:**
- [ ] Display offering player name
- [ ] Show resources offered (icons + counts)
- [ ] Show resources requested (icons + counts)
- [ ] Accept button calls `acceptTrade` from GameContext
- [ ] Decline button calls `declineTrade` from GameContext
- [ ] Optional: Counter button (opens ProposeTradeModal with reversed offer)

**Files to modify:**
- `frontend/src/components/Game/IncomingTradeModal.tsx` (full implementation)
- `frontend/src/context/GameContext.tsx` (add `acceptTrade`, `declineTrade` methods)

**Data attributes:**
- `data-cy="incoming-trade-modal"`
- `data-cy="accept-trade-btn"`
- `data-cy="decline-trade-btn"`

**Backend:** Already complete (`handleRespondTrade`, `AcceptTrade`, `DeclineTrade` commands)

**Validation:**
- E2E test: incoming trade → accept/decline works
- `make typecheck` passes
- `make lint` passes
- `make e2e` passes

---

### PRIORITY 3: CODE QUALITY & POLISH

#### Task 3.1: Remove Stub/TODO Comments
**File:** `frontend/src/components/Game/Game.tsx`

**Lines to address:**
- Line 256: `// TODO: Hook up to trade send` (bank trade)
- Line 270: `// TODO: Hook up to trade send` (propose trade)
- Line 37: `// Simulated offer for stub incoming modal demo`

**Action:**
- After Tasks 2.3 and 2.4 are complete, remove stub comments
- Remove demo incoming trade simulation logic

**Files to modify:**
- `frontend/src/components/Game/Game.tsx`

**Validation:**
- `make lint` passes with no TODO warnings
- E2E tests confirm real trade flow works

---

#### Task 3.2: Fix ESLint useMemo Warnings
**File:** `frontend/src/components/Game/Game.tsx`

**Current warnings:**
```
96:9  warning  The 'players' logical expression could make the dependencies of useMemo Hook (at line 101) change on every render.
96:9  warning  The 'players' logical expression could make the dependencies of useMemo Hook (at line 107) change on every render.
```

**Action:**
- Wrap `players` initialization in its own `useMemo` hook
- Ensure stable dependencies for dependent `useMemo` hooks

**Files to modify:**
- `frontend/src/components/Game/Game.tsx`

**Validation:**
- `make lint` passes with 0 warnings
- No functional changes

---

### PRIORITY 4: OPTIONAL ENHANCEMENTS

#### Task 4.1: 3D Board Visualization
**Spec:** `specs/3d-board.md`
**Priority:** HIGH (per spec) but OPTIONAL (game fully playable without)

**Current state:** 2D SVG board rendering works correctly

**Required implementation:**
- [ ] Install React Three Fiber (`npm install three @react-three/fiber @react-three/drei`)
- [ ] Create 3D hex tile component with resource-specific heights/colors
- [ ] Implement 3D vertices (spheres) and edges (cylinders)
- [ ] Add OrbitControls for camera rotation/zoom
- [ ] Preserve all `data-cy` attributes for Playwright compatibility
- [ ] Ensure 60fps performance on mid-range hardware
- [ ] Lazy load Three.js bundle (code split)

**Files to create:**
- `frontend/src/components/Board/Board3D.tsx` (new)
- `frontend/src/components/Board/HexTile3D.tsx` (new)
- `frontend/src/components/Board/Vertex3D.tsx` (new)
- `frontend/src/components/Board/Edge3D.tsx` (new)

**Files to modify:**
- `frontend/package.json` (add Three.js dependencies)
- `frontend/src/components/Game/Game.tsx` (toggle between 2D/3D board)

**Validation:**
- All existing E2E tests pass with 3D board
- Performance: 60fps on mid-range hardware
- Accessibility: all interactive elements retain `data-cy` attributes

**Notes:**
- This is a significant visual enhancement but not required for gameplay
- Consider feature flag to toggle 2D/3D rendering
- May require separate E2E tests for 3D-specific interactions

----

## VALIDATION CHECKLIST

After each task:
- [ ] `make test-backend` passes (all Go unit tests)
- [ ] `make build` passes (backend + frontend builds)
- [ ] `make typecheck` passes (TypeScript compilation)
- [ ] `make lint` passes (Go vet + ESLint)
- [ ] `make e2e` passes (Playwright tests for affected features)

----

## BLOCKERS

**RESOLVED (2026-01-20):**
- ✅ 2 longest road test failures (fixed)
- ✅ Flaky resource distribution tests (stable)
- ✅ Catastrophic handler breakage (resolved)
- ✅ TypeScript build errors (all 55+ errors fixed)
- ✅ Lint errors (ProposeTradeModal, IncomingTradeModal, GameContext)

**NO CURRENT BLOCKERS**

----

## NOTES

- Backend implementation is **COMPLETE** for all 8 priority specs
- All Go unit tests pass
- Focus now on **E2E test coverage** and **frontend UI completion**
- ProposeTradeModal and IncomingTradeModal are stubs but do not block gameplay (just testing)
- Knight and Road Building cards need UI integration to complete dev card flow
- 3D board is optional enhancement, not required for playable game

----

# Plan Validation (2026-01-20)

**Gap Analysis Complete:**

1. **Backend:** ✅ No gaps. All game logic implemented and tested.
2. **Frontend UI:** ⚠️ 4 components stub/incomplete (ProposeTradeModal, IncomingTradeModal, Knight→robber, RoadBuilding→free placement)
3. **E2E Tests:** ⚠️ 6 of 8 specs have placeholder tests only; 2 specs missing tests entirely

**Prioritization Rationale:**
- E2E tests are highest priority to validate end-to-end user flows
- Frontend UI stubs block E2E test implementation (must complete first where needed)
- Code quality/polish tasks are low priority (game is playable, just needs cleanup)
- 3D board is optional enhancement (game fully playable with 2D board)

**Next Steps:**
1. Complete frontend UI stubs (Priority 2: Tasks 2.1-2.4)
2. Implement comprehensive E2E tests (Priority 1: Tasks 1.1-1.6)
3. Code cleanup (Priority 3: Tasks 3.1-3.2)
4. Consider 3D board enhancement (Priority 4: Task 4.1)

This plan is complete and actionable as of 2026-01-20.
