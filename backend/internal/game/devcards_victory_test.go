package game

import (
	pbb "settlers_from_catan/gen/proto/catan/v1"
	"strconv"
	"testing"
)

func TestPlayDevCard_VictoryStatus(t *testing.T) {
	tests := map[string]struct {
		settlements int
		cities      int
		vpCards     int32
		wantStatus  pbb.GameStatus
	}{
		"victory reached sets finished status": {
			settlements: 2,
			cities:      2,
			vpCards:     4,
			wantStatus:  pbb.GameStatus_GAME_STATUS_FINISHED,
		},
		"no victory keeps playing status": {
			settlements: 1,
			cities:      1,
			vpCards:     1,
			wantStatus:  pbb.GameStatus_GAME_STATUS_PLAYING,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Setup
			board := &pbb.BoardState{Vertices: []*pbb.Vertex{}}
			addBuildings := func(playerID string, buildingType pbb.BuildingType, count int) {
				t.Helper()
				for i := 0; i < count; i++ {
					board.Vertices = append(board.Vertices, &pbb.Vertex{
						Id:       playerID + ":" + strconv.Itoa(i),
						Building: &pbb.Building{OwnerId: playerID, Type: buildingType},
					})
				}
			}
			addBuildings("p1", pbb.BuildingType_BUILDING_TYPE_SETTLEMENT, tt.settlements)
			addBuildings("p1", pbb.BuildingType_BUILDING_TYPE_CITY, tt.cities)

			state := &pbb.GameState{
				Players: []*pbb.PlayerState{{
					Id:                "p1",
					VictoryPointCards: tt.vpCards,
					DevCardCount:      1,
					DevCards:          map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT): 1},
				}},
				Board:       board,
				Status:      pbb.GameStatus_GAME_STATUS_PLAYING,
				CurrentTurn: 0,
			}

			// Execute
			err := PlayDevCard(state, "p1", pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT, nil, nil)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if state.Status != tt.wantStatus {
				t.Fatalf("expected status %v, got %v", tt.wantStatus, state.Status)
			}
		})
	}
}
