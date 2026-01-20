package game

import (
	pbb "settlers_from_catan/gen/proto/catan/v1"
	"testing"
)

func TestBuyDevCardDeductsResourcesAndAddsCard(t *testing.T) {
	state := &pbb.GameState{
		DevCardDeck: InitDevCardDeck(),
		Players: []*pbb.PlayerState{{
			Id:        "P1",
			Resources: &pbb.ResourceCount{Ore: 1, Wheat: 1, Sheep: 1},
		}},
	}
	cardType, err := BuyDevCard(state, "P1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cardType == pbb.DevCardType_DEV_CARD_TYPE_UNSPECIFIED {
		t.Fatalf("expected valid card type, got unspecified")
	}
	p := state.Players[0]
	if p.Resources.Ore != 0 || p.Resources.Wheat != 0 || p.Resources.Sheep != 0 {
		t.Errorf("expected resources to be deducted, got: %+v", p.Resources)
	}
	if p.DevCardCount != 1 {
		t.Errorf("expected DevCardCount 1, got %d", p.DevCardCount)
	}
	if p.DevCards == nil {
		t.Errorf("expected DevCards to be initialized")
	}
}

func TestBuyDevCardInsufficientResources(t *testing.T) {
	state := &pbb.GameState{
		DevCardDeck: InitDevCardDeck(),
		Players:     []*pbb.PlayerState{{Id: "P2", Resources: &pbb.ResourceCount{Ore: 0, Wheat: 1, Sheep: 1}}},
	}
	_, err := BuyDevCard(state, "P2")
	if err == nil {
		t.Fatalf("expected error for insufficient resources, got nil")
	}
}

func TestPlayDevCardDecrementCount(t *testing.T) {
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{{
			Id:                "P3",
			VictoryPointCards: 1,
			DevCardCount:      1,
			DevCards:          map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT): 1},
		}},
	}
	err := PlayDevCard(state, "P3", pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := state.Players[0]
	if p.VictoryPointCards != 1 {
		t.Errorf("expected VictoryPointCards to remain 1 for scoring, got %d", p.VictoryPointCards)
	}
	if p.DevCardCount != 0 {
		t.Errorf("expected DevCardCount 0, got %d", p.DevCardCount)
	}
}

func TestPlayDevCardNoCardError(t *testing.T) {
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{{
			Id:                "P4",
			VictoryPointCards: 0,
			DevCardCount:      1,
			DevCards:          map[int32]int32{},
		}},
	}
	err := PlayDevCard(state, "P4", pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT, nil, nil)
	if err == nil {
		t.Fatalf("expected error for no VP card, got nil")
	}
}

func TestPlayKnightIncrementsKnightCount(t *testing.T) {
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{{
			Id:            "P5",
			KnightsPlayed: 0,
			DevCardCount:  1,
			DevCards:      map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 1},
		}},
	}
	err := PlayDevCard(state, "P5", pbb.DevCardType_DEV_CARD_TYPE_KNIGHT, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error playing Knight: %v", err)
	}
	p := state.Players[0]
	if p.KnightsPlayed != 1 {
		t.Errorf("expected KnightsPlayed to be 1, got %d", p.KnightsPlayed)
	}
	if p.DevCardCount != 0 {
		t.Errorf("expected DevCardCount to be 0, got %d", p.DevCardCount)
	}
}

func TestPlayKnightTriggersLargestArmyCheck(t *testing.T) {
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{
			{
				Id:            "P1",
				KnightsPlayed: 2,
				DevCardCount:  1,
				DevCards:      map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 1},
			},
			{
				Id:            "P2",
				KnightsPlayed: 1,
			},
		},
	}
	// Playing Knight should bring P1 to 3 knights and award Largest Army
	err := PlayDevCard(state, "P1", pbb.DevCardType_DEV_CARD_TYPE_KNIGHT, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Players[0].KnightsPlayed != 3 {
		t.Errorf("expected P1 KnightsPlayed to be 3, got %d", state.Players[0].KnightsPlayed)
	}
	if state.LargestArmyPlayerId == nil || *state.LargestArmyPlayerId != "P1" {
		t.Errorf("expected P1 to have Largest Army, got %v", state.LargestArmyPlayerId)
	}
}

func TestPlayYearOfPlentyGrants2Resources(t *testing.T) {
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{{
			Id:           "P6",
			DevCardCount: 1,
			DevCards:     map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_YEAR_OF_PLENTY): 1},
			Resources:    &pbb.ResourceCount{},
		}},
	}
	resources := []pbb.Resource{pbb.Resource_RESOURCE_WOOD, pbb.Resource_RESOURCE_BRICK}
	err := PlayDevCard(state, "P6", pbb.DevCardType_DEV_CARD_TYPE_YEAR_OF_PLENTY, nil, resources)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p := state.Players[0]
	if p.Resources.Wood != 1 || p.Resources.Brick != 1 {
		t.Errorf("expected Wood=1 Brick=1, got Wood=%d Brick=%d", p.Resources.Wood, p.Resources.Brick)
	}
}

func TestPlayMonopolyCollectsResources(t *testing.T) {
	targetRes := pbb.Resource_RESOURCE_WHEAT
	state := &pbb.GameState{
		Players: []*pbb.PlayerState{
			{
				Id:           "P7",
				DevCardCount: 1,
				DevCards:     map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_MONOPOLY): 1},
				Resources:    &pbb.ResourceCount{Wheat: 0},
			},
			{
				Id:        "P8",
				Resources: &pbb.ResourceCount{Wheat: 3},
			},
			{
				Id:        "P9",
				Resources: &pbb.ResourceCount{Wheat: 2},
			},
		},
	}
	err := PlayDevCard(state, "P7", pbb.DevCardType_DEV_CARD_TYPE_MONOPOLY, &targetRes, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// P7 should collect all wheat from P8 and P9
	if state.Players[0].Resources.Wheat != 5 {
		t.Errorf("expected P7 to have 5 wheat, got %d", state.Players[0].Resources.Wheat)
	}
	if state.Players[1].Resources.Wheat != 0 {
		t.Errorf("expected P8 to have 0 wheat, got %d", state.Players[1].Resources.Wheat)
	}
	if state.Players[2].Resources.Wheat != 0 {
		t.Errorf("expected P9 to have 0 wheat, got %d", state.Players[2].Resources.Wheat)
	}
}

func TestPlayKnightSameTurnRestriction(t *testing.T) {
	state := &pbb.GameState{
		TurnCounter: 5, // Current turn counter
		Players: []*pbb.PlayerState{{
			Id:            "P1",
			KnightsPlayed: 0,
			DevCardCount:  1,
			DevCards:      map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 1},
			// Simulate card purchased this turn
			DevCardsPurchasedTurn: map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 5},
		}},
	}

	// Should not be able to play knight card purchased this turn
	err := PlayDevCard(state, "P1", pbb.DevCardType_DEV_CARD_TYPE_KNIGHT, nil, nil)
	if err == nil {
		t.Fatalf("expected error for playing dev card purchased same turn")
	}
	if err.Error() != "cannot play development card purchased this turn" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Player should still have the card
	if state.Players[0].DevCards[int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT)] != 1 {
		t.Errorf("card should not have been removed from hand")
	}
}

func TestPlayKnightPreviousTurnAllowed(t *testing.T) {
	state := &pbb.GameState{
		TurnCounter: 6, // Current turn counter
		Players: []*pbb.PlayerState{{
			Id:            "P1",
			KnightsPlayed: 0,
			DevCardCount:  1,
			DevCards:      map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 1},
			// Card purchased previous turn
			DevCardsPurchasedTurn: map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_KNIGHT): 5},
		}},
	}

	// Should be able to play knight card purchased previous turn
	err := PlayDevCard(state, "P1", pbb.DevCardType_DEV_CARD_TYPE_KNIGHT, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error playing knight from previous turn: %v", err)
	}

	// Player should have played the card
	if state.Players[0].KnightsPlayed != 1 {
		t.Errorf("expected KnightsPlayed to be 1, got %d", state.Players[0].KnightsPlayed)
	}
	if state.Players[0].DevCardCount != 0 {
		t.Errorf("expected DevCardCount to be 0, got %d", state.Players[0].DevCardCount)
	}
}

func TestPlayVictoryPointCardSameTurnAllowed(t *testing.T) {
	state := &pbb.GameState{
		TurnCounter: 5, // Current turn counter
		Players: []*pbb.PlayerState{{
			Id:                "P1",
			VictoryPointCards: 1,
			DevCardCount:      1,
			DevCards:          map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT): 1},
			// VP card purchased this turn - should still be playable
			DevCardsPurchasedTurn: map[int32]int32{int32(pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT): 5},
		}},
	}

	// VP cards should be playable same turn they're purchased
	err := PlayDevCard(state, "P1", pbb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error playing VP card same turn: %v", err)
	}

	// VP card should have been played
	if state.Players[0].VictoryPointCards != 1 {
		t.Errorf("expected VictoryPointCards to remain 1, got %d", state.Players[0].VictoryPointCards)
	}
}

func TestBuyDevCardTracksPurchaseTurn(t *testing.T) {
	state := &pbb.GameState{
		TurnCounter: 10,
		DevCardDeck: InitDevCardDeck(),
		Players: []*pbb.PlayerState{{
			Id:        "P1",
			Resources: &pbb.ResourceCount{Ore: 1, Wheat: 1, Sheep: 1},
		}},
	}

	cardType, err := BuyDevCard(state, "P1")
	if err != nil {
		t.Fatalf("unexpected error buying dev card: %v", err)
	}

	// Check that purchase turn was tracked
	p := state.Players[0]
	if p.DevCardsPurchasedTurn == nil {
		t.Fatalf("DevCardsPurchasedTurn map should be initialized")
	}
	if p.DevCardsPurchasedTurn[int32(cardType)] != state.TurnCounter {
		t.Errorf("expected purchase turn %d, got %d", state.TurnCounter, p.DevCardsPurchasedTurn[int32(cardType)])
	}
}

func TestPlayRoadBuildingCard_AllowsTwoFreeRoads(t *testing.T) {
	tests := map[string]struct {
		setup              func(*pbb.GameState)
		wantRemainingRoads int32
	}{
		"playing road building card sets remaining to 2": {
			setup: func(state *pbb.GameState) {
				// Player already has the road building card
				state.Players[0].DevCards[int32(pbb.DevCardType_DEV_CARD_TYPE_ROAD_BUILDING)] = 1
				state.Players[0].DevCardCount = 1
			},
			wantRemainingRoads: 2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			state := &pbb.GameState{
				TurnCounter: 5,
				Players: []*pbb.PlayerState{{
					Id:       "p1",
					DevCards: make(map[int32]int32),
				}},
			}
			tt.setup(state)

			// Play the road building card
			err := PlayDevCard(state, "p1", pbb.DevCardType_DEV_CARD_TYPE_ROAD_BUILDING, nil, nil)
			if err != nil {
				t.Fatalf("unexpected error playing road building card: %v", err)
			}

			// Check that remaining roads is set correctly
			p := state.Players[0]
			if p.RoadBuildingRoadsRemaining != tt.wantRemainingRoads {
				t.Errorf("expected %d remaining roads, got %d", tt.wantRemainingRoads, p.RoadBuildingRoadsRemaining)
			}
		})
	}
}

func TestRoadBuildingCard_SkipsResourceCost(t *testing.T) {
	// Create a minimal board with edges and vertices
	board := GenerateBoard()
	state := &pbb.GameState{
		Board:       board,
		Status:      pbb.GameStatus_GAME_STATUS_PLAYING,
		TurnPhase:   pbb.TurnPhase_TURN_PHASE_BUILD,
		CurrentTurn: 0,
		Players: []*pbb.PlayerState{{
			Id: "p1",
			Resources: &pbb.ResourceCount{
				Wood:  0, // No resources for road building
				Brick: 0,
			},
			RoadBuildingRoadsRemaining: 2, // Has road building active
		}},
	}

	// Place a settlement first to satisfy connectivity
	targetVertex := board.Vertices[0] // First vertex
	targetVertex.Building = &pbb.Building{
		Type:    pbb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: "p1",
	}

	// Find an edge connected to that vertex
	var targetEdge *pbb.Edge
	for _, edge := range board.Edges {
		for _, vId := range edge.Vertices {
			if vId == targetVertex.Id {
				targetEdge = edge
				break
			}
		}
		if targetEdge != nil {
			break
		}
	}

	if targetEdge == nil {
		t.Fatalf("could not find edge connected to vertex")
	}

	// Should be able to place road without resources
	err := PlaceRoad(state, "p1", targetEdge.Id)
	if err != nil {
		t.Fatalf("expected road building to skip resource cost, got error: %v", err)
	}

	// Check that road was placed
	if targetEdge.Road == nil || targetEdge.Road.OwnerId != "p1" {
		t.Errorf("expected road to be placed")
	}

	// Check that remaining count was decremented
	if state.Players[0].RoadBuildingRoadsRemaining != 1 {
		t.Errorf("expected remaining roads to be decremented to 1, got %d", state.Players[0].RoadBuildingRoadsRemaining)
	}

	// Check that resources weren't deducted (they were 0 to begin with)
	p := state.Players[0]
	if p.Resources.Wood != 0 || p.Resources.Brick != 0 {
		t.Errorf("expected resources to remain 0, got wood: %d, brick: %d", p.Resources.Wood, p.Resources.Brick)
	}
}
