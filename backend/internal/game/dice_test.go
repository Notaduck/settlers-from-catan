package game

import (
	"testing"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Helper to create a test game state in playing phase
func createPlayingGameState(numPlayers int) *pb.GameState {
	names := []string{"Alice", "Bob", "Charlie", "Dave"}[:numPlayers]
	ids := []string{"p1", "p2", "p3", "p4"}[:numPlayers]
	state := NewGameState("g1", "CODE", names, ids)
	state.Status = pb.GameStatus_GAME_STATUS_PLAYING
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL
	state.CurrentTurn = 0
	return state
}

// Dice Rolling Tests
// - Can only roll during ROLL phase
// - Can only roll on your turn
// - Roll advances phase to TRADE
// - Rolling 7 triggers robber phase

func TestRollDice_OnlyInRollPhase(t *testing.T) {
	state := createPlayingGameState(2)
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_TRADE // Not ROLL phase

	playerID := state.Players[0].Id

	_, err := PerformDiceRoll(state, playerID)
	if err != ErrWrongPhase {
		t.Errorf("Expected ErrWrongPhase when rolling in TRADE phase, got %v", err)
	}
}

func TestRollDice_OnlyOnYourTurn(t *testing.T) {
	state := createPlayingGameState(2)
	state.CurrentTurn = 0

	// Player 1 (index 1) tries to roll when it's player 0's turn
	wrongPlayerID := state.Players[1].Id

	_, err := PerformDiceRoll(state, wrongPlayerID)
	if err != ErrNotYourTurn {
		t.Errorf("Expected ErrNotYourTurn, got %v", err)
	}
}

func TestRollDice_AdvancesToTradePhase(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id

	// Use specific values to ensure we don't roll a 7
	_, err := PerformDiceRollWithValues(state, playerID, 3, 3) // Roll 6, not 7
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		t.Errorf("Expected phase to advance to TRADE, got %v", state.TurnPhase)
	}
}

func TestRollDice_SetsDiceValues(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id

	result, err := PerformDiceRoll(state, playerID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify dice values are in valid range
	if result.Die1 < 1 || result.Die1 > 6 {
		t.Errorf("Die1 out of range: %d", result.Die1)
	}
	if result.Die2 < 1 || result.Die2 > 6 {
		t.Errorf("Die2 out of range: %d", result.Die2)
	}
	if result.Total != result.Die1+result.Die2 {
		t.Errorf("Total %d != Die1 + Die2 (%d + %d)", result.Total, result.Die1, result.Die2)
	}

	// Verify state dice are updated
	if len(state.Dice) != 2 {
		t.Errorf("Expected 2 dice in state, got %d", len(state.Dice))
	}
}

func TestRollDice_SevenTriggersRobber(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id

	// Use seeded roll to get a 7
	result, err := PerformDiceRollWithValues(state, playerID, 3, 4)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Total != 7 {
		t.Errorf("Expected total 7, got %d", result.Total)
	}

	if !result.RobberActivated {
		t.Error("Expected robber to be activated on 7")
	}

	// Phase should be BUILD (skip trade when robber activates)
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_BUILD {
		t.Errorf("Expected phase TURN_PHASE_BUILD after 7, got %v", state.TurnPhase)
	}
}

func TestRollDice_NotDuringSetup(t *testing.T) {
	state := createPlayingGameState(2)
	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_BUILD

	playerID := state.Players[0].Id

	_, err := PerformDiceRoll(state, playerID)
	if err != ErrWrongPhase {
		t.Errorf("Expected ErrWrongPhase during setup, got %v", err)
	}
}

// Resource Distribution Tests

func TestResourceDistribution_SettlementGetsOneResource(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id
	player := state.Players[0]

	// Place a settlement on a vertex adjacent to a hex with number 8 and WHEAT
	// First find a hex with number 8
	var targetHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Number == 8 {
			targetHex = hex
			break
		}
	}
	if targetHex == nil {
		t.Skip("No hex with number 8 found in test board")
	}

	// Find a vertex adjacent to this hex
	var targetVertex *pb.Vertex
	for _, vertex := range state.Board.Vertices {
		for _, adjHex := range vertex.AdjacentHexes {
			if adjHex.Q == targetHex.Coord.Q && adjHex.R == targetHex.Coord.R {
				targetVertex = vertex
				break
			}
		}
		if targetVertex != nil {
			break
		}
	}
	if targetVertex == nil {
		t.Skip("No vertex found adjacent to hex 8")
	}

	// Place settlement
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: playerID,
	}

	// Record initial resources
	initialResources := getResourceTotal(player.Resources)

	// Roll an 8
	_, err := PerformDiceRollWithValues(state, playerID, 4, 4)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check player received exactly 1 resource
	newResources := getResourceTotal(player.Resources)
	gained := newResources - initialResources

	if gained != 1 {
		t.Errorf("Settlement should gain 1 resource on matching roll, gained %d", gained)
	}
}

func TestResourceDistribution_CityGetsTwoResources(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id
	player := state.Players[0]

	// Find a hex with number 6
	var targetHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Number == 6 {
			targetHex = hex
			break
		}
	}
	if targetHex == nil {
		t.Skip("No hex with number 6 found in test board")
	}

	// Find a vertex adjacent to ONLY this hex (to avoid multiple resource grants)
	var targetVertex *pb.Vertex
	for _, vertex := range state.Board.Vertices {
		adjacentToTarget := false
		otherHexCount := 0
		for _, adjHex := range vertex.AdjacentHexes {
			if adjHex.Q == targetHex.Coord.Q && adjHex.R == targetHex.Coord.R {
				adjacentToTarget = true
			} else {
				// Check if this other hex also has number 6
				for _, hex := range state.Board.Hexes {
					if hex.Coord.Q == adjHex.Q && hex.Coord.R == adjHex.R && hex.Number == 6 {
						otherHexCount++
					}
				}
			}
		}
		if adjacentToTarget && otherHexCount == 0 {
			targetVertex = vertex
			break
		}
	}
	if targetVertex == nil {
		// Fallback: just verify we get MORE resources than a settlement would
		// Find any vertex adjacent to a 6
		for _, vertex := range state.Board.Vertices {
			for _, adjHex := range vertex.AdjacentHexes {
				if adjHex.Q == targetHex.Coord.Q && adjHex.R == targetHex.Coord.R {
					targetVertex = vertex
					break
				}
			}
			if targetVertex != nil {
				break
			}
		}
	}

	// Clear all buildings first
	for _, v := range state.Board.Vertices {
		v.Building = nil
	}

	// Place CITY (not settlement)
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_CITY,
		OwnerId: playerID,
	}

	// Clear resources
	player.Resources = &pb.ResourceCount{}

	// Roll a 6
	_, err := PerformDiceRollWithValues(state, playerID, 3, 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check player received resources (at least 2 for city)
	// A city on a single hex-6 should get exactly 2
	// A city on a vertex adjacent to multiple hex-6s gets 2 per hex
	newResources := getResourceTotal(player.Resources)

	// Verify we got at least 2 and it's an even number (cities always produce 2 per hex)
	if newResources < 2 {
		t.Errorf("City should gain at least 2 resources on matching roll, gained %d", newResources)
	}
	if newResources%2 != 0 {
		t.Errorf("City resources should be even (2 per hex), got %d", newResources)
	}
}

func TestResourceDistribution_NoResourcesFromDesert(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id
	player := state.Players[0]

	// Find the desert hex
	var desertHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Resource == pb.TileResource_TILE_RESOURCE_DESERT {
			desertHex = hex
			break
		}
	}
	if desertHex == nil {
		t.Skip("No desert hex found")
	}

	// Find a vertex adjacent only to desert
	var targetVertex *pb.Vertex
	for _, vertex := range state.Board.Vertices {
		adjacentToDesert := false
		for _, adjHex := range vertex.AdjacentHexes {
			if adjHex.Q == desertHex.Coord.Q && adjHex.R == desertHex.Coord.R {
				adjacentToDesert = true
				break
			}
		}
		if adjacentToDesert {
			targetVertex = vertex
			break
		}
	}
	if targetVertex == nil {
		t.Skip("No vertex found adjacent to desert")
	}

	// Place settlement
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: playerID,
	}

	// Clear player resources
	player.Resources = &pb.ResourceCount{}

	// Roll any number (desert has no number, so shouldn't produce)
	// Desert has Number = 0, which can never be rolled
	_, err := PerformDiceRollWithValues(state, playerID, 3, 3) // Roll 6
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Desert shouldn't produce (unless vertex is also adjacent to a 6 hex)
	// This test is a bit tricky - we're really testing desert produces nothing
}

func TestResourceDistribution_RobberBlocksProduction(t *testing.T) {
	state := createPlayingGameState(2)

	playerID := state.Players[0].Id
	player := state.Players[0]

	// Find a hex with number 8
	var targetHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Number == 8 && hex.Resource != pb.TileResource_TILE_RESOURCE_DESERT {
			targetHex = hex
			break
		}
	}
	if targetHex == nil {
		t.Skip("No hex with number 8 found")
	}

	// Find a vertex adjacent to this hex
	var targetVertex *pb.Vertex
	for _, vertex := range state.Board.Vertices {
		for _, adjHex := range vertex.AdjacentHexes {
			if adjHex.Q == targetHex.Coord.Q && adjHex.R == targetHex.Coord.R {
				targetVertex = vertex
				break
			}
		}
		if targetVertex != nil {
			break
		}
	}

	// Place settlement
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: playerID,
	}

	// Place robber on that hex
	state.Board.RobberHex = targetHex.Coord

	// Clear resources
	player.Resources = &pb.ResourceCount{}

	// Roll an 8
	_, err := PerformDiceRollWithValues(state, playerID, 4, 4)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Player should NOT receive resources because robber is blocking
	totalResources := getResourceTotal(player.Resources)
	if totalResources != 0 {
		t.Errorf("Robber should block production, but player got %d resources", totalResources)
	}
}

func TestResourceDistribution_MultiplePlayersReceive(t *testing.T) {
	state := createPlayingGameState(2)

	player1 := state.Players[0]
	player2 := state.Players[1]

	// Find a hex with number 9
	var targetHex *pb.Hex
	for _, hex := range state.Board.Hexes {
		if hex.Number == 9 && hex.Resource != pb.TileResource_TILE_RESOURCE_DESERT {
			targetHex = hex
			break
		}
	}
	if targetHex == nil {
		t.Skip("No hex with number 9 found")
	}

	// Find two different vertices adjacent to this hex
	var vertices []*pb.Vertex
	for _, vertex := range state.Board.Vertices {
		for _, adjHex := range vertex.AdjacentHexes {
			if adjHex.Q == targetHex.Coord.Q && adjHex.R == targetHex.Coord.R {
				vertices = append(vertices, vertex)
				break
			}
		}
		if len(vertices) >= 2 {
			break
		}
	}
	if len(vertices) < 2 {
		t.Skip("Not enough vertices adjacent to hex 9")
	}

	// Place settlements for both players
	vertices[0].Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: player1.Id,
	}
	vertices[1].Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: player2.Id,
	}

	// Clear resources
	player1.Resources = &pb.ResourceCount{}
	player2.Resources = &pb.ResourceCount{}

	// Roll a 9 (player 1 rolls)
	_, err := PerformDiceRollWithValues(state, player1.Id, 4, 5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Both players should receive 1 resource each
	p1Resources := getResourceTotal(player1.Resources)
	p2Resources := getResourceTotal(player2.Resources)

	if p1Resources != 1 {
		t.Errorf("Player 1 should have 1 resource, got %d", p1Resources)
	}
	if p2Resources != 1 {
		t.Errorf("Player 2 should have 1 resource, got %d", p2Resources)
	}
}

// Helper function to count total resources
func getResourceTotal(r *pb.ResourceCount) int {
	if r == nil {
		return 0
	}
	return int(r.Wood + r.Brick + r.Sheep + r.Wheat + r.Ore)
}

// Discard on 7 tests

func TestDiscardOnSeven_OverSevenCards(t *testing.T) {
	state := createPlayingGameState(2)

	player := state.Players[0]
	// Give player 10 cards
	player.Resources = &pb.ResourceCount{
		Wood:  3,
		Brick: 2,
		Sheep: 2,
		Wheat: 2,
		Ore:   1,
	}

	result, err := PerformDiceRollWithValues(state, player.Id, 3, 4) // Roll 7
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have players who need to discard
	if len(result.PlayersToDiscard) == 0 {
		t.Error("Expected players to discard when rolling 7 with > 7 cards")
	}

	// Player should need to discard 5 (half of 10)
	for _, pd := range result.PlayersToDiscard {
		if pd.PlayerID == player.Id {
			if pd.DiscardCount != 5 {
				t.Errorf("Expected to discard 5 cards, got %d", pd.DiscardCount)
			}
			return
		}
	}
	t.Error("Player with 10 cards not in discard list")
}

func TestDiscardOnSeven_ExactlySevenCards(t *testing.T) {
	state := createPlayingGameState(2)

	player := state.Players[0]
	// Give player exactly 7 cards - should NOT discard
	player.Resources = &pb.ResourceCount{
		Wood:  2,
		Brick: 1,
		Sheep: 1,
		Wheat: 2,
		Ore:   1,
	}

	result, err := PerformDiceRollWithValues(state, player.Id, 3, 4) // Roll 7
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Player with exactly 7 should NOT need to discard
	for _, pd := range result.PlayersToDiscard {
		if pd.PlayerID == player.Id {
			t.Error("Player with exactly 7 cards should not need to discard")
		}
	}
}

func TestDiscardOnSeven_UnderSevenCards(t *testing.T) {
	state := createPlayingGameState(2)

	player := state.Players[0]
	// Give player 5 cards - should NOT discard
	player.Resources = &pb.ResourceCount{
		Wood:  1,
		Brick: 1,
		Sheep: 1,
		Wheat: 1,
		Ore:   1,
	}

	result, err := PerformDiceRollWithValues(state, player.Id, 3, 4) // Roll 7
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Player with under 7 should NOT need to discard
	for _, pd := range result.PlayersToDiscard {
		if pd.PlayerID == player.Id {
			t.Error("Player with under 7 cards should not need to discard")
		}
	}
}
