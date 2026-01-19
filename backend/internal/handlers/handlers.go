package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"

	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// JSON marshaler that emits default/zero values (important for coordinates like q=0, r=0)
var jsonMarshaler = protojson.MarshalOptions{
	EmitDefaultValues: true,
}

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db  *sqlx.DB
	hub *hub.Hub
}

// NewHandler creates a new handler
func NewHandler(db *sqlx.DB, h *hub.Hub) *Handler {
	return &Handler{
		db:  db,
		hub: h,
	}
}

// HandleWebSocket upgrades HTTP to WebSocket
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.URL.Query().Get("token")
	if sessionToken == "" {
		http.Error(w, "Missing session token", http.StatusUnauthorized)
		return
	}

	// Find player by session token
	var player struct {
		ID     string `db:"id"`
		GameID string `db:"game_id"`
		Name   string `db:"name"`
	}
	err := h.db.Get(&player, "SELECT id, game_id, name FROM players WHERE session_token = ?", sessionToken)
	if err != nil {
		http.Error(w, "Invalid session token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := hub.NewClient(h.hub, conn, player.ID, player.GameID)

	// Set up message handler
	client.OnMessage = func(message []byte) {
		h.handleGameMessage(client, message)
	}

	h.hub.Register(client)

	// Mark player as connected
	h.db.Exec("UPDATE players SET connected = 1, last_seen = CURRENT_TIMESTAMP WHERE id = ?", player.ID)

	go client.WritePump()
	go client.ReadPump()

	// Send current game state
	h.sendGameState(client)
}

func (h *Handler) handleGameMessage(client *hub.Client, message []byte) {
	// Parse protobuf-ts oneof format
	var msg struct {
		Message struct {
			OneofKind      string          `json:"oneofKind"`
			StartGame      json.RawMessage `json:"startGame,omitempty"`
			RollDice       json.RawMessage `json:"rollDice,omitempty"`
			BuildStructure json.RawMessage `json:"buildStructure,omitempty"`
			EndTurn        json.RawMessage `json:"endTurn,omitempty"`
			PlayerReady    json.RawMessage `json:"playerReady,omitempty"`
		} `json:"message"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Invalid message format: %v", err)
		h.sendError(client, "INVALID_FORMAT", "Invalid message format")
		return
	}

	log.Printf("Received message %s from player %s", msg.Message.OneofKind, client.PlayerID)

	switch msg.Message.OneofKind {
	case "startGame":
		h.handleStartGame(client)
	case "rollDice":
		h.handleRollDice(client)
	case "buildStructure":
		h.handleBuildStructure(client, msg.Message.BuildStructure)
	case "endTurn":
		h.handleEndTurn(client)
	case "playerReady":
		h.handlePlayerReady(client, msg.Message.PlayerReady)
	default:
		log.Printf("Unknown message type: %s", msg.Message.OneofKind)
		h.sendError(client, "UNKNOWN_TYPE", "Unknown message type")
	}
}

func (h *Handler) handleStartGame(client *hub.Client) {
	// Verify player is host
	var isHost bool
	h.db.Get(&isHost, "SELECT is_host FROM players WHERE id = ?", client.PlayerID)
	if !isHost {
		h.sendError(client, "NOT_HOST", "Only the host can start the game")
		return
	}

	// Get game state
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)

	var state map[string]interface{}
	json.Unmarshal([]byte(stateJSON), &state)

	// Check player count
	players, _ := state["players"].([]interface{})
	if len(players) < 2 {
		h.sendError(client, "NOT_ENOUGH_PLAYERS", "Need at least 2 players to start")
		return
	}

	// Check all players are ready
	for _, p := range players {
		player := p.(map[string]interface{})
		isReady, _ := player["isReady"].(bool)
		if !isReady {
			h.sendError(client, "PLAYERS_NOT_READY", "All players must be ready to start")
			return
		}
	}

	// Update status to setup phase
	state["status"] = "GAME_STATUS_SETUP"
	state["currentTurn"] = 0
	state["turnPhase"] = "TURN_PHASE_BUILD" // Setup phase: players place initial settlements

	updatedJSON, _ := json.Marshal(state)
	h.db.Exec("UPDATE games SET state = ?, status = ? WHERE id = ?", string(updatedJSON), "setup", client.GameID)

	// Broadcast game started
	h.broadcastGameStarted(client.GameID, state)
}

func (h *Handler) handlePlayerReady(client *hub.Client, payload json.RawMessage) {
	var req struct {
		Ready bool `json:"ready"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Invalid ready request")
		return
	}

	log.Printf("Player %s setting ready=%v for game %s", client.PlayerID, req.Ready, client.GameID)

	// Get game state
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)

	var state map[string]interface{}
	json.Unmarshal([]byte(stateJSON), &state)

	// Find and update player's ready state
	players := state["players"].([]interface{})
	for i, p := range players {
		player := p.(map[string]interface{})
		if player["id"] == client.PlayerID {
			player["isReady"] = req.Ready
			players[i] = player
			break
		}
	}
	state["players"] = players

	updatedJSON, _ := json.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)

	// Broadcast ready change to all players in game
	h.broadcastPlayerReadyChanged(client.GameID, client.PlayerID, req.Ready)

	// Also send updated game state so all clients have the latest
	h.broadcastGameState(client.GameID, state)
}

func (h *Handler) handleRollDice(client *hub.Client) {
	// Get game state as protobuf
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)

	var state game.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}

	// Perform dice roll (includes validation and resource distribution)
	result, err := game.PerformDiceRoll(&state, client.PlayerID)
	if err != nil {
		switch err {
		case game.ErrNotYourTurn:
			h.sendError(client, "NOT_YOUR_TURN", "It's not your turn")
		case game.ErrWrongPhase:
			h.sendError(client, "WRONG_PHASE", "You can only roll dice during the roll phase")
		default:
			h.sendError(client, "ROLL_ERROR", err.Error())
		}
		return
	}

	// Save updated state
	updatedJSON, err := protojson.Marshal(&state)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to marshal game state")
		return
	}
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)

	// Broadcast dice rolled with resource distribution info
	h.broadcastDiceRolledWithResources(client.GameID, client.PlayerID, result)

	// Also broadcast full game state
	var stateMap map[string]interface{}
	json.Unmarshal(updatedJSON, &stateMap)
	h.broadcastGameState(client.GameID, stateMap)
}

func (h *Handler) handleBuildStructure(client *hub.Client, payload json.RawMessage) {
	var req struct {
		StructureType string `json:"structureType"`
		Location      string `json:"location"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "INVALID_PAYLOAD", "Invalid build request")
		return
	}

	// Get game state as protobuf
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)

	var state game.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to parse game state")
		return
	}

	var err error
	var broadcastType string

	// Handle based on structure type and game status
	switch req.StructureType {
	case "STRUCTURE_TYPE_SETTLEMENT":
		if state.Status == game.GameStatusSetup {
			err = game.PlaceSetupSettlement(&state, client.PlayerID, req.Location)
			broadcastType = "building"
		} else {
			err = game.PlaceSettlement(&state, client.PlayerID, req.Location)
			broadcastType = "building"
		}
	case "STRUCTURE_TYPE_CITY":
		err = game.PlaceCity(&state, client.PlayerID, req.Location)
		broadcastType = "building"
	case "STRUCTURE_TYPE_ROAD":
		if state.Status == game.GameStatusSetup {
			err = game.PlaceSetupRoad(&state, client.PlayerID, req.Location)
			broadcastType = "road"
		} else {
			err = game.PlaceRoad(&state, client.PlayerID, req.Location)
			broadcastType = "road"
		}
	default:
		h.sendError(client, "INVALID_STRUCTURE", "Unknown structure type")
		return
	}

	if err != nil {
		// Convert game errors to client errors
		errorCode := "BUILD_ERROR"
		switch err {
		case game.ErrNotYourTurn:
			errorCode = "NOT_YOUR_TURN"
		case game.ErrWrongPhase:
			errorCode = "WRONG_PHASE"
		case game.ErrInvalidVertex:
			errorCode = "INVALID_LOCATION"
		case game.ErrInvalidEdge:
			errorCode = "INVALID_LOCATION"
		case game.ErrVertexOccupied, game.ErrEdgeOccupied:
			errorCode = "LOCATION_OCCUPIED"
		case game.ErrDistanceRule:
			errorCode = "DISTANCE_RULE"
		case game.ErrMustConnectToOwned:
			errorCode = "MUST_CONNECT"
		case game.ErrInsufficientResources:
			errorCode = "INSUFFICIENT_RESOURCES"
		case game.ErrCannotUpgrade:
			errorCode = "CANNOT_UPGRADE"
		case game.ErrMustPlaceSettlementFirst:
			errorCode = "MUST_PLACE_SETTLEMENT_FIRST"
		case game.ErrRoadMustConnectToSetup:
			errorCode = "ROAD_MUST_CONNECT"
		case game.ErrMaxSettlementsReached:
			errorCode = "MAX_SETTLEMENTS"
		case game.ErrMaxRoadsReached:
			errorCode = "MAX_ROADS"
		}
		h.sendError(client, errorCode, err.Error())
		return
	}

	// Save updated state
	updatedJSON, err := protojson.Marshal(&state)
	if err != nil {
		h.sendError(client, "INTERNAL_ERROR", "Failed to marshal game state")
		return
	}
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)

	// Broadcast the placement
	if broadcastType == "building" {
		h.broadcastBuildingPlaced(client.GameID, client.PlayerID, req.StructureType, req.Location)
	} else {
		h.broadcastRoadPlaced(client.GameID, client.PlayerID, req.Location)
	}

	// Also broadcast full game state (important for setup phase turn changes)
	var stateMap map[string]interface{}
	json.Unmarshal(updatedJSON, &stateMap)
	h.broadcastGameState(client.GameID, stateMap)
}

func (h *Handler) handleEndTurn(client *hub.Client) {
	// Get game state
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", client.GameID)

	var state map[string]interface{}
	json.Unmarshal([]byte(stateJSON), &state)

	// Verify it's this player's turn
	currentTurn := int(state["currentTurn"].(float64))
	players := state["players"].([]interface{})
	currentPlayer := players[currentTurn].(map[string]interface{})

	if currentPlayer["id"] != client.PlayerID {
		h.sendError(client, "NOT_YOUR_TURN", "It's not your turn")
		return
	}

	// Move to next player
	nextTurn := (currentTurn + 1) % len(players)
	state["currentTurn"] = nextTurn
	state["turnPhase"] = "TURN_PHASE_ROLL"
	state["dice"] = []int{0, 0}

	updatedJSON, _ := json.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), client.GameID)

	// Broadcast updated game state
	h.broadcastGameState(client.GameID, state)
}

func (h *Handler) sendError(client *hub.Client, code, message string) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "error",
			"error": map[string]interface{}{
				"code":    code,
				"message": message,
			},
		},
	}
	data, _ := json.Marshal(msg)
	client.Send(data)
}

func (h *Handler) broadcastGameStarted(gameID string, state map[string]interface{}) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "gameStarted",
			"gameStarted": map[string]interface{}{
				"state": state,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastDiceRolled(gameID, playerID string, die1, die2, total int) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "diceRolled",
			"diceRolled": map[string]interface{}{
				"playerId": playerID,
				"values":   []int{die1, die2},
				"total":    total,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastGameState(gameID string, state map[string]interface{}) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "gameState",
			"gameState": map[string]interface{}{
				"state": state,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastPlayerReadyChanged(gameID, playerID string, isReady bool) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "playerReadyChanged",
			"playerReadyChanged": map[string]interface{}{
				"playerId": playerID,
				"isReady":  isReady,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastBuildingPlaced(gameID, playerID, buildingType, vertexID string) {
	// Convert structure type to building type
	bType := "BUILDING_TYPE_SETTLEMENT"
	if buildingType == "STRUCTURE_TYPE_CITY" {
		bType = "BUILDING_TYPE_CITY"
	}

	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "buildingPlaced",
			"buildingPlaced": map[string]interface{}{
				"playerId":     playerID,
				"buildingType": bType,
				"vertexId":     vertexID,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastDiceRolledWithResources(gameID, playerID string, result *game.DiceRollResult) {
	// Build resource distribution list
	var resourcesDistributed []map[string]interface{}
	for pID, resources := range result.ResourcesGained {
		resourcesDistributed = append(resourcesDistributed, map[string]interface{}{
			"playerId": pID,
			"resources": map[string]interface{}{
				"wood":  resources.Wood,
				"brick": resources.Brick,
				"sheep": resources.Sheep,
				"wheat": resources.Wheat,
				"ore":   resources.Ore,
			},
		})
	}

	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "diceRolled",
			"diceRolled": map[string]interface{}{
				"playerId":             playerID,
				"values":               []int{result.Die1, result.Die2},
				"resourcesDistributed": resourcesDistributed,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) broadcastRoadPlaced(gameID, playerID, edgeID string) {
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "roadPlaced",
			"roadPlaced": map[string]interface{}{
				"playerId": playerID,
				"edgeId":   edgeID,
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func randInt(min, max int) int {
	// Simple random using crypto/rand
	var b [1]byte
	rand.Read(b[:])
	return min + int(b[0])%(max-min)
}

func (h *Handler) sendGameState(client *hub.Client) {
	var gameState string
	err := h.db.Get(&gameState, "SELECT state FROM games WHERE id = ?", client.GameID)
	if err != nil {
		log.Printf("Failed to get game state: %v", err)
		return
	}

	// Wrap in server message format
	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "gameState",
			"gameState": map[string]interface{}{
				"state": json.RawMessage(gameState),
			},
		},
	}
	data, _ := json.Marshal(msg)
	client.Send(data)
}

// generateCode creates a random 6-character uppercase code
func generateCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return strings.ToUpper(hex.EncodeToString(b))
}

// generateSessionToken creates a secure session token
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HandleCreateGame handles POST /api/games
func (h *Handler) HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerName string `json:"playerName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" || len(req.PlayerName) > 20 {
		http.Error(w, "Player name must be 1-20 characters", http.StatusBadRequest)
		return
	}

	gameID := uuid.New().String()
	code := generateCode()
	playerID := uuid.New().String()
	sessionToken := generateSessionToken()

	// Create initial game state with just this player
	gameState := game.NewGameState(gameID, code, []string{req.PlayerName}, []string{playerID})
	stateJSON, err := jsonMarshaler.Marshal(gameState)
	if err != nil {
		log.Printf("Failed to marshal game state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert game
	_, err = h.db.Exec(
		"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
		gameID, code, string(stateJSON), "waiting",
	)
	if err != nil {
		log.Printf("Failed to create game: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert player
	_, err = h.db.Exec(
		"INSERT INTO players (id, game_id, name, session_token, color, is_host) VALUES (?, ?, ?, ?, ?, ?)",
		playerID, gameID, req.PlayerName, sessionToken, "red", true,
	)
	if err != nil {
		log.Printf("Failed to create player: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"gameId":       gameID,
		"code":         code,
		"sessionToken": sessionToken,
		"playerId":     playerID,
	})
}

// HandleGameRoutes handles /api/games/:code routes
func (h *Handler) HandleGameRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/games/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		http.Error(w, "Invalid route", http.StatusBadRequest)
		return
	}

	code := parts[0]

	if len(parts) > 1 && parts[1] == "join" {
		h.handleJoinGame(w, r, code)
		return
	}

	h.handleGetGame(w, r, code)
}

func (h *Handler) handleJoinGame(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PlayerName string `json:"playerName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerName == "" || len(req.PlayerName) > 20 {
		http.Error(w, "Player name must be 1-20 characters", http.StatusBadRequest)
		return
	}

	// Find game by code
	var gameRow struct {
		ID     string `db:"id"`
		State  string `db:"state"`
		Status string `db:"status"`
	}
	err := h.db.Get(&gameRow, "SELECT id, state, status FROM games WHERE code = ?", strings.ToUpper(code))
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if gameRow.Status != "waiting" {
		http.Error(w, "Game already started", http.StatusBadRequest)
		return
	}

	// Count existing players
	var playerCount int
	h.db.Get(&playerCount, "SELECT COUNT(*) FROM players WHERE game_id = ?", gameRow.ID)
	if playerCount >= 4 {
		http.Error(w, "Game is full", http.StatusBadRequest)
		return
	}

	// Get existing players
	var existingPlayers []struct {
		ID    string `db:"id"`
		Name  string `db:"name"`
		Color string `db:"color"`
	}
	h.db.Select(&existingPlayers, "SELECT id, name, color FROM players WHERE game_id = ?", gameRow.ID)

	// Assign color
	colors := []string{"red", "blue", "green", "orange"}
	usedColors := make(map[string]bool)
	for _, p := range existingPlayers {
		usedColors[p.Color] = true
	}
	var newColor string
	for _, c := range colors {
		if !usedColors[c] {
			newColor = c
			break
		}
	}

	playerID := uuid.New().String()
	sessionToken := generateSessionToken()

	// Insert new player
	_, err = h.db.Exec(
		"INSERT INTO players (id, game_id, name, session_token, color, is_host) VALUES (?, ?, ?, ?, ?, ?)",
		playerID, gameRow.ID, req.PlayerName, sessionToken, newColor, false,
	)
	if err != nil {
		log.Printf("Failed to create player: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update game state with new player
	h.updateGameStateWithPlayer(gameRow.ID, playerID, req.PlayerName, newColor)

	// Build response with all players
	var allPlayers []struct {
		ID     string `db:"id"`
		Name   string `db:"name"`
		Color  string `db:"color"`
		IsHost bool   `db:"is_host"`
	}
	h.db.Select(&allPlayers, "SELECT id, name, color, is_host FROM players WHERE game_id = ?", gameRow.ID)

	players := make([]map[string]interface{}, 0, len(allPlayers))
	for _, p := range allPlayers {
		players = append(players, map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"color":  p.Color,
			"isHost": p.IsHost,
		})
	}

	// Broadcast player joined to existing connections
	h.broadcastPlayerJoined(gameRow.ID, playerID, req.PlayerName, newColor)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"gameId":       gameRow.ID,
		"sessionToken": sessionToken,
		"playerId":     playerID,
		"players":      players,
	})
}

func (h *Handler) updateGameStateWithPlayer(gameID, playerID, playerName, color string) {
	// This updates the game state JSON to include the new player
	// In a production system, you'd want to be more careful about concurrent updates
	var stateJSON string
	h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID)

	var state map[string]interface{}
	json.Unmarshal([]byte(stateJSON), &state)

	players, ok := state["players"].([]interface{})
	if !ok {
		players = []interface{}{}
	}

	colorMap := map[string]string{
		"red":    "PLAYER_COLOR_RED",
		"blue":   "PLAYER_COLOR_BLUE",
		"green":  "PLAYER_COLOR_GREEN",
		"orange": "PLAYER_COLOR_ORANGE",
	}

	newPlayer := map[string]interface{}{
		"id":            playerID,
		"name":          playerName,
		"color":         colorMap[color],
		"resources":     map[string]interface{}{},
		"devCardCount":  0,
		"knightsPlayed": 0,
		"victoryPoints": 0,
		"connected":     false,
		"isReady":       false,
		"isHost":        false,
	}

	players = append(players, newPlayer)
	state["players"] = players

	updatedJSON, _ := json.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), gameID)
}

func (h *Handler) broadcastPlayerJoined(gameID, playerID, playerName, color string) {
	colorMap := map[string]string{
		"red":    "PLAYER_COLOR_RED",
		"blue":   "PLAYER_COLOR_BLUE",
		"green":  "PLAYER_COLOR_GREEN",
		"orange": "PLAYER_COLOR_ORANGE",
	}

	msg := map[string]interface{}{
		"message": map[string]interface{}{
			"oneofKind": "playerJoined",
			"playerJoined": map[string]interface{}{
				"player": map[string]interface{}{
					"id":            playerID,
					"name":          playerName,
					"color":         colorMap[color],
					"resources":     map[string]interface{}{},
					"devCardCount":  0,
					"knightsPlayed": 0,
					"victoryPoints": 0,
					"connected":     false,
					"isReady":       false,
					"isHost":        false,
				},
			},
		},
	}
	data, _ := json.Marshal(msg)
	h.hub.BroadcastToGame(gameID, data)
}

func (h *Handler) handleGetGame(w http.ResponseWriter, r *http.Request, code string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var gameRow struct {
		Code   string `db:"code"`
		Status string `db:"status"`
		State  string `db:"state"`
	}
	err := h.db.Get(&gameRow, "SELECT code, status, state FROM games WHERE code = ?", strings.ToUpper(code))
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Parse the full game state
	var gameState map[string]interface{}
	json.Unmarshal([]byte(gameRow.State), &gameState)

	// Get players with isReady status from parsed state
	players, _ := gameState["players"].([]interface{})
	playerList := make([]map[string]interface{}, len(players))
	for i, p := range players {
		player := p.(map[string]interface{})
		playerList[i] = map[string]interface{}{
			"name":    player["name"],
			"color":   player["color"],
			"isReady": player["isReady"],
			"isHost":  player["isHost"],
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":        gameRow.Code,
		"status":      gameRow.Status,
		"playerCount": len(players),
		"players":     playerList,
		"board":       gameState["board"],
	})
}
