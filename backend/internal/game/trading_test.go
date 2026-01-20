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
	state.Board = GenerateBoard()
	// Happy path - default 4:1 ratio
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
	state.Board = GenerateBoard()
	err = BankTrade(state, "me", &catanv1.ResourceCount{Wood: 3}, catanv1.Resource_RESOURCE_BRICK)
	if err == nil {
		t.Error("Allowed bank trade with insufficient resources")
	}
}

func TestBankTradeWithGenericPort(t *testing.T) {
	state := basicGameState(makePlayer("me", &catanv1.ResourceCount{Wood: 3}))
	state.Board = GenerateBoard()

	// Find a generic port and give player access
	var genericPort *catanv1.Port
	for _, p := range state.Board.Ports {
		if p.Type == catanv1.PortType_PORT_TYPE_GENERIC {
			genericPort = p
			break
		}
	}
	if genericPort == nil {
		t.Fatal("No generic port found on board")
	}

	// Place settlement on port vertex
	for _, v := range state.Board.Vertices {
		if v.Id == genericPort.Location[0] {
			v.Building = &catanv1.Building{
				Type:    catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
				OwnerId: "me",
			}
			break
		}
	}

	// Should succeed with 3:1 ratio
	err := BankTrade(state, "me", &catanv1.ResourceCount{Wood: 3}, catanv1.Resource_RESOURCE_BRICK)
	if err != nil {
		t.Fatalf("BankTrade with generic port failed: %v", err)
	}
	if state.Players[0].Resources.Wood != 0 {
		t.Errorf("Expected 0 wood remaining, got %d", state.Players[0].Resources.Wood)
	}
	if state.Players[0].Resources.Brick != 1 {
		t.Errorf("Expected 1 brick received, got %d", state.Players[0].Resources.Brick)
	}
}

func TestBankTradeWithSpecificPort(t *testing.T) {
	state := basicGameState(makePlayer("me", &catanv1.ResourceCount{Wheat: 2}))
	state.Board = GenerateBoard()

	// Find wheat-specific port and give player access
	var wheatPort *catanv1.Port
	for _, p := range state.Board.Ports {
		if p.Type == catanv1.PortType_PORT_TYPE_SPECIFIC && p.Resource == catanv1.Resource_RESOURCE_WHEAT {
			wheatPort = p
			break
		}
	}
	if wheatPort == nil {
		t.Fatal("No wheat-specific port found on board")
	}

	// Place settlement on port vertex
	for _, v := range state.Board.Vertices {
		if v.Id == wheatPort.Location[0] {
			v.Building = &catanv1.Building{
				Type:    catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
				OwnerId: "me",
			}
			break
		}
	}

	// Should succeed with 2:1 ratio for wheat
	err := BankTrade(state, "me", &catanv1.ResourceCount{Wheat: 2}, catanv1.Resource_RESOURCE_BRICK)
	if err != nil {
		t.Fatalf("BankTrade with specific port failed: %v", err)
	}
	if state.Players[0].Resources.Wheat != 0 {
		t.Errorf("Expected 0 wheat remaining, got %d", state.Players[0].Resources.Wheat)
	}
	if state.Players[0].Resources.Brick != 1 {
		t.Errorf("Expected 1 brick received, got %d", state.Players[0].Resources.Brick)
	}
}

func TestBankTradeSpecificPortWrongResource(t *testing.T) {
	state := basicGameState(makePlayer("me", &catanv1.ResourceCount{Wood: 2}))
	state.Board = GenerateBoard()

	// Find wheat-specific port and give player access
	var wheatPort *catanv1.Port
	for _, p := range state.Board.Ports {
		if p.Type == catanv1.PortType_PORT_TYPE_SPECIFIC && p.Resource == catanv1.Resource_RESOURCE_WHEAT {
			wheatPort = p
			break
		}
	}
	if wheatPort == nil {
		t.Fatal("No wheat-specific port found on board")
	}

	// Place settlement on port vertex
	for _, v := range state.Board.Vertices {
		if v.Id == wheatPort.Location[0] {
			v.Building = &catanv1.Building{
				Type:    catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
				OwnerId: "me",
			}
			break
		}
	}

	// Should fail with 2:1 ratio for wood (wheat port doesn't help with wood)
	err := BankTrade(state, "me", &catanv1.ResourceCount{Wood: 2}, catanv1.Resource_RESOURCE_BRICK)
	if err == nil {
		t.Error("BankTrade should fail - specific port doesn't apply to different resource")
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
