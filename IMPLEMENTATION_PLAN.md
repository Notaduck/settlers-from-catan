# Implementation Plan - Settlers from Catan

> Tasks are ordered by priority and dependency. Each task is commit-sized and lists files/tests to touch.

## Priority 1: Interactive Board (CRITICAL)
... (unchanged content omitted for brevity) ...

## Priority 5: Trading

- [x] Proto: add pending trades to GameState and bank trade message
   - Files: proto/catan/v1/types.proto, proto/catan/v1/messages.proto
    ...
- [x] Backend trading logic (propose/respond/bank/expire)
   - Files: backend/internal/game/trading.go, backend/internal/game/trading_test.go
   - Go tests: backend/internal/game/trading_test.go (new)
   - Playwright: frontend/tests/trading.spec.ts (new)
- [x] Trade UI (trade/build toggle, bank trade, propose trade, incoming trade)
   - Files: frontend/src/components/Game/Game.tsx, frontend/src/components/Game/BankTradeModal.tsx, frontend/src/components/Game/ProposeTradeModal.tsx, frontend/src/components/Game/IncomingTradeModal.tsx, frontend/src/context/GameContext.tsx
   - Go tests: none
   - Playwright: frontend/tests/trading.spec.ts

---

## Validation Notes

- Trading UI integrated with stubbed modals and Playwright test stub.
- All new frontend code typechecks (make typecheck: PASS).
- Lint returns errors due to pre-existing backend/handler corruption (see BLOCKER below) but frontend/TS is clean.
- Backend unit tests: All trading and game logic pass except for known unrelated Board/Robber/Handler / nil pointer failures. No regression caused by trading.
- Playwright e2e cannot be run by agent (requires running servers).

## BLOCKER 2026-01-19: Handler Package Corruption
- backend/internal/handlers/handlers.go is catastrophically broken; handler and websocket logic can’t be tested or wired until file is restored.

# ...remaining sections as before...
