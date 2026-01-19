package game

import (
	"math/rand"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Building costs in Catan
var buildingCosts = map[string]map[pb.TileResource]int{
	"settlement": {
		pb.TileResource_TILE_RESOURCE_WOOD:  1,
		pb.TileResource_TILE_RESOURCE_BRICK: 1,
		pb.TileResource_TILE_RESOURCE_SHEEP: 1,
		pb.TileResource_TILE_RESOURCE_WHEAT: 1,
	},
	"city": {
		pb.TileResource_TILE_RESOURCE_WHEAT: 2,
		pb.TileResource_TILE_RESOURCE_ORE:   3,
	},
	"road": {
		pb.TileResource_TILE_RESOURCE_WOOD:  1,
		pb.TileResource_TILE_RESOURCE_BRICK: 1,
	},
	"development_card": {
		pb.TileResource_TILE_RESOURCE_SHEEP: 1,
		pb.TileResource_TILE_RESOURCE_WHEAT: 1,
		pb.TileResource_TILE_RESOURCE_ORE:   1,
	},
}

// Victory points for different structures/bonuses
var victoryPoints = map[string]int{
	"settlement":   1,
	"city":         2,
	"road":         0,
	"longest_road": 2,
	"largest_army": 2,
}

// GetBuildingCosts returns the resource costs for a building type
func GetBuildingCosts(building string) map[pb.TileResource]int {
	if costs, ok := buildingCosts[building]; ok {
		return costs
	}
	return map[pb.TileResource]int{}
}

// GetVictoryPoints returns the VP value for a building/bonus type
func GetVictoryPoints(building string) int {
	if vp, ok := victoryPoints[building]; ok {
		return vp
	}
	return 0
}

// CanAfford checks if a player can afford to build something
func CanAfford(resources map[pb.TileResource]int, building string) bool {
	costs := GetBuildingCosts(building)
	for resource, cost := range costs {
		if resources[resource] < cost {
			return false
		}
	}
	return true
}

// RollDice simulates rolling two dice
func RollDice() int {
	die1 := rand.Intn(6) + 1
	die2 := rand.Intn(6) + 1
	return die1 + die2
}

// ShouldActivateRobber returns true if the roll triggers the robber
func ShouldActivateRobber(roll int) bool {
	return roll == 7
}

// MustDiscardOnSeven returns whether a player must discard and how many
func MustDiscardOnSeven(cardCount int) (bool, int) {
	if cardCount > 7 {
		return true, cardCount / 2
	}
	return false, 0
}

// QualifiesForLongestRoad checks if a road length qualifies for the bonus
func QualifiesForLongestRoad(roadLength int) bool {
	return roadLength >= 5
}

// QualifiesForLargestArmy checks if knight count qualifies for the bonus
func QualifiesForLargestArmy(knightCount int) bool {
	return knightCount >= 3
}

// IsVictorious checks if a player has won
func IsVictorious(vp int) bool {
	return vp >= 10
}

// CalculateVictoryPoints calculates total VP for a player
func CalculateVictoryPoints(settlements, cities int, hasLongestRoad, hasLargestArmy bool, vpCards int) int {
	total := settlements*GetVictoryPoints("settlement") + cities*GetVictoryPoints("city") + vpCards
	if hasLongestRoad {
		total += GetVictoryPoints("longest_road")
	}
	if hasLargestArmy {
		total += GetVictoryPoints("largest_army")
	}
	return total
}

// SecondSettlementGivesResources returns whether second settlement gives resources
func SecondSettlementGivesResources() bool {
	return true
}

// GetMaxPlayers returns the maximum number of players
func GetMaxPlayers() int {
	return 4
}

// GetMinPlayers returns the minimum number of players to start
func GetMinPlayers() int {
	return 2
}

// GetInitialResourceBank returns the starting resource counts
func GetInitialResourceBank() map[pb.TileResource]int {
	return map[pb.TileResource]int{
		pb.TileResource_TILE_RESOURCE_WOOD:  19,
		pb.TileResource_TILE_RESOURCE_BRICK: 19,
		pb.TileResource_TILE_RESOURCE_SHEEP: 19,
		pb.TileResource_TILE_RESOURCE_WHEAT: 19,
		pb.TileResource_TILE_RESOURCE_ORE:   19,
	}
}

// GetMaxSettlements returns the max settlements per player
func GetMaxSettlements() int {
	return 5
}

// GetMaxCities returns the max cities per player
func GetMaxCities() int {
	return 4
}

// GetMaxRoads returns the max roads per player
func GetMaxRoads() int {
	return 15
}

// DeductResources removes building costs from a player's resources
func DeductResources(resources *pb.ResourceCount, building string) {
	costs := GetBuildingCosts(building)
	for resource, cost := range costs {
		switch resource {
		case pb.TileResource_TILE_RESOURCE_WOOD:
			resources.Wood -= int32(cost)
		case pb.TileResource_TILE_RESOURCE_BRICK:
			resources.Brick -= int32(cost)
		case pb.TileResource_TILE_RESOURCE_SHEEP:
			resources.Sheep -= int32(cost)
		case pb.TileResource_TILE_RESOURCE_WHEAT:
			resources.Wheat -= int32(cost)
		case pb.TileResource_TILE_RESOURCE_ORE:
			resources.Ore -= int32(cost)
		}
	}
}

// ResourceCountToMap converts ResourceCount to a map for easier checking
func ResourceCountToMap(rc *pb.ResourceCount) map[pb.TileResource]int {
	return map[pb.TileResource]int{
		pb.TileResource_TILE_RESOURCE_WOOD:  int(rc.Wood),
		pb.TileResource_TILE_RESOURCE_BRICK: int(rc.Brick),
		pb.TileResource_TILE_RESOURCE_SHEEP: int(rc.Sheep),
		pb.TileResource_TILE_RESOURCE_WHEAT: int(rc.Wheat),
		pb.TileResource_TILE_RESOURCE_ORE:   int(rc.Ore),
	}
}
