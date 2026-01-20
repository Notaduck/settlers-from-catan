package game

import (
	"errors"
	"math/rand"
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// InitDevCardDeck creates a shuffled deck of 25 development cards
func InitDevCardDeck() []pb.DevCardType {
	deck := make([]pb.DevCardType, 0, 25)
	// 14 Knights
	for i := 0; i < 14; i++ {
		deck = append(deck, pb.DevCardType_DEV_CARD_TYPE_KNIGHT)
	}
	// 5 Victory Points
	for i := 0; i < 5; i++ {
		deck = append(deck, pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT)
	}
	// 2 Road Building
	for i := 0; i < 2; i++ {
		deck = append(deck, pb.DevCardType_DEV_CARD_TYPE_ROAD_BUILDING)
	}
	// 2 Year of Plenty
	for i := 0; i < 2; i++ {
		deck = append(deck, pb.DevCardType_DEV_CARD_TYPE_YEAR_OF_PLENTY)
	}
	// 2 Monopoly
	for i := 0; i < 2; i++ {
		deck = append(deck, pb.DevCardType_DEV_CARD_TYPE_MONOPOLY)
	}
	// Shuffle
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

// BuyDevCard deducts resources and draws a dev card for the player
func BuyDevCard(state *pb.GameState, playerID string) (pb.DevCardType, error) {
	// Find player
	var p *pb.PlayerState
	for _, pl := range state.Players {
		if pl.Id == playerID {
			p = pl
			break
		}
	}
	if p == nil {
		return pb.DevCardType_DEV_CARD_TYPE_UNSPECIFIED, errors.New("player not found")
	}

	// Check deck not empty
	if len(state.DevCardDeck) == 0 {
		return pb.DevCardType_DEV_CARD_TYPE_UNSPECIFIED, errors.New("dev card deck is empty")
	}

	// Deduct resources (ore, wheat, sheep)
	if !CanAfford(ResourceCountToMap(p.Resources), "development_card") {
		return pb.DevCardType_DEV_CARD_TYPE_UNSPECIFIED, errors.New("insufficient resources")
	}
	DeductResources(p.Resources, "development_card")

	// Draw card from deck
	cardType := state.DevCardDeck[0]
	state.DevCardDeck = state.DevCardDeck[1:]

	// Initialize dev_cards map if needed
	if p.DevCards == nil {
		p.DevCards = make(map[int32]int32)
	}
	// Initialize dev_cards_purchased_turn map if needed
	if p.DevCardsPurchasedTurn == nil {
		p.DevCardsPurchasedTurn = make(map[int32]int32)
	}

	// Add to player's hand
	p.DevCards[int32(cardType)] = p.DevCards[int32(cardType)] + 1
	p.DevCardCount++

	// Track when this card was purchased
	p.DevCardsPurchasedTurn[int32(cardType)] = state.TurnCounter

	// Track VP cards separately for victory calculation
	if cardType == pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT {
		p.VictoryPointCards = p.GetVictoryPointCards() + 1
	}

	return cardType, nil
}

// PlayDevCard plays a card from hand, applying its effect
func PlayDevCard(state *pb.GameState, playerID string, cardType pb.DevCardType, targetResource *pb.Resource, resources []pb.Resource) error {
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

	// Check player has the card
	if p.DevCards == nil || p.DevCards[int32(cardType)] == 0 {
		return errors.New("you don't have that card")
	}

	// Check timing rules - cards purchased this turn cannot be played (except Victory Point cards)
	if cardType != pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT {
		if p.DevCardsPurchasedTurn != nil && p.DevCardsPurchasedTurn[int32(cardType)] == state.TurnCounter {
			return errors.New("cannot play development card purchased this turn")
		}
	}

	// Apply card effect
	switch cardType {
	case pb.DevCardType_DEV_CARD_TYPE_KNIGHT:
		// Knight effect: move robber (handled separately in MoveRobber)
		// Just increment knight count here
		p.KnightsPlayed++
		RecalculateLargestArmy(state)

	case pb.DevCardType_DEV_CARD_TYPE_VICTORY_POINT:
		// VP cards are revealed, decrement hidden count
		if p.VictoryPointCards > 0 {
			p.VictoryPointCards--
		}
		// VP already counted in CheckVictory

	case pb.DevCardType_DEV_CARD_TYPE_ROAD_BUILDING:
		// Road building allows 2 free roads (handled in UI/handler)
		// This just validates the card was played

	case pb.DevCardType_DEV_CARD_TYPE_YEAR_OF_PLENTY:
		// Year of Plenty: take 2 resources from bank
		if len(resources) != 2 {
			return errors.New("year of plenty requires exactly 2 resources")
		}
		for _, res := range resources {
			AddResource(p.Resources, res, 1)
		}

	case pb.DevCardType_DEV_CARD_TYPE_MONOPOLY:
		// Monopoly: collect all of one resource from all players
		if targetResource == nil {
			return errors.New("monopoly requires a target resource")
		}
		totalCollected := int32(0)
		for _, pl := range state.Players {
			if pl.Id == playerID {
				continue // Skip self
			}
			amount := GetResourceAmount(pl.Resources, *targetResource)
			if amount > 0 {
				RemoveResource(pl.Resources, *targetResource, amount)
				totalCollected += amount
			}
		}
		AddResource(p.Resources, *targetResource, totalCollected)

	default:
		return errors.New("unknown card type")
	}

	// Remove card from hand
	p.DevCards[int32(cardType)]--
	if p.DevCards[int32(cardType)] == 0 {
		delete(p.DevCards, int32(cardType))
	}
	p.DevCardCount--

	// Also remove from purchase turn tracking when played
	if p.DevCardsPurchasedTurn != nil && p.DevCardsPurchasedTurn[int32(cardType)] > 0 {
		p.DevCardsPurchasedTurn[int32(cardType)]--
		if p.DevCardsPurchasedTurn[int32(cardType)] == 0 {
			delete(p.DevCardsPurchasedTurn, int32(cardType))
		}
	}

	// Check victory after playing card (in case of VP card)
	CheckVictory(state)

	return nil
}

// RecalculateLargestArmy updates the largest army holder based on current knight counts
func RecalculateLargestArmy(state *pb.GameState) {
	maxKnights := int32(0)
	claimers := []string{}

	// Find max knights and all players who have that count
	for _, p := range state.Players {
		if p.KnightsPlayed > maxKnights {
			maxKnights = p.KnightsPlayed
			claimers = []string{p.Id}
		} else if p.KnightsPlayed == maxKnights && maxKnights > 0 {
			claimers = append(claimers, p.Id)
		}
	}

	// Award to player with >= 3 knights
	if maxKnights >= 3 {
		// In case of tie, keep current holder if they're in the tie
		current := ""
		if state.LargestArmyPlayerId != nil {
			current = *state.LargestArmyPlayerId
		}
		for _, c := range claimers {
			if c == current {
				state.LargestArmyPlayerId = &current
				return
			}
		}
		// Otherwise, award to first claimer
		state.LargestArmyPlayerId = &claimers[0]
	} else {
		// No one qualifies, clear the bonus
		state.LargestArmyPlayerId = nil
	}
}

// Helper functions for resource manipulation
func GetResourceAmount(rc *pb.ResourceCount, res pb.Resource) int32 {
	if rc == nil {
		return 0
	}
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		return rc.Wood
	case pb.Resource_RESOURCE_BRICK:
		return rc.Brick
	case pb.Resource_RESOURCE_SHEEP:
		return rc.Sheep
	case pb.Resource_RESOURCE_WHEAT:
		return rc.Wheat
	case pb.Resource_RESOURCE_ORE:
		return rc.Ore
	default:
		return 0
	}
}

func AddResource(rc *pb.ResourceCount, res pb.Resource, amount int32) {
	if rc == nil || amount <= 0 {
		return
	}
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		rc.Wood += amount
	case pb.Resource_RESOURCE_BRICK:
		rc.Brick += amount
	case pb.Resource_RESOURCE_SHEEP:
		rc.Sheep += amount
	case pb.Resource_RESOURCE_WHEAT:
		rc.Wheat += amount
	case pb.Resource_RESOURCE_ORE:
		rc.Ore += amount
	}
}

func RemoveResource(rc *pb.ResourceCount, res pb.Resource, amount int32) {
	if rc == nil || amount <= 0 {
		return
	}
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		rc.Wood -= amount
	case pb.Resource_RESOURCE_BRICK:
		rc.Brick -= amount
	case pb.Resource_RESOURCE_SHEEP:
		rc.Sheep -= amount
	case pb.Resource_RESOURCE_WHEAT:
		rc.Wheat -= amount
	case pb.Resource_RESOURCE_ORE:
		rc.Ore -= amount
	}
}
