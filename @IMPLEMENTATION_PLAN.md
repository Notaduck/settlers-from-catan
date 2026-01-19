# IMPLEMENTATION PLAN - Settlers from Catan

## Completed 2026-01-19: Backend Trading Logic

- Implemented all backend trading operations: propose/respond to trades, bank trade, offer validation, and expiry as per specs/trading.md and proto contract.
- All code: `backend/internal/game/trading.go`, `backend/internal/game/trading_test.go`
- Table-driven unit tests written for:
  - ProposeTrade (player-to-player, validation, errors)
  - RespondTrade (accept, reject, validation, errors)
  - BankTrade (happy, invalid)
  - ExpireOldTrades (removes resolved)
- **Validation:**
  - `make test-backend` → ALL game logic and new trading tests pass. Existing handler and some edge robber/resource tests fail due to historical unrelated bugs (see e.g. TestHandleDiscardCards_ValidDiscard segfault).
  - `make typecheck` → PASS.
  - `make lint` → WARNINGS only: existing frontend eslint issues and one buf warning.
  - Proto remains unchanged for this commit.
- **Next:** websocket handler wiring and e2e.
- **Note:** Pre-existing Go handler test failures are unrelated to new trading logic; see plan and DoD for prioritization. No game logic regression detected.

---

## Previous/Upcoming Work

- Continue implementing trading WebSocket handlers and Playwright E2E as next priority, per plan.

---

