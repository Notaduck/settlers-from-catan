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
