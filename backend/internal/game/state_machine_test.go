package game

import (
	"testing"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Setup Phase Tests
// Catan setup follows a "snake draft":
// Round 1: P1 → P2 → P3 → P4 (forward)
// Round 2: P4 → P3 → P2 → P1 (reverse)
// Each player places 1 settlement + 1 road per round

func TestSetupPlacementOrder_SnakeDraft(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob", "Charlie", "Dave"}, []string{"p1", "p2", "p3", "p4"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{
		Round:            1,
		PlacementsInTurn: 0,
	}
	state.CurrentTurn = 0

	// Expected turn order for setup:
	// Round 1: 0, 1, 2, 3 (forward)
	// Round 2: 3, 2, 1, 0 (reverse)
	expectedOrder := []int32{0, 1, 2, 3, 3, 2, 1, 0}

	for i, expected := range expectedOrder {
		currentIdx := GetSetupTurnIndex(state, i)
		if currentIdx != expected {
			t.Errorf("Placement %d: expected player index %d, got %d", i, expected, currentIdx)
		}
	}
}

func TestSetupPlacementOrder_ThreePlayers(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob", "Charlie"}, []string{"p1", "p2", "p3"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP

	// 3 players: 0,1,2,2,1,0
	expectedOrder := []int32{0, 1, 2, 2, 1, 0}

	for i, expected := range expectedOrder {
		currentIdx := GetSetupTurnIndex(state, i)
		if currentIdx != expected {
			t.Errorf("Placement %d: expected player index %d, got %d", i, expected, currentIdx)
		}
	}
}

func TestSetupSettlement_NoResourceCost(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}

	// Player has no resources
	player := state.Players[0]
	if player.Resources.Wood != 0 || player.Resources.Brick != 0 {
		t.Fatal("Test setup: player should start with 0 resources")
	}

	// Pick a valid vertex
	vertexID := state.Board.Vertices[0].Id

	err := PlaceSetupSettlement(state, "p1", vertexID)
	if err != nil {
		t.Errorf("Setup settlement should not require resources, got error: %v", err)
	}

	// Verify resources unchanged (still 0)
	if player.Resources.Wood != 0 || player.Resources.Brick != 0 {
		t.Error("Setup settlement should not deduct resources")
	}
}

func TestSetupSecondSettlement_GrantsResources(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP

	// Manually set a hex with known resource for testing
	// Find a non-desert hex and set it to wood
	var woodHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Resource != pb.TileResource_TILE_RESOURCE_DESERT {
			hex.Resource = pb.TileResource_TILE_RESOURCE_WOOD
			woodHex = hex
			break
		}
	}
	if woodHex == nil {
		t.Fatal("Could not find non-desert hex")
	}

	// Find a vertex adjacent to this hex
	var targetVertex *pb.Vertex
	for _, v := range state.Board.Vertices {
		for _, adj := range v.AdjacentHexes {
			if adj.Q == woodHex.Coord.Q && adj.R == woodHex.Coord.R {
				targetVertex = v
				break
			}
		}
		if targetVertex != nil {
			break
		}
	}

	if targetVertex == nil {
		t.Fatal("Test setup: couldn't find vertex adjacent to wood hex")
	}

	// Place first settlement (round 1 - should NOT grant resources)
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}
	state.CurrentTurn = 0

	// Use a different vertex for first settlement
	firstVertex := state.Board.Vertices[0]
	if firstVertex.Id == targetVertex.Id {
		firstVertex = state.Board.Vertices[1]
	}
	_ = PlaceSetupSettlement(state, "p1", firstVertex.Id)

	player := state.Players[0]
	woodAfterFirst := player.Resources.Wood

	// Place second settlement (round 2 - should grant resources from adjacent hexes)
	state.SetupPhase = &pb.SetupPhase{Round: 2, PlacementsInTurn: 0}
	err := PlaceSetupSettlement(state, "p1", targetVertex.Id)
	if err != nil {
		t.Fatalf("Failed to place second settlement: %v", err)
	}

	// Should have gained wood from the adjacent wood hex
	if player.Resources.Wood <= woodAfterFirst {
		t.Errorf("Second settlement should grant resources from adjacent hexes. Wood before: %d, after: %d",
			woodAfterFirst, player.Resources.Wood)
	}
}

func TestSetupToPlayingTransition(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}

	// 2 players × 2 rounds = 4 total turns
	totalTurns := len(state.Players) * 2

	if !IsSetupComplete(state, totalTurns) {
		t.Error("Setup should be complete after all turns")
	}

	// Transition should move to PLAYING status
	TransitionToPlaying(state)

	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		t.Errorf("Expected PLAYING status after setup, got %v", state.Status)
	}

	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_ROLL {
		t.Errorf("Expected ROLL phase after setup, got %v", state.TurnPhase)
	}

	if state.CurrentTurn != 0 {
		t.Errorf("Expected turn to reset to player 0 after setup, got %d", state.CurrentTurn)
	}
}

func TestSetupPhase_MustPlaceSettlementBeforeRoad(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}

	// Get a valid edge
	edgeID := state.Board.Edges[0].Id

	// Try to place road before settlement
	err := PlaceSetupRoad(state, "p1", edgeID)
	if err == nil {
		t.Error("Should not be able to place road before settlement in setup")
	}
}

func TestSetupPhase_RoadMustConnectToSettlement(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}

	// Place settlement at a known vertex
	settlementVertex := state.Board.Vertices[0]
	_ = PlaceSetupSettlement(state, "p1", settlementVertex.Id)

	// Find an edge that is NOT connected to this vertex
	var disconnectedEdge *pb.Edge
	for _, edge := range state.Board.Edges {
		connected := false
		for _, vID := range edge.Vertices {
			if vID == settlementVertex.Id {
				connected = true
				break
			}
		}
		if !connected {
			disconnectedEdge = edge
			break
		}
	}

	if disconnectedEdge == nil {
		t.Skip("Could not find a disconnected edge for this test")
	}

	// Try to place road at disconnected edge
	err := PlaceSetupRoad(state, "p1", disconnectedEdge.Id)
	if err == nil {
		t.Error("Road must connect to the just-placed settlement in setup")
	}
}

func TestSetupPhase_WrongPlayer(t *testing.T) {
	state := NewGameState("g1", "CODE", []string{"Alice", "Bob"}, []string{"p1", "p2"})
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{Round: 1, PlacementsInTurn: 0}
	state.CurrentTurn = 0 // Player 0's turn

	// Player 2 tries to place
	vertexID := state.Board.Vertices[0].Id
	err := PlaceSetupSettlement(state, "p2", vertexID)
	if err == nil {
		t.Error("Should not allow placement when it's not the player's turn")
	}
}
