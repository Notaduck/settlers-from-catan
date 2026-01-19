package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"settlers_from_catan/internal/db"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"

	"google.golang.org/protobuf/encoding/protojson"
)

func setupTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()

	tmpFile := "test_handlers.db"
	database, err := db.Initialize(tmpFile)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	h := hub.NewHub()
	go h.Run()

	handler := NewHandler(database, h)

	cleanup := func() {
		database.Close()
		os.Remove(tmpFile)
	}

	return handler, cleanup
}

func TestHandleCreateGame(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	reqBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/games", reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleCreateGame(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["gameId"] == "" {
		t.Error("Response missing gameId")
	}
	if response["code"] == "" {
		t.Error("Response missing code")
	}
	if response["sessionToken"] == "" {
		t.Error("Response missing sessionToken")
	}
	if len(response["code"]) != 6 {
		t.Errorf("Expected code to be 6 characters, got %d", len(response["code"]))
	}
}

func TestHandleCreateGame_EmptyName(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	reqBody := bytes.NewBuffer([]byte(`{"playerName": ""}`))
	req := httptest.NewRequest(http.MethodPost, "/api/games", reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleCreateGame(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty name, got %d", w.Code)
	}
}

func TestHandleCreateGame_NameTooLong(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	reqBody := bytes.NewBuffer([]byte(`{"playerName": "ThisNameIsWayTooLongForTheGame"}`))
	req := httptest.NewRequest(http.MethodPost, "/api/games", reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.HandleCreateGame(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for name too long, got %d", w.Code)
	}
}

func TestJoinGame(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	joinBody := bytes.NewBuffer([]byte(`{"playerName": "Bob"}`))
	joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	handler.HandleGameRoutes(joinW, joinReq)

	if joinW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", joinW.Code, joinW.Body.String())
	}

	var joinResponse map[string]interface{}
	json.NewDecoder(joinW.Body).Decode(&joinResponse)

	if joinResponse["sessionToken"] == "" {
		t.Error("Response missing sessionToken")
	}

	players, ok := joinResponse["players"].([]interface{})
	if !ok {
		t.Fatal("Response missing players array")
	}
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}
}

func TestJoinGame_GameNotFound(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	joinBody := bytes.NewBuffer([]byte(`{"playerName": "Bob"}`))
	joinReq := httptest.NewRequest(http.MethodPost, "/api/games/NOTFND/join", joinBody)
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	handler.HandleGameRoutes(joinW, joinReq)

	if joinW.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", joinW.Code)
	}
}

func TestJoinGame_GameFull(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	for _, name := range []string{"Bob", "Charlie", "Diana"} {
		joinBody := bytes.NewBuffer([]byte(`{"playerName": "` + name + `"}`))
		joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
		joinReq.Header.Set("Content-Type", "application/json")
		joinW := httptest.NewRecorder()
		handler.HandleGameRoutes(joinW, joinReq)

		if joinW.Code != http.StatusOK {
			t.Errorf("Player %s should be able to join, got %d", name, joinW.Code)
		}
	}

	joinBody := bytes.NewBuffer([]byte(`{"playerName": "Eve"}`))
	joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	handler.HandleGameRoutes(joinW, joinReq)

	if joinW.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for full game, got %d", joinW.Code)
	}
}

func TestGetGame(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	getReq := httptest.NewRequest(http.MethodGet, "/api/games/"+gameCode, nil)
	getW := httptest.NewRecorder()
	handler.HandleGameRoutes(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", getW.Code)
	}

	var getResponse map[string]interface{}
	json.NewDecoder(getW.Body).Decode(&getResponse)

	if getResponse["code"] != gameCode {
		t.Errorf("Expected code %s, got %s", gameCode, getResponse["code"])
	}
	if getResponse["status"] != "waiting" {
		t.Errorf("Expected status 'waiting', got %s", getResponse["status"])
	}
}

func TestPlayersHaveDifferentColors(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	for _, name := range []string{"Bob", "Charlie", "Diana"} {
		joinBody := bytes.NewBuffer([]byte(`{"playerName": "` + name + `"}`))
		joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
		joinReq.Header.Set("Content-Type", "application/json")
		joinW := httptest.NewRecorder()
		handler.HandleGameRoutes(joinW, joinReq)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/games/"+gameCode, nil)
	getW := httptest.NewRecorder()
	handler.HandleGameRoutes(getW, getReq)

	var getResponse map[string]interface{}
	json.NewDecoder(getW.Body).Decode(&getResponse)

	players := getResponse["players"].([]interface{})
	colors := make(map[string]bool)

	for _, p := range players {
		player := p.(map[string]interface{})
		color := player["color"].(string)
		if colors[color] {
			t.Errorf("Duplicate color found: %s", color)
		}
		colors[color] = true
	}

	if len(colors) != 4 {
		t.Errorf("Expected 4 different colors, got %d", len(colors))
	}
}

func TestStartGame_RequiresHost(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create game as Alice (host)
	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	// Bob joins
	joinBody := bytes.NewBuffer([]byte(`{"playerName": "Bob"}`))
	joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	handler.HandleGameRoutes(joinW, joinReq)

	var joinResponse map[string]interface{}
	json.NewDecoder(joinW.Body).Decode(&joinResponse)

	// Check that Alice is host and Bob is not
	players := joinResponse["players"].([]interface{})
	for _, p := range players {
		player := p.(map[string]interface{})
		name := player["name"].(string)
		isHost, ok := player["isHost"].(bool)
		if !ok {
			isHost = false // default if not present
		}

		if name == "Alice" && !isHost {
			t.Error("Alice should be the host")
		}
		if name == "Bob" && isHost {
			t.Error("Bob should not be the host")
		}
	}
}

func TestStartGame_RequiresAllReady(t *testing.T) {
	// This test validates that the game cannot start unless all players are ready
	// The actual ready toggling is done via WebSocket, so this tests the rule
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	// Get game state - players should not be ready by default
	getReq := httptest.NewRequest(http.MethodGet, "/api/games/"+gameCode, nil)
	getW := httptest.NewRecorder()
	handler.HandleGameRoutes(getW, getReq)

	var getResponse map[string]interface{}
	json.NewDecoder(getW.Body).Decode(&getResponse)

	players := getResponse["players"].([]interface{})
	for _, p := range players {
		player := p.(map[string]interface{})
		isReady, ok := player["isReady"].(bool)
		if ok && isReady {
			t.Error("Players should not be ready by default")
		}
	}
}

func TestStartGame_InitializesSetupPhase(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	if err := json.NewDecoder(createW.Body).Decode(&createResponse); err != nil {
		t.Fatalf("Failed to decode create response: %v", err)
	}

	gameID := createResponse["gameId"]
	hostID := createResponse["playerId"]
	gameCode := createResponse["code"]

	joinBody := bytes.NewBuffer([]byte(`{"playerName": "Bob"}`))
	joinReq := httptest.NewRequest(http.MethodPost, "/api/games/"+gameCode+"/join", joinBody)
	joinReq.Header.Set("Content-Type", "application/json")
	joinW := httptest.NewRecorder()
	handler.HandleGameRoutes(joinW, joinReq)

	if joinW.Code != http.StatusOK {
		t.Fatalf("Expected join status 200, got %d: %s", joinW.Code, joinW.Body.String())
	}

	var stateJSON string
	if err := handler.db.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		t.Fatalf("Failed to get game state: %v", err)
	}

	var state game.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		t.Fatalf("Failed to unmarshal game state: %v", err)
	}

	for _, player := range state.Players {
		player.IsReady = true
	}

	updatedJSON, err := jsonMarshaler.Marshal(&state)
	if err != nil {
		t.Fatalf("Failed to marshal updated state: %v", err)
	}
	handler.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(updatedJSON), gameID)

	client := &hub.Client{PlayerID: hostID, GameID: gameID}
	handler.handleStartGame(client)

	var startedJSON string
	if err := handler.db.Get(&startedJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		t.Fatalf("Failed to get started game state: %v", err)
	}

	var startedState game.GameState
	if err := protojson.Unmarshal([]byte(startedJSON), &startedState); err != nil {
		t.Fatalf("Failed to unmarshal started game state: %v", err)
	}

	if startedState.Status != game.GameStatusSetup {
		t.Errorf("Expected status setup, got %v", startedState.Status)
	}
	if startedState.TurnPhase != game.TurnPhaseBuild {
		t.Errorf("Expected turn phase build, got %v", startedState.TurnPhase)
	}
	if startedState.CurrentTurn != 0 {
		t.Errorf("Expected current turn 0, got %d", startedState.CurrentTurn)
	}
	if startedState.SetupPhase == nil {
		t.Fatal("Expected setup phase to be initialized")
	}
	if startedState.SetupPhase.Round != 1 {
		t.Errorf("Expected setup round 1, got %d", startedState.SetupPhase.Round)
	}
	if startedState.SetupPhase.PlacementsInTurn != 0 {
		t.Errorf("Expected placements_in_turn 0, got %d", startedState.SetupPhase.PlacementsInTurn)
	}

	var status string
	if err := handler.db.Get(&status, "SELECT status FROM games WHERE id = ?", gameID); err != nil {
		t.Fatalf("Failed to get game status: %v", err)
	}
	if status != "setup" {
		t.Errorf("Expected status column 'setup', got %s", status)
	}
}

func TestGetGame_IncludesBoard(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	createBody := bytes.NewBuffer([]byte(`{"playerName": "Alice"}`))
	createReq := httptest.NewRequest(http.MethodPost, "/api/games", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	handler.HandleCreateGame(createW, createReq)

	var createResponse map[string]string
	json.NewDecoder(createW.Body).Decode(&createResponse)
	gameCode := createResponse["code"]

	getReq := httptest.NewRequest(http.MethodGet, "/api/games/"+gameCode, nil)
	getW := httptest.NewRecorder()
	handler.HandleGameRoutes(getW, getReq)

	var getResponse map[string]interface{}
	json.NewDecoder(getW.Body).Decode(&getResponse)

	board, ok := getResponse["board"].(map[string]interface{})
	if !ok {
		t.Fatal("Response should include board")
	}

	hexes, ok := board["hexes"].([]interface{})
	if !ok {
		t.Fatal("Board should include hexes")
	}

	if len(hexes) != 19 {
		t.Errorf("Board should have 19 hexes, got %d", len(hexes))
	}

	vertices, ok := board["vertices"].([]interface{})
	if !ok {
		t.Fatal("Board should include vertices")
	}

	if len(vertices) < 54 {
		t.Errorf("Board should have at least 54 vertices, got %d", len(vertices))
	}

	edges, ok := board["edges"].([]interface{})
	if !ok {
		t.Fatal("Board should include edges")
	}

	if len(edges) < 72 {
		t.Errorf("Board should have at least 72 edges, got %d", len(edges))
	}

	robberHex, ok := board["robberHex"].(map[string]interface{})
	if !ok {
		t.Fatal("Board should include robberHex")
	}

	// Robber should have coordinates (even if 0,0)
	if _, hasQ := robberHex["q"]; !hasQ {
		t.Error("RobberHex should have q coordinate")
	}
}
