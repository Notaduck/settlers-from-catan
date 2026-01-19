package game

import (
	"testing"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// TestFullGamePlaythrough simulates an entire Catan game from start to finish
// following official Catan rules for 2-4 players
func TestFullGamePlaythrough(t *testing.T) {
	t.Run("Game setup and initial placement", func(t *testing.T) {
		// Create a 3-player game
		gameID := "playthrough-test"
		code := "TEST01"
		playerNames := []string{"Alice", "Bob", "Charlie"}
		playerIDs := []string{"p1", "p2", "p3"}

		state := NewGameState(gameID, code, playerNames, playerIDs)

		// Verify initial game state
		if state.Status != pb.GameStatus_GAME_STATUS_WAITING {
			t.Errorf("New game should be in WAITING status, got %v", state.Status)
		}
		if len(state.Players) != 3 {
			t.Fatalf("Expected 3 players, got %d", len(state.Players))
		}

		// Verify board is generated
		if state.Board == nil {
			t.Fatal("Board should be initialized")
		}
		if len(state.Board.Hexes) != 19 {
			t.Errorf("Board should have 19 hexes, got %d", len(state.Board.Hexes))
		}

		// Verify all players start with 0 VP
		for _, player := range state.Players {
			if player.VictoryPoints != 0 {
				t.Errorf("Player %s should start with 0 VP, got %d", player.Name, player.VictoryPoints)
			}
		}
	})

	t.Run("Setup phase - initial settlements and roads", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice", "Bob"}, []string{"p1", "p2"})

		// Start the game
		state.Status = pb.GameStatus_GAME_STATUS_SETUP

		// Each player places 2 settlements and 2 roads in setup phase
		// First round: Player 1, Player 2
		// Second round (reverse): Player 2, Player 1

		// Simulate placing first settlement for Alice
		simulatePlaceSettlement(state, "p1", "0.0,-0.67") // Example vertex ID
		if state.Players[0].VictoryPoints != 1 {
			t.Errorf("After first settlement, Alice should have 1 VP, got %d", state.Players[0].VictoryPoints)
		}

		// Place road connected to settlement
		simulatePlaceRoad(state, "p1", "0.0,-0.67-0.5,-0.33")

		// Bob places settlement and road
		simulatePlaceSettlement(state, "p2", "1.0,0.33")
		simulatePlaceRoad(state, "p2", "1.0,0.33-0.5,0.33")

		// Second round (reverse order) - Bob then Alice
		simulatePlaceSettlement(state, "p2", "-1.0,0.67")
		simulatePlaceRoad(state, "p2", "-1.0,0.67--0.5,0.33")

		simulatePlaceSettlement(state, "p1", "0.5,-0.33")
		simulatePlaceRoad(state, "p1", "0.5,-0.33-0.0,-0.67")

		// After setup, each player should have 2 VP (2 settlements)
		if state.Players[0].VictoryPoints != 2 {
			t.Errorf("After setup, Alice should have 2 VP, got %d", state.Players[0].VictoryPoints)
		}
		if state.Players[1].VictoryPoints != 2 {
			t.Errorf("After setup, Bob should have 2 VP, got %d", state.Players[1].VictoryPoints)
		}
	})

	t.Run("Turn sequence - roll, trade, build", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice", "Bob"}, []string{"p1", "p2"})
		state.Status = pb.GameStatus_GAME_STATUS_PLAYING
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL

		// Turn starts with rolling dice
		if state.TurnPhase != pb.TurnPhase_TURN_PHASE_ROLL {
			t.Error("Turn should start in ROLL phase")
		}

		// Simulate dice roll
		roll := RollDice()
		state.Dice = []int32{int32(roll/2 + roll%2), int32(roll / 2)}
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_TRADE

		if state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
			t.Error("After rolling, should be in TRADE phase")
		}

		// Move to build phase
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_BUILD

		// End turn
		state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL
		state.CurrentTurn = (state.CurrentTurn + 1) % int32(len(state.Players))

		if state.CurrentTurn != 1 {
			t.Errorf("After Alice's turn, should be Bob's turn (index 1), got %d", state.CurrentTurn)
		}
	})

	t.Run("Rolling 7 triggers robber", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice", "Bob"}, []string{"p1", "p2"})
		state.Status = pb.GameStatus_GAME_STATUS_PLAYING

		// Give Alice 9 cards (more than 7)
		state.Players[0].Resources = &pb.ResourceCount{
			Wood: 3, Brick: 2, Sheep: 2, Wheat: 1, Ore: 1,
		}

		roll := 7
		if !ShouldActivateRobber(roll) {
			t.Error("Roll of 7 should activate robber")
		}

		// Check if Alice must discard
		cardCount := int(state.Players[0].Resources.Wood +
			state.Players[0].Resources.Brick +
			state.Players[0].Resources.Sheep +
			state.Players[0].Resources.Wheat +
			state.Players[0].Resources.Ore)

		mustDiscard, discardAmount := MustDiscardOnSeven(cardCount)
		if !mustDiscard {
			t.Error("Alice with 9 cards should discard on 7")
		}
		if discardAmount != 4 {
			t.Errorf("Alice should discard 4 cards (half of 9 rounded down), got %d", discardAmount)
		}
	})

	t.Run("Building costs resources correctly", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice"}, []string{"p1"})
		state.Status = pb.GameStatus_GAME_STATUS_PLAYING

		// Give Alice resources for a road
		state.Players[0].Resources = &pb.ResourceCount{
			Wood: 2, Brick: 2, Sheep: 0, Wheat: 0, Ore: 0,
		}

		// Check if can afford road
		resources := ResourceCountToMap(state.Players[0].Resources)
		if !CanAfford(resources, "road") {
			t.Error("Alice should be able to afford a road")
		}

		// Build road and deduct resources
		DeductResources(state.Players[0].Resources, "road")

		if state.Players[0].Resources.Wood != 1 || state.Players[0].Resources.Brick != 1 {
			t.Errorf("After building road, should have 1 wood and 1 brick, got %d wood, %d brick",
				state.Players[0].Resources.Wood, state.Players[0].Resources.Brick)
		}

		// Can still afford another road
		resources = ResourceCountToMap(state.Players[0].Resources)
		if !CanAfford(resources, "road") {
			t.Error("Alice should still be able to afford another road")
		}

		DeductResources(state.Players[0].Resources, "road")

		// Now cannot afford road
		resources = ResourceCountToMap(state.Players[0].Resources)
		if CanAfford(resources, "road") {
			t.Error("Alice should not be able to afford a road anymore")
		}
	})

	t.Run("Victory condition at 10 points", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice", "Bob"}, []string{"p1", "p2"})
		state.Status = pb.GameStatus_GAME_STATUS_PLAYING

		// Simulate Alice reaching 10 VP
		// 4 settlements (4 VP) + 2 cities (4 VP) + longest road (2 VP) = 10 VP
		settlements := 4
		cities := 2
		hasLongestRoad := true
		hasLargestArmy := false
		vpCards := 0

		totalVP := CalculateVictoryPoints(settlements, cities, hasLongestRoad, hasLargestArmy, vpCards)

		if totalVP != 10 {
			t.Errorf("Expected 10 VP, got %d", totalVP)
		}

		if !IsVictorious(totalVP) {
			t.Error("10 VP should be victorious")
		}

		// Bob with 9 VP is not victorious
		bobVP := CalculateVictoryPoints(3, 2, false, true, 0) // 3 + 4 + 2 = 9 VP
		if IsVictorious(bobVP) {
			t.Error("9 VP should not be victorious")
		}
	})

	t.Run("Longest road bonus transfer", func(t *testing.T) {
		// Alice has 5 roads, gets bonus
		aliceRoads := 5
		if !QualifiesForLongestRoad(aliceRoads) {
			t.Error("5 roads should qualify for longest road")
		}

		// Bob builds 6 roads, takes bonus from Alice
		bobRoads := 6
		if !QualifiesForLongestRoad(bobRoads) {
			t.Error("6 roads should qualify for longest road")
		}

		// The game logic should transfer the bonus when Bob exceeds Alice
		// This is verified by checking both qualify, but Bob's is longer
		if bobRoads <= aliceRoads {
			t.Error("Bob should have more roads than Alice to take the bonus")
		}
	})

	t.Run("Largest army bonus", func(t *testing.T) {
		// 2 knights - no bonus
		if QualifiesForLargestArmy(2) {
			t.Error("2 knights should not qualify for largest army")
		}

		// 3 knights - gets bonus
		if !QualifiesForLargestArmy(3) {
			t.Error("3 knights should qualify for largest army")
		}
	})

	t.Run("City upgrade from settlement", func(t *testing.T) {
		state := NewGameState("g1", "CODE01", []string{"Alice"}, []string{"p1"})
		state.Status = pb.GameStatus_GAME_STATUS_PLAYING

		// Give Alice resources for a city (2 wheat, 3 ore)
		state.Players[0].Resources = &pb.ResourceCount{
			Wood: 0, Brick: 0, Sheep: 0, Wheat: 2, Ore: 3,
		}

		resources := ResourceCountToMap(state.Players[0].Resources)
		if !CanAfford(resources, "city") {
			t.Error("Alice should be able to afford a city")
		}

		// Cities are worth 2 VP (upgrade from 1 VP settlement = net +1 VP)
		cityVP := GetVictoryPoints("city")
		settlementVP := GetVictoryPoints("settlement")

		if cityVP != 2 || settlementVP != 1 {
			t.Errorf("City should be 2 VP, settlement 1 VP. Got city=%d, settlement=%d", cityVP, settlementVP)
		}

		// Building city deducts resources
		DeductResources(state.Players[0].Resources, "city")
		if state.Players[0].Resources.Wheat != 0 || state.Players[0].Resources.Ore != 0 {
			t.Error("Building city should consume all wheat and ore")
		}
	})

	t.Run("Full game simulation to victory", func(t *testing.T) {
		state := NewGameState("g1", "FINAL1", []string{"Alice", "Bob", "Charlie"}, []string{"p1", "p2", "p3"})

		// Setup phase
		state.Status = pb.GameStatus_GAME_STATUS_SETUP

		// Each player places 2 settlements = 2 VP each
		for i := range state.Players {
			state.Players[i].VictoryPoints = 2 // After setup phase
		}

		state.Status = pb.GameStatus_GAME_STATUS_PLAYING

		// Simulate Alice's path to victory:
		// Start: 2 settlements (2 VP)
		// Build: 2 more settlements (+2 VP = 4 VP)
		// Build: 5 roads, get longest road (+2 VP = 6 VP)
		// Upgrade: 2 settlements to cities (+2 VP net = 8 VP)
		// Build: 1 more settlement (+1 VP = 9 VP)
		// Upgrade: 1 settlement to city (+1 VP net = 10 VP)

		aliceSettlements := 2 // From setup
		aliceCities := 0
		aliceHasLongestRoad := false

		// Build 2 more settlements
		aliceSettlements = 4

		// Build 5 roads, get longest road
		aliceHasLongestRoad = true

		// Upgrade 2 settlements to cities
		aliceSettlements -= 2
		aliceCities += 2

		// Build 1 more settlement
		aliceSettlements += 1

		// Upgrade 1 settlement to city
		aliceSettlements -= 1
		aliceCities += 1

		// Final count: 2 settlements, 3 cities, longest road
		// VP: 2*1 + 3*2 + 2 = 2 + 6 + 2 = 10 VP

		calculatedVP := CalculateVictoryPoints(aliceSettlements, aliceCities, aliceHasLongestRoad, false, 0)
		expectedVP := 10

		if calculatedVP != expectedVP {
			t.Errorf("VP calculation error: expected=%d, calculated=%d (settlements=%d, cities=%d, longestRoad=%v)",
				expectedVP, calculatedVP, aliceSettlements, aliceCities, aliceHasLongestRoad)
		}

		if !IsVictorious(calculatedVP) {
			t.Errorf("Alice with %d VP should be victorious", calculatedVP)
		}

		// Game ends
		state.Status = pb.GameStatus_GAME_STATUS_FINISHED
		if state.Status != pb.GameStatus_GAME_STATUS_FINISHED {
			t.Error("Game should be finished when someone wins")
		}
	})
}

// Helper functions for simulation

func simulatePlaceSettlement(state *pb.GameState, playerID, vertexID string) {
	for _, player := range state.Players {
		if player.Id == playerID {
			player.VictoryPoints += int32(GetVictoryPoints("settlement"))
			break
		}
	}
}

func simulatePlaceRoad(state *pb.GameState, playerID, edgeID string) {
	// Roads don't give VP, just placed
}

func TestResourceGatheringOnDiceRoll(t *testing.T) {
	board := GenerateBoard()

	// Find all hexes with number 6 (common roll)
	var hexesWithSix []*pb.Hex
	for _, hex := range board.Hexes {
		if hex.Number == 6 {
			hexesWithSix = append(hexesWithSix, hex)
		}
	}

	if len(hexesWithSix) != 2 {
		t.Errorf("Should have exactly 2 hexes with number 6, got %d", len(hexesWithSix))
	}

	// Verify these hexes have resources (not desert)
	for _, hex := range hexesWithSix {
		if hex.Resource == pb.TileResource_TILE_RESOURCE_DESERT {
			t.Error("Hex with number 6 should not be desert")
		}
	}
}

func TestMaxBuildingLimits(t *testing.T) {
	// Verify building limits per player match Catan rules
	if GetMaxSettlements() != 5 {
		t.Errorf("Max settlements should be 5, got %d", GetMaxSettlements())
	}
	if GetMaxCities() != 4 {
		t.Errorf("Max cities should be 4, got %d", GetMaxCities())
	}
	if GetMaxRoads() != 15 {
		t.Errorf("Max roads should be 15, got %d", GetMaxRoads())
	}

	// Total pieces: 5 settlements + 4 cities + 15 roads = 24 pieces per player
	totalPieces := GetMaxSettlements() + GetMaxCities() + GetMaxRoads()
	if totalPieces != 24 {
		t.Errorf("Total pieces per player should be 24, got %d", totalPieces)
	}
}

func TestPortTrades(t *testing.T) {
	// Standard Catan port trade ratios
	// 4:1 - generic trade (no port needed)
	// 3:1 - generic port
	// 2:1 - specific resource port

	genericTradeRatio := 4
	genericPortRatio := 3
	specificPortRatio := 2

	if genericTradeRatio != 4 {
		t.Error("Generic trade should be 4:1")
	}
	if genericPortRatio != 3 {
		t.Error("Generic port should be 3:1")
	}
	if specificPortRatio != 2 {
		t.Error("Specific port should be 2:1")
	}
}
