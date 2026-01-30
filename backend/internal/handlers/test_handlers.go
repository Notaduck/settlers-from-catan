package handlers

import (
	"encoding/json"
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Helper to personalize state for each player during test broadcasts
func (h *Handler) broadcastGameStatePersonalized(gameID string, state *pb.GameState) {
	clients := h.hub.GetClientsForGame(gameID)
	if len(clients) == 0 || state == nil {
		return
	}
	for _, client := range clients {
		if client == nil {
			continue
		}
		personalized := redactedGameStateForPlayer(state, client.PlayerID)
		m := map[string]interface{}{
			"type":    "game_state",
			"payload": personalized,
		}
		msg, err := json.Marshal(m)
		if err == nil {
			client.Send(msg)
		}
	}
}

// Helper for deep copy and dev card redaction in tests
func redactedGameStateForPlayer(src *pb.GameState, playerID string) *pb.GameState {
	if src == nil {
		return nil
	}
	b, _ := json.Marshal(src)
	var out pb.GameState
	_ = json.Unmarshal(b, &out)
	for i, p := range out.Players {
		if p.Id != playerID {
			p.DevCards = nil
			p.DevCardCount = 0
			p.DevCardsPurchasedTurn = nil
			out.Players[i] = p
		}
	}
	return &out
}
