package game

import (
	"errors"

	pb "settlers_from_catan/gen/proto/catan/v1"
)

// Error definitions for state machine operations
var (
	ErrNotYourTurn              = errors.New("it's not your turn")
	ErrWrongPhase               = errors.New("cannot perform this action in current phase")
	ErrInvalidVertex            = errors.New("invalid vertex ID")
	ErrInvalidEdge              = errors.New("invalid edge ID")
	ErrVertexOccupied           = errors.New("vertex is already occupied")
	ErrEdgeOccupied             = errors.New("edge is already occupied")
	ErrDistanceRule             = errors.New("settlement violates distance rule (must be 2+ edges from other settlements)")
	ErrMustConnectToOwned       = errors.New("structure must connect to your existing structures")
	ErrInsufficientResources    = errors.New("insufficient resources")
	ErrCannotUpgrade            = errors.New("can only upgrade your own settlement to city")
	ErrMustPlaceSettlementFirst = errors.New("must place settlement before road in setup")
	ErrRoadMustConnectToSetup   = errors.New("road must connect to just-placed settlement")
	ErrMaxSettlementsReached    = errors.New("maximum settlements reached")
	ErrMaxRoadsReached          = errors.New("maximum roads reached")
	ErrPlayerNotFound           = errors.New("player not found")
	ErrNotHost                  = errors.New("only host can start the game")
	ErrPlayersNotReady          = errors.New("all players must be ready")
	ErrNotEnoughPlayers         = errors.New("not enough players to start")
)

// GetSetupTurnIndex returns the player index for a given turn number (0-indexed)
// Snake draft: Round 1 forward (0,1,2,3), Round 2 reverse (3,2,1,0)
// Each turn, a player places one settlement + one road
// For 4 players: turns 0,1,2,3 are players 0,1,2,3 (round 1)
//
//	turns 4,5,6,7 are players 3,2,1,0 (round 2)
func GetSetupTurnIndex(state *pb.GameState, turnNumber int) int32 {
	numPlayers := len(state.Players)
	if numPlayers == 0 {
		return 0
	}

	if turnNumber < numPlayers {
		// Round 1: forward order
		return int32(turnNumber)
	} else {
		// Round 2: reverse order
		adjustedTurn := turnNumber - numPlayers
		return int32((numPlayers - 1) - adjustedTurn)
	}
}

// IsSetupComplete checks if all setup turns have been completed
func IsSetupComplete(state *pb.GameState, completedTurns int) bool {
	numPlayers := len(state.Players)
	// Total turns needed: numPlayers * 2 rounds
	expectedTurns := numPlayers * 2
	return completedTurns >= expectedTurns
}

// TransitionToPlaying moves the game from setup to playing phase
func TransitionToPlaying(state *pb.GameState) {
	state.Status = pb.GameStatus_GAME_STATUS_PLAYING
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL
	state.CurrentTurn = 0
	state.SetupPhase = nil
}

// SetPlayerReady updates a player's ready status in the lobby.
func SetPlayerReady(state *pb.GameState, playerID string, ready bool) error {
	if state.Status != pb.GameStatus_GAME_STATUS_WAITING {
		return ErrWrongPhase
	}
	player := getPlayerByID(state, playerID)
	if player == nil {
		return ErrPlayerNotFound
	}
	player.IsReady = ready
	return nil
}

// StartGame transitions the game from waiting to setup if requirements are met.
func StartGame(state *pb.GameState, playerID string) error {
	if state.Status != pb.GameStatus_GAME_STATUS_WAITING {
		return ErrWrongPhase
	}
	player := getPlayerByID(state, playerID)
	if player == nil {
		return ErrPlayerNotFound
	}
	if !player.IsHost {
		return ErrNotHost
	}
	if len(state.Players) < GetMinPlayers() {
		return ErrNotEnoughPlayers
	}
	for _, p := range state.Players {
		if !p.IsReady {
			return ErrPlayersNotReady
		}
	}

	state.Status = pb.GameStatus_GAME_STATUS_SETUP
	state.SetupPhase = &pb.SetupPhase{
		Round:            1,
		PlacementsInTurn: 0,
	}
	state.CurrentTurn = 0
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL
	return nil
}

// EndTurn advances to the next player's turn.
func EndTurn(state *pb.GameState, playerID string) error {
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return ErrWrongPhase
	}
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	if state.Players[currentPlayerIdx].Id != playerID {
		return ErrNotYourTurn
	}
	if state.TurnPhase == pb.TurnPhase_TURN_PHASE_ROLL {
		return ErrWrongPhase
	}
	if len(state.Players) == 0 {
		return ErrNotEnoughPlayers
	}
	state.CurrentTurn = (state.CurrentTurn + 1) % int32(len(state.Players))
	state.TurnPhase = pb.TurnPhase_TURN_PHASE_ROLL
	state.Dice = []int32{0, 0}
	return nil
}

// PlaceSetupSettlement places a settlement during setup phase (no resource cost)
func PlaceSetupSettlement(state *pb.GameState, playerID, vertexID string) error {
	// Verify it's setup phase
	if state.Status != pb.GameStatus_GAME_STATUS_SETUP {
		return ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	if state.Players[currentPlayerIdx].Id != playerID {
		return ErrNotYourTurn
	}

	// Verify we need a settlement (placements_in_turn should be 0)
	if state.SetupPhase != nil && state.SetupPhase.PlacementsInTurn != 0 {
		return ErrWrongPhase
	}

	// Find the vertex
	var targetVertex *pb.Vertex
	for _, v := range state.Board.Vertices {
		if v.Id == vertexID {
			targetVertex = v
			break
		}
	}
	if targetVertex == nil {
		return ErrInvalidVertex
	}

	// Check vertex is empty
	if targetVertex.Building != nil {
		return ErrVertexOccupied
	}

	// Check distance rule
	if violatesDistanceRule(state.Board, vertexID) {
		return ErrDistanceRule
	}

	// Place the settlement
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: playerID,
	}

	// Update victory points
	for _, p := range state.Players {
		if p.Id == playerID {
			p.VictoryPoints++
			break
		}
	}

	// Grant resources for second settlement (round 2)
	if state.SetupPhase != nil && state.SetupPhase.Round == 2 {
		grantResourcesForVertex(state, playerID, targetVertex)
	}

	// Update setup phase state
	if state.SetupPhase != nil {
		state.SetupPhase.PlacementsInTurn = 1 // Now need to place road
	}

	return nil
}

// PlaceSetupRoad places a road during setup phase (no resource cost)
func PlaceSetupRoad(state *pb.GameState, playerID, edgeID string) error {
	// Verify it's setup phase
	if state.Status != pb.GameStatus_GAME_STATUS_SETUP {
		return ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	if state.Players[currentPlayerIdx].Id != playerID {
		return ErrNotYourTurn
	}

	// Verify we need a road (placements_in_turn should be 1, meaning settlement was placed)
	if state.SetupPhase == nil || state.SetupPhase.PlacementsInTurn != 1 {
		return ErrMustPlaceSettlementFirst
	}

	// Find the edge
	var targetEdge *pb.Edge
	for _, e := range state.Board.Edges {
		if e.Id == edgeID {
			targetEdge = e
			break
		}
	}
	if targetEdge == nil {
		return ErrInvalidEdge
	}

	// Check edge is empty
	if targetEdge.Road != nil {
		return ErrEdgeOccupied
	}

	// Road must connect to a settlement owned by this player
	// In setup, it must specifically connect to the just-placed settlement
	connectsToOwned := false
	for _, vID := range targetEdge.Vertices {
		for _, v := range state.Board.Vertices {
			if v.Id == vID && v.Building != nil && v.Building.OwnerId == playerID {
				connectsToOwned = true
				break
			}
		}
		if connectsToOwned {
			break
		}
	}
	if !connectsToOwned {
		return ErrRoadMustConnectToSetup
	}

	// Place the road
	targetEdge.Road = &pb.Road{
		OwnerId: playerID,
	}

	// Move to next player or next round
	advanceSetupTurn(state)

	return nil
}

// advanceSetupTurn advances to the next player in setup or transitions to playing
func advanceSetupTurn(state *pb.GameState) {
	numPlayers := len(state.Players)
	if numPlayers == 0 || state.SetupPhase == nil {
		return
	}

	round := state.SetupPhase.Round

	if round == 1 {
		// Round 1: forward order
		if int(state.CurrentTurn) == numPlayers-1 {
			// End of round 1, start round 2 with same player (snake)
			state.SetupPhase.Round = 2
			state.SetupPhase.PlacementsInTurn = 0
		} else {
			// Move to next player
			state.CurrentTurn++
			state.SetupPhase.PlacementsInTurn = 0
		}
	} else {
		// Round 2: reverse order
		if state.CurrentTurn == 0 {
			// End of setup, transition to playing
			TransitionToPlaying(state)
		} else {
			// Move to previous player (reverse)
			state.CurrentTurn--
			state.SetupPhase.PlacementsInTurn = 0
		}
	}
}

// violatesDistanceRule checks if placing a settlement at vertexID would violate the distance rule
func violatesDistanceRule(board *pb.BoardState, vertexID string) bool {
	// Find all edges connected to this vertex
	adjacentVertexIDs := make(map[string]bool)
	for _, edge := range board.Edges {
		for i, vID := range edge.Vertices {
			if vID == vertexID {
				// The other vertex on this edge is adjacent
				otherIdx := 1 - i
				if otherIdx >= 0 && otherIdx < len(edge.Vertices) {
					adjacentVertexIDs[edge.Vertices[otherIdx]] = true
				}
			}
		}
	}

	// Check if any adjacent vertex has a building
	for _, vertex := range board.Vertices {
		if adjacentVertexIDs[vertex.Id] && vertex.Building != nil {
			return true
		}
	}

	return false
}

// grantResourcesForVertex gives resources to a player based on adjacent hexes
func grantResourcesForVertex(state *pb.GameState, playerID string, vertex *pb.Vertex) {
	var player *pb.PlayerState
	for _, p := range state.Players {
		if p.Id == playerID {
			player = p
			break
		}
	}
	if player == nil || player.Resources == nil {
		return
	}

	for _, hexCoord := range vertex.AdjacentHexes {
		// Find the hex
		for _, hex := range state.Board.Hexes {
			if hex.Coord.Q == hexCoord.Q && hex.Coord.R == hexCoord.R {
				// Grant 1 resource for this hex (skip desert)
				addResourceForTile(player.Resources, hex.Resource)
				break
			}
		}
	}
}

// addResourceForTile adds 1 resource to the count based on tile type
func addResourceForTile(resources *pb.ResourceCount, tileResource pb.TileResource) {
	switch tileResource {
	case pb.TileResource_TILE_RESOURCE_WOOD:
		resources.Wood++
	case pb.TileResource_TILE_RESOURCE_BRICK:
		resources.Brick++
	case pb.TileResource_TILE_RESOURCE_SHEEP:
		resources.Sheep++
	case pb.TileResource_TILE_RESOURCE_WHEAT:
		resources.Wheat++
	case pb.TileResource_TILE_RESOURCE_ORE:
		resources.Ore++
		// Desert gives nothing
	}
}
