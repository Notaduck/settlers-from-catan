package handlers

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsMarshal = protojson.MarshalOptions{
	UseEnumNumbers: true,
}

type clientEnvelope struct {
	Message struct {
		OneofKind    string          `json:"oneofKind"`
		JoinGame     json.RawMessage `json:"joinGame,omitempty"`
		StartGame    json.RawMessage `json:"startGame,omitempty"`
		RollDice     json.RawMessage `json:"rollDice,omitempty"`
		BuildStruct  json.RawMessage `json:"buildStructure,omitempty"`
		ProposeTrade json.RawMessage `json:"proposeTrade,omitempty"`
		RespondTrade json.RawMessage `json:"respondTrade,omitempty"`
		MoveRobber   json.RawMessage `json:"moveRobber,omitempty"`
		EndTurn      json.RawMessage `json:"endTurn,omitempty"`
		PlayDevCard  json.RawMessage `json:"playDevCard,omitempty"`
		PlayerReady  json.RawMessage `json:"playerReady,omitempty"`
		DiscardCards json.RawMessage `json:"discardCards,omitempty"`
		BankTrade    json.RawMessage `json:"bankTrade,omitempty"`
		SetTurnPhase json.RawMessage `json:"setTurnPhase,omitempty"`
		BuyDevCard   json.RawMessage `json:"buyDevCard,omitempty"`
	} `json:"message"`
}

type serverEnvelope struct {
	Message serverMessage `json:"message"`
}

type serverMessage struct {
	OneofKind string          `json:"oneofKind"`
	GameState *gameStateWire  `json:"gameState,omitempty"`
	GameOver  json.RawMessage `json:"gameOver,omitempty"`
	Error     json.RawMessage `json:"error,omitempty"`
}

type gameStateWire struct {
	State json.RawMessage `json:"state"`
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

	// Broadcast updated game state to connected clients
	h.broadcastGameStatePersonalized(gameID, &state)
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
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/join") {
		h.HandleJoinGame(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := extractGameCode(r.URL.Path)
	if code == "" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	_, state, err := h.loadGameByCode(code)
	if err != nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	players := make([]*catanv1.PlayerInfo, 0, len(state.Players))
	for _, p := range state.Players {
		players = append(players, &catanv1.PlayerInfo{
			Id:    p.Id,
			Name:  p.Name,
			Color: p.Color,
		})
	}

	resp := &catanv1.GameInfoResponse{
		Code:        code,
		Status:      state.Status,
		PlayerCount: int32(len(state.Players)),
		Players:     players,
	}

	w.Header().Set("Content-Type", "application/json")
	resJSON, err := protojson.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(resJSON)
}

func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	var player struct {
		ID     string `db:"id"`
		GameID string `db:"game_id"`
	}
	if err := h.db.Get(&player, "SELECT id, game_id FROM players WHERE session_token = ?", token); err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := hub.NewClient(h.hub, conn, player.ID, player.GameID)
	client.OnMessage = func(payload []byte) {
		h.handleClientMessage(client, payload)
	}

	h.hub.Register(client)

	// Mark player connected in DB + game state, then broadcast state.
	state, err := h.loadGameState(player.GameID)
	if err == nil && state != nil {
		markPlayerConnected(state, player.ID, true)
		_ = h.saveGameState(player.GameID, state)
		h.broadcastGameStatePersonalized(player.GameID, state)
	}
	_, _ = h.db.Exec("UPDATE players SET connected = 1, last_seen = CURRENT_TIMESTAMP WHERE id = ?", player.ID)

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) HandleGrantResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type resourceDelta struct {
		Wood  int32 `json:"wood"`
		Brick int32 `json:"brick"`
		Sheep int32 `json:"sheep"`
		Wheat int32 `json:"wheat"`
		Ore   int32 `json:"ore"`
	}
	type apiRequest struct {
		GameCode  string        `json:"gameCode"`
		PlayerID  string        `json:"playerId"`
		Resources resourceDelta `json:"resources"`
	}

	var req apiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.GameCode == "" || req.PlayerID == "" {
		http.Error(w, "missing gameCode or playerId", http.StatusBadRequest)
		return
	}
	if req.Resources.Wood < 0 || req.Resources.Brick < 0 || req.Resources.Sheep < 0 || req.Resources.Wheat < 0 || req.Resources.Ore < 0 {
		http.Error(w, "resource values must be non-negative", http.StatusBadRequest)
		return
	}

	gameID, state, err := h.loadGameByCode(req.GameCode)
	if err != nil || state == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var target *catanv1.PlayerState
	for _, player := range state.Players {
		if player.Id == req.PlayerID {
			target = player
			break
		}
	}
	if target == nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}
	if target.Resources == nil {
		target.Resources = &catanv1.ResourceCount{}
	}
	target.Resources.Wood += req.Resources.Wood
	target.Resources.Brick += req.Resources.Brick
	target.Resources.Sheep += req.Resources.Sheep
	target.Resources.Wheat += req.Resources.Wheat
	target.Resources.Ore += req.Resources.Ore

	if err := h.saveGameState(gameID, state); err != nil {
		http.Error(w, "failed to persist game state", http.StatusInternalServerError)
		return
	}
	h.broadcastGameStatePersonalized(gameID, state)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"playerId": req.PlayerID,
		"resources": map[string]int32{
			"wood":  target.Resources.Wood,
			"brick": target.Resources.Brick,
			"sheep": target.Resources.Sheep,
			"wheat": target.Resources.Wheat,
			"ore":   target.Resources.Ore,
		},
	})
}

func (h *Handler) handleClientMessage(client *hub.Client, payload []byte) {
	if client == nil || client.GameID == "" || len(payload) == 0 {
		return
	}

	var envelope clientEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		h.sendError(client, "bad_request", "invalid client message")
		return
	}

	switch envelope.Message.OneofKind {
	case "playerReady":
		var msg catanv1.PlayerReadyMessage
		if err := protojson.Unmarshal(envelope.Message.PlayerReady, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid ready payload")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.SetPlayerReady(state, client.PlayerID, msg.Ready)
		})
	case "startGame":
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.StartGame(state, client.PlayerID)
		})
	case "buildStructure":
		var msg catanv1.BuildStructureMessage
		if err := protojson.Unmarshal(envelope.Message.BuildStruct, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid build payload")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return applyBuildStructure(state, client.PlayerID, &msg)
		})
	case "rollDice":
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			_, err := game.PerformDiceRoll(state, client.PlayerID)
			return err
		})
	case "endTurn":
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.EndTurn(state, client.PlayerID)
		})
	case "setTurnPhase":
		h.handleSetTurnPhase(client, envelope.Message.SetTurnPhase)
	case "proposeTrade":
		var msg catanv1.ProposeTradeMessage
		if err := protojson.Unmarshal(envelope.Message.ProposeTrade, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid trade payload")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			_, err := game.ProposeTrade(state, client.PlayerID, msg.TargetId, msg.Offering, msg.Requesting)
			return err
		})
	case "respondTrade":
		var msg catanv1.RespondTradeMessage
		if err := protojson.Unmarshal(envelope.Message.RespondTrade, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid trade response")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.RespondTrade(state, msg.TradeId, client.PlayerID, msg.Accept)
		})
	case "bankTrade":
		var msg catanv1.BankTradeMessage
		if err := protojson.Unmarshal(envelope.Message.BankTrade, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid bank trade")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.BankTrade(state, client.PlayerID, msg.Offering, msg.ResourceRequested)
		})
	case "buyDevCard":
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			_, err := game.BuyDevCard(state, client.PlayerID)
			return err
		})
	case "playDevCard":
		var msg catanv1.PlayDevCardMessage
		if err := protojson.Unmarshal(envelope.Message.PlayDevCard, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid dev card")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.PlayDevCard(state, client.PlayerID, msg.CardType, msg.TargetResource, msg.Resources)
		})
	case "discardCards":
		var msg catanv1.DiscardCardsMessage
		if err := protojson.Unmarshal(envelope.Message.DiscardCards, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid discard payload")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			return game.DiscardCards(state, client.PlayerID, msg.Resources)
		})
	case "moveRobber":
		var msg catanv1.MoveRobberMessage
		if err := protojson.Unmarshal(envelope.Message.MoveRobber, &msg); err != nil {
			h.sendError(client, "bad_request", "invalid robber payload")
			return
		}
		h.applyGameUpdate(client, func(state *catanv1.GameState) error {
			if msg.VictimId != nil && *msg.VictimId != "" {
				_, err := game.StealFromPlayer(state, client.PlayerID, *msg.VictimId)
				return err
			}
			return game.MoveRobber(state, client.PlayerID, msg.Hex)
		})
	default:
		h.sendError(client, "bad_request", "unknown message type")
	}
}

func (h *Handler) handleSetTurnPhase(client *hub.Client, payload []byte) {
	if client == nil || client.GameID == "" || len(payload) == 0 {
		return
	}
	var msg catanv1.SetTurnPhaseMessage
	if err := protojson.Unmarshal(payload, &msg); err != nil {
		h.sendError(client, "bad_request", "invalid turn phase payload")
		return
	}
	h.applyGameUpdate(client, func(state *catanv1.GameState) error {
		return game.SetTurnPhase(state, client.PlayerID, msg.Phase)
	})
}

func (h *Handler) HandleGrantDevCard(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleForceDiceRoll(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) HandleSetGameState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type apiRequest struct {
		GameCode string `json:"gameCode"`
		Status   string `json:"status"`
		Phase    string `json:"phase"`
	}

	var req apiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.GameCode == "" {
		http.Error(w, "missing gameCode", http.StatusBadRequest)
		return
	}
	if req.Status == "" && req.Phase == "" {
		http.Error(w, "missing status or phase", http.StatusBadRequest)
		return
	}

	gameID, state, err := h.loadGameByCode(req.GameCode)
	if err != nil || state == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	prevStatus := state.Status

	if req.Status != "" {
		status, ok := parseGameStatus(req.Status)
		if !ok {
			http.Error(w, "invalid status", http.StatusBadRequest)
			return
		}
		state.Status = status
	}
	if req.Phase != "" {
		phase, ok := parseTurnPhase(req.Phase)
		if !ok {
			http.Error(w, "invalid phase", http.StatusBadRequest)
			return
		}
		state.TurnPhase = phase
	}

	if err := h.saveGameState(gameID, state); err != nil {
		http.Error(w, "failed to persist game state", http.StatusInternalServerError)
		return
	}
	h.broadcastGameStatePersonalized(gameID, state)

	if prevStatus != catanv1.GameStatus_GAME_STATUS_FINISHED && state.Status == catanv1.GameStatus_GAME_STATUS_FINISHED {
		winnerID, ok := game.DetermineWinner(state)
		if !ok {
			winnerID = ""
		}
		h.broadcastGameOver(state, winnerID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": state.Status.String(),
		"phase":  state.TurnPhase.String(),
	})
}

func parseGameStatus(value string) (catanv1.GameStatus, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "WAITING", "GAME_STATUS_WAITING":
		return catanv1.GameStatus_GAME_STATUS_WAITING, true
	case "SETUP", "GAME_STATUS_SETUP":
		return catanv1.GameStatus_GAME_STATUS_SETUP, true
	case "PLAYING", "GAME_STATUS_PLAYING":
		return catanv1.GameStatus_GAME_STATUS_PLAYING, true
	case "FINISHED", "GAME_STATUS_FINISHED":
		return catanv1.GameStatus_GAME_STATUS_FINISHED, true
	default:
		return catanv1.GameStatus_GAME_STATUS_UNSPECIFIED, false
	}
}

func parseTurnPhase(value string) (catanv1.TurnPhase, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	switch normalized {
	case "ROLL", "TURN_PHASE_ROLL":
		return catanv1.TurnPhase_TURN_PHASE_ROLL, true
	case "TRADE", "TURN_PHASE_TRADE":
		return catanv1.TurnPhase_TURN_PHASE_TRADE, true
	case "BUILD", "TURN_PHASE_BUILD":
		return catanv1.TurnPhase_TURN_PHASE_BUILD, true
	case "UNSPECIFIED", "TURN_PHASE_UNSPECIFIED":
		return catanv1.TurnPhase_TURN_PHASE_UNSPECIFIED, true
	default:
		return catanv1.TurnPhase_TURN_PHASE_UNSPECIFIED, false
	}
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

func applyBuildStructure(state *catanv1.GameState, playerID string, msg *catanv1.BuildStructureMessage) error {
	if state == nil || msg == nil {
		return errors.New("invalid build payload")
	}

	switch msg.StructureType {
	case catanv1.StructureType_STRUCTURE_TYPE_SETTLEMENT:
		if state.Status == catanv1.GameStatus_GAME_STATUS_SETUP {
			return game.PlaceSetupSettlement(state, playerID, msg.Location)
		}
		return game.PlaceSettlement(state, playerID, msg.Location)
	case catanv1.StructureType_STRUCTURE_TYPE_ROAD:
		if state.Status == catanv1.GameStatus_GAME_STATUS_SETUP {
			return game.PlaceSetupRoad(state, playerID, msg.Location)
		}
		return game.PlaceRoad(state, playerID, msg.Location)
	case catanv1.StructureType_STRUCTURE_TYPE_CITY:
		return game.PlaceCity(state, playerID, msg.Location)
	default:
		return errors.New("invalid structure type")
	}
}

func markPlayerConnected(state *catanv1.GameState, playerID string, connected bool) {
	if state == nil {
		return
	}
	for _, p := range state.Players {
		if p.Id == playerID {
			p.Connected = connected
			return
		}
	}
}

func extractGameCode(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToUpper(parts[len(parts)-1])
}

func (h *Handler) loadGameByCode(code string) (string, *catanv1.GameState, error) {
	var (
		gameID    string
		stateJSON string
	)
	if err := h.db.Get(&gameID, "SELECT id FROM games WHERE code = ?", code); err != nil {
		return "", nil, err
	}
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		return "", nil, err
	}
	var state catanv1.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		return "", nil, err
	}
	return gameID, &state, nil
}

func (h *Handler) loadGameState(gameID string) (*catanv1.GameState, error) {
	var stateJSON string
	if err := h.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		return nil, err
	}
	var state catanv1.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (h *Handler) saveGameState(gameID string, state *catanv1.GameState) error {
	if state == nil {
		return errors.New("nil game state")
	}
	stateJSON, err := protojson.Marshal(state)
	if err != nil {
		return err
	}
	status := gameStatusToString(state.Status)
	_, err = h.db.Exec(
		"UPDATE games SET state = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		string(stateJSON),
		status,
		gameID,
	)
	return err
}

func gameStatusToString(status catanv1.GameStatus) string {
	switch status {
	case catanv1.GameStatus_GAME_STATUS_WAITING:
		return "waiting"
	case catanv1.GameStatus_GAME_STATUS_SETUP:
		return "setup"
	case catanv1.GameStatus_GAME_STATUS_PLAYING:
		return "playing"
	case catanv1.GameStatus_GAME_STATUS_FINISHED:
		return "finished"
	default:
		return "unknown"
	}
}

func (h *Handler) applyGameUpdate(client *hub.Client, apply func(state *catanv1.GameState) error) {
	if client == nil || client.GameID == "" {
		return
	}
	state, err := h.loadGameState(client.GameID)
	if err != nil {
		h.sendError(client, "load_failed", "failed to load game state")
		return
	}
	prevStatus := state.Status
	if err := apply(state); err != nil {
		h.sendError(client, "invalid_action", err.Error())
		return
	}
	if err := h.saveGameState(client.GameID, state); err != nil {
		h.sendError(client, "persist_failed", "failed to persist game state")
		return
	}
	h.broadcastGameStatePersonalized(client.GameID, state)

	if prevStatus != catanv1.GameStatus_GAME_STATUS_FINISHED && state.Status == catanv1.GameStatus_GAME_STATUS_FINISHED {
		winnerID, ok := game.DetermineWinner(state)
		if !ok {
			winnerID = ""
		}
		h.broadcastGameOver(state, winnerID)
	}
}

func (h *Handler) sendError(client *hub.Client, code, message string) {
	if client == nil {
		return
	}
	payload, err := wsMarshal.Marshal(&catanv1.ErrorPayload{Code: code, Message: message})
	if err != nil {
		return
	}
	envelope := serverEnvelope{
		Message: serverMessage{
			OneofKind: "error",
			Error:     payload,
		},
	}
	if msg, err := json.Marshal(envelope); err == nil {
		client.Send(msg)
	}
}

func (h *Handler) broadcastGameOver(state *catanv1.GameState, winnerID string) {
	if state == nil {
		return
	}
	payload := game.BuildGameOverPayload(state, winnerID)
	payloadJSON, err := wsMarshal.Marshal(payload)
	if err != nil {
		return
	}
	envelope := serverEnvelope{
		Message: serverMessage{
			OneofKind: "gameOver",
			GameOver:  payloadJSON,
		},
	}
	msg, err := json.Marshal(envelope)
	if err != nil {
		return
	}
	h.hub.BroadcastToGame(state.Id, msg)
}
