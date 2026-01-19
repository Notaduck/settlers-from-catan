package game

import (
	"math/rand"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// DiceRollResult contains the result of a dice roll
type DiceRollResult struct {
	Die1             int
	Die2             int
	Total            int
	RobberActivated  bool
	PlayersToDiscard []PlayerDiscard
	ResourcesGained  map[string]*pb.ResourceCount // playerID -> resources gained
}

// PlayerDiscard represents a player who needs to discard cards
type PlayerDiscard struct {
	PlayerID     string
	DiscardCount int
}

// PerformDiceRoll rolls dice and distributes resources
func PerformDiceRoll(state *pb.GameState, playerID string) (*DiceRollResult, error) {
	die1 := rand.Intn(6) + 1
	die2 := rand.Intn(6) + 1
	return PerformDiceRollWithValues(state, playerID, die1, die2)
}

// PerformDiceRollWithValues performs a dice roll with specific values (for testing)
func PerformDiceRollWithValues(state *pb.GameState, playerID string, die1, die2 int) (*DiceRollResult, error) {
	// Verify game is in playing status
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return nil, ErrWrongPhase
	}

	// Verify it's roll phase
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_ROLL {
		return nil, ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return nil, ErrNotYourTurn
	}
	if state.Players[currentPlayerIdx].Id != playerID {
		return nil, ErrNotYourTurn
	}

	total := die1 + die2

	result := &DiceRollResult{
		Die1:            die1,
		Die2:            die2,
		Total:           total,
		ResourcesGained: make(map[string]*pb.ResourceCount),
	}

	// Update state dice
	state.Dice = []int32{int32(die1), int32(die2)}

	// Check for 7 (robber)
	if total == 7 {
		result.RobberActivated = true
		result.PlayersToDiscard = calculatePlayersToDiscard(state)
		initializeRobberPhase(state, playerID, result.PlayersToDiscard)
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_BUILD // Skip trade, go to build (robber handling)
	} else {
		// Distribute resources
		distributeResources(state, total, result)
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_TRADE
	}

	return result, nil
}

// calculatePlayersToDiscard returns list of players who must discard on a 7
func calculatePlayersToDiscard(state *pb.GameState) []PlayerDiscard {
	var result []PlayerDiscard

	for _, player := range state.Players {
		cardCount := countTotalResources(player.Resources)
		if cardCount > 7 {
			discardCount := cardCount / 2
			result = append(result, PlayerDiscard{
				PlayerID:     player.Id,
				DiscardCount: discardCount,
			})
		}
	}

	return result
}

func initializeRobberPhase(state *pb.GameState, playerID string, discards []PlayerDiscard) {
	if state == nil {
		return
	}
	movePlayerID := playerID
	phase := &pb.RobberPhase{
		DiscardRequired:     map[string]int32{},
		MovePendingPlayerId: &movePlayerID,
	}
	for _, pd := range discards {
		if pd.DiscardCount <= 0 {
			continue
		}
		phase.DiscardPending = append(phase.DiscardPending, pd.PlayerID)
		phase.DiscardRequired[pd.PlayerID] = int32(pd.DiscardCount)
	}
	state.RobberPhase = phase
}

// countTotalResources returns the total number of resource cards a player has
func countTotalResources(r *pb.ResourceCount) int {
	if r == nil {
		return 0
	}
	return int(r.Wood + r.Brick + r.Sheep + r.Wheat + r.Ore)
}

// distributeResources gives resources to players with buildings on matching hexes
func distributeResources(state *pb.GameState, diceTotal int, result *DiceRollResult) {
	// Find all hexes with the rolled number
	for _, hex := range state.Board.Hexes {
		if hex.Number != int32(diceTotal) {
			continue
		}

		// Skip if robber is on this hex
		if state.Board.RobberHex != nil &&
			state.Board.RobberHex.Q == hex.Coord.Q &&
			state.Board.RobberHex.R == hex.Coord.R {
			continue
		}

		// Skip desert (produces nothing)
		if hex.Resource == pb.TileResource_TILE_RESOURCE_DESERT {
			continue
		}

		// Find all vertices adjacent to this hex with buildings
		for _, vertex := range state.Board.Vertices {
			if vertex.Building == nil {
				continue
			}

			// Check if vertex is adjacent to this hex
			isAdjacent := false
			for _, adjHex := range vertex.AdjacentHexes {
				if adjHex.Q == hex.Coord.Q && adjHex.R == hex.Coord.R {
					isAdjacent = true
					break
				}
			}
			if !isAdjacent {
				continue
			}

			// Determine resource count (city = 2, settlement = 1)
			resourceCount := 1
			if vertex.Building.Type == pb.BuildingType_BUILDING_TYPE_CITY {
				resourceCount = 2
			}

			// Find the player and add resources
			ownerID := vertex.Building.OwnerId
			for _, player := range state.Players {
				if player.Id == ownerID {
					addResourceToPlayer(player.Resources, hex.Resource, resourceCount)

					// Track in result
					if result.ResourcesGained[ownerID] == nil {
						result.ResourcesGained[ownerID] = &pb.ResourceCount{}
					}
					addResourceToCount(result.ResourcesGained[ownerID], hex.Resource, resourceCount)
					break
				}
			}
		}
	}
}

// addResourceToPlayer adds resources to a player's resource count
func addResourceToPlayer(resources *pb.ResourceCount, tileResource pb.TileResource, count int) {
	if resources == nil {
		return
	}
	switch tileResource {
	case pb.TileResource_TILE_RESOURCE_WOOD:
		resources.Wood += int32(count)
	case pb.TileResource_TILE_RESOURCE_BRICK:
		resources.Brick += int32(count)
	case pb.TileResource_TILE_RESOURCE_SHEEP:
		resources.Sheep += int32(count)
	case pb.TileResource_TILE_RESOURCE_WHEAT:
		resources.Wheat += int32(count)
	case pb.TileResource_TILE_RESOURCE_ORE:
		resources.Ore += int32(count)
	}
}

// addResourceToCount adds to a ResourceCount (for tracking gains)
func addResourceToCount(rc *pb.ResourceCount, tileResource pb.TileResource, count int) {
	switch tileResource {
	case pb.TileResource_TILE_RESOURCE_WOOD:
		rc.Wood += int32(count)
	case pb.TileResource_TILE_RESOURCE_BRICK:
		rc.Brick += int32(count)
	case pb.TileResource_TILE_RESOURCE_SHEEP:
		rc.Sheep += int32(count)
	case pb.TileResource_TILE_RESOURCE_WHEAT:
		rc.Wheat += int32(count)
	case pb.TileResource_TILE_RESOURCE_ORE:
		rc.Ore += int32(count)
	}
}
