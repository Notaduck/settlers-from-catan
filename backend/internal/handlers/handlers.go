package handlers

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
	"math/rand"
	"net/http"
	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
	"strings"
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
	type apiRequest struct {
		PlayerName string `json:"playerName"`
	}
	var req apiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerName == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	gameID := uuid.New().String()
	playerID := uuid.New().String()
	sessionToken := uuid.New().String()

	// Generate a random 6-char uppercase join code
	code := strings.ToUpper(randomCode(6))

	// Create game state
	state := game.NewGameState(gameID, code, []string{req.PlayerName}, []string{playerID})

	// Marshal game state to JSON (protojson)
	stateJSON, err := protojson.Marshal(state)
	if err != nil {
		http.Error(w, "failed to marshal game state", http.StatusInternalServerError)
		return
	}

	// Persist game row
	_, err = h.db.Exec(`INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)`,
		gameID, code, string(stateJSON), "waiting")
	if err != nil {
		http.Error(w, "failed to persist game", http.StatusInternalServerError)
		return
	}

	// Persist player row
	name := req.PlayerName
	color := int(state.Players[0].Color)
	_, err = h.db.Exec(`INSERT INTO players (id, game_id, name, color, session_token, is_host, connected) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		playerID, gameID, name, color, sessionToken, 1, 0)
	if err != nil {
		http.Error(w, "failed to persist player", http.StatusInternalServerError)
		return
	}

	resp := &catanv1.CreateGameResponse{
		GameId:       gameID,
		Code:         code,
		SessionToken: sessionToken,
		PlayerId:     playerID,
	}
	w.Header().Set("Content-Type", "application/json")
	resJSON, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(resJSON)
}

// HandleJoinGame allows a player to join an existing game by code.
func (h *Handler) HandleJoinGame(w http.ResponseWriter, r *http.Request) {
	type apiRequest struct {
		PlayerName string `json:"playerName"`
	}
	var req apiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlayerName == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Extract code from URL (expect /api/games/{code}/join)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	code := strings.ToUpper(parts[3])

	// Look up game by code
	var (
		gameID    string
		stateJSON string
	)
	err := h.db.Get(&gameID, "SELECT id FROM games WHERE code = ?", code)
	if err != nil || gameID == "" {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}
	err = h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID)
	if err != nil || stateJSON == "" {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	var state catanv1.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		http.Error(w, "failed to parse game state", http.StatusInternalServerError)
		return
	}

	// Assign first unused color
	allColors := []catanv1.PlayerColor{
		catanv1.PlayerColor_PLAYER_COLOR_RED,
		catanv1.PlayerColor_PLAYER_COLOR_BLUE,
		catanv1.PlayerColor_PLAYER_COLOR_GREEN,
		catanv1.PlayerColor_PLAYER_COLOR_ORANGE,
	}
	usedColors := map[catanv1.PlayerColor]bool{}
	for _, p := range state.Players {
		usedColors[p.Color] = true
	}
	var color catanv1.PlayerColor
	found := false
	for _, c := range allColors {
		if !usedColors[c] {
			color = c
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "game full", http.StatusBadRequest)
		return
	}

	playerID := uuid.New().String()
	sessionToken := uuid.New().String()

	// Add player to state
	newPlayer := &catanv1.PlayerState{
		Id:            playerID,
		Name:          req.PlayerName,
		Color:         color,
		Resources:     &catanv1.ResourceCount{},
		DevCardCount:  0,
		KnightsPlayed: 0,
		VictoryPoints: 0,
		Connected:     false,
		IsReady:       false,
		IsHost:        false,
	}
	state.Players = append(state.Players, newPlayer)

	// Persist updated state
	updatedStateJSON, err := protojson.Marshal(&state)
	if err != nil {
		http.Error(w, "failed to marshal state", http.StatusInternalServerError)
		return
	}
	_, err = h.db.Exec(`UPDATE games SET state = ? WHERE id = ?`, string(updatedStateJSON), gameID)
	if err != nil {
		http.Error(w, "failed to update game", http.StatusInternalServerError)
		return
	}

	// Add player to players table
	_, err = h.db.Exec(
		`INSERT INTO players (id, game_id, name, color, session_token, is_host, connected) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		playerID, gameID, req.PlayerName, int(color), sessionToken, 0, 0)
	if err != nil {
		http.Error(w, "failed to insert player", http.StatusInternalServerError)
		return
	}

	// Construct response, listing all players
	players := make([]*catanv1.PlayerInfo, len(state.Players))
	for i, p := range state.Players {
		players[i] = &catanv1.PlayerInfo{
			Id:    p.Id,
			Name:  p.Name,
			Color: p.Color,
		}
	}
	resp := &catanv1.JoinGameResponse{
		GameId:       gameID,
		SessionToken: sessionToken,
		PlayerId:     playerID,
		Players:      players,
	}

	w.Header().Set("Content-Type", "application/json")
	resJSON, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(resJSON)
}

func randomCode(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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
