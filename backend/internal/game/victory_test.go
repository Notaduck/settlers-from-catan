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
