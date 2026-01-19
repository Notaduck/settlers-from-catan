package game

import (
	"errors"
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// BuyDevCard deducts resources and draws a dev card for the player
func BuyDevCard(state *pb.GameState, playerID string) error {
	// Find player
	var p *pb.PlayerState
	for _, pl := range state.Players {
		if pl.Id == playerID {
			p = pl
			break
		}
	}
	if p == nil {
		return errors.New("player not found")
	}
	// Deduct resources (ore, wheat, sheep)
	if !CanAfford(ResourceCountToMap(p.Resources), "development_card") {
		return errors.New("insufficient resources")
	}
	DeductResources(p.Resources, "development_card")
	// Simulate deck: randomly draw a card (just VP for this logic)
	cardType := pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT // TODO: Proper deck logic
	if cardType == pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT {
		p.VictoryPointCards = p.GetVictoryPointCards() + 1
	}
	p.DevCardCount++
	return nil
}

// PlayDevCard plays a card from hand, applying its effect. Only VP relevant here.
func PlayDevCard(state *pb.GameState, playerID string, cardType pb.DevCardType) error {
	var p *pb.PlayerState
	for _, pl := range state.Players {
		if pl.Id == playerID {
			p = pl
			break
		}
	}
	if p == nil {
		return errors.New("player not found")
	}
	if cardType == pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT {
		if p.GetVictoryPointCards() == 0 {
			return errors.New("no VP card to play")
		}
		p.VictoryPointCards = p.GetVictoryPointCards() - 1
		// Reveal, score only at game over
	}
	// TODO: Handle other card types
	p.DevCardCount--
	return nil
}
