package game

import (
	"fmt"
	"github.com/google/uuid"
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// ProposeTrade creates a trade offer and appends it to pending_trades if valid.
func ProposeTrade(state *pb.GameState, proposerID string, targetID *string, offering, requesting *pb.ResourceCount) (string, error) {
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return "", ErrWrongPhase
	}
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		return "", ErrWrongPhase
	}
	currentPlayer := state.Players[state.CurrentTurn]
	if currentPlayer.Id != proposerID {
		return "", ErrNotYourTurn
	}

	if !canFulfillOffer(currentPlayer.Resources, offering) {
		return "", ErrInsufficientResources
	}

	trade := &pb.TradeOffer{
		Id:         uuid.New().String(),
		ProposerId: proposerID,
		TargetId:   targetID, // *string matches proto field
		Offering:   offering,
		Requesting: requesting,
		Status:     pb.TradeStatus_TRADE_STATUS_PENDING,
	}
	state.PendingTrades = append(state.PendingTrades, trade)
	return trade.Id, nil
}

// RespondTrade accepts or declines a trade offer. Accepted trades exchange resources.
func RespondTrade(state *pb.GameState, tradeID, responderID string, accept bool) error {
	for _, t := range state.PendingTrades {
		if t.Id == tradeID && t.Status == pb.TradeStatus_TRADE_STATUS_PENDING {
			if !accept {
				t.Status = pb.TradeStatus_TRADE_STATUS_REJECTED
				return nil
			}

			if t.TargetId != nil && (responderID == "" || *t.TargetId != responderID) {
				return ErrNotTradeParticipant
			}

			from := playerByID(state, t.ProposerId)
			to := playerByID(state, responderID)
			if from == nil || to == nil {
				return ErrInvalidPlayer
			}
			if !canFulfillOffer(from.Resources, t.Offering) || !canFulfillOffer(to.Resources, t.Requesting) {
				return ErrInsufficientResources
			}

			deductOffer(from.Resources, t.Offering)
			addOffer(from.Resources, t.Requesting)
			deductOffer(to.Resources, t.Requesting)
			addOffer(to.Resources, t.Offering)
			t.Status = pb.TradeStatus_TRADE_STATUS_ACCEPTED
			return nil
		}
	}
	return ErrTradeNotFound
}

// ExpireOldTrades removes all expired or resolved trades (after turn ends).
func ExpireOldTrades(state *pb.GameState) {
	pending := []*pb.TradeOffer{}
	for _, t := range state.PendingTrades {
		if t.Status == pb.TradeStatus_TRADE_STATUS_PENDING {
			pending = append(pending, t)
		}
	}
	state.PendingTrades = pending
}

// BankTrade exchanges 4 (or port ratio) identical resources for 1 of choice
func BankTrade(state *pb.GameState, playerID string, offering *pb.ResourceCount, requested pb.Resource) error {
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return ErrWrongPhase
	}
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		return ErrWrongPhase
	}
	currentPlayer := state.Players[state.CurrentTurn]
	if currentPlayer.Id != playerID {
		return ErrNotYourTurn
	}

	offerCount, offerRes := countSingleResourceOffer(offering)
	// TODO: Port support for ratios < 4
	if offerCount < 4 {
		return ErrInsufficientResources
	}
	if int(playerResource(currentPlayer.Resources, offerRes)) < offerCount {
		return ErrInsufficientResources
	}
	deductResource(currentPlayer.Resources, offerRes, offerCount)
	addResource(currentPlayer.Resources, requested, 1)
	return nil
}

// ========== Helpers ==========

func playerByID(state *pb.GameState, id string) *pb.PlayerState {
	for _, p := range state.Players {
		if p.Id == id {
			return p
		}
	}
	return nil
}

func canFulfillOffer(resources *pb.ResourceCount, offer *pb.ResourceCount) bool {
	return resources.Wood >= offer.Wood &&
		resources.Brick >= offer.Brick &&
		resources.Sheep >= offer.Sheep &&
		resources.Wheat >= offer.Wheat &&
		resources.Ore >= offer.Ore
}

func deductOffer(resources *pb.ResourceCount, offer *pb.ResourceCount) {
	resources.Wood -= offer.Wood
	resources.Brick -= offer.Brick
	resources.Sheep -= offer.Sheep
	resources.Wheat -= offer.Wheat
	resources.Ore -= offer.Ore
}

func addOffer(resources *pb.ResourceCount, offer *pb.ResourceCount) {
	resources.Wood += offer.Wood
	resources.Brick += offer.Brick
	resources.Sheep += offer.Sheep
	resources.Wheat += offer.Wheat
	resources.Ore += offer.Ore
}

// For bank trades: check only 1 nonzero resource is being offered
func countSingleResourceOffer(c *pb.ResourceCount) (int, pb.Resource) {
	var res pb.Resource
	cnt := 0
	for r, val := range map[pb.Resource]int32{
		pb.Resource_RESOURCE_WOOD:  c.Wood,
		pb.Resource_RESOURCE_BRICK: c.Brick,
		pb.Resource_RESOURCE_SHEEP: c.Sheep,
		pb.Resource_RESOURCE_WHEAT: c.Wheat,
		pb.Resource_RESOURCE_ORE:   c.Ore,
	} {
		if val > 0 {
			cnt = int(val)
			res = r
		}
	}
	return cnt, res
}

func playerResource(c *pb.ResourceCount, res pb.Resource) int32 {
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		return c.Wood
	case pb.Resource_RESOURCE_BRICK:
		return c.Brick
	case pb.Resource_RESOURCE_SHEEP:
		return c.Sheep
	case pb.Resource_RESOURCE_WHEAT:
		return c.Wheat
	case pb.Resource_RESOURCE_ORE:
		return c.Ore
	default:
		return 0
	}
}

func deductResource(c *pb.ResourceCount, res pb.Resource, n int) {
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		c.Wood -= int32(n)
	case pb.Resource_RESOURCE_BRICK:
		c.Brick -= int32(n)
	case pb.Resource_RESOURCE_SHEEP:
		c.Sheep -= int32(n)
	case pb.Resource_RESOURCE_WHEAT:
		c.Wheat -= int32(n)
	case pb.Resource_RESOURCE_ORE:
		c.Ore -= int32(n)
	}
}

func addResource(c *pb.ResourceCount, res pb.Resource, n int) {
	switch res {
	case pb.Resource_RESOURCE_WOOD:
		c.Wood += int32(n)
	case pb.Resource_RESOURCE_BRICK:
		c.Brick += int32(n)
	case pb.Resource_RESOURCE_SHEEP:
		c.Sheep += int32(n)
	case pb.Resource_RESOURCE_WHEAT:
		c.Wheat += int32(n)
	case pb.Resource_RESOURCE_ORE:
		c.Ore += int32(n)
	}
}

var (
	ErrInvalidPlayer       = fmt.Errorf("Invalid player")
	ErrTradeNotFound       = fmt.Errorf("Trade not found")
	ErrNotTradeParticipant = fmt.Errorf("Not a participant in this trade")
)
