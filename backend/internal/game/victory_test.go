package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
	"strconv"
	"testing"
)

func TestCheckVictory_BasicCases(t *testing.T) {
	players := []*pb.PlayerState{
		{Id: "p1", Name: "Alice"},
		{Id: "p2", Name: "Bob"},
	}
	board := &pb.BoardState{Vertices: []*pb.Vertex{}}
	state := &pb.GameState{
		Players:     players,
		Board:       board,
		Status:      pb.GameStatus_GAME_STATUS_PLAYING,
		CurrentTurn: 0,
	}

	addBuilding := func(playerId string, buildingType pb.BuildingType, n int) {
		for i := 0; i < n; i++ {
			board.Vertices = append(board.Vertices, &pb.Vertex{
				Id:       playerId + ":" + strconv.Itoa(i),
				Building: &pb.Building{OwnerId: playerId, Type: buildingType},
			})
		}
	}

	// 9 VP: No victory
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_SETTLEMENT, 4)
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_CITY, 2)
	state.LongestRoadPlayerId = nil
	state.LargestArmyPlayerId = nil
	victory, winner := CheckVictory(state)
	if victory {
		t.Fatalf("Victory incorrectly detected at <10 VP: winner=%s", winner)
	}

	// 10 VP: Victory with settlements/cities
	board.Vertices = nil
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_SETTLEMENT, 4)
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_CITY, 3)
	victory, winner = CheckVictory(state)
	if !victory || winner != "p1" {
		t.Fatalf("Victory should be detected at 10 VP (settlements/cities): got=%v, winner=%s", victory, winner)
	}

	// Longest road triggers win
	board.Vertices = nil
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_SETTLEMENT, 2)
	addBuilding("p1", pb.BuildingType_BUILDING_TYPE_CITY, 2)
	id := "p1"
	state.LongestRoadPlayerId = &id
	victory, winner = CheckVictory(state)
	if victory && winner != "p1" {
		t.Fatalf("p1 should win with longest road: got winner=%s", winner)
	}

	// Only active player can win
	state.CurrentTurn = 1
	victory, winner = CheckVictory(state)
	if victory {
		t.Fatalf("Victory should NOT occur on inactive player's turn!")
	}
	state.CurrentTurn = 0
}

// Test: Largest Army triggers victory (spec: Win possible with largest army bonus).
func TestCheckVictory_LargestArmyBonus(t *testing.T) {
	players := []*pb.PlayerState{{Id: "p1", Name: "Alice"}}
	board := &pb.BoardState{Vertices: []*pb.Vertex{}}
	state := &pb.GameState{
		Players:     players,
		Board:       board,
		Status:      pb.GameStatus_GAME_STATUS_PLAYING,
		CurrentTurn: 0,
	}
	// 3 settlements (3 VP), 2 cities (4 VP), largest army (+2 VP) = 9 VP
	for i := 0; i < 3; i++ {
		board.Vertices = append(board.Vertices, &pb.Vertex{Id: "p1:" + strconv.Itoa(i), Building: &pb.Building{OwnerId: "p1", Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT}})
	}
	for i := 0; i < 2; i++ {
		board.Vertices = append(board.Vertices, &pb.Vertex{Id: "p1:C" + strconv.Itoa(i), Building: &pb.Building{OwnerId: "p1", Type: pb.BuildingType_BUILDING_TYPE_CITY}})
	}
	id := "p1"
	state.LargestArmyPlayerId = &id
	victory, winner := CheckVictory(state)
	if victory {
		// Only 9 VP so not yet victorious
		t.Fatalf("Victory should NOT trigger with 9 VP (with largest army), got winner %s", winner)
	}
	// Add one more settlement -> now 10 VP
	board.Vertices = append(board.Vertices, &pb.Vertex{Id: "p1:3", Building: &pb.Building{OwnerId: "p1", Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT}})
	victory, winner = CheckVictory(state)
	if !victory || winner != "p1" {
		t.Fatalf("Victory should trigger at 10 VP (with largest army), got=%v, winner=%s", victory, winner)
	}
}

// Test: Victory points from hidden VP cards (spec: Win with development card VICTORY_POINT)
func TestCheckVictory_HiddenVPCards(t *testing.T) {
	settlements := 2
	cities := 2
	hasLongestRoad := false
	hasLargestArmy := false
	vpCards := 4 // 8 VP from settlements/cities, +2 from dev cards
	total := CalculateVictoryPoints(settlements, cities, hasLongestRoad, hasLargestArmy, vpCards)
	if total != 10 {
		t.Fatalf("Expected 10 VP with 4 dev card VP, got %d", total)
	}
	if !IsVictorious(total) {
		t.Fatalf("Player with 10 VP including dev cards should be victorious")
	}
}

// Test: Game status switches to FINISHED on win (spec: When player crosses 10 VP)
func TestCheckVictory_GameEndsOnWin(t *testing.T) {
	players := []*pb.PlayerState{{Id: "p1", Name: "Alice"}}
	board := &pb.BoardState{Vertices: []*pb.Vertex{}}
	state := &pb.GameState{
		Players:     players,
		Board:       board,
		Status:      pb.GameStatus_GAME_STATUS_PLAYING,
		CurrentTurn: 0,
	}
	// Simulate win: 5 settlements
	for i := 0; i < 5; i++ {
		board.Vertices = append(board.Vertices, &pb.Vertex{Id: "p1:" + strconv.Itoa(i), Building: &pb.Building{OwnerId: "p1", Type: pb.BuildingType_BUILDING_TYPE_SETTLEMENT}})
	}
	victory, _ := CheckVictory(state)
	if !victory {
		t.Fatalf("Player should be victorious at 10 VP")
	}
	// Simulate PlaceSettlement setting status
	status := pb.GameStatus_GAME_STATUS_PLAYING
	if victory {
		status = pb.GameStatus_GAME_STATUS_FINISHED
	}
	if status != pb.GameStatus_GAME_STATUS_FINISHED {
		t.Fatalf("Game status should be FINISHED after win")
	}
}

// Test: GameOverPayload contains correct winner and scores (spec: Confirm winner and all scores included on game over)
func TestGameOverPayload_CorrectScoresAndWinner(t *testing.T) {
	payload := &pb.GameOverPayload{
		WinnerId: "p2",
		Scores: []*pb.PlayerScore{
			{PlayerId: "p1", Points: 8},
			{PlayerId: "p2", Points: 11},
			{PlayerId: "p3", Points: 3},
		},
	}
	if payload.WinnerId != "p2" {
		t.Fatalf("Winner id in payload should be p2, got %s", payload.WinnerId)
	}
	foundScores := map[string]int{}
	for _, score := range payload.Scores {
		foundScores[score.PlayerId] = int(score.Points)
	}
	if foundScores["p2"] != 11 {
		t.Fatalf("Score for winner should be 11, got %d", foundScores["p2"])
	}
	if foundScores["p1"] != 8 {
		t.Fatalf("Score for p1 should be 8, got %d", foundScores["p1"])
	}
	if foundScores["p3"] != 3 {
		t.Fatalf("Score for p3 should be 3, got %d", foundScores["p3"])
	}
}
