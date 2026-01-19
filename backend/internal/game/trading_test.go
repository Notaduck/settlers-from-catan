package game

import (
	"settlers_from_catan/gen/proto/catan/v1"
	"testing"
)

func makePlayer(id string, rc *catanv1.ResourceCount) *catanv1.PlayerState {
	return &catanv1.PlayerState{Id: id, Resources: rc}
}

func basicGameState(players ...*catanv1.PlayerState) *catanv1.GameState {
	return &catanv1.GameState{
		Players:       players,
		CurrentTurn:   0,
		Status:        catanv1.GameStatus_GAME_STATUS_PLAYING,
		TurnPhase:     catanv1.TurnPhase_TURN_PHASE_TRADE,
		PendingTrades: []*catanv1.TradeOffer{},
	}
}

func TestProposeTrade(t *testing.T) {
	state := basicGameState(makePlayer("me", &catanv1.ResourceCount{Wood: 4}), makePlayer("them", &catanv1.ResourceCount{Sheep: 3}))
	// Happy path
	id, err := ProposeTrade(state, "me", nil, &catanv1.ResourceCount{Wood: 3}, &catanv1.ResourceCount{Sheep: 2})
	if err != nil || id == "" || len(state.PendingTrades) != 1 {
		t.Fatalf("ProposeTrade failed: %v (id: %v, trades: %d)", err, id, len(state.PendingTrades))
	}
	// Cannot propose if not your turn
	_, err = ProposeTrade(state, "them", nil, &catanv1.ResourceCount{}, &catanv1.ResourceCount{})
	if err == nil {
		t.Error("Should not allow propose trade if not your turn")
	}
	// Can't offer things you don't have
	_, err = ProposeTrade(state, "me", nil, &catanv1.ResourceCount{Brick: 2}, &catanv1.ResourceCount{})
	if err == nil {
		t.Error("Allowed offering resources you don't have")
	}
}

func TestRespondTrade_AcceptDecline(t *testing.T) {
	toID := "them"
	tradeID := "T1"
	off := &catanv1.ResourceCount{Wood: 2}
	req := &catanv1.ResourceCount{Sheep: 2}
	state := basicGameState(
		makePlayer("me", &catanv1.ResourceCount{Wood: 3, Brick: 1}),
		makePlayer(toID, &catanv1.ResourceCount{Sheep: 3, Brick: 1}),
	)
	propID, _ := ProposeTrade(state, "me", &toID, off, req)
	state.PendingTrades[0].Id = tradeID // for testing
	// Accept with correct resources
	err := RespondTrade(state, tradeID, toID, true)
	if err != nil || state.PendingTrades[0].Status != catanv1.TradeStatus_TRADE_STATUS_ACCEPTED {
		t.Fatalf("AcceptTrade failed: %v status=%v", err, state.PendingTrades[0].Status)
	}
	// Insufficient accepting
	state = basicGameState(
		makePlayer("me", &catanv1.ResourceCount{Wood: 2}),
		makePlayer(toID, &catanv1.ResourceCount{Sheep: 1}),
	)
	state.PendingTrades = append(state.PendingTrades, &catanv1.TradeOffer{
		Id: propID, ProposerId: "me", TargetId: &toID, Offering: off, Requesting: req, Status: catanv1.TradeStatus_TRADE_STATUS_PENDING,
	})
	err = RespondTrade(state, propID, toID, true)
	if err == nil {
		t.Error("AcceptTrade should fail if acceptor doesn't have requested resources")
	}
	// Decline
	state.PendingTrades[0].Status = catanv1.TradeStatus_TRADE_STATUS_PENDING
	err = RespondTrade(state, propID, toID, false)
	if err != nil || state.PendingTrades[0].Status != catanv1.TradeStatus_TRADE_STATUS_REJECTED {
		t.Errorf("DeclineTrade did not reject: err=%v status=%v", err, state.PendingTrades[0].Status)
	}
}

func TestBankTrade(t *testing.T) {
	state := basicGameState(makePlayer("me", &catanv1.ResourceCount{Wood: 4}))
	// Happy path
	err := BankTrade(state, "me", &catanv1.ResourceCount{Wood: 4}, catanv1.Resource_RESOURCE_BRICK)
	if err != nil {
		t.Fatalf("BankTrade failed: %v", err)
	}
	if got := state.Players[0].Resources.Brick; got != 1 {
		t.Errorf("BankTrade: bank did not give new resource, got %v", got)
	}
	if got := state.Players[0].Resources.Wood; got != 0 {
		t.Errorf("BankTrade: didn't deduct all Wood, got %v", got)
	}
	// Not enough for bank trade
	state = basicGameState(makePlayer("me", &catanv1.ResourceCount{Wood: 3}))
	err = BankTrade(state, "me", &catanv1.ResourceCount{Wood: 3}, catanv1.Resource_RESOURCE_BRICK)
	if err == nil {
		t.Error("Allowed bank trade with insufficient resources")
	}
}

func TestExpireOldTrades(t *testing.T) {
	pending := catanv1.TradeStatus_TRADE_STATUS_PENDING
	acc := catanv1.TradeStatus_TRADE_STATUS_ACCEPTED
	rej := catanv1.TradeStatus_TRADE_STATUS_REJECTED
	state := basicGameState(
		makePlayer("me", &catanv1.ResourceCount{Wood: 2}),
	)
	state.PendingTrades = []*catanv1.TradeOffer{
		{Id: "1", Status: pending},
		{Id: "2", Status: acc},
		{Id: "3", Status: rej},
	}
	ExpireOldTrades(state)
	if len(state.PendingTrades) != 1 || state.PendingTrades[0].Id != "1" {
		t.Errorf("ExpireOldTrades did not leave only pending: %+v", state.PendingTrades)
	}
}
