package handlers

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
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
		stateJSON, err := wsMarshal.Marshal(personalized)
		if err != nil {
			continue
		}
		envelope := serverEnvelope{
			Message: serverMessage{
				OneofKind: "gameState",
				GameState: &gameStateWire{State: stateJSON},
			},
		}
		msg, err := json.Marshal(envelope)
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
	out, ok := proto.Clone(src).(*pb.GameState)
	if !ok || out == nil {
		return nil
	}
	for i, p := range out.Players {
		if p.Id != playerID {
			p.DevCards = nil
			p.DevCardCount = 0
			p.DevCardsPurchasedTurn = nil
			out.Players[i] = p
		}
	}
	return out
}
