package handlers

// (rest of the existing code from handlers.go follows, as previously written)

// --- Trading handlers ---
// handleProposeTrade handles a player's trade proposal (player ↔ player).
func (h *Handler) handleProposeTrade(client *hub.Client, payload []byte) {
	var req catanv1.ProposeTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse propose trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	// Call game logic
	err := game.ProposeTrade(&state, client.PlayerID, req.GetTargetId(), req.GetOffering(), req.GetRequesting())
	if err != nil {
		h.sendError(client, "TRADE_PROPOSE_ERROR", err.Error())
		return
	}
	// Save updated state
	updatedJSON, _ := protojson.Marshal(&state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)
	// Broadcast new trade
	msg := catanv1.ServerMessage{
		Message: &catanv1.ServerMessage_TradeProposed{
			TradeProposed: &catanv1.TradeProposedPayload{
				Trade: state.GetPendingTrades()[len(state.GetPendingTrades())-1], // last added trade
			},
		},
	}
	data, _ := protojson.Marshal(&msg)
	h.hub.BroadcastToGame(client.GameID, data)
	// Always broadcast updated game state for robust sync
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleRespondTrade handles a player's response (accept/decline) to a trade.
func (h *Handler) handleRespondTrade(client *hub.Client, payload []byte) {
	var req catanv1.RespondTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse respond trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}
	var err error
	if req.Accept {
		err = game.AcceptTrade(&state, req.GetTradeId(), client.PlayerID)
	} else {
		err = game.DeclineTrade(&state, req.GetTradeId(), client.PlayerID)
	}
	if err != nil {
		h.sendError(client, "TRADE_RESPOND_ERROR", err.Error())
		return
	}
	updatedJSON, _ := protojson.Marshal(&state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)
	msg := catanv1.ServerMessage{
		Message: &catanv1.ServerMessage_TradeResolved{
			TradeResolved: &catanv1.TradeResolvedPayload{
				TradeId:    req.GetTradeId(),
				Accepted:   req.GetAccept(),
				AcceptedBy: &client.PlayerID,
			},
		},
	}
	data, _ := protojson.Marshal(&msg)
	h.hub.BroadcastToGame(client.GameID, data)
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleBankTrade handles a player's trade with the bank (4:1).
func (h *Handler) handleBankTrade(client *hub.Client, payload []byte) {
	var req catanv1.BankTradeMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse bank trade request")
		return
	}

	var stateJSON string
	var state game.GameState
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}
	if state.Status == game.GameStatusFinished {
		h.sendError(client, "GAME_OVER", "Game already finished")
		return
	}

	err := game.BankTrade(&state, client.PlayerID, req.GetOffering(), req.GetResourceRequested())
	if err != nil {
		h.sendError(client, "BANK_TRADE_ERROR", err.Error())
		return
	}
	updatedJSON, _ := protojson.Marshal(&state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)
	// No special payload; just broadcast updated state
	h.broadcastGameStateProto(client.GameID, &state)
}
