package handlers

import (
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
)

type Handler struct {
	db  *sqlx.DB
	hub *hub.Hub
}

func NewHandler(db *sqlx.DB, hub *hub.Hub) *Handler {
	return &Handler{
		db:  db,
		hub: hub,
	}
}

// --- HTTP Handlers ---

func (h *Handler) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) HandleGameRoutes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleGrantResources(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) handleClientMessage(client *hub.Client, payload []byte) {
	// stub: does nothing
}

func (h *Handler) handleSetTurnPhase(client *hub.Client, payload []byte) {
	// stub: does nothing
}

func (h *Handler) HandleGrantDevCard(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleForceDiceRoll(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleSetGameState(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// handleMoveRobber processes a MoveRobberMessage from a client
func (h *Handler) handleMoveRobber(client *hub.Client, payload []byte) {
	// Defensive: check args
	if client == nil || client.GameID == "" || len(payload) == 0 {
		return
	}
	// Retrieve game state from DB
	var stateJSON string
	err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)
	if err != nil || stateJSON == "" {
		return
	}
	var state catanv1.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		return
	}
	// Unmarshal move robber message
	var msg catanv1.MoveRobberMessage
	if err := protojson.Unmarshal(payload, &msg); err != nil {
		return
	}
	// Apply robber move or steal
	if msg.VictimId != nil && *msg.VictimId != "" {
		// Steal action
		_, _ = game.StealFromPlayer(&state, client.PlayerID, *msg.VictimId)
	} else if msg.Hex != nil {
		// Move robber
		_ = game.MoveRobber(&state, client.PlayerID, msg.Hex)
	} else {
		return
	}
	// Persist updated state
	newStateJSON, err := protojson.Marshal(&state)
	if err != nil {
		return
	}
	_, _ = h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(newStateJSON), client.GameID)
	// Broadcast updated state
	h.broadcastGameStatePersonalized(client.GameID, &state)
}
