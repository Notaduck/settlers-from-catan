package game

import (
	"errors"
	"math/rand"
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// DiscardCards removes the specified resources from the given player's hand for robber discards.
func DiscardCards(state *pb.GameState, playerID string, toDiscard *pb.ResourceCount) error {
	if state == nil || state.RobberPhase == nil {
		return errors.New("no robber phase active")
	}
	required, ok := state.RobberPhase.DiscardRequired[playerID]
	if !ok || required == 0 {
		return errors.New("player not required to discard")
	}
	have := countTotalResources(getPlayerByID(state, playerID).Resources)
	request := countTotalResources(toDiscard)
	if request != int(required) {
		return errors.New("incorrect discard count")
	}
	if request > have {
		return errors.New("cannot discard more cards than owned")
	}
	// Remove specified cards
	p := getPlayerByID(state, playerID)
	if err := removePlayerResources(p.Resources, toDiscard); err != nil {
		return err
	}
	// Mark as discarded
	removeFromSlice(&state.RobberPhase.DiscardPending, playerID)
	delete(state.RobberPhase.DiscardRequired, playerID)
	return nil
}

func removePlayerResources(r *pb.ResourceCount, toRemove *pb.ResourceCount) error {
	if r == nil || toRemove == nil {
		return errors.New("nil resource count")
	}
	fields := []struct {
		v  *int32
		rm int32
	}{
		{&r.Wood, toRemove.Wood},
		{&r.Brick, toRemove.Brick},
		{&r.Sheep, toRemove.Sheep},
		{&r.Wheat, toRemove.Wheat},
		{&r.Ore, toRemove.Ore},
	}
	for _, f := range fields {
		if *f.v < f.rm {
			return errors.New("not enough resources to remove")
		}
		*f.v -= f.rm
	}
	return nil
}

func removeFromSlice(s *[]string, val string) {
	out := (*s)[:0]
	for _, v := range *s {
		if v != val {
			out = append(out, v)
		}
	}
	*s = out
}

// MoveRobber moves the robber to the specified hex if valid
func MoveRobber(state *pb.GameState, playerID string, hex *pb.HexCoord) error {
	if state == nil || state.Board == nil {
		return errors.New("invalid state")
	}
	if state.RobberPhase == nil {
		return errors.New("not in robber phase")
	}
	if state.RobberPhase.MovePendingPlayerId == nil || *state.RobberPhase.MovePendingPlayerId != playerID {
		return errors.New("not player's turn to move robber")
	}
	if hex == nil {
		return errors.New("no destination specified")
	}
	if state.Board.RobberHex != nil && state.Board.RobberHex.Q == hex.Q && state.Board.RobberHex.R == hex.R {
		return errors.New("robber already on that hex")
	}
	// Confirm hex exists in board
	hexValid := false
	for _, h := range state.Board.Hexes {
		if h.Coord != nil && h.Coord.Q == hex.Q && h.Coord.R == hex.R {
			hexValid = true
			break
		}
	}
	if !hexValid {
		return errors.New("no such hex")
	}
	// Move robber
	state.Board.RobberHex = hex
	state.RobberPhase.MovePendingPlayerId = nil
	return nil
}

// StealFromPlayer attempts to transfer a random card from victim to thief
// If chooser is non-nil, use chooser(poolLen) for random selection (for testing).
func StealFromPlayer(state *pb.GameState, thiefID, victimID string, chooser ...func(n int) int) (pb.Resource, error) {
	if state == nil || state.Board == nil {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("invalid state")
	}
	if state.RobberPhase == nil {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("not in robber phase")
	}
	if state.RobberPhase.StealPendingPlayerId == nil || *state.RobberPhase.StealPendingPlayerId != thiefID {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("not thief's turn")
	}
	victim := getPlayerByID(state, victimID)
	thief := getPlayerByID(state, thiefID)
	if victim == nil || thief == nil {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("invalid player ids")
	}
	// Victim must be adjacent to robber hex
	if !playerIsAdjacentToRobberHex(state, victimID) {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("victim not adjacent to robber")
	}
	// Get victim's resources
	var resourcePool []pb.Resource
	for i := int32(0); i < victim.Resources.Wood; i++ {
		resourcePool = append(resourcePool, pb.Resource_RESOURCE_WOOD)
	}
	for i := int32(0); i < victim.Resources.Brick; i++ {
		resourcePool = append(resourcePool, pb.Resource_RESOURCE_BRICK)
	}
	for i := int32(0); i < victim.Resources.Sheep; i++ {
		resourcePool = append(resourcePool, pb.Resource_RESOURCE_SHEEP)
	}
	for i := int32(0); i < victim.Resources.Wheat; i++ {
		resourcePool = append(resourcePool, pb.Resource_RESOURCE_WHEAT)
	}
	for i := int32(0); i < victim.Resources.Ore; i++ {
		resourcePool = append(resourcePool, pb.Resource_RESOURCE_ORE)
	}
	if len(resourcePool) == 0 {
		return pb.Resource_RESOURCE_UNSPECIFIED, errors.New("victim has no resources")
	}
	var randIdx int
	if len(chooser) > 0 && chooser[0] != nil {
		randIdx = chooser[0](len(resourcePool))
	} else {
		randIdx = rand.Intn(len(resourcePool))
	}
	stolen := resourcePool[randIdx]
	switch stolen {
	case pb.Resource_RESOURCE_WOOD:
		victim.Resources.Wood--
		thief.Resources.Wood++
	case pb.Resource_RESOURCE_BRICK:
		victim.Resources.Brick--
		thief.Resources.Brick++
	case pb.Resource_RESOURCE_SHEEP:
		victim.Resources.Sheep--
		thief.Resources.Sheep++
	case pb.Resource_RESOURCE_WHEAT:
		victim.Resources.Wheat--
		thief.Resources.Wheat++
	case pb.Resource_RESOURCE_ORE:
		victim.Resources.Ore--
		thief.Resources.Ore++
	}
	state.RobberPhase.StealPendingPlayerId = nil
	return stolen, nil
}

func getPlayerByID(state *pb.GameState, id string) *pb.PlayerState {
	for _, p := range state.Players {
		if p.Id == id {
			return p
		}
	}
	return nil
}

// playerIsAdjacentToRobberHex returns true if any of the victim's settlements or cities are adjacent to the robber
func playerIsAdjacentToRobberHex(state *pb.GameState, playerID string) bool {
	robHex := state.Board.RobberHex
	if robHex == nil {
		return false
	}
	for _, v := range state.Board.Vertices {
		if v.Building != nil && v.Building.OwnerId == playerID {
			for _, h := range v.AdjacentHexes {
				if h.Q == robHex.Q && h.R == robHex.R {
					return true
				}
			}
		}
	}
	return false
}
