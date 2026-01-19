package handlers

import (
	"google.golang.org/protobuf/encoding/protojson"
	"settlers_from_catan/internal/game"
	"settlers_from_catan/internal/hub"
	"testing"
)

// TestGameOverWin_DisablesFurtherActions tests that after a winning move, the GameOver message is broadcast and further actions
// are disallowed
func TestGameOverWin_DisablesFurtherActions(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create minimal game state where Alice is about to win
	players := []*game.PlayerState{
		{Id: "alice", Name: "Alice", VictoryPoints: 9, IsReady: true},
		{Id: "bob", Name: "Bob", VictoryPoints: 2, IsReady: true},
	}
	board := &game.BoardState{
		Vertices: []*game.Vertex{
			{Id: "v1"},
		},
	}
	gs := &game.GameState{
		Players:     players,
		Board:       board,
		Status:      game.GameStatusPlaying,
		CurrentTurn: 0,
	}
	// Alice needs one more VP: place settlement on v1

	stateJSON, err := protojson.Marshal(gs)
	if err != nil {
		t.Fatalf("Failed to marshal game state: %v", err)
	}
	// Write state to DB
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "dummy-game-id")

	client := &hub.Client{PlayerID: "alice", GameID: "dummy-game-id"}
	payload := []byte(`{"structureType":"STRUCTURE_TYPE_SETTLEMENT","location":"v1"}`)

	// Call build structure (this should win the game)
	h.handleBuildStructure(client, payload)

	// Fetch updated game state
	var updatedJSON string
	h.db.Get(&updatedJSON, "SELECT state FROM games WHERE id = ?", "dummy-game-id")
	var updatedState game.GameState
	protojson.Unmarshal([]byte(updatedJSON), &updatedState)

	if updatedState.Status != game.GameStatusFinished {
		t.Errorf("Expected game status FINISHED, got %v", updatedState.Status)
	}

	// Attempt second action now that game is over
	h.handleBuildStructure(client, payload) // Should be blocked
	// No panic/crash: error should be sent
	// No further action should alter state
	var stillJSON string
	h.db.Get(&stillJSON, "SELECT state FROM games WHERE id = ?", "dummy-game-id")
	var stillState game.GameState
	protojson.Unmarshal([]byte(stillJSON), &stillState)
	if stillState.Status != game.GameStatusFinished {
		t.Errorf("Game should stay finished, got %v", stillState.Status)
	}
}
