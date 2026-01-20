package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
	"testing"
)

func makeEdge(id, owner string, v1, v2 string) *pb.Edge {
	return &pb.Edge{Id: id, Vertices: []string{v1, v2}, Road: &pb.Road{OwnerId: owner}}
}
func makeVertex(id string, owner *string) *pb.Vertex {
	var b *pb.Building
	if owner != nil {
		b = &pb.Building{OwnerId: *owner, Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT}
	}
	return &pb.Vertex{Id: id, Building: b}
}

func TestSingleRoadSegment(t *testing.T) {
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	e := makeEdge("E1", "p1", "A", "B")
	board := &pb.BoardState{Edges: []*pb.Edge{e}, Vertices: []*pb.Vertex{v1, v2}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	if lengths["p1"] != 1 {
		t.Errorf("Expected single road segment length 1, got %d", lengths["p1"])
	}
}

func TestTwoConnectedRoads(t *testing.T) {
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2}, Vertices: []*pb.Vertex{v1, v2, v3}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	if lengths["p1"] != 2 {
		t.Errorf("Expected two connected roads length 2, got %d", lengths["p1"])
	}
}

func TestBranchingRoads_MaxBranch(t *testing.T) {
	vA := makeVertex("A", nil)
	vB := makeVertex("B", nil)
	vC := makeVertex("C", nil)
	vD := makeVertex("D", nil)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	e3 := makeEdge("E3", "p1", "B", "D")
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2, e3}, Vertices: []*pb.Vertex{vA, vB, vC, vD}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	if lengths["p1"] != 3 {
		t.Errorf("Expected max branch of 3, got %d", lengths["p1"])
	}
}

func TestOpponentSettlementBlocksRoad(t *testing.T) {
	// Opponent settlement at B (middle vertex) blocks the road chain
	blockOwner := "opp"
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", &blockOwner) // Opponent settlement at connecting vertex
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	vC := makeVertex("C", nil)
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2}, Vertices: []*pb.Vertex{v1, v2, vC}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	// Roads are broken at B, so longest road is 1 (either E1 or E2, but not both)
	if lengths["p1"] != 1 {
		t.Errorf("Blocked road should only be length 1, got %d", lengths["p1"])
	}
}

func TestOwnSettlementDoesNotBlock(t *testing.T) {
	owner := "p1"
	v1 := makeVertex("A", &owner)
	v2 := makeVertex("B", nil)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	vC := makeVertex("C", nil)
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2}, Vertices: []*pb.Vertex{v1, v2, vC}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	if lengths["p1"] != 2 {
		t.Errorf("Own settlement should not block, got %d", lengths["p1"])
	}
}

func TestCircularRoadCountsCorrectly(t *testing.T) {
	vA := makeVertex("A", nil)
	vB := makeVertex("B", nil)
	vC := makeVertex("C", nil)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	e3 := makeEdge("E3", "p1", "C", "A")
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2, e3}, Vertices: []*pb.Vertex{vA, vB, vC}}
	lengths := GetLongestRoadLengths(board, board.Vertices)
	if lengths["p1"] != 3 {
		t.Errorf("Circle road should count max path 3, got %d", lengths["p1"])
	}
}

func TestAwardLongestRoadBonus(t *testing.T) {
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	v4 := makeVertex("D", nil)
	v5 := makeVertex("E", nil)
	v6 := makeVertex("F", nil)

	e1 := makeEdge("E1", "p2", "A", "B")

	e2 := makeEdge("E2", "p2", "B", "C")

	e3 := makeEdge("E3", "p2", "C", "D")

	e4 := makeEdge("E4", "p2", "D", "E")

	e5 := makeEdge("E5", "p2", "E", "F")
	board := &pb.BoardState{Edges: []*pb.Edge{e1, e2, e3, e4, e5}, Vertices: []*pb.Vertex{v1, v2, v3, v4, v5, v6}}
	state := &pb.GameState{Board: board, LongestRoadPlayerId: nil}
	pid := GetLongestRoadPlayerId(state)
	if pid != "p2" {
		t.Errorf("p2 should have longest road, got %s", pid)
	}
}

func TestTieKeepsCurrentHolder(t *testing.T) {
	// Both players have exactly 5 roads (minimum for longest road)
	// p1 is current holder, tie should keep p1
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	v4 := makeVertex("D", nil)
	v5 := makeVertex("E", nil)
	v6 := makeVertex("F", nil)
	v7 := makeVertex("G", nil)
	v8 := makeVertex("H", nil)
	v9 := makeVertex("I", nil)
	v10 := makeVertex("J", nil)
	v11 := makeVertex("K", nil)
	v12 := makeVertex("L", nil)

	// p1's road: A-B-C-D-E-F (5 segments)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	e3 := makeEdge("E3", "p1", "C", "D")
	e4 := makeEdge("E4", "p1", "D", "E")
	e5 := makeEdge("E5", "p1", "E", "F")

	// p2's road: G-H-I-J-K-L (5 segments)
	e6 := makeEdge("E6", "p2", "G", "H")
	e7 := makeEdge("E7", "p2", "H", "I")
	e8 := makeEdge("E8", "p2", "I", "J")
	e9 := makeEdge("E9", "p2", "J", "K")
	e10 := makeEdge("E10", "p2", "K", "L")

	board := &pb.BoardState{
		Edges:    []*pb.Edge{e1, e2, e3, e4, e5, e6, e7, e8, e9, e10},
		Vertices: []*pb.Vertex{v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12},
	}
	state := &pb.GameState{Board: board, LongestRoadPlayerId: ptr("p1")}
	pid := GetLongestRoadPlayerId(state)
	if pid != "p1" {
		t.Errorf("Tie should keep current holder, got %s", pid)
	}
}

func TestLongestRoadTransfer_NewPlayerExceedsHolder(t *testing.T) {
	// Create a simple board with one long road for p1 (5 roads) and space for p2 to build 6
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	v4 := makeVertex("D", nil)
	v5 := makeVertex("E", nil)
	v6 := makeVertex("F", nil)
	v7 := makeVertex("G", nil)
	v8 := makeVertex("H", nil)

	// p1's initial road: A-B-C-D-E-F (5 segments)
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	e3 := makeEdge("E3", "p1", "C", "D")
	e4 := makeEdge("E4", "p1", "D", "E")
	e5 := makeEdge("E5", "p1", "E", "F")

	// p2's road that will exceed p1: G-H (will extend to 6)
	e6 := makeEdge("E6", "p2", "G", "H")

	board := &pb.BoardState{
		Edges:    []*pb.Edge{e1, e2, e3, e4, e5, e6},
		Vertices: []*pb.Vertex{v1, v2, v3, v4, v5, v6, v7, v8},
	}

	state := &pb.GameState{
		Board:               board,
		LongestRoadPlayerId: ptr("p1"), // p1 initially has longest road
		Status:              pb.GameStatus_GAME_STATUS_PLAYING,
		TurnPhase:           pb.TurnPhase_TURN_PHASE_BUILD,
		CurrentTurn:         1, // p2's turn
		Players: []*pb.PlayerState{
			{Id: "p1", Resources: &pb.ResourceCount{}},
			{Id: "p2", Resources: &pb.ResourceCount{Wood: 10, Brick: 10}}, // p2 has resources
		},
	}

	// Place p2 settlement at H to prepare for extending road
	v8.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p2"}

	// Add more edges for p2 to extend to 6 total road segments
	// Let's use separate vertices for p2's extension
	v9 := makeVertex("I", nil)
	v10 := makeVertex("J", nil)
	v11 := makeVertex("K", nil)
	v12 := makeVertex("L", nil)
	v13 := makeVertex("M", nil)

	e7 := makeEdge("E7", "", "H", "I")
	e8 := makeEdge("E8", "", "I", "J")
	e9 := makeEdge("E9", "", "J", "K")
	e10 := makeEdge("E10", "", "K", "L")
	e11 := makeEdge("E11", "", "L", "M") // Add one more edge to make 6 segments

	board.Edges = append(board.Edges, e7, e8, e9, e10, e11)
	board.Vertices = append(board.Vertices, v9, v10, v11, v12, v13)

	// Place roads to give p2 a 6-road path: G-H-I-J-K-L-M (6 segments)
	e7.Road = &pb.Road{OwnerId: "p2"}
	e8.Road = &pb.Road{OwnerId: "p2"}
	e9.Road = &pb.Road{OwnerId: "p2"}
	e10.Road = &pb.Road{OwnerId: "p2"}
	e11.Road = &pb.Road{OwnerId: "p2"} // p2 now has 6 total road segments

	// Update longest road bonus
	UpdateLongestRoadBonus(state)

	// p2 should now have longest road with 6 segments vs p1's 5
	if state.LongestRoadPlayerId == nil || *state.LongestRoadPlayerId != "p2" {
		currentHolder := ""
		if state.LongestRoadPlayerId != nil {
			currentHolder = *state.LongestRoadPlayerId
		}
		t.Errorf("expected p2 to have longest road, got %s", currentHolder)
	}
}

func TestLongestRoadTransfer_BrokenByOpponentSettlement(t *testing.T) {
	// p1 has a 5-road path, p2 places settlement to break it, longest road should be lost
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	v4 := makeVertex("D", nil)
	v5 := makeVertex("E", nil)
	v6 := makeVertex("F", nil)

	// p1's road that will be broken: A-B-C-D-E-F
	e1 := makeEdge("E1", "p1", "A", "B")
	e2 := makeEdge("E2", "p1", "B", "C")
	e3 := makeEdge("E3", "p1", "C", "D")
	e4 := makeEdge("E4", "p1", "D", "E")
	e5 := makeEdge("E5", "p1", "E", "F")

	// Add a setup for p2 to legally connect to vertex C
	// Use more intermediate vertices to satisfy distance rule
	v7 := makeVertex("G", nil)
	v8 := makeVertex("H", nil)
	e6 := makeEdge("E6", "p2", "G", "H") // p2's road
	e7 := makeEdge("E7", "p2", "H", "C") // p2's road connecting to C

	board := &pb.BoardState{
		Edges:    []*pb.Edge{e1, e2, e3, e4, e5, e6, e7},
		Vertices: []*pb.Vertex{v1, v2, v3, v4, v5, v6, v7, v8},
	}

	state := &pb.GameState{
		Board:               board,
		LongestRoadPlayerId: ptr("p1"), // p1 initially has longest road
		Status:              pb.GameStatus_GAME_STATUS_PLAYING,
		TurnPhase:           pb.TurnPhase_TURN_PHASE_BUILD,
		CurrentTurn:         1, // p2's turn
		Players: []*pb.PlayerState{
			{Id: "p1", Resources: &pb.ResourceCount{}, VictoryPoints: 2},
			{Id: "p2", Resources: &pb.ResourceCount{Wood: 1, Brick: 1, Sheep: 1, Wheat: 1}},
		},
	}

	// Give p2 an initial settlement at G to satisfy connectivity requirement
	v7.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p2"}

	// p2 places settlement at vertex C, breaking p1's road
	err := PlaceSettlement(state, "p2", "C")
	if err != nil {
		t.Fatalf("unexpected error placing settlement: %v", err)
	}

	// UpdateLongestRoadBonus should have been called automatically by PlaceSettlement
	// p1's road is now broken into A-B (2 segments) and D-E-F (3 segments)
	// Neither player should have longest road anymore
	if state.LongestRoadPlayerId != nil {
		t.Errorf("expected no one to have longest road after breaking, got %s", *state.LongestRoadPlayerId)
	}
}

func TestLongestRoadUpdate_TriggersVictoryCheck(t *testing.T) {
	// Player at 8 VP gets longest road (2 VP) to reach 10 VP and win
	v1 := makeVertex("A", nil)
	v2 := makeVertex("B", nil)
	v3 := makeVertex("C", nil)
	v4 := makeVertex("D", nil)
	v5 := makeVertex("E", nil)
	v6 := makeVertex("F", nil)

	// Add additional vertices for 8 total VP (7 settlements + 1 city = 8 VP)
	v7 := makeVertex("G", nil)
	v8 := makeVertex("H", nil)
	v9 := makeVertex("I", nil)
	v10 := makeVertex("J", nil)
	v11 := makeVertex("K", nil)
	v12 := makeVertex("L", nil)
	v13 := makeVertex("M", nil)

	// Create a 5-edge path - initially only the edges exist, no roads placed yet
	e1 := &pb.Edge{Id: "E1", Vertices: []string{"A", "B"}} // No road initially
	e2 := &pb.Edge{Id: "E2", Vertices: []string{"B", "C"}}
	e3 := &pb.Edge{Id: "E3", Vertices: []string{"C", "D"}}
	e4 := &pb.Edge{Id: "E4", Vertices: []string{"D", "E"}}
	e5 := &pb.Edge{Id: "E5", Vertices: []string{"E", "F"}}

	board := &pb.BoardState{
		Edges:    []*pb.Edge{e1, e2, e3, e4, e5},
		Vertices: []*pb.Vertex{v1, v2, v3, v4, v5, v6, v7, v8, v9, v10, v11, v12, v13},
	}

	state := &pb.GameState{
		Board:       board,
		Status:      pb.GameStatus_GAME_STATUS_PLAYING,
		TurnPhase:   pb.TurnPhase_TURN_PHASE_BUILD,
		CurrentTurn: 0,
		Players: []*pb.PlayerState{
			{Id: "p1", Resources: &pb.ResourceCount{Wood: 10, Brick: 10}, VictoryPoints: 8}, // 8 VP, will get 2 more from longest road
		},
	}

	// Place buildings to actually have 8 VP on the board
	// 6 settlements (6 VP) + 1 city (2 VP) = 8 VP
	v1.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"}  // VP 1
	v7.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"}  // VP 2
	v8.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"}  // VP 3
	v9.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"}  // VP 4
	v10.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"} // VP 5
	v11.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT, OwnerId: "p1"} // VP 6
	v12.Building = &pb.Building{Type: pb.BuildingType_BUILDING_TYPE_CITY, OwnerId: "p1"}       // VP 8 (city = 2 VP)

	// Place first 4 roads (not enough for longest road yet)
	e1.Road = &pb.Road{OwnerId: "p1"}
	e2.Road = &pb.Road{OwnerId: "p1"}
	e3.Road = &pb.Road{OwnerId: "p1"}
	e4.Road = &pb.Road{OwnerId: "p1"}

	err := PlaceRoad(state, "p1", "E5")
	if err != nil {
		t.Fatalf("unexpected error placing final road: %v", err)
	}

	// Game should be finished due to victory
	if state.Status != pb.GameStatus_GAME_STATUS_FINISHED {
		t.Errorf("expected game to be finished after reaching 10 VP, got status %v", state.Status)
	}

	// p1 should have longest road
	if state.LongestRoadPlayerId == nil || *state.LongestRoadPlayerId != "p1" {
		currentHolder := ""
		if state.LongestRoadPlayerId != nil {
			currentHolder = *state.LongestRoadPlayerId
		}
		t.Errorf("expected p1 to have longest road, got %s", currentHolder)
	}
}
