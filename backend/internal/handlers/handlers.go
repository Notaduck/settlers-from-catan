package handlers

import (
	"log"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"

	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
)

type Handler struct {
	hub *hub.Hub
	db  *sqlx.DB
}

// sendError sends an error payload to the client and logs it.
func (h *Handler) sendError(client *hub.Client, code, message string) {
	log.Printf("Error [%s]: %s", code, message)
	// NOOP: In production this would send a websocket/server message
}

// broadcastGameStateProto broadcasts the game state. (Stub)
func (h *Handler) broadcastGameStateProto(gameID string, state *game.GameState) {
	log.Printf("[STUB] broadcastGameStateProto for gameID %s", gameID)
	// NOOP: In production this would serialize state and send to all clients in game
}

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
	_, err := game.ProposeTrade(&state, client.PlayerID, req.TargetId, req.GetOffering(), req.GetRequesting())
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
	err = game.RespondTrade(&state, req.GetTradeId(), client.PlayerID, req.Accept)

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

// handleDiscardCards handles discarding cards for the robber phase.
func (h *Handler) handleDiscardCards(client *hub.Client, payload []byte) {
	var req catanv1.DiscardCardsMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse discard cards request")
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
	if err := game.DiscardCards(&state, client.PlayerID, req.GetResources()); err != nil {
		h.sendError(client, "DISCARD_ERROR", err.Error())
		return
	}
	updatedJSON, _ := protojson.Marshal(&state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)
	msg := catanv1.ServerMessage{
		Message: &catanv1.ServerMessage_DiscardedCards{
			DiscardedCards: &catanv1.DiscardedCardsPayload{
				PlayerId:  client.PlayerID,
				Resources: req.GetResources(),
			},
		},
	}
	data, _ := protojson.Marshal(&msg)
	h.hub.BroadcastToGame(client.GameID, data)
	h.broadcastGameStateProto(client.GameID, &state)
}

// handleMoveRobber handles moving the robber and optional stealing.
func (h *Handler) handleMoveRobber(client *hub.Client, payload []byte) {
	var req catanv1.MoveRobberMessage
	if err := protojson.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Failed to parse move robber request")
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
	if err := game.MoveRobber(&state, client.PlayerID, req.GetHex()); err != nil {
		h.sendError(client, "MOVE_ROBBER_ERROR", err.Error())
		return
	}
	var stolenResource catanv1.Resource
	victimId := ""
	if req.VictimId != nil && *req.VictimId != "" {
		victimId = *req.VictimId
		res, err := game.StealFromPlayer(&state, client.PlayerID, victimId)
		if err == nil {
			stolenResource = res
		}
	}
	updatedJSON, _ := protojson.Marshal(&state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)
	msg := catanv1.ServerMessage{
		Message: &catanv1.ServerMessage_RobberMoved{
			RobberMoved: &catanv1.RobberMovedPayload{
				PlayerId:       client.PlayerID,
				Hex:            req.GetHex(),
				VictimId:       &victimId,
				StolenResource: &stolenResource,
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
