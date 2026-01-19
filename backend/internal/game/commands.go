package game

import (
	pb "settlers_from_catan/gen/proto/catan/v1"
)

// PlaceSettlement places a settlement during normal gameplay (costs resources)
func PlaceSettlement(state *pb.GameState, playerID, vertexID string) error {
	// Verify game is in playing status
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return ErrWrongPhase
	}

	// Verify it's build or trade phase
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_BUILD && state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		return ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	player := state.Players[currentPlayerIdx]
	if player.Id != playerID {
		return ErrNotYourTurn
	}

	// Check max settlements
	settlementCount := countPlayerSettlements(state.Board, playerID)
	if settlementCount >= GetMaxSettlements() {
		return ErrMaxSettlementsReached
	}

	// Check resources
	if !CanAfford(ResourceCountToMap(player.Resources), "settlement") {
		return ErrInsufficientResources
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

	// Check connectivity - must connect to player's road (in normal play)
	if !vertexConnectsToPlayerRoad(state.Board, vertexID, playerID) {
		return ErrMustConnectToOwned
	}

	// Deduct resources
	DeductResources(player.Resources, "settlement")

	// Place the settlement
	targetVertex.Building = &pb.Building{
		Type:    pb.BuildingType_BUILDING_TYPE_SETTLEMENT,
		OwnerId: playerID,
	}

	// Update victory points
	player.VictoryPoints++

	// Check victory after point-gaining settlement placement
	if victory, _ := CheckVictory(state); victory {
		state.Status = pb.GameStatus_GAME_STATUS_FINISHED
	}

	return nil
}

// PlaceCity upgrades a settlement to a city
func PlaceCity(state *pb.GameState, playerID, vertexID string) error {
	// Verify game is in playing status
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return ErrWrongPhase
	}

	// Verify it's build or trade phase
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_BUILD && state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		return ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	player := state.Players[currentPlayerIdx]
	if player.Id != playerID {
		return ErrNotYourTurn
	}

	// Check resources
	if !CanAfford(ResourceCountToMap(player.Resources), "city") {
		return ErrInsufficientResources
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

	// Check there's a settlement owned by this player
	if targetVertex.Building == nil {
		return ErrCannotUpgrade
	}
	if targetVertex.Building.OwnerId != playerID {
		return ErrCannotUpgrade
	}
	if targetVertex.Building.Type != pb.BuildingType_BUILDING_TYPE_SETTLEMENT {
		return ErrCannotUpgrade // Already a city
	}

	// Deduct resources
	DeductResources(player.Resources, "city")

	// Upgrade to city
	targetVertex.Building.Type = pb.BuildingType_BUILDING_TYPE_CITY

	// Update victory points (city is worth 2, settlement was worth 1, so +1)
	player.VictoryPoints++

	// Check victory after city placement
	if victory, _ := CheckVictory(state); victory {
		state.Status = pb.GameStatus_GAME_STATUS_FINISHED
	}

	return nil
}

// PlaceRoad places a road during normal gameplay
func PlaceRoad(state *pb.GameState, playerID, edgeID string) error {
	// Verify game is in playing status
	if state.Status != pb.GameStatus_GAME_STATUS_PLAYING {
		return ErrWrongPhase
	}

	// Verify it's build or trade phase
	if state.TurnPhase != pb.TurnPhase_TURN_PHASE_BUILD && state.TurnPhase != pb.TurnPhase_TURN_PHASE_TRADE {
		return ErrWrongPhase
	}

	// Verify it's this player's turn
	currentPlayerIdx := state.CurrentTurn
	if currentPlayerIdx < 0 || int(currentPlayerIdx) >= len(state.Players) {
		return ErrNotYourTurn
	}
	player := state.Players[currentPlayerIdx]
	if player.Id != playerID {
		return ErrNotYourTurn
	}

	// Check max roads
	roadCount := countPlayerRoads(state.Board, playerID)
	if roadCount >= GetMaxRoads() {
		return ErrMaxRoadsReached
	}

	// Check resources
	if !CanAfford(ResourceCountToMap(player.Resources), "road") {
		return ErrInsufficientResources
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

	// Check connectivity - must connect to player's road, settlement, or city
	if !edgeConnectsToPlayerStructure(state.Board, targetEdge, playerID) {
		return ErrMustConnectToOwned
	}

	// Deduct resources
	DeductResources(player.Resources, "road")

	// Place the road
	targetEdge.Road = &pb.Road{
		OwnerId: playerID,
	}

	// Check victory after possible bonus (e.g., longest road transferred)
	if victory, _ := CheckVictory(state); victory {
		state.Status = pb.GameStatus_GAME_STATUS_FINISHED
	}

	return nil
}

// GetBuildingAtVertex returns the building at a vertex, or nil if empty
func GetBuildingAtVertex(board *pb.BoardState, vertexID string) *pb.Building {
	for _, v := range board.Vertices {
		if v.Id == vertexID {
			return v.Building
		}
	}
	return nil
}

// Helper functions

func countPlayerSettlements(board *pb.BoardState, playerID string) int {
	count := 0
	for _, v := range board.Vertices {
		if v.Building != nil && v.Building.OwnerId == playerID &&
			v.Building.Type == pb.BuildingType_BUILDING_TYPE_SETTLEMENT {
			count++
		}
	}
	return count
}

func countPlayerRoads(board *pb.BoardState, playerID string) int {
	count := 0
	for _, e := range board.Edges {
		if e.Road != nil && e.Road.OwnerId == playerID {
			count++
		}
	}
	return count
}

func vertexConnectsToPlayerRoad(board *pb.BoardState, vertexID, playerID string) bool {
	for _, edge := range board.Edges {
		if edge.Road != nil && edge.Road.OwnerId == playerID {
			for _, vID := range edge.Vertices {
				if vID == vertexID {
					return true
				}
			}
		}
	}
	return false
}

func edgeConnectsToPlayerStructure(board *pb.BoardState, edge *pb.Edge, playerID string) bool {
	// Check if edge connects to player's settlement/city
	for _, vID := range edge.Vertices {
		for _, v := range board.Vertices {
			if v.Id == vID && v.Building != nil && v.Building.OwnerId == playerID {
				return true
			}
		}
	}

	// Check if edge connects to player's road
	for _, vID := range edge.Vertices {
		for _, otherEdge := range board.Edges {
			if otherEdge.Road != nil && otherEdge.Road.OwnerId == playerID {
				for _, otherVID := range otherEdge.Vertices {
					if otherVID == vID {
						return true
					}
				}
			}
		}
	}

	return false
}
