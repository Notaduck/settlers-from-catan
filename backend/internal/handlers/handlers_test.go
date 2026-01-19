package handlers

func TestHandleProposeTrade_BroadcastsNewTrade(t *testing.T) {
	h := &Handler{} // TODO: replace with real setup
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{
			{Id: "p1"}, {Id: "p2"},
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "trade-game-id")
	client := &hub.Client{PlayerID: "p1", GameID: "trade-game-id"}
	payload, _ := protojson.Marshal(&catanv1.ProposeTradeMessage{
		TargetId:   protoString("p2"),
		Offering:   &catanv1.ResourceCount{Wood: 2},
		Requesting: &catanv1.ResourceCount{Brick: 1},
	})
	h.handleProposeTrade(client, payload)
	// Add assertions for DB and outgoing messages (TODO)
}

func TestHandleRespondTrade_AcceptUpdatesState(t *testing.T) {
	h := &Handler{} // TODO: replace with real setup
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{
			{Id: "p1"}, {Id: "p2"},
		},
		PendingTrades: []*catanv1.TradeOffer{
			{TradeId: "t1", FromId: "p1", ToId: protoString("p2"), Offering: &catanv1.ResourceCount{Wood: 2}, Requesting: &catanv1.ResourceCount{Brick: 1}, Status: catanv1.TradeStatus_TRADE_STATUS_PENDING},
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "trade-game-id")
	client := &hub.Client{PlayerID: "p2", GameID: "trade-game-id"}
	payload, _ := protojson.Marshal(&catanv1.RespondTradeMessage{
		TradeId: "t1", Accept: true,
	})
	h.handleRespondTrade(client, payload)
	// Add assertions for DB and outgoing messages (TODO)
}

func protoString(s string) *string { return &s }
