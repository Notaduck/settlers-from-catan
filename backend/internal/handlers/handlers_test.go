package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"

	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/db"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
)

type gameStateEnvelope struct {
	Message struct {
		OneofKind string          `json:"oneofKind"`
		GameState json.RawMessage `json:"gameState"`
	} `json:"message"`
}

type gameStatePayload struct {
	State json.RawMessage `json:"state"`
}

func TestHandleCreateGame_PersistsStateAndPlayer(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	handler := NewHandler(database, hub.NewHub())

	payload := []byte(`{"playerName":"Alice"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/games", bytes.NewBuffer(payload))
	recorder := httptest.NewRecorder()

	handler.HandleCreateGame(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var resp catanv1.CreateGameResponse
	if err := protojson.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.GetGameId() == "" || resp.GetSessionToken() == "" || resp.GetPlayerId() == "" {
		t.Fatalf(
			"response missing ids: gameID=%q sessionToken=%q playerID=%q",
			resp.GetGameId(),
			resp.GetSessionToken(),
			resp.GetPlayerId(),
		)
	}

	var gameCount int
	if err := database.Get(&gameCount, "SELECT COUNT(*) FROM games"); err != nil {
		t.Fatalf("failed to query games: %v", err)
	}
	if gameCount != 1 {
		t.Fatalf("expected 1 game row, got %d", gameCount)
	}

	var playerCount int
	if err := database.Get(&playerCount, "SELECT COUNT(*) FROM players"); err != nil {
		t.Fatalf("failed to query players: %v", err)
	}
	if playerCount != 1 {
		t.Fatalf("expected 1 player row, got %d", playerCount)
	}

	var stateJSON string
	if err := database.Get(&stateJSON, "SELECT state FROM games WHERE id = ?", resp.GetGameId()); err != nil {
		t.Fatalf("failed to load game state: %v", err)
	}
	var state catanv1.GameState
	if err := protojson.Unmarshal([]byte(stateJSON), &state); err != nil {
		t.Fatalf("failed to parse game state: %v", err)
	}
	if len(state.Players) != 1 || state.Players[0].Name != "Alice" {
		t.Fatalf("unexpected players in state: %+v", state.Players)
	}
}

func TestHandleJoinGame_BroadcastsGameState(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	h := hub.NewHub()
	go h.Run()

	handler := NewHandler(database, h)
	server := httptest.NewServer(buildMux(handler))
	defer server.Close()

	createResp := createGameViaHTTP(t, server.URL, "Host")

	wsURL := strings.Replace(server.URL, "http", "ws", 1) + "/ws?token=" + createResp.GetSessionToken()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to connect websocket: %v", err)
	}
	defer conn.Close()

	readGameState(t, conn, 1)

	joinGameViaHTTP(t, server.URL, createResp.GetCode(), "Guest")

	readGameState(t, conn, 2)
}

func TestHandleMoveRobber_AllowsStealWithoutHex(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	h := hub.NewHub()
	go h.Run()

	handler := NewHandler(database, h)

	gameID := "game-robber"
	code := "ROB001"
	thiefID := "p1"
	victimID := "p2"
	state := game.NewGameState(gameID, code, []string{"Thief", "Victim"}, []string{thiefID, victimID})
	if state.Board == nil || state.Board.RobberHex == nil {
		t.Fatal("expected robber hex in board state")
	}

	var victimVertex *catanv1.Vertex
	for _, vertex := range state.Board.Vertices {
		for _, hex := range vertex.AdjacentHexes {
			if hex.Q == state.Board.RobberHex.Q && hex.R == state.Board.RobberHex.R {
				victimVertex = vertex
				break
			}
		}
		if victimVertex != nil {
			break
		}
	}
	if victimVertex == nil {
		t.Fatal("expected vertex adjacent to robber hex")
	}
	victimVertex.Building = &catanv1.Building{
		OwnerId: victimID,
		Type:    catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT,
	}

	for _, player := range state.Players {
		if player.Id == victimID {
			player.Resources = &catanv1.ResourceCount{Wood: 1}
		} else if player.Id == thiefID {
			player.Resources = &catanv1.ResourceCount{}
		}
	}

	state.RobberPhase = &catanv1.RobberPhase{
		StealPendingPlayerId: &thiefID,
	}

	stateJSON, err := protojson.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}
	if _, err := database.Exec(
		"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
		gameID,
		code,
		string(stateJSON),
		"playing",
	); err != nil {
		t.Fatalf("failed to insert game state: %v", err)
	}

	client := hub.NewClient(h, &websocket.Conn{}, thiefID, gameID)
	h.Register(client)

	payload, err := protojson.Marshal(&catanv1.MoveRobberMessage{
		VictimId: &victimID,
	})
	if err != nil {
		t.Fatalf("failed to marshal move robber payload: %v", err)
	}

	handler.handleMoveRobber(client, payload)

	var updatedStateJSON string
	if err := database.Get(&updatedStateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		t.Fatalf("failed to load updated state: %v", err)
	}
	var updatedState catanv1.GameState
	if err := protojson.Unmarshal([]byte(updatedStateJSON), &updatedState); err != nil {
		t.Fatalf("failed to parse updated state: %v", err)
	}

	var updatedVictim, updatedThief *catanv1.PlayerState
	for _, player := range updatedState.Players {
		switch player.Id {
		case victimID:
			updatedVictim = player
		case thiefID:
			updatedThief = player
		}
	}
	if updatedVictim == nil || updatedThief == nil {
		t.Fatalf("missing updated player states")
	}
	if updatedVictim.Resources.GetWood() != 0 {
		t.Fatalf("expected victim wood to be 0, got %d", updatedVictim.Resources.GetWood())
	}
	if updatedThief.Resources.GetWood() != 1 {
		t.Fatalf("expected thief wood to be 1, got %d", updatedThief.Resources.GetWood())
	}
	if updatedState.RobberPhase != nil {
		t.Fatalf("expected robber phase to clear after steal")
	}
}

func TestHandleClientMessage_BuildStructureUsesProtoJSON(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	h := hub.NewHub()
	go h.Run()

	handler := NewHandler(database, h)

	gameID := "game-setup"
	code := "SET001"
	playerID := "p1"
	state := game.NewGameState(gameID, code, []string{"Host"}, []string{playerID})
	state.Status = game.GameStatusSetup
	state.CurrentTurn = 0
	state.SetupPhase = &catanv1.SetupPhase{
		Round:            1,
		PlacementsInTurn: 0,
	}

	vertexID := state.Board.Vertices[0].Id

	stateJSON, err := protojson.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}
	if _, err := database.Exec(
		"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
		gameID,
		code,
		string(stateJSON),
		"setup",
	); err != nil {
		t.Fatalf("failed to insert game state: %v", err)
	}

	client := hub.NewClient(h, &websocket.Conn{}, playerID, gameID)
	h.Register(client)

	message := []byte(`{"message":{"oneofKind":"buildStructure","buildStructure":{"structureType":1,"location":"` + vertexID + `"}}}`)
	handler.handleClientMessage(client, message)

	var updatedStateJSON string
	if err := database.Get(&updatedStateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
		t.Fatalf("failed to load updated state: %v", err)
	}

	var updatedState catanv1.GameState
	if err := protojson.Unmarshal([]byte(updatedStateJSON), &updatedState); err != nil {
		t.Fatalf("failed to parse updated state: %v", err)
	}

	var placed *catanv1.Vertex
	for _, vertex := range updatedState.Board.Vertices {
		if vertex.Id == vertexID {
			placed = vertex
			break
		}
	}
	if placed == nil || placed.Building == nil {
		t.Fatalf("expected settlement placed at vertex %s", vertexID)
	}
	if placed.Building.Type != catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT {
		t.Fatalf("expected settlement type, got %v", placed.Building.Type)
	}
	if placed.Building.OwnerId != playerID {
		t.Fatalf("expected owner %s, got %s", playerID, placed.Building.OwnerId)
	}
}

func TestHandleSetTurnPhase_UpdatesTurnPhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	h := hub.NewHub()
	go h.Run()

	handler := NewHandler(database, h)

	tests := map[string]struct {
		setup     func(*catanv1.GameState)
		playerID  string
		phase     catanv1.TurnPhase
		wantPhase catanv1.TurnPhase
		wantErr   bool
	}{
		"trade to build": {
			playerID:  "p1",
			phase:     catanv1.TurnPhase_TURN_PHASE_BUILD,
			wantPhase: catanv1.TurnPhase_TURN_PHASE_BUILD,
		},
		"not your turn": {
			setup: func(state *catanv1.GameState) {
				state.CurrentTurn = 1
			},
			playerID:  "p1",
			phase:     catanv1.TurnPhase_TURN_PHASE_BUILD,
			wantPhase: catanv1.TurnPhase_TURN_PHASE_TRADE,
			wantErr:   true,
		},
	}

	gameIndex := 0
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gameIndex++
			gameID := fmt.Sprintf("game-phase-%d", gameIndex)
			code := fmt.Sprintf("TP%03d", gameIndex)
			state := game.NewGameState(gameID, code, []string{"Alice", "Bob"}, []string{"p1", "p2"})
			state.Status = game.GameStatusPlaying
			state.CurrentTurn = 0
			state.TurnPhase = catanv1.TurnPhase_TURN_PHASE_TRADE
			if tt.setup != nil {
				tt.setup(state)
			}

			stateJSON, err := protojson.Marshal(state)
			if err != nil {
				t.Fatalf("failed to marshal state: %v", err)
			}
			if _, err := database.Exec(
				"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
				gameID,
				code,
				string(stateJSON),
				"playing",
			); err != nil {
				t.Fatalf("failed to insert game state: %v", err)
			}

			client := hub.NewClient(h, &websocket.Conn{}, tt.playerID, gameID)
			h.Register(client)

			payload, err := protojson.Marshal(&catanv1.SetTurnPhaseMessage{Phase: tt.phase})
			if err != nil {
				t.Fatalf("failed to marshal set turn phase payload: %v", err)
			}

			handler.handleSetTurnPhase(client, payload)

			var updatedStateJSON string
			if err := database.Get(&updatedStateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
				t.Fatalf("failed to load updated state: %v", err)
			}
			var updatedState catanv1.GameState
			if err := protojson.Unmarshal([]byte(updatedStateJSON), &updatedState); err != nil {
				t.Fatalf("failed to parse updated state: %v", err)
			}

			if updatedState.TurnPhase != tt.wantPhase {
				t.Fatalf("expected phase %v, got %v", tt.wantPhase, updatedState.TurnPhase)
			}
			if tt.wantErr && updatedState.TurnPhase != catanv1.TurnPhase_TURN_PHASE_TRADE {
				t.Fatalf("expected phase to remain trade on error")
			}
		})
	}
}

func TestHandleGrantResources_AddsResources(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	handler := NewHandler(database, hub.NewHub())

	type resourceDelta struct {
		Wood  int32 `json:"wood"`
		Brick int32 `json:"brick"`
		Sheep int32 `json:"sheep"`
		Wheat int32 `json:"wheat"`
		Ore   int32 `json:"ore"`
	}
	type resourceSnapshot struct {
		Wood  int32
		Brick int32
		Sheep int32
		Wheat int32
		Ore   int32
	}

	tests := map[string]struct {
		initial *catanv1.ResourceCount
		delta   resourceDelta
		want    resourceSnapshot
	}{
		"adds to existing resources": {
			initial: &catanv1.ResourceCount{Wood: 1, Brick: 2},
			delta:   resourceDelta{Wood: 3, Ore: 2},
			want:    resourceSnapshot{Wood: 4, Brick: 2, Ore: 2},
		},
		"initializes resources when nil": {
			initial: nil,
			delta:   resourceDelta{Sheep: 2, Wheat: 1},
			want:    resourceSnapshot{Sheep: 2, Wheat: 1},
		},
	}

	testIndex := 0
	for name, tt := range tests {
		testIndex++
		index := testIndex
		t.Run(name, func(t *testing.T) {
			gameID := fmt.Sprintf("game-grant-%d", index)
			code := fmt.Sprintf("GR%03d", index)
			playerID := "p1"
			state := game.NewGameState(gameID, code, []string{"Alice"}, []string{playerID})
			state.Status = game.GameStatusPlaying
			state.Players[0].Resources = tt.initial

			stateJSON, err := protojson.Marshal(state)
			if err != nil {
				t.Fatalf("failed to marshal state: %v", err)
			}
			if _, err := database.Exec(
				"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
				gameID,
				code,
				string(stateJSON),
				"playing",
			); err != nil {
				t.Fatalf("failed to insert game state: %v", err)
			}

			payload := map[string]any{
				"gameCode": code,
				"playerId": playerID,
				"resources": map[string]int32{
					"wood":  tt.delta.Wood,
					"brick": tt.delta.Brick,
					"sheep": tt.delta.Sheep,
					"wheat": tt.delta.Wheat,
					"ore":   tt.delta.Ore,
				},
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/test/grant-resources", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			handler.HandleGrantResources(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", recorder.Code)
			}

			var updatedStateJSON string
			if err := database.Get(&updatedStateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
				t.Fatalf("failed to load updated state: %v", err)
			}
			var updatedState catanv1.GameState
			if err := protojson.Unmarshal([]byte(updatedStateJSON), &updatedState); err != nil {
				t.Fatalf("failed to parse updated state: %v", err)
			}

			updated := updatedState.Players[0].Resources
			if updated == nil {
				t.Fatalf("expected resources to be initialized")
			}
			got := resourceSnapshot{
				Wood:  updated.Wood,
				Brick: updated.Brick,
				Sheep: updated.Sheep,
				Wheat: updated.Wheat,
				Ore:   updated.Ore,
			}
			if got != tt.want {
				t.Fatalf("unexpected resources: got=%+v want=%+v", got, tt.want)
			}
		})
	}
}

func TestHandleSetGameState_UpdatesStatusAndPhase(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	handler := NewHandler(database, hub.NewHub())

	tests := map[string]struct {
		status     string
		phase      string
		wantStatus catanv1.GameStatus
		wantPhase  catanv1.TurnPhase
	}{
		"updates status to finished": {
			status:     "FINISHED",
			wantStatus: catanv1.GameStatus_GAME_STATUS_FINISHED,
			wantPhase:  catanv1.TurnPhase_TURN_PHASE_ROLL,
		},
		"updates phase to build": {
			phase:      "BUILD",
			wantStatus: catanv1.GameStatus_GAME_STATUS_PLAYING,
			wantPhase:  catanv1.TurnPhase_TURN_PHASE_BUILD,
		},
	}

	testIndex := 0
	for name, tt := range tests {
		testIndex++
		index := testIndex
		t.Run(name, func(t *testing.T) {
			gameID := fmt.Sprintf("game-state-%d", index)
			code := fmt.Sprintf("GS%03d", index)
			playerID := "p1"
			state := game.NewGameState(gameID, code, []string{"Alice"}, []string{playerID})
			state.Status = game.GameStatusPlaying
			state.TurnPhase = catanv1.TurnPhase_TURN_PHASE_ROLL

			stateJSON, err := protojson.Marshal(state)
			if err != nil {
				t.Fatalf("failed to marshal state: %v", err)
			}
			if _, err := database.Exec(
				"INSERT INTO games (id, code, state, status) VALUES (?, ?, ?, ?)",
				gameID,
				code,
				string(stateJSON),
				"playing",
			); err != nil {
				t.Fatalf("failed to insert game state: %v", err)
			}

			payload := map[string]string{
				"gameCode": code,
			}
			if tt.status != "" {
				payload["status"] = tt.status
			}
			if tt.phase != "" {
				payload["phase"] = tt.phase
			}
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/test/set-game-state", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()

			handler.HandleSetGameState(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", recorder.Code)
			}

			var updatedStateJSON string
			if err := database.Get(&updatedStateJSON, "SELECT state FROM games WHERE id = ?", gameID); err != nil {
				t.Fatalf("failed to load updated state: %v", err)
			}
			var updatedState catanv1.GameState
			if err := protojson.Unmarshal([]byte(updatedStateJSON), &updatedState); err != nil {
				t.Fatalf("failed to parse updated state: %v", err)
			}

			if updatedState.Status != tt.wantStatus {
				t.Fatalf("expected status %v, got %v", tt.wantStatus, updatedState.Status)
			}
			if updatedState.TurnPhase != tt.wantPhase {
				t.Fatalf("expected phase %v, got %v", tt.wantPhase, updatedState.TurnPhase)
			}
		})
	}
}

func setupTestDB(t *testing.T) (*sqlx.DB, func()) {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	database, err := db.Initialize(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	return database, func() { _ = database.Close() }
}

func buildMux(handler *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/games", handler.HandleCreateGame)
	mux.HandleFunc("/api/games/", handler.HandleGameRoutes)
	mux.HandleFunc("/ws", handler.HandleWebSocket)
	return mux
}

func createGameViaHTTP(t *testing.T, baseURL, playerName string) *catanv1.CreateGameResponse {
	t.Helper()
	body := map[string]string{"playerName": playerName}
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/api/games", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var createResp catanv1.CreateGameResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read create response: %v", err)
	}
	if err := protojson.Unmarshal(bodyBytes, &createResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	return &createResp
}

func joinGameViaHTTP(t *testing.T, baseURL, code, playerName string) *catanv1.JoinGameResponse {
	t.Helper()
	body := map[string]string{"playerName": playerName}
	data, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/api/games/"+code+"/join", "application/json", bytes.NewBuffer(data))
	if err != nil {
		t.Fatalf("failed to join game: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var joinResp catanv1.JoinGameResponse
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read join response: %v", err)
	}
	if err := protojson.Unmarshal(bodyBytes, &joinResp); err != nil {
		t.Fatalf("failed to decode join response: %v", err)
	}
	return &joinResp
}

func readGameState(t *testing.T, conn *websocket.Conn, expectedPlayers int) {
	t.Helper()
	for i := 0; i < 5; i++ {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read websocket message: %v", err)
		}
		var envelope gameStateEnvelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			t.Fatalf("failed to unmarshal websocket message: %v", err)
		}
		if envelope.Message.OneofKind != "gameState" || len(envelope.Message.GameState) == 0 {
			continue
		}
		var payload gameStatePayload
		if err := json.Unmarshal(envelope.Message.GameState, &payload); err != nil {
			t.Fatalf("failed to unmarshal game state payload: %v", err)
		}
		var state catanv1.GameState
		if err := protojson.Unmarshal(payload.State, &state); err != nil {
			t.Fatalf("failed to decode game state: %v", err)
		}
		if len(state.Players) == expectedPlayers {
			return
		}
	}
	t.Fatalf("expected game state with %d players", expectedPlayers)
}
