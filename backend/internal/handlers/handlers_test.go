package handlers

import (
	"google.golang.org/protobuf/encoding/protojson"
	catanv1 "settlers_from_catan/gen/proto/catan/v1"
	"settlers_from_catan/internal/hub"
	"testing"
)

// ... Existing tests above ...

func TestHandleDiscardCards_ValidDiscard(t *testing.T) {
	// h, cleanup := setupTestHandler(t) // FIXME: Setup stubbed for compilation
	h := &Handler{} // Dummy handler, replace with real setup if available
	cleanup := func() {}
	defer cleanup()
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{
			{Id: "p1", Resources: &catanv1.ResourceCount{Wood: 3, Brick: 1}},
		},
		RobberPhase: &catanv1.RobberPhase{
			DiscardPending:  []string{"p1"},
			DiscardRequired: map[string]int32{"p1": 2},
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "robber-game-id")
	client := &hub.Client{PlayerID: "p1", GameID: "robber-game-id"}
	payload, _ := protojson.Marshal(&catanv1.DiscardCardsMessage{
		Resources: &catanv1.ResourceCount{Wood: 2},
	})
	h.handleDiscardCards(client, payload)
	var updatedJSON string
	h.db.Get(&updatedJSON, "SELECT state FROM games WHERE id = ?", "robber-game-id")
	var updatedState catanv1.GameState
	protojson.Unmarshal([]byte(updatedJSON), &updatedState)
	if len(updatedState.RobberPhase.DiscardPending) != 0 {
		t.Errorf("discard pending should be empty after discard")
	}
	if got := updatedState.Players[0].Resources.Wood; got != 1 {
		t.Errorf("Player should have 1 wood left, got %d", got)
	}
}

func TestHandleDiscardCards_InvalidCount(t *testing.T) {
	// h, cleanup := setupTestHandler(t) // FIXME: Setup stubbed for compilation
	h := &Handler{} // Dummy handler, replace with real setup if available
	cleanup := func() {}
	defer cleanup()
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{
			{Id: "p2", Resources: &catanv1.ResourceCount{Wood: 2}},
		},
		RobberPhase: &catanv1.RobberPhase{
			DiscardPending:  []string{"p2"},
			DiscardRequired: map[string]int32{"p2": 2},
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "robber-game-id")
	client := &hub.Client{PlayerID: "p2", GameID: "robber-game-id"}
	// Discard only 1 instead of 2
	payload, _ := protojson.Marshal(&catanv1.DiscardCardsMessage{
		Resources: &catanv1.ResourceCount{Wood: 1},
	})
	h.handleDiscardCards(client, payload)
	var updatedJSON string
	h.db.Get(&updatedJSON, "SELECT state FROM games WHERE id = ?", "robber-game-id")
	var updatedState catanv1.GameState
	protojson.Unmarshal([]byte(updatedJSON), &updatedState)
	if len(updatedState.RobberPhase.DiscardPending) != 1 {
		t.Errorf("discard pending should still have player after bad discard")
	}
	if got := updatedState.Players[0].Resources.Wood; got != 2 {
		t.Errorf("Resources should not change on error, got %d", got)
	}
}

func TestHandleMoveRobber_ValidMoveAndSteal(t *testing.T) {
	// h, cleanup := setupTestHandler(t) // FIXME: Setup stubbed for compilation
	h := &Handler{} // Dummy handler, replace with real setup if available
	cleanup := func() {}
	defer cleanup()
	robHex := &catanv1.HexCoord{Q: 0, R: 0}
	newHex := &catanv1.HexCoord{Q: 1, R: 1}
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{{Id: "roller"}, {Id: "vic", Resources: &catanv1.ResourceCount{Wood: 1}}},
		Board: &catanv1.BoardState{
			Hexes:     []*catanv1.Hex{{Coord: robHex}, {Coord: newHex}},
			RobberHex: robHex,
			Vertices: []*catanv1.Vertex{
				{Id: "v1", Building: &catanv1.Building{OwnerId: "vic", Type: catanv1.BuildingType_BUILDING_TYPE_SETTLEMENT}, AdjacentHexes: []*catanv1.HexCoord{newHex}},
			},
		},
		RobberPhase: &catanv1.RobberPhase{
			MovePendingPlayerId:  strptr("roller"),
			StealPendingPlayerId: strptr("roller"),
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "robber-game-id")
	client := &hub.Client{PlayerID: "roller", GameID: "robber-game-id"}
	payload, _ := protojson.Marshal(&catanv1.MoveRobberMessage{
		Hex:      newHex,
		VictimId: strptr("vic"),
	})
	h.handleMoveRobber(client, payload)
	var updatedJSON string
	h.db.Get(&updatedJSON, "SELECT state FROM games WHERE id = ?", "robber-game-id")
	var updatedState catanv1.GameState
	protojson.Unmarshal([]byte(updatedJSON), &updatedState)
	if updatedState.Board.RobberHex.Q != newHex.Q || updatedState.Board.RobberHex.R != newHex.R {
		t.Errorf("Robber should be on new hex")
	}
	if got := updatedState.Players[1].Resources.Wood; got != 0 {
		t.Errorf("Victim should have 0 wood after steal, got %d", got)
	}
}

func TestHandleMoveRobber_InvalidMove(t *testing.T) {
	// h, cleanup := setupTestHandler(t) // FIXME: Setup stubbed for compilation
	h := &Handler{} // Dummy handler, replace with real setup if available
	cleanup := func() {}
	defer cleanup()
	robHex := &catanv1.HexCoord{Q: 0, R: 0}
	state := &catanv1.GameState{
		Players: []*catanv1.PlayerState{{Id: "roller"}},
		Board: &catanv1.BoardState{
			Hexes:     []*catanv1.Hex{{Coord: robHex}},
			RobberHex: robHex,
		},
		RobberPhase: &catanv1.RobberPhase{
			MovePendingPlayerId: strptr("roller"),
		},
	}
	stateJSON, _ := protojson.Marshal(state)
	h.db.Exec("UPDATE games SET state = ? WHERE id = ?", string(stateJSON), "robber-game-id")
	client := &hub.Client{PlayerID: "roller", GameID: "robber-game-id"}
	// Try to move to same hex
	payload, _ := protojson.Marshal(&catanv1.MoveRobberMessage{
		Hex: robHex,
	})
	h.handleMoveRobber(client, payload)
	var updatedJSON string
	h.db.Get(&updatedJSON, "SELECT state FROM games WHERE id = ?", "robber-game-id")
	var updatedState catanv1.GameState
	protojson.Unmarshal([]byte(updatedJSON), &updatedState)
	if updatedState.Board.RobberHex.Q != robHex.Q || updatedState.Board.RobberHex.R != robHex.R {
		t.Errorf("Robber should not move on invalid move")
	}
}

// Helpers
func strptr(s string) *string { return &s }
