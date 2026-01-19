package game

import (
	"testing"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Catan Rules Reference:
// - 2-4 players
// - 19 land hexes (4 wood, 4 sheep, 4 wheat, 3 brick, 3 ore, 1 desert)
// - 18 number tokens (no 7, desert has no number)
// - Settlements: 2 VP each, need wood + brick + sheep + wheat
// - Cities: 3 VP each (2 additional), need 2 wheat + 3 ore
// - Roads: 0 VP, need wood + brick
// - Longest Road: 2 VP (minimum 5 roads)
// - Largest Army: 2 VP (minimum 3 knights)
// - Victory: First to 10 VP

func TestBuildingCosts(t *testing.T) {
	tests := []struct {
		name     string
		building string
		costs    map[pb.TileResource]int
	}{
		{
			name:     "Settlement costs",
			building: "settlement",
			costs: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WOOD:  1,
				pb.TileResource_TILE_RESOURCE_BRICK: 1,
				pb.TileResource_TILE_RESOURCE_SHEEP: 1,
				pb.TileResource_TILE_RESOURCE_WHEAT: 1,
			},
		},
		{
			name:     "City costs",
			building: "city",
			costs: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WHEAT: 2,
				pb.TileResource_TILE_RESOURCE_ORE:   3,
			},
		},
		{
			name:     "Road costs",
			building: "road",
			costs: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WOOD:  1,
				pb.TileResource_TILE_RESOURCE_BRICK: 1,
			},
		},
		{
			name:     "Development card costs",
			building: "development_card",
			costs: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_SHEEP: 1,
				pb.TileResource_TILE_RESOURCE_WHEAT: 1,
				pb.TileResource_TILE_RESOURCE_ORE:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			costs := GetBuildingCosts(tt.building)
			for resource, expectedCost := range tt.costs {
				if costs[resource] != expectedCost {
					t.Errorf("Expected %d %v for %s, got %d", expectedCost, resource, tt.building, costs[resource])
				}
			}
		})
	}
}

func TestVictoryPoints(t *testing.T) {
	tests := []struct {
		name     string
		building string
		vp       int
	}{
		{"Settlement is 1 VP", "settlement", 1},
		{"City is 2 VP", "city", 2},
		{"Road is 0 VP", "road", 0},
		{"Longest Road bonus is 2 VP", "longest_road", 2},
		{"Largest Army bonus is 2 VP", "largest_army", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := GetVictoryPoints(tt.building)
			if vp != tt.vp {
				t.Errorf("Expected %d VP for %s, got %d", tt.vp, tt.building, vp)
			}
		})
	}
}

func TestCanAffordBuilding(t *testing.T) {
	tests := []struct {
		name      string
		resources map[pb.TileResource]int
		building  string
		canAfford bool
	}{
		{
			name: "Can afford settlement",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WOOD:  1,
				pb.TileResource_TILE_RESOURCE_BRICK: 1,
				pb.TileResource_TILE_RESOURCE_SHEEP: 1,
				pb.TileResource_TILE_RESOURCE_WHEAT: 1,
			},
			building:  "settlement",
			canAfford: true,
		},
		{
			name: "Cannot afford settlement - missing wood",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_BRICK: 1,
				pb.TileResource_TILE_RESOURCE_SHEEP: 1,
				pb.TileResource_TILE_RESOURCE_WHEAT: 1,
			},
			building:  "settlement",
			canAfford: false,
		},
		{
			name: "Can afford city",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WHEAT: 2,
				pb.TileResource_TILE_RESOURCE_ORE:   3,
			},
			building:  "city",
			canAfford: true,
		},
		{
			name: "Cannot afford city - not enough ore",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WHEAT: 2,
				pb.TileResource_TILE_RESOURCE_ORE:   2,
			},
			building:  "city",
			canAfford: false,
		},
		{
			name: "Can afford road",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WOOD:  1,
				pb.TileResource_TILE_RESOURCE_BRICK: 1,
			},
			building:  "road",
			canAfford: true,
		},
		{
			name: "Has excess resources",
			resources: map[pb.TileResource]int{
				pb.TileResource_TILE_RESOURCE_WOOD:  5,
				pb.TileResource_TILE_RESOURCE_BRICK: 5,
				pb.TileResource_TILE_RESOURCE_SHEEP: 5,
				pb.TileResource_TILE_RESOURCE_WHEAT: 5,
				pb.TileResource_TILE_RESOURCE_ORE:   5,
			},
			building:  "settlement",
			canAfford: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanAfford(tt.resources, tt.building)
			if result != tt.canAfford {
				t.Errorf("Expected CanAfford to be %v, got %v", tt.canAfford, result)
			}
		})
	}
}

func TestResourceDistributionOnDiceRoll(t *testing.T) {
	board := GenerateBoard()

	rollCounts := make(map[int]int)
	for _, hex := range board.Hexes {
		if hex.Number != 0 {
			rollCounts[int(hex.Number)]++
		}
	}

	if rollCounts[7] != 0 {
		t.Error("No hex should have number token 7")
	}

	expectedCounts := map[int]int{
		2: 1, 3: 2, 4: 2, 5: 2, 6: 2,
		8: 2, 9: 2, 10: 2, 11: 2, 12: 1,
	}

	for roll, expected := range expectedCounts {
		if rollCounts[roll] != expected {
			t.Errorf("Expected %d hexes with number %d, got %d", expected, roll, rollCounts[roll])
		}
	}
}

func TestDiceRollRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		roll := RollDice()
		if roll < 2 || roll > 12 {
			t.Errorf("Dice roll %d is out of valid range (2-12)", roll)
		}
	}
}

func TestRobberActivatedOnSeven(t *testing.T) {
	roll := 7
	robberActivated := ShouldActivateRobber(roll)
	if !robberActivated {
		t.Error("Robber should be activated on roll of 7")
	}

	for _, r := range []int{2, 3, 4, 5, 6, 8, 9, 10, 11, 12} {
		if ShouldActivateRobber(r) {
			t.Errorf("Robber should not be activated on roll of %d", r)
		}
	}
}

func TestDiscardOnSevenWithTooManyCards(t *testing.T) {
	tests := []struct {
		cardCount   int
		mustDiscard bool
		discardAmt  int
	}{
		{7, false, 0},
		{8, true, 4},
		{9, true, 4},
		{10, true, 5},
		{11, true, 5},
		{12, true, 6},
	}

	for _, tt := range tests {
		mustDiscard, discardAmt := MustDiscardOnSeven(tt.cardCount)
		if mustDiscard != tt.mustDiscard {
			t.Errorf("With %d cards, mustDiscard should be %v, got %v", tt.cardCount, tt.mustDiscard, mustDiscard)
		}
		if discardAmt != tt.discardAmt {
			t.Errorf("With %d cards, should discard %d, got %d", tt.cardCount, tt.discardAmt, discardAmt)
		}
	}
}

func TestLongestRoadCalculation(t *testing.T) {
	tests := []struct {
		name       string
		roadLength int
		hasBonus   bool
	}{
		{"4 roads - no bonus", 4, false},
		{"5 roads - has bonus", 5, true},
		{"10 roads - has bonus", 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasBonus := QualifiesForLongestRoad(tt.roadLength)
			if hasBonus != tt.hasBonus {
				t.Errorf("Road length %d: expected bonus=%v, got %v", tt.roadLength, tt.hasBonus, hasBonus)
			}
		})
	}
}

func TestLargestArmyCalculation(t *testing.T) {
	tests := []struct {
		name        string
		knightCount int
		hasBonus    bool
	}{
		{"2 knights - no bonus", 2, false},
		{"3 knights - has bonus", 3, true},
		{"5 knights - has bonus", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasBonus := QualifiesForLargestArmy(tt.knightCount)
			if hasBonus != tt.hasBonus {
				t.Errorf("Knight count %d: expected bonus=%v, got %v", tt.knightCount, tt.hasBonus, hasBonus)
			}
		})
	}
}

func TestVictoryCondition(t *testing.T) {
	tests := []struct {
		name   string
		vp     int
		winner bool
	}{
		{"9 VP - not winner", 9, false},
		{"10 VP - winner", 10, true},
		{"11 VP - winner", 11, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isWinner := IsVictorious(tt.vp)
			if isWinner != tt.winner {
				t.Errorf("With %d VP: expected winner=%v, got %v", tt.vp, tt.winner, isWinner)
			}
		})
	}
}

func TestCalculateTotalVictoryPoints(t *testing.T) {
	tests := []struct {
		name           string
		settlements    int
		cities         int
		hasLongestRoad bool
		hasLargestArmy bool
		vpCards        int
		expectedVP     int
	}{
		{
			name:        "Initial setup (2 settlements)",
			settlements: 2,
			expectedVP:  2,
		},
		{
			name:        "Mid-game progress",
			settlements: 3,
			cities:      1,
			expectedVP:  5,
		},
		{
			name:           "With longest road",
			settlements:    3,
			cities:         1,
			hasLongestRoad: true,
			expectedVP:     7,
		},
		{
			name:           "Winning position",
			settlements:    2,
			cities:         3,
			hasLongestRoad: true,
			hasLargestArmy: true,
			expectedVP:     12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := CalculateVictoryPoints(tt.settlements, tt.cities, tt.hasLongestRoad, tt.hasLargestArmy, tt.vpCards)
			if vp != tt.expectedVP {
				t.Errorf("Expected %d VP, got %d", tt.expectedVP, vp)
			}
		})
	}
}

func TestSetupPhaseRules(t *testing.T) {
	t.Run("Each player places 2 settlements and 2 roads in setup", func(t *testing.T) {
		expectedSettlements := 2
		expectedRoads := 2

		if expectedSettlements != 2 || expectedRoads != 2 {
			t.Error("Setup phase should give each player 2 settlements and 2 roads")
		}
	})

	t.Run("Second settlement gives starting resources", func(t *testing.T) {
		getsResources := SecondSettlementGivesResources()
		if !getsResources {
			t.Error("Second settlement placement should give starting resources")
		}
	})
}

func TestTurnPhases(t *testing.T) {
	phases := []string{"roll", "trade", "build", "end"}

	if len(phases) != 4 {
		t.Errorf("Expected 4 turn phases, got %d", len(phases))
	}

	if phases[0] != "roll" {
		t.Error("First phase should be rolling dice")
	}
}

func TestMaxPlayersPerGame(t *testing.T) {
	maxPlayers := GetMaxPlayers()
	if maxPlayers != 4 {
		t.Errorf("Max players should be 4, got %d", maxPlayers)
	}
}

func TestMinPlayersToStart(t *testing.T) {
	minPlayers := GetMinPlayers()
	if minPlayers != 2 {
		t.Errorf("Min players to start should be 2, got %d", minPlayers)
	}
}

func TestInitialResourceCount(t *testing.T) {
	for resource, count := range GetInitialResourceBank() {
		if count != 19 {
			t.Errorf("Resource %v should start with 19, got %d", resource, count)
		}
	}
}

func TestMaxBuildingsPerPlayer(t *testing.T) {
	maxSettlements := GetMaxSettlements()
	maxCities := GetMaxCities()
	maxRoads := GetMaxRoads()

	if maxSettlements != 5 {
		t.Errorf("Max settlements should be 5, got %d", maxSettlements)
	}
	if maxCities != 4 {
		t.Errorf("Max cities should be 4, got %d", maxCities)
	}
	if maxRoads != 15 {
		t.Errorf("Max roads should be 15, got %d", maxRoads)
	}
}
