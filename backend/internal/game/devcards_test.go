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
