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
	if p.VictoryPointCards != 0 {
		t.Errorf("expected VictoryPointCards to decrement to 0, got %d", p.VictoryPointCards)
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
