# Trading System

**Priority**: MEDIUM - Important for strategy but game playable without

## Overview

Settlers has three trading mechanisms:

1. **Player-to-player**: Negotiate any resource exchange
2. **Bank trade (4:1)**: Trade 4 of same resource for 1 of any
3. **Port trade (3:1 or 2:1)**: Better ratios at port locations

This spec covers player and bank trading. Ports are separate spec.

## Acceptance Criteria

### Trade Phase

- [ ] Trade phase occurs after roll, before build
- [ ] "Trade" and "Build" buttons to switch phases
- [ ] Can only trade during TRADE phase of your turn
- [ ] Can make multiple trades per turn

### Bank Trade (4:1)

- [ ] "Trade with Bank" button opens bank trade modal
- [ ] Select 4 of one resource to give
- [ ] Select 1 resource to receive
- [ ] Confirm executes trade immediately
- [ ] Resources update in player panel

### Player Trade - Proposing

- [ ] "Propose Trade" button opens trade proposal modal
- [ ] Select resources to offer (from your hand)
- [ ] Select resources to request
- [ ] Can target specific player OR broadcast to all
- [ ] Submit sends `propose_trade` message

### Player Trade - Receiving

- [ ] Incoming trade shows as modal/toast for target players
- [ ] Shows: "{Player} offers {X} for {Y}"
- [ ] Buttons: "Accept" / "Decline" / "Counter"
- [ ] Accept executes trade
- [ ] Decline dismisses (notifies proposer)
- [ ] Counter opens counter-proposal UI

### Trade Resolution

- [ ] On accept: resources transfer, both players notified
- [ ] On decline: proposer notified
- [ ] Trade offer expires at end of turn
- [ ] Only one active trade proposal at a time (per player)

### Backend Implementation Needed

- [ ] `handleProposeTrade` handler
- [ ] `handleRespondTrade` handler
- [ ] `ProposeTrade(state, from, to, offer, request)` command
- [ ] `AcceptTrade(state, tradeId, accepterId)` command
- [ ] `DeclineTrade(state, tradeId, declinerId)` command
- [ ] Validate players have resources before transfer

### Data Attributes (for Playwright)

- [ ] `data-cy="trade-phase-btn"` to enter trade phase
- [ ] `data-cy="bank-trade-btn"` open bank trade
- [ ] `data-cy="propose-trade-btn"` open proposal
- [ ] `data-cy="trade-offer-{resource}"` offer selection
- [ ] `data-cy="trade-request-{resource}"` request selection
- [ ] `data-cy="incoming-trade-modal"` for received offers
- [ ] `data-cy="accept-trade-btn"` / `data-cy="decline-trade-btn"`

## Required Go Unit Tests

File: `backend/internal/game/trading_test.go` (new)

```go
// Test: ProposeTrade creates pending trade offer
// Test: ProposeTrade fails if not your turn
// Test: ProposeTrade fails if not in trade phase
// Test: ProposeTrade fails if you lack offered resources
// Test: AcceptTrade transfers resources correctly
// Test: AcceptTrade fails if accepter lacks requested resources
// Test: AcceptTrade fails if offer expired
// Test: DeclineTrade removes pending offer
// Test: BankTrade requires 4 of same resource
// Test: BankTrade fails if insufficient resources
```

## Required Playwright E2E Tests

File: `frontend/tests/trading.spec.ts`

```typescript
// Test: Bank trade 4:1 works
// Test: Cannot bank trade without 4 resources
// Test: Player can propose trade to another
// Test: Trade recipient sees offer modal
// Test: Accepting trade transfers resources
// Test: Declining trade notifies proposer
// Test: Cannot trade outside trade phase
```

## Implementation Notes

- Trade offers stored in `gameState.pendingTrades` array
- Each trade has unique ID for accept/decline reference
- Consider: counter-offers as new proposals (simpler than nested)
- UI should prevent proposing trades you can't fulfill
