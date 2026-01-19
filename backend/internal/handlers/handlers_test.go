package handlers

import (
	"bytes"
	"encoding/json"
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
	if err := unmarshalOptions.Unmarshal([]byte(stateJSON), &state); err != nil {
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
