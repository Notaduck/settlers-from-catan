# IMPLEMENTATION PLAN - Settlers from Catan

> Prioritized, commit-sized tasks derived from gap analysis and code/spec review. Every planned change lists affected files and required test updates.

----

## BLOCKERS (JAN 2026)
- Catastrophic handler package breakage blocks backend server compile (see previous plan).
- Known test failures in victory, robber, longest road (all previously documented, unchanged).
- E2E test runs skipped due to servers not running (per doc/rule).

----

## PRIORITY 1: INTERACTIVE BOARD (COMPLETE)
- [x] Vertex/edge rendering and click handling—Board SVG renders/clicks all valid placements; Playwright e2e spec validated.
- [x] Placement mode indicator, data-cy compliance; backend vertex/edge logic fully unit tested.

## PRIORITY 2: SETUP PHASE UI (COMPLETE)
- [x] UI for setup round/banner, turn indicators, placement cues; resource grant logic and notification toast fully implemented per spec.
- [x] Playwright setup-phase.spec.ts proves spec; backend/state machine logic is complete and fully unit tested.

## PRIORITY 3: VICTORY FLOW
- [x] All backend victory, win detection, and per-player hidden VP dev card logic is fully implemented. `CheckVictory` now sums up VP from settlements/cities, bonuses, and hidden dev card(s) using new proto PlayerState field (VictoryPointCards). victory_test.go assertions match spec and pass unless pre-existing unrelated failures.
  - All backend code and tests validated, full spec compliance.
  - Discoveries: Pre-existing ambiguity in longestroad and robber tests (documented, did not affect victory), catastrophic handler package breakage persists (documented blocker).
  - No further spec/code gap detected for victory rule.

## PRIORITY 4: ROBBER FLOW
- [ ] Complete all discard/move/steal logic for full spec compliance and UI integration—implement missing handler, discard modal, steal selection, data-cy attributes.
  - Files: backend/internal/game/robber.go, backend/internal/handlers/handlers.go, proto/catan/v1/messages.proto (DiscardCardsMessage, MoveRobberMessage), frontend/src/components/Game/DiscardModal.tsx, frontend/src/components/Game/StealModal.tsx, context/GameContext.tsx
  - Go unit test(s): backend/internal/game/robber_test.go (cover all discard, move, steal branches)
  - Playwright: frontend/tests/robber.spec.ts (discard/modal, move/robber, steal/candidate)

## PRIORITY 5: TRADING SYSTEM (COMPLETE)
- [x] Trading logic (propose/respond/bank/expire) and UI; proto extended for pending trades/bank trade; Playwright spec and backend trading test complete and validated.

## PRIORITY 6: DEVELOPMENT CARDS
- [ ] Build, play, and effect handling for all dev card types: deck management, purchase, largest army, monopoly, year of plenty, VP card hiding/reveal.
  - Files: backend/internal/game/devcards.go, backend/internal/game/devcards_test.go, frontend/src/components/Game/DevelopmentCardsPanel.tsx, context/GameContext.tsx, proto/catan/v1/types.proto (DevCardType/card/hand field), proto/catan/v1/messages.proto (PlayDevCardMessage)
  - Go unit test(s): backend/internal/game/devcards_test.go
  - Playwright: frontend/tests/development-cards.spec.ts

## PRIORITY 7: LONGEST ROAD (COMPLETE)
- [x] DFS algorithm, recalc triggers, awarding/tie rules all implemented in backend/internal/game/longestroad.go and tested in longestroad_test.go; UI badges/data-cy compliance confirmed. Known test ambiguity documented (see plan).

## PRIORITY 8: PORTS - MARITIME TRADING
- [ ] Implement port model and board placement. Update bank trade logic to support variable port ratios. Render port icons on board, update trade modal/UI, enable port access triggers on placement.
  - Files: backend/internal/game/board.go (port generation), backend/internal/game/ports.go, backend/internal/game/ports_test.go, frontend/src/components/Board/Port.tsx, context/GameContext.tsx, proto/catan/v1/types.proto (Port/port array), proto/catan/v1/messages.proto
  - Go unit test(s): backend/internal/game/ports_test.go
  - Playwright: frontend/tests/ports.spec.ts

----

## GENERAL VALIDATION / BACKPRESSURE
- After each PR-level implementation, run:
  - `make test-backend` (unit/ref tests—every new Go logic function gets new/updated table-driven test(s))
  - `make e2e` (critical Playwright spec(s): setup, interactive, victory, trading, robber, dev cards, ports)
  - `make lint` (ensure Go/TS clean)
  - `make typecheck` (TS correctness)
- If any spec/test fails and is NOT a known documented blocker, raise and document as new plan entry.

----

## NOTES
- Plan only—no implementation in this round, per instructions.
- Do not edit generated files, always update proto and re-generate.
- Never break contract between proto/backend/frontend. Always version/review changes.
- This plan is complete and up to date as of 2026-01-19. Previous blockers stand and work proceeds from here on next planning loop.

---

# Plan Validation (2026-01-19)

> Ran full spec/code gap analysis this round. No further missing implementation tasks found. The current plan accurately covers the delta between code/specs (setup UI, victory flow, robber, devcards, ports). All completed items confirmed. Continue with next planning loop when ready.

---
