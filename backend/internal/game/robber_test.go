package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
	"testing"
)

func makeTestPlayer(id string, resources *pb.ResourceCount) *pb.PlayerState {
	return &pb.PlayerState{
		Id:        id,
		Resources: resources,
	}
}

func makeRobberVertex(owner string, adj []*pb.HexCoord) *pb.Vertex {
	return &pb.Vertex{
		Building:      &pb.Building{OwnerId: owner, Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT},
		AdjacentHexes: adj,
	}
}

func TestMoveRobber(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	newHex := &pb.HexCoord{Q: 1, R: 1}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robberHex}, {Coord: newHex}},
			RobberHex: robberHex,
		},
		RobberPhase: &pb.RobberPhase{
			MovePendingPlayerId: ptr("p1"),
		},
		Players: []*pb.PlayerState{{Id: "p1"}},
	}

	err := MoveRobber(state, "p1", newHex)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if state.Board.RobberHex.Q != newHex.Q || state.Board.RobberHex.R != newHex.R {
		t.Errorf("Robber not moved correctly")
	}
	state.RobberPhase = &pb.RobberPhase{MovePendingPlayerId: ptr("p1")}
	// Same hex
	state.Board.RobberHex = newHex
	state.RobberPhase.MovePendingPlayerId = ptr("p1")
	if err := MoveRobber(state, "p1", newHex); err == nil {
		t.Error("Should fail if moving to same hex")
	}
	// Wrong player
	state.Board.RobberHex = robberHex
	state.RobberPhase.MovePendingPlayerId = ptr("p1")
	if err := MoveRobber(state, "p2", newHex); err == nil {
		t.Error("Should fail if not current player")
	}
}

func TestMoveRobber_RequiresDiscards(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	newHex := &pb.HexCoord{Q: 1, R: 1}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robberHex}, {Coord: newHex}},
			RobberHex: robberHex,
		},
		Players: []*pb.PlayerState{{Id: "p1", Resources: &pb.ResourceCount{Wood: 1}}},
		RobberPhase: &pb.RobberPhase{
			DiscardPending:      []string{"p1"},
			DiscardRequired:     map[string]int32{"p1": 1},
			MovePendingPlayerId: ptr("p1"),
		},
	}

	if err := MoveRobber(state, "p1", newHex); err == nil {
		t.Fatal("Expected move to fail before discards complete")
	}
	if err := DiscardCards(state, "p1", &pb.ResourceCount{Wood: 1}); err != nil {
		t.Fatalf("Unexpected discard error: %v", err)
	}
	if err := MoveRobber(state, "p1", newHex); err != nil {
		t.Fatalf("Expected move to succeed after discards: %v", err)
	}
	if state.RobberPhase != nil {
		t.Errorf("Expected robber phase to clear with no victims")
	}
}

func TestStealFromPlayer(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	adjHex := &pb.HexCoord{Q: 0, R: 0}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robberHex}},
			RobberHex: robberHex,
			Vertices: []*pb.Vertex{
				makeRobberVertex("vic", []*pb.HexCoord{adjHex}),
			},
		},
		Players: []*pb.PlayerState{
			{Id: "thief", Resources: &pb.ResourceCount{}},
			{Id: "vic", Resources: &pb.ResourceCount{Wood: 2}},
		},
		RobberPhase: &pb.RobberPhase{StealPendingPlayerId: ptr("thief")},
	}
	got, err := StealFromPlayer(state, "thief", "vic")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if got != pb.Resource_RESOURCE_WOOD {
		t.Errorf("Should steal wood resource (only one)")
	}
	// Victim has no more cards
	state.Players[1].Resources.Wood = 0
	state.RobberPhase = &pb.RobberPhase{StealPendingPlayerId: ptr("thief")}
	state.RobberPhase.StealPendingPlayerId = ptr("thief")
	if _, err := StealFromPlayer(state, "thief", "vic"); err == nil {
		t.Error("Should fail when victim empty")
	}
	// Wrong player id
	state.Players[1].Resources.Wood = 1
	state.RobberPhase = &pb.RobberPhase{StealPendingPlayerId: ptr("thief")}
	state.RobberPhase.StealPendingPlayerId = ptr("thief")
	if _, err := StealFromPlayer(state, "nope", "vic"); err == nil {
		t.Error("Should fail with bad thief id")
	}
}

func TestMoveRobber_SetsStealPending(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	newHex := &pb.HexCoord{Q: 1, R: 1}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robberHex}, {Coord: newHex}},
			RobberHex: robberHex,
			Vertices: []*pb.Vertex{
				makeRobberVertex("vic", []*pb.HexCoord{newHex}),
			},
		},
		Players: []*pb.PlayerState{{Id: "thief"}, {Id: "vic"}},
		RobberPhase: &pb.RobberPhase{
			MovePendingPlayerId: ptr("thief"),
			DiscardRequired:     map[string]int32{},
		},
	}

	if err := MoveRobber(state, "thief", newHex); err != nil {
		t.Fatalf("Unexpected move error: %v", err)
	}
	if state.RobberPhase == nil || state.RobberPhase.StealPendingPlayerId == nil {
		t.Fatal("Expected steal pending after move")
	}
	if *state.RobberPhase.StealPendingPlayerId != "thief" {
		t.Errorf("Expected steal pending for thief")
	}
}

func TestStealFromPlayer_NotAdjacent(t *testing.T) {
	robHex := &pb.HexCoord{Q: 1, R: 1}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robHex}},
			RobberHex: robHex,
			Vertices: []*pb.Vertex{
				makeRobberVertex("vic", []*pb.HexCoord{{Q: 2, R: 2}}), // Not adjacent
			},
		},
		Players: []*pb.PlayerState{
			{Id: "thief", Resources: &pb.ResourceCount{}},
			{Id: "vic", Resources: &pb.ResourceCount{Wood: 1}},
		},
		RobberPhase: &pb.RobberPhase{StealPendingPlayerId: ptr("thief")},
	}
	_, err := StealFromPlayer(state, "thief", "vic")
	if err == nil {
		t.Error("Should fail—not adjacent")
	}
}

func TestStealFromPlayer_ClearsRobberPhase(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	adjHex := &pb.HexCoord{Q: 0, R: 0}
	state := &pb.GameState{
		Board: &pb.BoardState{
			Hexes:     []*pb.Hex{{Coord: robberHex}},
			RobberHex: robberHex,
			Vertices: []*pb.Vertex{
				makeRobberVertex("vic", []*pb.HexCoord{adjHex}),
			},
		},
		Players: []*pb.PlayerState{
			{Id: "thief", Resources: &pb.ResourceCount{}},
			{Id: "vic", Resources: &pb.ResourceCount{Wood: 1}},
		},
		RobberPhase: &pb.RobberPhase{StealPendingPlayerId: ptr("thief")},
	}

	if _, err := StealFromPlayer(state, "thief", "vic"); err != nil {
		t.Fatalf("Unexpected steal error: %v", err)
	}
	if state.RobberPhase != nil {
		t.Errorf("Expected robber phase to clear after steal")
	}
}

func TestDiscardCards(t *testing.T) {
	state := &pb.GameState{
		Players: []*pb.PlayerState{
			{Id: "p1", Resources: &pb.ResourceCount{Wood: 5, Ore: 2}},
		},
		RobberPhase: &pb.RobberPhase{
			DiscardPending:  []string{"p1"},
			DiscardRequired: map[string]int32{"p1": 4},
		},
	}
	toDiscard := &pb.ResourceCount{Wood: 3, Ore: 1}
	err := DiscardCards(state, "p1", toDiscard)
	if err != nil {
		t.Fatalf("Expected no error: %v", err)
	}
	if countTotalResources(state.Players[0].Resources) != 3 {
		t.Errorf("Player should have 3 cards left, got %d", countTotalResources(state.Players[0].Resources))
	}
	if len(state.RobberPhase.DiscardPending) != 0 {
		t.Errorf("DiscardPending should be empty")
	}
}

func TestDiscardCards_Errors(t *testing.T) {
	state := &pb.GameState{
		Players: []*pb.PlayerState{
			{Id: "p1", Resources: &pb.ResourceCount{Wood: 3}},
		},
		RobberPhase: &pb.RobberPhase{
			DiscardPending:  []string{"p1"},
			DiscardRequired: map[string]int32{"p1": 2},
		},
	}
	bad1 := &pb.ResourceCount{Wood: 3} // wrong count
	err := DiscardCards(state, "p1", bad1)
	if err == nil {
		t.Error("Should error for discarding too many")
	}
	bad2 := &pb.ResourceCount{Wood: 1} // not enough
	err = DiscardCards(state, "p1", bad2)
	if err == nil {
		t.Error("Should error for discarding too few")
	}
}

// Test helpers
func ptr(s string) *string { return &s }

// Test: table-driven for all resource types for stealing
func TestStealFromPlayer_MultiResource(t *testing.T) {
	robberHex := &pb.HexCoord{Q: 0, R: 0}
	adjHex := &pb.HexCoord{Q: 0, R: 0}
	victimID := "victim"
	thiefID := "thief"
	for i := 0; i < 5; i++ {
		victim := &pb.PlayerState{Id: victimID, Resources: &pb.ResourceCount{Wood: 1, Brick: 1, Sheep: 1, Wheat: 1, Ore: 1}}
		thief := &pb.PlayerState{Id: thiefID, Resources: &pb.ResourceCount{}}
		state := &pb.GameState{
			Board: &pb.BoardState{
				Hexes:     []*pb.Hex{{Coord: robberHex}},
				RobberHex: robberHex,
				Vertices:  []*pb.Vertex{makeRobberVertex(victimID, []*pb.HexCoord{adjHex})},
			},
			Players:     []*pb.PlayerState{thief, victim},
			RobberPhase: &pb.RobberPhase{StealPendingPlayerId: ptr(thiefID)},
		}
		chooser := func(n int) int { return i % n }
		stolenRes, err := StealFromPlayer(state, thiefID, victimID, chooser)
		if err != nil {
			t.Errorf("Expected successful steal, got %v", err)
		}
		// The test just ensures one card is removed/transferred and is plausible
		nV := countTotalResources(state.Players[1].Resources)
		nT := countTotalResources(state.Players[0].Resources)
		if nV != 4 || nT != 1 {
			t.Errorf("Expected victim 4, thief 1, got %d/%d", nV, nT)
		}
		_ = stolenRes // checked for error
	}
}
